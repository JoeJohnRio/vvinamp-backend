package session

import (
	"context"
	"fmt"
)

// GraphQLSessionHelper provides helper functions for GraphQL resolvers to validate sessions
type GraphQLSessionHelper struct {
	sessionMgr *SessionManager
}

// NewGraphQLSessionHelper creates a new helper for GraphQL session validation
func NewGraphQLSessionHelper() *GraphQLSessionHelper {
	return &GraphQLSessionHelper{
		sessionMgr: NewSessionManager(),
	}
}

// RequireSessionForResolver validates that the current request has a valid session
// This function can be called from within GraphQL resolvers to enforce authentication
func (h *GraphQLSessionHelper) RequireSessionForResolver(ctx context.Context) (string, error) {
	// Check if session validation was already performed by middleware
	if sessionValid, ok := ctx.Value("session_valid").(bool); ok {
		if !sessionValid {
			sessionError, _ := ctx.Value("session_error").(string)
			if sessionError != "" {
				return "", NewSessionError("UNAUTHORIZED", sessionError)
			}
			return "", ErrNoSession
		}
	}

	// Get the session manager from context
	sessionMgr, ok := ctx.Value("session_manager").(*SessionManager)
	if !ok {
		sessionMgr = h.sessionMgr // fallback to default instance
	}

	// Validate the session
	userID, err := sessionMgr.ValidateSession(ctx)
	if err != nil {
		return "", err
	}

	return userID, nil
}

// GetCurrentUserID retrieves the current user's ID from the session
// Returns empty string and error if no valid session exists
func (h *GraphQLSessionHelper) GetCurrentUserID(ctx context.Context) (string, error) {
	// First check if current_user_id was set by middleware
	if userID, ok := ctx.Value("current_user_id").(string); ok && userID != "" {
		return userID, nil
	}

	// Fallback to validating session directly
	return h.RequireSessionForResolver(ctx)
}

// IsAuthenticated checks if the current request is authenticated
func (h *GraphQLSessionHelper) IsAuthenticated(ctx context.Context) bool {
	_, err := h.GetCurrentUserID(ctx)
	return err == nil
}

// GraphQLError represents an error that can be returned from GraphQL resolvers
type GraphQLError struct {
	Message string
	Code    string
}

func (e *GraphQLError) Error() string {
	return fmt.Sprintf("GraphQL Error [%s]: %s", e.Code, e.Message)
}

// NewGraphQLError creates a new GraphQL error
func NewGraphQLError(message, code string) *GraphQLError {
	return &GraphQLError{
		Message: message,
		Code:    code,
	}
}
