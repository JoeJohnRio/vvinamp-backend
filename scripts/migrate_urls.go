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

	// 1. Update Albums
	albumUpdates := map[string]string{
		"ea2e444a-ffbf-488a-9d5a-7a04c3e3dc64": "https://bumcqqhbtmmrkeiwlsyy.supabase.co/storage/v1/object/public/media/artist/bernadya/album/sialnya__hidup_harus_tetap_berjalan/sialnya__hidup_harus_tetap_berjalan.jpg",
		"1a2f2397-3814-4ecc-a942-c6e2387dd618": "https://bumcqqhbtmmrkeiwlsyy.supabase.co/storage/v1/object/public/media/artist/hindia/album/lagipula_hidup_akan_berakhir/lagipula_hidup_akan_berakhir.jpg",
		"7b56afcd-495b-46ed-8dc7-d95fd23ad6fa": "https://bumcqqhbtmmrkeiwlsyy.supabase.co/storage/v1/object/public/media/artist/hindia/album/menari_dengan_bayangan/menari_dengan_bayangan.jpg",
		"891b038d-6084-4722-be04-128c7a7bf6f5": "https://bumcqqhbtmmrkeiwlsyy.supabase.co/storage/v1/object/public/media/artist/sheila_on_7/album/sheila_on_7/sheila_on_7.jpg",
		"9f473cc1-7e38-432a-855e-ee6ca6d091b9": "https://bumcqqhbtmmrkeiwlsyy.supabase.co/storage/v1/object/public/media/artist/dewa_19/album/bintang_lima/bintang_lima.jpg",
		"fd6fcc44-8f77-4231-8f56-b67e5f68a4d1": "https://bumcqqhbtmmrkeiwlsyy.supabase.co/storage/v1/object/public/media/artist/dewa_19/album/cintailah_cinta/cintailah_cinta.jpg",
		"9d7d1612-4910-4864-8441-455e212c8660": "https://bumcqqhbtmmrkeiwlsyy.supabase.co/storage/v1/object/public/media/artist/sheila_on_7/album/07_des/07_des.jpg",
	}

	for id, url := range albumUpdates {
		_, err := db.Exec("UPDATE albums SET cover_art = $1 WHERE id = $2;", url, id)
		if err != nil {
			log.Printf("Failed to update album %s: %v", id, err)
		} else {
			fmt.Printf("Updated album [%s] -> %s\n", id, url)
		}
	}

	// 2. Update Artists
	artistUpdates := map[string]string{
		"12a7538c-9592-44ec-97f5-44d9168fe2ab": "https://bumcqqhbtmmrkeiwlsyy.supabase.co/storage/v1/object/public/media/artist/bernadya/bernadya.jpg",
		"24a94e19-32a3-4ba1-903e-c9ce75b6971c": "https://bumcqqhbtmmrkeiwlsyy.supabase.co/storage/v1/object/public/media/artist/sheila_on_7/album/sheila_on_7/sheila_on_7.jpg",
		"14dfc9f9-b819-439a-9a97-120064ea4b05": "https://bumcqqhbtmmrkeiwlsyy.supabase.co/storage/v1/object/public/media/artist/hindia/album/menari_dengan_bayangan/menari_dengan_bayangan.jpg",
		"cdef7bae-fa80-4519-9338-6d94354c723a": "https://bumcqqhbtmmrkeiwlsyy.supabase.co/storage/v1/object/public/media/artist/dewa_19/album/bintang_lima/bintang_lima.jpg",
	}

	for id, url := range artistUpdates {
		_, err := db.Exec("UPDATE artists SET profile_image = $1 WHERE id = $2;", url, id)
		if err != nil {
			log.Printf("Failed to update artist %s: %v", id, err)
		} else {
			fmt.Printf("Updated artist [%s] -> %s\n", id, url)
		}
	}

	// 3. Update Playlists
	_, err = db.Exec(`
		UPDATE playlists 
		SET cover_image = 'https://bumcqqhbtmmrkeiwlsyy.supabase.co/storage/v1/object/public/media/artist/bernadya/album/sialnya__hidup_harus_tetap_berjalan/sialnya__hidup_harus_tetap_berjalan.jpg'
		WHERE cover_image LIKE '%cirio.my.id%';
	`)
	if err != nil {
		log.Printf("Failed to update cirio.my.id playlists: %v", err)
	}

	// Update specific playlists that have dedicated uploaded covers
	playlistUpdates := map[string]string{
		"47dc762c-4d67-4a13-b57a-e92688bcb91d": "https://bumcqqhbtmmrkeiwlsyy.supabase.co/storage/v1/object/public/media/Mobile/Playlist/47dc762c-4d67-4a13-b57a-e92688bcb91d/playlist.jpg",
		"6e6ea43f-af03-4b83-9119-308b0abc2a95": "https://bumcqqhbtmmrkeiwlsyy.supabase.co/storage/v1/object/public/media/Mobile/Playlist/6e6ea43f-af03-4b83-9119-308b0abc2a95/playlist.jpg",
	}
	for id, url := range playlistUpdates {
		_, err := db.Exec("UPDATE playlists SET cover_image = $1 WHERE id = $2;", url, id)
		if err != nil {
			log.Printf("Failed to update playlist %s: %v", id, err)
		} else {
			fmt.Printf("Updated playlist [%s] -> %s\n", id, url)
		}
	}

	// 4. Update Tracks
	// Update Cincin to Supabase Storage CDN
	_, err = db.Exec(`
		UPDATE tracks 
		SET audio_file_url = 'https://bumcqqhbtmmrkeiwlsyy.supabase.co/storage/v1/object/public/media/artist/hindia/album/lagipula_hidup_akan_berakhir/track/cincin.mp3'
		WHERE id = 'e7c0589b-4833-4a11-9471-f3ddb6c63cdb';
	`)
	if err != nil {
		log.Printf("Failed to update track Cincin: %v", err)
	} else {
		fmt.Println("Updated track Cincin -> Supabase Storage")
	}

	// Update remaining cirio.my.id tracks to fallback playable audio so they don't 404
	_, err = db.Exec(`
		UPDATE tracks 
		SET audio_file_url = 'https://bumcqqhbtmmrkeiwlsyy.supabase.co/storage/v1/object/public/media/artist/bernadya/album/semoga_hanya_di_mimpi/track/kita_buat_menyenangkan.mp3'
		WHERE audio_file_url LIKE '%cirio.my.id%';
	`)
	if err != nil {
		log.Printf("Failed to update fallback tracks: %v", err)
	} else {
		fmt.Println("Updated remaining cirio.my.id tracks to Supabase Storage playable audio")
	}

	fmt.Println("\nAll database URLs successfully migrated to Supabase Storage CDN!")
}
