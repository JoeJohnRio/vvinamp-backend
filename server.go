package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/codegen/testserver/compliant-int/generated-default"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/JoeJohnRio/youtube-music/graph"
	"github.com/JoeJohnRio/youtube-music/internal/auth"
	database "github.com/JoeJohnRio/youtube-music/internal/pkg/db/migrations/mysql"
	"github.com/JoeJohnRio/youtube-music/internal/pkg/db/seeds"
	"github.com/JoeJohnRio/youtube-music/internal/repository"
	"github.com/go-chi/chi/v5"

	"flag"

	_ "github.com/go-sql-driver/mysql"
)

const defaultPort = "8080"

func main() {

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	router := chi.NewRouter()

	router.Use(auth.Middleware())

	database.InitDB()
	database.Migrate()

	albumRepo := repository.NewAlbumRepository(database.Db)

	resolver := &graph.Resolver{
		AlbumRepo: albumRepo,
	}

	// Create server config WITH the resolver
	server := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{
		Resolvers: resolver, // This is where the resolver is actually used
	}))

	// server := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{}}))
	router.Handle("/", playground.Handler("GraphQL playground", "/query"))
	router.Handle("/query", server)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, router))

	// godotenv.Load()
	// handleArgs()
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
