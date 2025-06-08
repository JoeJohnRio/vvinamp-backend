package server

import (
	"context"
	"crypto/tls"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"time"
	"vvinamp/internal/pkg/db/seeds"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"flag"

	"golang.org/x/crypto/acme/autocert"

	_ "github.com/go-sql-driver/mysql"
)

const defaultPort = "8080"

// Target is parameters to get all mux's dependencies
type Target struct {
	fx.In
	Environment string `name:"env"`
	Port        string `name:"port"`
	Lc          fx.Lifecycle
	Logger      *zap.Logger
}

// New is constructor to create Mux server on specific addr and port
func New(target Target) *gin.Engine {
	var man *autocert.Manager
	var server *http.Server
	r := gin.New()

	// zap.Logger integration with gin
	r.Use(ginzap.Ginzap(target.Logger, time.RFC3339, true))
	r.Use(ginzap.RecoveryWithZap(target.Logger, true))

	if target.Environment != "local" {
		host := ""
		man = &autocert.Manager{
			Prompt: autocert.AcceptTOS,
			Cache:  autocert.DirCache("certs"),
		}

		server = &http.Server{
			Addr:    host + ":443",
			Handler: r,
			TLSConfig: &tls.Config{
				GetCertificate: man.GetCertificate,
			},
		}
	} else {
		host := "localhost"
		server = &http.Server{
			Addr:    host + ":" + target.Port,
			Handler: r,
		}
	}

	target.Lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			if target.Environment != "local" {
				target.Logger.Info("Starting HTTPS server at " + server.Addr)
				go server.ListenAndServeTLS("", "")
				go http.ListenAndServe(":80", man.HTTPHandler(nil))
			} else {
				target.Logger.Info("Starting HTTP server at " + server.Addr)
				go server.ListenAndServe()
			}
			return nil
		},
		OnStop: func(ctx context.Context) error {
			target.Logger.Info("Stopping HTTPS server.")
			return server.Shutdown(ctx)
		},
	})

	return r
}

func handleArgs() {
	flag.Parse()
	args := flag.Args()

	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASS")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	name := os.Getenv("DB_NAME")

	if len(args) >= 1 {
		switch args[0] {
		case "seed":
			connString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local&multiStatements=true",
				user, pass, host, port, name)
			// connect DB
			db, err := sql.Open("mysql", connString)
			if err != nil {
				log.Fatalf("Error opening DB: %v", err)
			}

			seeds.Execute(
				db,
				"ArtistsSeed",
				"GenresSeed",
				"AlbumsSeed",
			)
			// seeds.Execute(db, args[1:]...)
			os.Exit(0)
		}
	}
}
