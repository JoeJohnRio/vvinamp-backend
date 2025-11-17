package auth

import (
	"context"
	"log"
	"strings"
	"time"
	"vvinamp/constant"
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
}

// SetSessionToResolver implements the AuthResponseWriter interface
func (arw *AuthResponseWriter) SetSessionToResolver(userID string) {
	log.Println("joel123 SetSessionToResolver")
	arw.sessionToResolver = userID
}

func (arw *AuthResponseWriter) GetAuth() *Auth {
	log.Println("joel123 GetAuth")
	return arw.auth
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
	}
}

// GenerateTokens creates both access and refresh tokens
func (a *Auth) GenerateTokens(userID string) (accessToken, refreshToken string, err error) {
	log.Println("joel123", userID)
	// Generate access token (short-lived)
	accessClaims := &model.Token{
		UserID: userID,
		Type:   "access",
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(15000 * time.Minute).Unix(),
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

	log.Println("joel123 access token", accessToken)
	return accessToken, refreshToken, nil
}

// Middleware for GraphQL resolver to pass session into ctx and receive
// session from ctx set in request.Cache
func (a *Auth) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get JWT from Authorization: Bearer <token>
		log.Println("joel123 Middleware Authentication")
		authHeader := c.GetHeader("Authorization")
		tokenPart := ""
		if strings.HasPrefix(authHeader, "Bearer ") {

			tokenPart = strings.TrimPrefix(authHeader, "Bearer ")
		}

		log.Println("joel123 Bearer Authentication", tokenPart)

		if tokenPart != "" {
			tk := &model.Token{}
			token, err := jwt.ParseWithClaims(tokenPart, tk, func(token *jwt.Token) (interface{}, error) {
				return []byte(a.tokenPassword), nil
			})
			if err == nil && token.Valid && !time.Unix(tk.ExpiresAt, 0).Before(time.Now()) {
				// Session logic: Write the user ID from token into the request context
				ctx := context.WithValue(c.Request.Context(), constant.Session, tk.UserID)
				c.Request = c.Request.WithContext(ctx)
			}
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
