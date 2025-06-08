package controller

import (
	context "context"

	"vvinamp/database/mysql"
	"vvinamp/graphql"
	"vvinamp/graphql/resolvers"
	"vvinamp/package/user"

	"github.com/99designs/gqlgen/handler"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// GraphQLController handle the graphql request, parse request to schema and return results
type GraphQLController struct {
	graphiQLEnable bool
	auth           *Auth
	mysql          *mysql.Connection
	user           user.Service
	logger         *zap.Logger
}

// GraphQLControllerTarget is parameter object for geting all GraphQLController's dependency
type GraphQLControllerTarget struct {
	fx.In
	GraphiQLEnable bool `name:"graphiql_enable"`
	Auth           *Auth
	Postgresql     *mysql.Connection
	User           user.Service
	Logger         *zap.Logger
}

// NewGraphQLController is a constructor for GraphQLController
func NewGraphQLController(target GraphQLControllerTarget) Result {
	return Result{
		Controller: &GraphQLController{
			graphiQLEnable: target.GraphiQLEnable,
			auth:           target.Auth,
			mysql:          target.Postgresql,
			user:           target.User,
			logger:         target.Logger,
		},
	}
}

// GrqphQL is defining as the GraphQL handler
func (m *GraphQLController) GraphQL() gin.HandlerFunc {
	h := handler.GraphQL(graphql.NewExecutableSchema(graphql.Config{Resolvers: &resolvers.Resolver{}}))

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// GraphiQL is defining as the GraphiQL Page handler
func (m *GraphQLController) GraphiQL() gin.HandlerFunc {
	h := handler.Playground("GraphQL", "/")

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// Register is function to register all controller's endpoint handler
func (m *GraphQLController) Register(r *gin.Engine) {
	r.Use(m.mysql.Middleware()).
		Use(m.Middleware()).
		Use(m.auth.Middleware()).
		POST("/v1/graphql", m.GraphQL())
	if !m.graphiQLEnable {
		r.GET("/v1/graphiql", m.GraphiQL())
	}
}

// Middleware for GraphQL resolver to pass services into ctx
func (m *GraphQLController) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, user.Key, m.user)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
