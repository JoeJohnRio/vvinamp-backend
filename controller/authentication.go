package controller

import (
	"context"
	"net/http"
	"time"
	"vvinamp/constant"
	"vvinamp/model"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type key string

type authResponseWriter struct {
	gin.ResponseWriter
	tokenPassword     string
	sessionToResolver string
	sessionFromCookie string
	auth              *Auth
}

func (arw *authResponseWriter) Write(data []byte) (n int, err error) {
	// Generate tokens when session is established during login
	if arw.sessionToResolver != "" && arw.sessionToResolver != arw.sessionFromCookie {
		accessToken, refreshToken, err := arw.auth.GenerateTokens(arw.sessionToResolver)
		if err != nil {
			arw.auth.logger.Error("Failed to generate tokens", zap.Error(err))
			return arw.ResponseWriter.Write(data)
		}

		// Set access token cookie (short-lived)
		http.SetCookie(arw, &http.Cookie{
			Name:     "access_token",
			Value:    accessToken,
			HttpOnly: true,
			Secure:   true, // Use HTTPS in production
			SameSite: http.SameSiteStrictMode,
			MaxAge:   15 * 60, // 15 minutes
			Path:     "/",
		})

		// Set refresh token cookie (long-lived)
		http.SetCookie(arw, &http.Cookie{
			Name:     "refresh_token",
			Value:    refreshToken,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
			MaxAge:   7 * 24 * 60 * 60, // 7 days
			Path:     "/",
		})
	}

	return arw.ResponseWriter.Write(data)
}

// Auth is class provided middleware for read/write token in to request/response cache
type Auth struct {
	tokenPassword string
	logger        *zap.Logger
}

// AuthTarget is parameter object for getting all Auth's dependency
type AuthTarget struct {
	fx.In
	TokenPassword string `name:"token_password"`
	Logger        *zap.Logger
}

// NewAuth is construct for Auth
func NewAuth(target AuthTarget) *Auth {
	return &Auth{
		tokenPassword: target.TokenPassword,
		logger:        target.Logger,
	}
}

// GenerateTokens creates both access and refresh tokens
func (a *Auth) GenerateTokens(userID string) (accessToken, refreshToken string, err error) {
	// Generate access token (short-lived)
	accessClaims := &model.Token{
		UserID: userID,
		Type:   "access",
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(15 * time.Minute).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}

	accessTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessToken, err = accessTokenObj.SignedString([]byte(a.tokenPassword))
	if err != nil {
		return "", "", err
	}

	// Generate refresh token (long-lived)
	refreshClaims := &model.Token{
		UserID: userID,
		Type:   "refresh",
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(7 * 24 * time.Hour).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}

	refreshTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshToken, err = refreshTokenObj.SignedString([]byte(a.tokenPassword))
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// Middleware for GraphQL resolver to pass session into ctx and receive
// session from ctx set in request.Cache
func (a *Auth) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		arw := &authResponseWriter{
			ResponseWriter:    c.Writer,
			tokenPassword:     a.tokenPassword,
			sessionToResolver: "",
			sessionFromCookie: "",
			auth:              a,
		}

		// Try to get access token first
		tokenPart, _ := c.Cookie("access_token")
		if tokenPart == "" {
			// Fallback to old token cookie for backwards compatibility
			tokenPart, _ = c.Cookie("token")
		}

		if tokenPart != "" {
			tk := &model.Token{}
			token, err := jwt.ParseWithClaims(tokenPart, tk, func(token *jwt.Token) (interface{}, error) {
				return []byte(a.tokenPassword), nil
			})

			// Malformed token
			if err != nil {
				a.logger.Error("malformed token", zap.Error(err))
			} else {
				// Token is valid
				if token.Valid && !time.Unix(tk.ExpiresAt, 0).Before(time.Now()) {
					arw.sessionFromCookie = tk.UserID
					arw.sessionToResolver = tk.UserID
				}
			}
		}

		// Pass session to context - fixed the reference issue
		ctx := context.WithValue(c.Request.Context(), constant.Session, arw.sessionToResolver)
		c.Request = c.Request.WithContext(ctx)
		c.Writer = arw
		c.Next()
	}
}

// SetSession allows resolvers to set a new session
func SetSession(ctx context.Context, userID string) context.Context {
	if arw, ok := ctx.Value("authResponseWriter").(*authResponseWriter); ok {
		arw.sessionToResolver = userID
	}
	return context.WithValue(ctx, constant.Session, userID)
}
