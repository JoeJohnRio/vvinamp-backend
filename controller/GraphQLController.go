package controller

import (
	context "context"
	"log"

	"vvinamp/controller/auth"
	"vvinamp/database/postgre"
	"vvinamp/graphql"
	"vvinamp/graphql/resolvers"
	"vvinamp/package/history"
	"vvinamp/package/user"

	"github.com/99designs/gqlgen/graphql/handler"
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
			logger:         target.Logger,
		},
	}
}

// GrqphQL is defining as the GraphQL handler
// GrqphQL is defining as the GraphQL handler
func (m *GraphQLController) GraphQL() gin.HandlerFunc {
	// 1. Create a new Server instance using handler.New()
	srv := handler.New(graphql.NewExecutableSchema(graphql.Config{Resolvers: &resolvers.Resolver{}}))

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
	r.Use(m.auth.Middleware()).
		Use(m.postgre.Middleware()).
		Use(m.Middleware()).
		POST("/v1/graphql", m.GraphQL())
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

		// Pass the response writer for session management
		if arw, ok := c.Writer.(*auth.AuthResponseWriter); ok {
			ctx = context.WithValue(ctx, "authResponseWriter", arw)
		}

		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
