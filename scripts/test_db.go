package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	dsn := os.Getenv("POSTGRE_URL")
	if dsn == "" {
		log.Fatal("POSTGRE_URL is not set")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("sql.Open failed: %v", err)
	}
	defer db.Close()

	var artistsCount, albumsCount, tracksCount, playlistsCount int
	db.QueryRow("SELECT COUNT(*) FROM artists").Scan(&artistsCount)
	db.QueryRow("SELECT COUNT(*) FROM albums").Scan(&albumsCount)
	db.QueryRow("SELECT COUNT(*) FROM tracks").Scan(&tracksCount)
	db.QueryRow("SELECT COUNT(*) FROM playlists").Scan(&playlistsCount)

	fmt.Printf("DATABASE VERIFIED:\n")
	fmt.Printf(" - Artists: %d\n", artistsCount)
	fmt.Printf(" - Albums: %d\n", albumsCount)
	fmt.Printf(" - Tracks: %d\n", tracksCount)
	fmt.Printf(" - Playlists: %d\n", playlistsCount)
}
