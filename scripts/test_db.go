package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	dsn := "host=aws-1-eu-west-3.pooler.supabase.com port=5432 user=postgres.bumcqqhbtmmrkeiwlsyy password=***REMOVED*** dbname=postgres sslmode=require"
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
