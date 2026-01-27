package session

import (
	"context"
	"fmt"
	"log"
	"time"
	"vvinamp/constant"

	"github.com/gin-gonic/gin"
)

// SessionManager handles session operations
type SessionManager struct {
	// You can add session store configurations here if needed
	// For now, we'll use context-based session management
}

// NewSessionManager creates a new session manager
func NewSessionManager() *SessionManager {
	return &SessionManager{}
}

// Session represents a user session
type Session struct {
	UserID    string
	CreatedAt time.Time
	ExpiresAt time.Time
}

// CreateSession creates a new session for a user
func (sm *SessionManager) CreateSession(ctx context.Context, userID string) context.Context {
	session := &Session{
		UserID:    userID,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour), // Session expires in 24 hours
	}

	// Store session in context
	ctx = context.WithValue(ctx, constant.Session, session.UserID)
	return ctx
}

// ValidateSession validates if a session exists and is valid
func (sm *SessionManager) ValidateSession(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(constant.Session).(string)
	if !ok || userID == "" {
		return "", ErrInvalidSession
	}

	// In a real implementation, you might want to check if the session has expired
	// For now, we just validate that a user ID exists in the context

	return userID, nil
}

// IsAuthenticated checks if the current request has a valid session
func (sm *SessionManager) IsAuthenticated(c *gin.Context) bool {
	ctx := c.Request.Context()
	_, err := sm.ValidateSession(ctx)
	return err == nil
}

// RequireSession is a middleware that ensures a valid session exists
func (sm *SessionManager) RequireSession() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		userID, err := sm.ValidateSession(ctx)
		if err != nil {
			log.Printf("Session validation failed: %v", err)

			// Set error in context for GraphQL resolvers to handle
			ctx = context.WithValue(ctx, "session_error", err.Error())
			c.Request = c.Request.WithContext(ctx)

			// Return GraphQL-formatted error response
			c.JSON(200, gin.H{
				"errors": []gin.H{
					{
						"message": err.Error(),
						"extensions": gin.H{
							"code": err.(*SessionError).Code,
						},
					},
				},
				"data": nil,
			})
			c.Abort()
			return
		}

		// Add user ID to context for downstream handlers
		ctx = context.WithValue(ctx, "current_user_id", userID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// RequireSessionForQuery is a specialized middleware for GraphQL queries that need authentication
func (sm *SessionManager) RequireSessionForQuery() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// Check if this is a GraphQL request
		if c.Request.URL.Path == "/v1/graphql" {
			// For GraphQL, we'll set up the error in context and let the resolver handle it
			userID, err := sm.ValidateSession(ctx)
			if err != nil {
				log.Printf("GraphQL session validation failed: %v", err)

				// Set error in context for GraphQL resolvers to handle
				ctx = context.WithValue(ctx, "session_error", err.Error())
				ctx = context.WithValue(ctx, "session_valid", false)
				c.Request = c.Request.WithContext(ctx)

				// Continue to the next handler - GraphQL resolvers will check for session validity
				c.Next()
				return
			}

			// Session is valid, add user info to context
			ctx = context.WithValue(ctx, "current_user_id", userID)
			ctx = context.WithValue(ctx, "session_valid", true)
			c.Request = c.Request.WithContext(ctx)
		}

		c.Next()
	}
}

// GetUserIDFromContext retrieves the user ID from the request context
func (sm *SessionManager) GetUserIDFromContext(ctx context.Context) (string, error) {
	userID, err := sm.ValidateSession(ctx)
	if err != nil {
		return "", err
	}
	return userID, nil
}

// DestroySession removes the session from context
func (sm *SessionManager) DestroySession(ctx context.Context) context.Context {
	return context.WithValue(ctx, constant.Session, nil)
}

// SessionError represents a session-related error
type SessionError struct {
	Code    string
	Message string
}

func (e *SessionError) Error() string {
	return fmt.Sprintf("Session Error [%s]: %s", e.Code, e.Message)
}

// NewSessionError creates a new session error
func NewSessionError(code, message string) *SessionError {
	return &SessionError{
		Code:    code,
		Message: message,
	}
}

// Predefined session errors
var (
	ErrInvalidSession = NewSessionError("INVALID_SESSION", "Invalid or expired session")
	ErrNoSession      = NewSessionError("NO_SESSION", "No session found")
	ErrUnauthorized   = NewSessionError("UNAUTHORIZED", "Unauthorized access")
)
