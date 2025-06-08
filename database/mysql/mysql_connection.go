package mysql

import (
	"context"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Target contains injected dependencies from Fx
type Target struct {
	fx.In
	MySQLURL string `name:"mysql_url" optional:"true"`
	Lc       fx.Lifecycle
	Logger   *zap.Logger
}

// Connection wraps the MySQL DB client
type Connection struct {
	DB *gorm.DB
}

// New creates a new GORM MySQL connection
func New(target Target) (*Connection, error) {
	if target.MySQLURL == "" {
		return nil, nil
	}

	db, err := gorm.Open(mysql.Open(target.MySQLURL), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	target.Lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			target.Logger.Info("Connecting to MySQL at " + target.MySQLURL)
			return sqlDB.PingContext(ctx)
		},
		OnStop: func(ctx context.Context) error {
			target.Logger.Info("Disconnecting MySQL at " + target.MySQLURL)
			return sqlDB.Close()
		},
	})

	return &Connection{DB: db}, nil
}

type ctxKey string

const (
	mysqlKey ctxKey = "mysql_client"
)

func (m *Connection) Client() *gorm.DB {
	return m.DB
}

// Middleware injects the MySQL DB into the Gin context
func (m *Connection) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if m != nil {
			ctx := context.WithValue(c.Request.Context(), mysqlKey, m.DB)
			c.Request = c.Request.WithContext(ctx)
		}
		c.Next()
	}
}

// WithContext returns a new context with the DB injected
func (m *Connection) WithContext(ctx context.Context) context.Context {
	if m != nil {
		return context.WithValue(ctx, mysqlKey, m.DB)
	}
	return ctx
}

// ForContext extracts the DB client from context
func ForContext(ctx context.Context) *gorm.DB {
	db, ok := ctx.Value(mysqlKey).(*gorm.DB)
	if !ok {
		panic("context does not contain MySQL client")
	}
	return db
}
