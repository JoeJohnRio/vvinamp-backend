package controller

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	"vvinamp/controller/auth"
	"vvinamp/database/postgre"
	redisdb "vvinamp/database/redis"
	"vvinamp/graphql/generated"
	"vvinamp/graphql/resolvers"
	"vvinamp/internal/pkg/session"
	youtubeService "vvinamp/internal/services/youtube"
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
	graphiQLEnable  bool
	auth            *auth.Auth
	postgre         *postgre.Connection
	user            user.Service
	history         history.Service
	song            song.Service
	artist          artistPackage.ArtistService
	search          search.Service
	streaming       *streaming.Service
	playlist        playlist.Service
	youtubeSearch   *youtubeService.SearchService
	youtubeDownload *youtubeService.DownloadService
	logger          *zap.Logger
	corsAllowOrigins    []string
	corsAllowCredentials bool
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
	Redis          *redisdb.Connection
}

// NewGraphQLController is a constructor for GraphQLController
func NewGraphQLController(target GraphQLControllerTarget) Result {
	ytDlpPath := os.Getenv("YT_DLP_PATH")
	if ytDlpPath == "" {
		for _, candidate := range []string{"yt-dlp", "/usr/bin/yt-dlp", "/usr/local/bin/yt-dlp", "./bin/yt-dlp", "bin/yt-dlp"} {
			if _, err := exec.LookPath(candidate); err == nil {
				ytDlpPath = candidate
				break
			}
			if _, err := os.Stat(candidate); err == nil {
				ytDlpPath = candidate
				break
			}
		}
		if ytDlpPath == "" {
			ytDlpPath = "yt-dlp"
		}
	}

	storagePath := os.Getenv("STORAGE_PATH")
	if storagePath == "" {
		storagePath = "./media/downloads"
	}

	maxDownloadSize := int64(500)
	if env := os.Getenv("MAX_DOWNLOAD_SIZE"); env != "" {
		if parsed, err := strconv.ParseInt(env, 10, 64); err == nil {
			maxDownloadSize = parsed
		}
	}

	downloadTimeout := 600
	if env := os.Getenv("DOWNLOAD_TIMEOUT"); env != "" {
		if parsed, err := strconv.Atoi(env); err == nil {
			downloadTimeout = parsed
		}
	}

	return Result{
		Controller: &GraphQLController{
			graphiQLEnable:  target.GraphiQLEnable,
			auth:            target.Auth,
			postgre:         target.Postgresql,
			user:            target.User,
			history:         target.History,
			song:            target.Song,
			artist:          target.Artist,
			search:          target.Search,
			streaming:       target.Streaming,
			playlist:        target.Playlist,
			youtubeSearch:   youtubeService.NewSearchService(ytDlpPath, target.Redis),
			youtubeDownload: youtubeService.NewDownloadService(storagePath, ytDlpPath, downloadTimeout, maxDownloadSize),
			logger:          target.Logger,
			corsAllowOrigins:    parseAllowedOrigins(os.Getenv("CORS_ALLOW_ORIGINS")),
			corsAllowCredentials: parseBoolEnv(os.Getenv("CORS_ALLOW_CREDENTIALS"), true),
		},
	}
}

func parseAllowedOrigins(env string) []string {
	if env == "" {
		return []string{
			"http://localhost:8080",
			"http://localhost:8081",
			"http://localhost:8082",
			"http://localhost:3000",
			"http://localhost:5173",
			"http://127.0.0.1:8080",
			"http://127.0.0.1:8081",
			"https://cirio.my.id",
			"https://vvinamp.cirio.my.id",
			"https://vvinamp-project.vercel.app",
		}
	}
	parts := strings.Split(env, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		t := strings.TrimSpace(p)
		if t != "" {
			out = append(out, t)
		}
	}
	return out
}

func parseBoolEnv(env string, def bool) bool {
	if env == "" {
		return def
	}
	v := strings.ToLower(strings.TrimSpace(env))
	return v == "1" || v == "true" || v == "yes"
}

// CORSMiddleware handles CORS preflight and sets CORS headers for actual requests
func (m *GraphQLController) CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		// Always add Vary: Origin
		c.Header("Vary", "Origin")

		// If no origin provided, continue
		if origin == "" {
			c.Next()
			return
		}

		// Check if origin is allowed
		allowed := false
		for _, ao := range m.corsAllowOrigins {
			if ao == "*" || ao == origin {
				allowed = true
				break
			}
		}
		if !allowed && (strings.HasSuffix(origin, ".vercel.app") || strings.HasSuffix(origin, ".cirio.my.id") || origin == "https://cirio.my.id") {
			allowed = true
		}

		if allowed {
			// If credentials are enabled we must not return '*'
			if m.corsAllowCredentials {
				c.Header("Access-Control-Allow-Origin", origin)
			} else {
				// If allow credentials is false and origin list contains '*', allow wildcard
				if contains(m.corsAllowOrigins, "*") {
					c.Header("Access-Control-Allow-Origin", "*")
				} else {
					c.Header("Access-Control-Allow-Origin", origin)
				}
			}

			c.Header("Access-Control-Allow-Methods", "POST, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Apollo-Require-Preflight")
			if m.corsAllowCredentials {
				c.Header("Access-Control-Allow-Credentials", "true")
			}
			c.Header("Access-Control-Max-Age", "86400")
		}

		// If this is a preflight request, respond immediately without auth
		if c.Request.Method == "OPTIONS" {
			// Successful preflight
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func contains(arr []string, s string) bool {
	for _, a := range arr {
		if a == s {
			return true
		}
	}
	return false
}

// GrqphQL is defining as the GraphQL handler
func (m *GraphQLController) GraphQL() gin.HandlerFunc {
	// Create resolver with required services
	resolver := resolvers.NewResolver(m.artist, m.user, m.history, m.song, m.streaming, m.playlist, m.youtubeSearch, m.youtubeDownload)

	// 1. Create a new Server instance using handler.New()
	srv := handler.New(generated.NewExecutableSchema(generated.Config{Resolvers: resolver}))

	// Enable introspection for development and schema exploration
	srv.Use(extension.Introspection{})

	// 2. Explicitly add the necessary transports for your application.
	// This makes the handler modular and production-ready.
	srv.AddTransport(transport.GET{})           // Supports GET requests
	srv.AddTransport(transport.POST{})          // Supports standard JSON POST requests
	srv.AddTransport(transport.MultipartForm{}) // Supports file uploads
	srv.AddTransport(transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second,
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	})

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
	// OPTIONS preflight — no auth required, must return CORS headers
	r.OPTIONS("/v1/graphql", CORSMiddleware())
	r.OPTIONS("/query", CORSMiddleware())

	graphqlHandlers := []gin.HandlerFunc{
		CORSMiddleware(),
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
		m.GraphQL(),
	}

	r.POST("/v1/graphql", graphqlHandlers...)
	r.POST("/query", graphqlHandlers...)

	if !m.graphiQLEnable {
		r.GET("/v1/graphiql", m.GraphiQL())
	}
}

// Middleware for GraphQL resolver to pass services into ctx
func (m *GraphQLController) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
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
