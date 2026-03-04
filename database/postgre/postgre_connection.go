package postgre

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Target struct {
	fx.In
	// IMPORTANT: make sure this matches your Fx provider name exactly.
	// If your logs say [name="postgre_url"], use postgre_url here and in config.
	PostgresURL string `name:"postgre_url"`
	Lc          fx.Lifecycle
	Logger      *zap.Logger
}

type Connection struct {
	db *gorm.DB
}

func New(t Target) (*Connection, error) {
	if t.PostgresURL == "" {
		return nil, fmt.Errorf("postgre_url is empty")
	}

	db, err := gorm.Open(postgres.Open(t.PostgresURL), &gorm.Config{
		Logger: nil, // Disable GORM logger
	})
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("sql DB handle: %w", err)
	}

	// Fail fast if the DB is not reachable
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}

	t.Lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			t.Logger.Info("PostgreSQL connected", zap.String("dsn", t.PostgresURL))
			return nil
		},
		OnStop: func(ctx context.Context) error {
			t.Logger.Info("PostgreSQL disconnect")
			return sqlDB.Close()
		},
	})

	return &Connection{db: db}, nil
}

// Return error if not initialized, so callers won't nil-deref.
func (c *Connection) Client() (*gorm.DB, error) {
	if c == nil || c.db == nil {
		return nil, fmt.Errorf("postgres connection not initialized")
	}
	return c.db, nil
}

type ctxKey string

const postgreKey ctxKey = "postgre_client"

func (c *Connection) Middleware() gin.HandlerFunc {
	return func(g *gin.Context) {
		if c != nil && c.db != nil {
			ctx := context.WithValue(g.Request.Context(), postgreKey, c.db)
			g.Request = g.Request.WithContext(ctx)
		}
		g.Next()
	}
}

func (c *Connection) WithContext(ctx context.Context) context.Context {
	if c != nil && c.db != nil {
		return context.WithValue(ctx, postgreKey, c.db)
	}
	return ctx
}

// If you want to keep this panic, ensure Middleware always runs before handlers.
// Alternatively, return nil and let callers handle it.
func ForContext(ctx context.Context) *gorm.DB {
	db, _ := ctx.Value(postgreKey).(*gorm.DB)
	return db
}
