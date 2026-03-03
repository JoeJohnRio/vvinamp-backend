package controller

import (
	"bytes"
	"context"
	"io"
	"log"
	"strings"

	"vvinamp/controller/auth"
	"vvinamp/database/postgre"
	"vvinamp/graphql/generated"
	"vvinamp/graphql/resolvers"
	"vvinamp/internal/pkg/session"
	artistPackage "vvinamp/package/artist/service"
	"vvinamp/package/history"
	"vvinamp/package/playlist"
	"vvinamp/package/search"
	"vvinamp/package/song"
	"vvinamp/package/streaming"
	"vvinamp/package/user"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport" // New import for transports
	oldHandler "github.com/99designs/gqlgen/handler"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// GraphQLController handle the graphql request, parse request to schema and return results
type GraphQLController struct {
	graphiQLEnable bool
	auth           *auth.Auth
	postgre        *postgre.Connection
	user           user.Service
	history        history.Service
	song           song.Service
	artist         artistPackage.ArtistService
	search         search.Service
	streaming      *streaming.Service
	playlist       playlist.Service
	logger         *zap.Logger
}

// GraphQLControllerTarget is parameter object for geting all GraphQLController's dependency
type GraphQLControllerTarget struct {
	fx.In
	GraphiQLEnable bool `name:"graphiql_enable"`
	Auth           *auth.Auth
	Postgresql     *postgre.Connection
	History        history.Service
	User           user.Service
	Song           song.Service
	Artist         artistPackage.ArtistService
	Search         search.Service
	Streaming      *streaming.Service
	Playlist       playlist.Service
	Logger         *zap.Logger
}

// NewGraphQLController is a constructor for GraphQLController
func NewGraphQLController(target GraphQLControllerTarget) Result {
	return Result{
		Controller: &GraphQLController{
			graphiQLEnable: target.GraphiQLEnable,
			auth:           target.Auth,
			postgre:        target.Postgresql,
			user:           target.User,
			history:        target.History,
			song:           target.Song,
			artist:         target.Artist,
			search:         target.Search,
			streaming:      target.Streaming,
			playlist:       target.Playlist,
			logger:         target.Logger,
		},
	}
}

// GrqphQL is defining as the GraphQL handler
func (m *GraphQLController) GraphQL() gin.HandlerFunc {
	// Create resolver with required services
	resolver := resolvers.NewResolver(m.artist, m.user, m.history, m.song, m.streaming, m.playlist)

	// 1. Create a new Server instance using handler.New()
	srv := handler.New(generated.NewExecutableSchema(generated.Config{Resolvers: resolver}))

	// Enable introspection for development and schema exploration
	srv.Use(extension.Introspection{})

	// 2. Explicitly add the necessary transports for your application.
	// This makes the handler modular and production-ready.
	srv.AddTransport(transport.GET{})           // Supports GET requests
	srv.AddTransport(transport.POST{})          // Supports standard JSON POST requests
	srv.AddTransport(transport.MultipartForm{}) // Supports file uploads
	srv.AddTransport(transport.Websocket{})     // Supports GraphQL Subscriptions (if needed)

	// Note: You can also use handler.NewDefaultServer() for a simpler (but less customizable)
	// setup, but handler.New() is preferred for production.

	// 3. Return a gin.HandlerFunc that calls the new handler's ServeHTTP method
	return func(c *gin.Context) {
		srv.ServeHTTP(c.Writer, c.Request)
	}
}

// GraphiQL is defining as the GraphiQL Page handler
func (m *GraphQLController) GraphiQL() gin.HandlerFunc {
	h := oldHandler.Playground("GraphQL", "/")

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// Register is function to register all controller's endpoint handler
func (m *GraphQLController) Register(r *gin.Engine) {
	// Single route with conditional authentication
	r.POST("/v1/graphql",
		m.auth.Middleware(),
		m.postgre.Middleware(),
		m.Middleware(),
		// Conditional authentication middleware
		func(c *gin.Context) {
			// Read request body to determine if this is a public operation
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err != nil {
				c.Next()
				return
			}
			bodyStr := string(bodyBytes)

			// Reset the body for further processing
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			// Check if this is a public operation (login, register, refreshToken, introspection, etc.)
			isPublicOperation := strings.Contains(bodyStr, "login") ||
				strings.Contains(bodyStr, "register") ||
				strings.Contains(bodyStr, "refreshToken") ||
				(strings.Contains(bodyStr, "__schema")) || // Allow schema introspection
				(strings.Contains(bodyStr, "__type")) // Allow type introspection

			// If this is NOT a public operation, validate session
			if !isPublicOperation {
				// Validate session for protected operations
				ctx := c.Request.Context()
				sessionMgr := session.NewSessionManager()
				_, err := sessionMgr.ValidateSession(ctx)
				if err != nil {
					// Return HTTP 401 for unauthorized access
					c.Header("WWW-Authenticate", "Bearer")
					c.JSON(401, gin.H{
						"errors": []gin.H{
							{
								"message": "Unauthorized: Invalid or expired session",
								"extensions": gin.H{
									"code": "UNAUTHORIZED",
								},
							},
						},
						"data": nil,
					})
					c.Abort()
					return
				} else {
					// Session is valid, add flag to context
					ctx = context.WithValue(ctx, "session_valid", true)
					c.Request = c.Request.WithContext(ctx)
				}
			} else {
				// For public operations, set session_valid to true to allow the resolvers to proceed
				ctx := c.Request.Context()
				ctx = context.WithValue(ctx, "session_valid", true)
				c.Request = c.Request.WithContext(ctx)
			}

			c.Next()
		},
		m.GraphQL())

	if !m.graphiQLEnable {
		r.GET("/v1/graphiql", m.GraphiQL())
	}
}

// Middleware for GraphQL resolver to pass services into ctx
func (m *GraphQLController) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		log.Println("joel123 Middleware graphqlcontroller.go")
		ctx = context.WithValue(ctx, user.Key, m.user)
		ctx = context.WithValue(ctx, history.Key, m.history)
		ctx = context.WithValue(ctx, song.Key, m.song)
		ctx = context.WithValue(ctx, artistPackage.Key, m.artist)
		ctx = context.WithValue(ctx, search.Key, m.search)
		ctx = context.WithValue(ctx, playlist.Key, m.playlist)

		// Pass the response writer for session management
		if arw, ok := c.Writer.(*auth.AuthResponseWriter); ok {
			ctx = context.WithValue(ctx, "authResponseWriter", arw)
		}

		// Add session manager to context for resolvers that need to validate sessions
		sessionMgr := session.NewSessionManager()
		ctx = context.WithValue(ctx, "session_manager", sessionMgr)

		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
