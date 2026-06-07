package auth

import (
	"strings"
	"time"
	"vvinamp/internal/pkg/session"
	"vvinamp/model"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type key string

const UserIDKey key = "userID"

type AuthResponseWriter struct {
	gin.ResponseWriter
	tokenPassword     string
	sessionToResolver string
	auth              *Auth
}
type Auth struct {
	tokenPassword string
	logger        *zap.Logger
	sessionMgr    *session.SessionManager
}

// SetSessionToResolver implements the AuthResponseWriter interface
func (arw *AuthResponseWriter) SetSessionToResolver(userID string) {
	arw.sessionToResolver = userID
}

func (arw *AuthResponseWriter) GetAuth() *Auth {
	return arw.auth
}

func (arw *AuthResponseWriter) GetTokenPassword() string {
	return arw.tokenPassword
}

// Auth is class provided middleware for read/write token in to request/response cache

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
		sessionMgr:    session.NewSessionManager(),
	}
}

// GenerateTokens creates both access and refresh tokens
func (a *Auth) GenerateTokens(userID string) (accessToken, refreshToken string, err error) {
	// Generate access token (short-lived)
	accessClaims := &model.Token{
		UserID: userID,
		Type:   "access",
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(60 * time.Minute).Unix(),
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
		// Get JWT from Authorization: Bearer <token>
		authHeader := c.GetHeader("Authorization")
		tokenPart := ""
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenPart = strings.TrimPrefix(authHeader, "Bearer ")
		}

		if tokenPart != "" {
			tk := &model.Token{}
			token, err := jwt.ParseWithClaims(tokenPart, tk, func(token *jwt.Token) (interface{}, error) {
				return []byte(a.tokenPassword), nil
			})
			if err == nil && token.Valid && !time.Unix(tk.ExpiresAt, 0).Before(time.Now()) {

				// Create session using session manager
				ctx := a.sessionMgr.CreateSession(c.Request.Context(), tk.UserID)
				c.Request = c.Request.WithContext(ctx)
			} else {
				// Even if token is invalid, continue with the request but without session
				// The session validation will happen later in specific routes
			}
		} else {
			// No token provided, continue without session
		}

		arw := &AuthResponseWriter{
			ResponseWriter:    c.Writer,
			tokenPassword:     a.tokenPassword, // pass secrets/configs from your Auth struct
			sessionToResolver: "",
			auth:              a,
		}
		c.Writer = arw // <-- This is the critical line!
		// Continue to next Handler
		c.Next()
	}
}

// SetSession allows resolvers to set a new session
// func SetSession(ctx context.Context, userID string) context.Context {
// 	if arw, ok := ctx.Value("authResponseWriter").(*authResponseWriter); ok {
// 		arw.sessionToResolver = userID
// 	}
// 	return context.WithValue(ctx, constant.Session, userID)
// }
