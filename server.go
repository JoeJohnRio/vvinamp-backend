package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/JoeJohnRio/youtube-music/graphql"
	"github.com/JoeJohnRio/youtube-music/graphql/resolvers"
	"github.com/JoeJohnRio/youtube-music/internal/pkg/db/seeds"
	"github.com/JoeJohnRio/youtube-music/internal/repository"
	"github.com/JoeJohnRio/youtube-music/internal/repository/album"
	"github.com/go-chi/chi/v5"

	"flag"

	_ "github.com/go-sql-driver/mysql"
)

const defaultPort = "8080"

func main() {

	// Set up the port
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	// Create router
	router := chi.NewRouter()

	// Uncomment this if you are using auth middleware
	// router.Use(auth.Middleware())

	// Initialize DB connection
	db, err := sql.Open("mysql", "root:password@tcp(localhost)/youtube_music_clone")
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	// Initialize repositories
	albumRepo := album.NewAlbumRepository(db)

	// Set up the root repository with injected dependencies
	repo := &repository.Repository{
		Album: albumRepo,
	}

	// Create resolver with repositories
	resolver := &resolvers.Resolver{
		Repo: repo,
	}

	// Create GraphQL server
	server := handler.NewDefaultServer(graphql.NewExecutableSchema(graphql.Config{
		Resolvers: resolver,
	}))

	// Set up the routes
	router.Handle("/", playground.Handler("GraphQL playground", "/graphql"))
	router.Handle("/graphql", server)

	// Log server start
	log.Printf("Starting server on http://localhost:%s", port)

	// Start the server
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
