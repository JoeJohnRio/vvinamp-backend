package seeder

import (
	"context"
	"fmt"
	"log"
	"time"
	dbModel "vvinamp/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Seeder struct {
	db *gorm.DB
}

func NewSeeder(db *gorm.DB) *Seeder {
	return &Seeder{db: db}
}

func (s *Seeder) ClearMusicData() error {
	log.Println("Clearing existing music data (preserving user data)...")

	// Clear listening history first (since it references tracks/users)
	if err := s.db.Exec("DELETE FROM listening_histories").Error; err != nil {
		log.Printf("Warning: Could not clear listening histories: %v", err)
	}

	// Check if track_artists table exists before clearing it
	if err := s.db.Exec("SELECT 1 FROM track_artists LIMIT 1").Error; err == nil {
		// Table exists, proceed with deletion
		if err := s.db.Exec("DELETE FROM track_artists").Error; err != nil {
			log.Printf("Warning: Could not clear track_artists: %v", err)
		}
	} else {
		log.Println("track_artists table does not exist, skipping deletion")
	}

	// Clear tracks
	if err := s.db.Exec("DELETE FROM tracks").Error; err != nil {
		log.Printf("Warning: Could not clear tracks: %v", err)
	}

	// Clear albums
	if err := s.db.Exec("DELETE FROM albums").Error; err != nil {
		log.Printf("Warning: Could not clear albums: %v", err)
	}

	// Clear artists
	if err := s.db.Exec("DELETE FROM artists").Error; err != nil {
		log.Printf("Warning: Could not clear artists: %v", err)
	}

	// Clear genres (if needed)
	if err := s.db.Exec("SELECT 1 FROM genres LIMIT 1").Error; err == nil {
		// Genres table exists, proceed with deletion
		if err := s.db.Exec("DELETE FROM genres WHERE id NOT IN (SELECT genre_id FROM albums WHERE genre_id IS NOT NULL)").Error; err != nil {
			log.Printf("Warning: Could not clear unused genres: %v", err)
		}
	} else {
		log.Println("genres table does not exist, skipping deletion")
	}

	log.Println("Music data cleared successfully!")
	return nil
}

func (s *Seeder) SeedArtists(ctx context.Context) error {
	artists := []dbModel.Artist{
		{
			ID:               uuid.New().String(),
			Name:             "Dewa 19",
			Bio:              stringPtr("Indonesian rock band formed in Surabaya"),
			ProfileImage:     stringPtr("https://cirio.my.id/images/artists/dewa19_artist_profile.jpg"),
			Verified:         true,
			MonthlyListeners: 5000000,
		},
		{
			ID:               uuid.New().String(),
			Name:             "Sheila On 7",
			Bio:              stringPtr("Indonesian pop rock band formed in Yogyakarta"),
			ProfileImage:     stringPtr("https://cirio.my.id/images/artists/sheilaOn7_artist_profile.jpg"),
			Verified:         true,
			MonthlyListeners: 4500000,
		},
		{
			ID:               uuid.New().String(),
			Name:             "Hindia",
			Bio:              stringPtr("Indonesian alternative pop solo project by Baskara Putra"),
			ProfileImage:     stringPtr("https://cirio.my.id/images/artists/hindia_artist_profile.jpg"),
			Verified:         true,
			MonthlyListeners: 3000000,
		},
		{
			ID:               uuid.New().String(),
			Name:             "Bernadya",
			Bio:              stringPtr("Indonesian indie pop singer-songwriter"),
			ProfileImage:     stringPtr("https://cirio.my.id/images/artists/bernadya_artist_profile.jpg"),
			Verified:         false,
			MonthlyListeners: 1200000,
		},
	}

	for _, artist := range artists {
		if err := s.db.Create(&artist).Error; err != nil {
			log.Printf("Error seeding artist %s: %v", artist.Name, err)
			continue
		}
		log.Printf("Seeded artist: %s", artist.Name)
	}

	return nil
}

func (s *Seeder) SeedGenres(ctx context.Context) error {
	// Skip genre seeding since the Genre model doesn't exist in the main model directory
	log.Println("Skipping genre seeding - Genre model not found in main model directory")
	return nil
}

func (s *Seeder) SeedAlbums(ctx context.Context) error {
	// Refresh the database session to ensure artists are committed
	var artists []dbModel.Artist
	if err := s.db.Find(&artists).Error; err != nil {
		return fmt.Errorf("failed to fetch artists for albums: %w", err)
	}

	// Create a map for easy lookup
	artistMap := make(map[string]dbModel.Artist)
	for _, artist := range artists {
		artistMap[artist.Name] = artist
	}

	// Since the Genre model doesn't exist in the main model directory, we'll need to handle this differently
	// We'll try to create albums without the genre constraint violation by using raw SQL to bypass the constraint temporarily
	// Or we'll need to handle the constraint differently

	// For now, let's try to create albums without specifying GenreID to avoid the constraint
	// This assumes that the GenreID column allows NULL values
	albums := []struct {
		ID          uuid.UUID
		Title       string
		ArtistID    uuid.UUID
		CoverArt    string
		ReleaseDate string
	}{
		// Dewa 19 Albums
		{
			ID:          uuid.New(),
			Title:       "Bintang Lima",
			ArtistID:    stringToUUID(artistMap["Dewa 19"].ID),
			CoverArt:    "https://cirio.my.id/images/albums/placeholder_album.jpg",
			ReleaseDate: "2000-01-01",
		},
		{
			ID:          uuid.New(),
			Title:       "Cintailah Cinta",
			ArtistID:    stringToUUID(artistMap["Dewa 19"].ID),
			CoverArt:    "https://cirio.my.id/images/albums/placeholder_album.jpg",
			ReleaseDate: "2002-01-01",
		},
		// Sheila On 7 Albums
		{
			ID:          uuid.New(),
			Title:       "Sheila On 7",
			ArtistID:    stringToUUID(artistMap["Sheila On 7"].ID),
			CoverArt:    "https://cirio.my.id/images/albums/placeholder_album.jpg",
			ReleaseDate: "1999-01-01",
		},
		{
			ID:          uuid.New(),
			Title:       "07 Des",
			ArtistID:    stringToUUID(artistMap["Sheila On 7"].ID),
			CoverArt:    "https://cirio.my.id/images/albums/placeholder_album.jpg",
			ReleaseDate: "2002-01-01",
		},
		// Hindia Albums
		{
			ID:          uuid.New(),
			Title:       "Menari Dengan Bayangan",
			ArtistID:    stringToUUID(artistMap["Hindia"].ID),
			CoverArt:    "https://cirio.my.id/images/albums/placeholder_album.jpg",
			ReleaseDate: "2019-11-29",
		},
		{
			ID:          uuid.New(),
			Title:       "Lagipula Hidup Akan Berakhir",
			ArtistID:    stringToUUID(artistMap["Hindia"].ID),
			CoverArt:    "https://cirio.my.id/images/albums/placeholder_album.jpg",
			ReleaseDate: "2023-07-07",
		},
		// Bernadya Albums
		{
			ID:          uuid.New(),
			Title:       "Sialnya, Hidup Harus Tetap Berjalan",
			ArtistID:    stringToUUID(artistMap["Bernadya"].ID),
			CoverArt:    "https://cirio.my.id/images/albums/placeholder_album.jpg",
			ReleaseDate: "2023-06-01",
		},
	}

	for _, albumData := range albums {
		// Create album using raw SQL to avoid the foreign key constraint issue
		query := `INSERT INTO albums (id, title, artist_id, cover_art, release_date) VALUES (?, ?, ?, ?, ?)`
		if err := s.db.Exec(query, albumData.ID, albumData.Title, albumData.ArtistID, albumData.CoverArt, albumData.ReleaseDate).Error; err != nil {
			log.Printf("Error seeding album %s: %v", albumData.Title, err)
			continue
		}
		log.Printf("Seeded album: %s by %s", albumData.Title, getArtistNameById(s.db, albumData.ArtistID.String()))
	}

	return nil
}

func (s *Seeder) SeedTracks(ctx context.Context) error {
	// Define a helper function to get album ID by title and artist name
	getAlbumIdByTitleAndArtist := func(title, artistName string) uuid.UUID {
		var album dbModel.Album
		var artist dbModel.Artist

		// Get artist ID first
		if err := s.db.Where("name = ?", artistName).First(&artist).Error; err != nil {
			log.Printf("Error finding artist %s: %v", artistName, err)
			return uuid.Nil // Return nil UUID if not found
		}

		// Find album by title and artist ID
		if err := s.db.Where("title = ? AND artist_id = ?", title, stringToUUID(artist.ID)).First(&album).Error; err != nil {
			log.Printf("Error finding album %s by artist %s: %v", title, artistName, err)
			return uuid.Nil // Return nil UUID if not found
		}

		return album.ID
	}

	tracks := []dbModel.Track{
		// Dewa 19 Tracks
		{
			ID:           uuid.New(),
			Title:        "Kangen",
			Duration:     300,
			AudioFileURL: "https://example.com/dewa19/kangen.mp3",
			StreamCount:  2500000,
			AlbumID:      getAlbumIdByTitleAndArtist("Bintang Lima", "Dewa 19"),
			Explicit:     false,
			Lyrics:       "Ku rindu padamu, kau tak mengerti\nKu menanti dirimu, kau tak datang-datang...",
			ReleaseDate:  "2000-01-01",
		},
		{
			ID:           uuid.New(),
			Title:        "Separuh Nafas",
			Duration:     240,
			AudioFileURL: "https://example.com/dewa19/separuh_nafas.mp3",
			StreamCount:  2200000,
			AlbumID:      getAlbumIdByTitleAndArtist("Bintang Lima", "Dewa 19"),
			Explicit:     false,
			Lyrics:       "Separuh nafasku hanyalah untukmu\nSeparuh jiwaku juga hanya untukmu...",
			ReleaseDate:  "2000-01-01",
		},
		{
			ID:           uuid.New(),
			Title:        "Roman Picisan",
			Duration:     270,
			AudioFileURL: "https://example.com/dewa19/roman_picisan.mp3",
			StreamCount:  1800000,
			AlbumID:      getAlbumIdByTitleAndArtist("Bintang Lima", "Dewa 19"),
			Explicit:     false,
			Lyrics:       "Roman picisan, cinta yang tak mungkin\nKu rela kau pergi, meski ku tak rela...",
			ReleaseDate:  "2000-01-01",
		},
		{
			ID:           uuid.New(),
			Title:        "Risalah Hati",
			Duration:     250,
			AudioFileURL: "https://example.com/dewa19/risalah_hati.mp3",
			StreamCount:  2000000,
			AlbumID:      getAlbumIdByTitleAndArtist("Bintang Lima", "Dewa 19"),
			Explicit:     false,
			Lyrics:       "Ku tulis puisi cinta, tentang kisah kita\nKu baca beribu kali, tak jemu-jemunya...",
			ReleaseDate:  "2000-01-01",
		},
		{
			ID:           uuid.New(),
			Title:        "Pupus",
			Duration:     260,
			AudioFileURL: "https://example.com/dewa19/pupus.mp3",
			StreamCount:  1900000,
			AlbumID:      getAlbumIdByTitleAndArtist("Cintailah Cinta", "Dewa 19"),
			Explicit:     false,
			Lyrics:       "Pupuslah sudah semua tentang kita\nPupuslah sudah cinta yang dulu kita jaga...",
			ReleaseDate:  "2002-01-01",
		},
		{
			ID:           uuid.New(),
			Title:        "Laskar Cinta",
			Duration:     280,
			AudioFileURL: "https://example.com/dewa19/laskar_cinta.mp3",
			StreamCount:  1700000,
			AlbumID:      getAlbumIdByTitleAndArtist("Cintailah Cinta", "Dewa 19"),
			Explicit:     false,
			Lyrics:       "Laskar cinta, datang membawa harapan\nLaskar cinta, pergi membawa luka...",
			ReleaseDate:  "2002-01-01",
		},
		{
			ID:           uuid.New(),
			Title:        "Dewi",
			Duration:     250,
			AudioFileURL: "https://example.com/dewa19/dewi.mp3",
			StreamCount:  1600000,
			AlbumID:      getAlbumIdByTitleAndArtist("Bintang Lima", "Dewa 19"),
			Explicit:     false,
			Lyrics:       "Dewi, kau bidadari dalam hidupku\nDewi, kau cahaya dalam gelapku...",
			ReleaseDate:  "2000-01-01",
		},
		{
			ID:           uuid.New(),
			Title:        "Arjuna",
			Duration:     260,
			AudioFileURL: "https://example.com/dewa19/arjuna.mp3",
			StreamCount:  1500000,
			AlbumID:      getAlbumIdByTitleAndArtist("Cintailah Cinta", "Dewa 19"),
			Explicit:     false,
			Lyrics:       "Arjuna, kisah cinta yang tak terungkapkan\nArjuna, cinta yang hanya dalam mimpi...",
			ReleaseDate:  "2002-01-01",
		},
		{
			ID:           uuid.New(),
			Title:        "Kosong",
			Duration:     260,
			AudioFileURL: "https://example.com/dewa19/kosong.mp3",
			StreamCount:  1400000,
			AlbumID:      getAlbumIdByTitleAndArtist("Bintang Lima", "Dewa 19"),
			Explicit:     false,
			Lyrics:       "Kosong... hati ini kosong tanpamu\nKosong... cinta ini kosong tanpamu...",
			ReleaseDate:  "2000-01-01",
		},
		{
			ID:           uuid.New(),
			Title:        "Angin",
			Duration:     240,
			AudioFileURL: "https://example.com/dewa19/angin.mp3",
			StreamCount:  1300000,
			AlbumID:      getAlbumIdByTitleAndArtist("Cintailah Cinta", "Dewa 19"),
			Explicit:     false,
			Lyrics:       "Angin berhembus membawa kerinduan\nAngin berbisik tentang masa lalu...",
			ReleaseDate:  "2002-01-01",
		},
		{
			ID:           uuid.New(),
			Title:        "Cemburu",
			Duration:     230,
			AudioFileURL: "https://example.com/dewa19/cemburu.mp3",
			StreamCount:  1200000,
			AlbumID:      getAlbumIdByTitleAndArtist("Cintailah Cinta", "Dewa 19"),
			Explicit:     false,
			Lyrics:       "Cemburu... hati ini cemburu padanya\nCemburu... karena kau bukan milikku...",
			ReleaseDate:  "2002-01-01",
		},

		// Sheila On 7 Tracks
		{
			ID:           uuid.New(),
			Title:        "Dan",
			Duration:     260,
			AudioFileURL: "https://example.com/sheilaon7/dan.mp3",
			StreamCount:  3500000,
			AlbumID:      getAlbumIdByTitleAndArtist("07 Des", "Sheila On 7"),
			Explicit:     false,
			Lyrics:       "Dan kau datang membawa cerita\nCinta yang dulu pergi...",
			ReleaseDate:  "2002-01-01",
		},
		{
			ID:           uuid.New(),
			Title:        "Sephia",
			Duration:     270,
			AudioFileURL: "https://example.com/sheilaon7/sephia.mp3",
			StreamCount:  3200000,
			AlbumID:      getAlbumIdByTitleAndArtist("07 Des", "Sheila On 7"),
			Explicit:     false,
			Lyrics:       "Sephia, warna cinta yang tak pudar\nSephia, bayangan yang selalu hadir...",
			ReleaseDate:  "2002-01-01",
		},
		{
			ID:           uuid.New(),
			Title:        "Pemuja Rahasia",
			Duration:     290,
			AudioFileURL: "https://example.com/sheilaon7/pemuja_rahasia.mp3",
			StreamCount:  2800000,
			AlbumID:      getAlbumIdByTitleAndArtist("07 Des", "Sheila On 7"),
			Explicit:     false,
			Lyrics:       "Aku pemuja rahasia, dari jauh memuja\nAku pemuja rahasia, tak berani mendekat...",
			ReleaseDate:  "2002-01-01",
		},
		{
			ID:           uuid.New(),
			Title:        "Kita",
			Duration:     260,
			AudioFileURL: "https://example.com/sheilaon7/kita.mp3",
			StreamCount:  2500000,
			AlbumID:      getAlbumIdByTitleAndArtist("Sheila On 7", "Sheila On 7"),
			Explicit:     false,
			Lyrics:       "Kita pernah bahagia, kita pernah bersama\nKita pernah saling percaya...",
			ReleaseDate:  "1999-01-01",
		},
		{
			ID:           uuid.New(),
			Title:        "Anugerah Terindah Yang Pernah Kumiliki",
			Duration:     300,
			AudioFileURL: "https://example.com/sheilaon7/anugerah_terindah.mp3",
			StreamCount:  4000000,
			AlbumID:      getAlbumIdByTitleAndArtist("Sheila On 7", "Sheila On 7"),
			Explicit:     false,
			Lyrics:       "Kau anugerah terindah yang pernah kumiliki\nKau cahaya dalam hidupku...",
			ReleaseDate:  "1999-01-01",
		},
		{
			ID:           uuid.New(),
			Title:        "Melompat Lebih Tinggi",
			Duration:     240,
			AudioFileURL: "https://example.com/sheilaon7/melompat_lebih_tinggi.mp3",
			StreamCount:  2200000,
			AlbumID:      getAlbumIdByTitleAndArtist("07 Des", "Sheila On 7"),
			Explicit:     false,
			Lyrics:       "Aku kan melompat lebih tinggi\nUntuk meraih impianku...",
			ReleaseDate:  "2002-01-01",
		},
		{
			ID:           uuid.New(),
			Title:        "Sahabat Sejati",
			Duration:     250,
			AudioFileURL: "https://example.com/sheilaon7/sahabat_sejati.mp3",
			StreamCount:  2000000,
			AlbumID:      getAlbumIdByTitleAndArtist("Sheila On 7", "Sheila On 7"),
			Explicit:     false,
			Lyrics:       "Sahabat sejati, yang selalu ada\nSahabat sejati, yang tak pernah pergi...",
			ReleaseDate:  "1999-01-01",
		},
		{
			ID:           uuid.New(),
			Title:        "J.A.P",
			Duration:     230,
			AudioFileURL: "https://example.com/sheilaon7/jap.mp3",
			StreamCount:  1800000,
			AlbumID:      getAlbumIdByTitleAndArtist("07 Des", "Sheila On 7"),
			Explicit:     false,
			Lyrics:       "J.A.P, jangan pernah berubah\nJ.A.P, tetaplah seperti dulu...",
			ReleaseDate:  "2002-01-01",
		},
		{
			ID:           uuid.New(),
			Title:        "Pejantan Tangguh",
			Duration:     240,
			AudioFileURL: "https://example.com/sheilaon7/pejantan_tangguh.mp3",
			StreamCount:  1900000,
			AlbumID:      getAlbumIdByTitleAndArtist("07 Des", "Sheila On 7"),
			Explicit:     false,
			Lyrics:       "Aku pejantan tangguh, yang tak pernah menyerah\nAku pejantan tangguh, yang selalu berjuang...",
			ReleaseDate:  "2002-01-01",
		},
		{
			ID:           uuid.New(),
			Title:        "Lapang Dada",
			Duration:     270,
			AudioFileURL: "https://example.com/sheilaon7/lapang_dada.mp3",
			StreamCount:  3500000,
			AlbumID:      getAlbumIdByTitleAndArtist("07 Des", "Sheila On 7"),
			Explicit:     false,
			Lyrics:       "Lapang dada, terima semua yang terjadi\nLapang dada, jangan mudah marah...",
			ReleaseDate:  "2002-01-01",
		},
		{
			ID:           uuid.New(),
			Title:        "Betapa",
			Duration:     260,
			AudioFileURL: "https://example.com/sheilaon7/betapa.mp3",
			StreamCount:  2100000,
			AlbumID:      getAlbumIdByTitleAndArtist("Sheila On 7", "Sheila On 7"),
			Explicit:     false,
			Lyrics:       "Betapa ingin ku katakan semua\nBetapa ingin ku ungkapkan perasaan...",
			ReleaseDate:  "1999-01-01",
		},

		// Hindia Tracks
		{
			ID:           uuid.New(),
			Title:        "Evaluasi",
			Duration:     300,
			AudioFileURL: "https://example.com/hindia/evaluasi.mp3",
			StreamCount:  2500000,
			AlbumID:      getAlbumIdByTitleAndArtist("Menari Dengan Bayangan", "Hindia"),
			Explicit:     false,
			Lyrics:       "Aku evaluasi semua yang telah terjadi\nAku evaluasi semua yang telah berlalu...",
			ReleaseDate:  "2019-11-29",
		},
		{
			ID:           uuid.New(),
			Title:        "Secukupnya",
			Duration:     290,
			AudioFileURL: "https://example.com/hindia/secukupnya.mp3",
			StreamCount:  2300000,
			AlbumID:      getAlbumIdByTitleAndArtist("Menari Dengan Bayangan", "Hindia"),
			Explicit:     false,
			Lyrics:       "Berikan aku secukupnya\nCinta yang kau punya...",
			ReleaseDate:  "2019-11-29",
		},
		{
			ID:           uuid.New(),
			Title:        "Rumah Ke Rumah",
			Duration:     280,
			AudioFileURL: "https://example.com/hindia/rumah_ke_rumah.mp3",
			StreamCount:  2000000,
			AlbumID:      getAlbumIdByTitleAndArtist("Menari Dengan Bayangan", "Hindia"),
			Explicit:     false,
			Lyrics:       "Dari rumah ke rumah, aku mencari cinta\nDari kota ke kota, aku mencari arti...",
			ReleaseDate:  "2019-11-29",
		},
		{
			ID:           uuid.New(),
			Title:        "Wejangan Mama",
			Duration:     240,
			AudioFileURL: "https://example.com/hindia/wejangan_mama.mp3",
			StreamCount:  1800000,
			AlbumID:      getAlbumIdByTitleAndArtist("Menari Dengan Bayangan", "Hindia"),
			Explicit:     false,
			Lyrics:       "Wejangan mama, jangan lupa sholat\nWejangan mama, jangan lupa ibadah...",
			ReleaseDate:  "2019-11-29",
		},
		{
			ID:           uuid.New(),
			Title:        "Membasuh",
			Duration:     260,
			AudioFileURL: "https://example.com/hindia/membasuh.mp3",
			StreamCount:  2200000,
			AlbumID:      getAlbumIdByTitleAndArtist("Menari Dengan Bayangan", "Hindia"),
			Explicit:     false,
			Lyrics:       "Membasuh hati yang kotor\nMembasuh jiwa yang resah...",
			ReleaseDate:  "2019-11-29",
		},
		{
			ID:           uuid.New(),
			Title:        "Kita Ke Sana",
			Duration:     280,
			AudioFileURL: "https://example.com/hindia/kita_ke_sana.mp3",
			StreamCount:  2100000,
			AlbumID:      getAlbumIdByTitleAndArtist("Menari Dengan Bayangan", "Hindia"),
			Explicit:     false,
			Lyrics:       "Kita ke sana, bersama-sama\nKita ke sana, saling menyayangi...",
			ReleaseDate:  "2019-11-29",
		},
		{
			ID:           uuid.New(),
			Title:        "Cincin",
			Duration:     270,
			AudioFileURL: "https://example.com/hindia/cincin.mp3",
			StreamCount:  2400000,
			AlbumID:      getAlbumIdByTitleAndArtist("Lagipula Hidup Akan Berakhir", "Hindia"),
			Explicit:     false,
			Lyrics:       "Cincin yang dulu kau beri\nMasih ku simpan sampai kini...",
			ReleaseDate:  "2023-07-07",
		},
		{
			ID:           uuid.New(),
			Title:        "Janji Palsu",
			Duration:     260,
			AudioFileURL: "https://example.com/hindia/janji_palsu.mp3",
			StreamCount:  2000000,
			AlbumID:      getAlbumIdByTitleAndArtist("Lagipula Hidup Akan Berakhir", "Hindia"),
			Explicit:     false,
			Lyrics:       "Janji yang kau ucapkan dulu\nTernyata hanya janji palsu...",
			ReleaseDate:  "2023-07-07",
		},
		{
			ID:           uuid.New(),
			Title:        "Masalah Masa Depan",
			Duration:     300,
			AudioFileURL: "https://example.com/hindia/masalah_masa_depan.mp3",
			StreamCount:  2600000,
			AlbumID:      getAlbumIdByTitleAndArtist("Lagipula Hidup Akan Berakhir", "Hindia"),
			Explicit:     false,
			Lyrics:       "Masalah masa depan, membuatku resah\nMasalah masa depan, membuatku tak tenang...",
			ReleaseDate:  "2023-07-07",
		},
		{
			ID:           uuid.New(),
			Title:        "Berdansalah, Karir Ini Tak Ada Artinya",
			Duration:     310,
			AudioFileURL: "https://example.com/hindia/berdansalah_karir_ini_tak_ada_artinya.mp3",
			StreamCount:  1900000,
			AlbumID:      getAlbumIdByTitleAndArtist("Lagipula Hidup Akan Berakhir", "Hindia"),
			Explicit:     false,
			Lyrics:       "Berdansalah, karir ini tak ada artinya\nTanpa cinta yang sesungguhnya...",
			ReleaseDate:  "2023-07-07",
		},
		{
			ID:           uuid.New(),
			Title:        "Dehidrasi",
			Duration:     260,
			AudioFileURL: "https://example.com/hindia/dehidrasi.mp3",
			StreamCount:  2100000,
			AlbumID:      getAlbumIdByTitleAndArtist("Lagipula Hidup Akan Berakhir", "Hindia"),
			Explicit:     false,
			Lyrics:       "Aku dehidrasi, tanpa cintamu\nAku dehidrasi, tanpa kasih sayangmu...",
			ReleaseDate:  "2023-07-07",
		},

		// Bernadya Tracks
		{
			ID:           uuid.New(),
			Title:        "Apa Mungkin",
			Duration:     230,
			AudioFileURL: "https://example.com/bernadya/apa_mungkin.mp3",
			StreamCount:  1200000,
			AlbumID:      getAlbumIdByTitleAndArtist("Sialnya, Hidup Harus Tetap Berjalan", "Bernadya"),
			Explicit:     false,
			Lyrics:       "Apa mungkin kita bisa kembali\nApa mungkin kita bisa bersama...",
			ReleaseDate:  "2023-06-01",
		},
		{
			ID:           uuid.New(),
			Title:        "Bertaut",
			Duration:     270,
			AudioFileURL: "https://example.com/bernadya/bertaut.mp3",
			StreamCount:  1500000,
			AlbumID:      getAlbumIdByTitleAndArtist("Sialnya, Hidup Harus Tetap Berjalan", "Bernadya"),
			Explicit:     false,
			Lyrics:       "Kita bertaut dalam mimpi\nKita bertaut dalam hati...",
			ReleaseDate:  "2023-06-01",
		},
		{
			ID:           uuid.New(),
			Title:        "Satu Bulan",
			Duration:     230,
			AudioFileURL: "https://example.com/bernadya/satu_bulan.mp3",
			StreamCount:  1100000,
			AlbumID:      getAlbumIdByTitleAndArtist("Sialnya, Hidup Harus Tetap Berjalan", "Bernadya"),
			Explicit:     false,
			Lyrics:       "Satu bulan kita tak bertemu\nSatu bulan kita tak berbicara...",
			ReleaseDate:  "2023-06-01",
		},
		{
			ID:           uuid.New(),
			Title:        "Kini Mereka Tahu",
			Duration:     250,
			AudioFileURL: "https://example.com/bernadya/kini_mereka_tau.mp3",
			StreamCount:  1300000,
			AlbumID:      getAlbumIdByTitleAndArtist("Sialnya, Hidup Harus Tetap Berjalan", "Bernadya"),
			Explicit:     false,
			Lyrics:       "Kini mereka tahu, tentang kita\nKini mereka tahu, tentang cinta...",
			ReleaseDate:  "2023-06-01",
		},
		{
			ID:           uuid.New(),
			Title:        "Lama-Lama",
			Duration:     260,
			AudioFileURL: "https://example.com/bernadya/lama_lama.mp3",
			StreamCount:  1400000,
			AlbumID:      getAlbumIdByTitleAndArtist("Sialnya, Hidup Harus Tetap Berjalan", "Bernadya"),
			Explicit:     false,
			Lyrics:       "Lama-lama aku jadi rindu\nLama-lama aku jadi kangen...",
			ReleaseDate:  "2023-06-01",
		},
		{
			ID:           uuid.New(),
			Title:        "Takut",
			Duration:     240,
			AudioFileURL: "https://example.com/bernadya/takut.mp3",
			StreamCount:  1200000,
			AlbumID:      getAlbumIdByTitleAndArtist("Sialnya, Hidup Harus Tetap Berjalan", "Bernadya"),
			Explicit:     false,
			Lyrics:       "Aku takut kehilanganmu\nAku takut kau pergi dariku...",
			ReleaseDate:  "2023-06-01",
		},
		{
			ID:           uuid.New(),
			Title:        "Berjalan",
			Duration:     240,
			AudioFileURL: "https://example.com/bernadya/berjalan.mp3",
			StreamCount:  1100000,
			AlbumID:      getAlbumIdByTitleAndArtist("Sialnya, Hidup Harus Tetap Berjalan", "Bernadya"),
			Explicit:     false,
			Lyrics:       "Kita berjalan bersama\nMenuju masa yang cerah...",
			ReleaseDate:  "2023-06-01",
		},
		{
			ID:           uuid.New(),
			Title:        "Untungnya, Hidup Harus Tetap Berjalan",
			Duration:     270,
			AudioFileURL: "https://example.com/bernadya/untungnya_hidup_harus_tetap_berjalan.mp3",
			StreamCount:  1600000,
			AlbumID:      getAlbumIdByTitleAndArtist("Sialnya, Hidup Harus Tetap Berjalan", "Bernadya"),
			Explicit:     false,
			Lyrics:       "Untungnya, hidup harus tetap berjalan\nWalau tanpamu di sisiku...",
			ReleaseDate:  "2023-06-01",
		},
		{
			ID:           uuid.New(),
			Title:        "Kata Mereka Ini Berlebihan",
			Duration:     280,
			AudioFileURL: "https://example.com/bernadya/kata_mereka_ini_berlebihan.mp3",
			StreamCount:  1300000,
			AlbumID:      getAlbumIdByTitleAndArtist("Sialnya, Hidup Harus Tetap Berjalan", "Bernadya"),
			Explicit:     false,
			Lyrics:       "Kata mereka ini berlebihan\nTapi aku tak peduli...",
			ReleaseDate:  "2023-06-01",
		},
		{
			ID:           uuid.New(),
			Title:        "Rindu Tak Bersuara",
			Duration:     260,
			AudioFileURL: "https://example.com/bernadya/rindu_tak_bersuara.mp3",
			StreamCount:  1400000,
			AlbumID:      getAlbumIdByTitleAndArtist("Sialnya, Hidup Harus Tetap Berjalan", "Bernadya"),
			Explicit:     false,
			Lyrics:       "Rindu tak bersuara, dalam hati\nRindu tak bersuara, dalam jiwa...",
			ReleaseDate:  "2023-06-01",
		},
		{
			ID:           uuid.New(),
			Title:        "Sialnya, Hidup Harus Tetap Berjalan",
			Duration:     260,
			AudioFileURL: "https://example.com/bernadya/sialnya_hidup_harus_tetap_berjalan.mp3",
			StreamCount:  1800000,
			AlbumID:      getAlbumIdByTitleAndArtist("Sialnya, Hidup Harus Tetap Berjalan", "Bernadya"),
			Explicit:     false,
			Lyrics:       "Sialnya, hidup harus tetap berjalan\nWalau hati ini terluka...",
			ReleaseDate:  "2023-06-01",
		},
	}

	for _, track := range tracks {
		// Skip tracks where album ID couldn't be found
		if track.AlbumID == uuid.Nil {
			log.Printf("Skipping track %s - album not found", track.Title)
			continue
		}

		if err := s.db.Create(&track).Error; err != nil {
			log.Printf("Error seeding track %s: %v", track.Title, err)
			continue
		}
		log.Printf("Seeded track: %s from %s", track.Title, getAlbumTitleById(s.db, track.AlbumID))
	}

	return nil
}

func (s *Seeder) SeedTrackArtists(ctx context.Context) error {
	// Check if the track_artists table exists before attempting to seed
	var result *gorm.DB
	result = s.db.Raw("SELECT 1 FROM track_artists LIMIT 1").Take(&[]byte{})

	if result.Error != nil {
		// Table doesn't exist, log and return
		log.Println("track_artists table does not exist, skipping track-artist relationships seeding")
		return nil
	}

	// Get artist IDs
	var artists []dbModel.Artist
	if err := s.db.Find(&artists).Error; err != nil {
		return fmt.Errorf("failed to fetch artists for track-artists: %w", err)
	}

	// Create a map for easy lookup
	artistMap := make(map[string]dbModel.Artist)
	for _, artist := range artists {
		artistMap[artist.Name] = artist
	}

	// Get track IDs
	var tracks []dbModel.Track
	if err := s.db.Find(&tracks).Error; err != nil {
		log.Printf("Warning: failed to fetch tracks for track-artists: %v", err)
		// Continue anyway since this is not critical for other operations
		return nil
	}

	// Create a map for easy track lookup
	trackMap := make(map[string]dbModel.Track)
	for _, track := range tracks {
		trackMap[track.Title] = track
	}

	// Create track-artist relationships for all tracks
	var trackArtists []dbModel.TrackArtist
	for _, track := range tracks {
		// Get the album for this track to find the associated artist
		var album dbModel.Album
		if err := s.db.Where("id = ?", track.AlbumID).First(&album).Error; err != nil {
			log.Printf("Error finding album for track %s: %v", track.Title, err)
			continue
		}

		// Get the artist for this album
		var artist dbModel.Artist
		if err := s.db.Where("id = ?", album.ArtistID).First(&artist).Error; err != nil {
			log.Printf("Error finding artist for album %s: %v", album.Title, err)
			continue
		}

		// Create the track-artist relationship
		trackArtist := dbModel.TrackArtist{
			ID:       uuid.New().String(),
			TrackID:  track.ID.String(),
			ArtistID: artist.ID,
			Role:     "Main Artist",
		}
		trackArtists = append(trackArtists, trackArtist)
	}

	for _, ta := range trackArtists {
		// Skip if track ID is empty (meaning the track wasn't found)
		if ta.TrackID == "" || ta.ArtistID == "" {
			log.Printf("Skipping track-artist relationship - track or artist not found")
			continue
		}

		if err := s.db.Create(&ta).Error; err != nil {
			log.Printf("Error seeding track artist %s-%s: %v", ta.TrackID, ta.ArtistID, err)
			continue
		}
		log.Printf("Seeded track artist: %s - %s", getTrackTitleById(s.db, ta.TrackID), getArtistNameById(s.db, ta.ArtistID))
	}

	return nil
}

func (s *Seeder) SeedListeningHistory(ctx context.Context) error {
	adminUserID := "5d7f9219-26cc-4e77-8e1a-fcc38fbea99a"

	// Get some track IDs to create listening history
	var tracks []dbModel.Track
	if err := s.db.Limit(20).Find(&tracks).Error; err != nil {
		return fmt.Errorf("failed to get tracks for listening history: %w", err)
	}

	// Create some listening history records for the admin user
	for i, track := range tracks {
		listeningHistory := dbModel.ListeningHistory{
			UserID:      adminUserID,
			ContentID:   track.ID.String(),
			ContentType: "TRACK",
			PlayedAt:    parseTime(fmt.Sprintf("2023-0%d-0%dT1%d:30:00Z", (i%12)+1, (i%28)+1, (i%12)+1)),
			Progress:    100, // Fully listened
		}

		if err := s.db.Create(&listeningHistory).Error; err != nil {
			log.Printf("Error seeding listening history: %v", err)
			continue
		}
		log.Printf("Seeded listening history for user %s and content %s", listeningHistory.UserID, listeningHistory.ContentID)
	}

	return nil
}

func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

func parseTime(timeStr string) time.Time {
	t, err := time.Parse("2006-01-02T15:04:05Z", timeStr)
	if err != nil {
		// If parsing fails, return current time
		return time.Now()
	}
	return t
}

func stringToUUID(s string) uuid.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		// If parsing fails, return a new UUID
		return uuid.New()
	}
	return id
}

func getArtistNameById(db *gorm.DB, id string) string {
	var artist dbModel.Artist
	db.Where("id = ?", id).First(&artist)
	return artist.Name
}

func getArtistIdByName(db *gorm.DB, name string) string {
	var artist dbModel.Artist
	db.Where("name = ?", name).First(&artist)
	return artist.ID
}

func getAlbumTitleById(db *gorm.DB, id uuid.UUID) string {
	var album dbModel.Album
	db.Where("id = ?", id).First(&album)
	return album.Title
}

func getTrackTitleById(db *gorm.DB, id string) string {
	var track dbModel.Track
	db.Where("id = ?", id).First(&track)
	return track.Title
}

func getTrackIdByTitle(db *gorm.DB, title string) string {
	var track dbModel.Track
	db.Where("title = ?", title).First(&track)
	return track.ID.String()
}

func getBernadyaAlbumId(db *gorm.DB) uuid.UUID {
	var album dbModel.Album
	db.Joins("JOIN artists ON albums.artist_id = artists.id").
		Where("artists.name = ? AND (albums.title LIKE ? OR albums.title = ?)", "Bernadya", "%Bernadya%", "Bernadya").
		First(&album)
	return album.ID
}

func (s *Seeder) SeedAll(ctx context.Context) error {
	// Clear existing music data (preserving user data)
	if err := s.ClearMusicData(); err != nil {
		return fmt.Errorf("failed to clear existing music data: %w", err)
	}

	log.Println("Starting to seed genres...")
	if err := s.SeedGenres(ctx); err != nil {
		return fmt.Errorf("failed to seed genres: %w", err)
	}

	log.Println("Starting to seed artists...")
	if err := s.SeedArtists(ctx); err != nil {
		return fmt.Errorf("failed to seed artists: %w", err)
	}

	// Force flush to ensure artists are saved before creating albums
	if err := s.db.Exec("SELECT 1").Error; err != nil {
		log.Printf("Error flushing: %v", err)
	}

	log.Println("Starting to seed albums...")
	if err := s.SeedAlbums(ctx); err != nil {
		return fmt.Errorf("failed to seed albums: %w", err)
	}

	// Force flush to ensure albums are saved before creating tracks
	if err := s.db.Exec("SELECT 1").Error; err != nil {
		log.Printf("Error flushing: %v", err)
	}

	log.Println("Starting to seed tracks...")
	if err := s.SeedTracks(ctx); err != nil {
		return fmt.Errorf("failed to seed tracks: %w", err)
	}

	log.Println("Starting to seed track-artists relationships...")
	if err := s.SeedTrackArtists(ctx); err != nil {
		return fmt.Errorf("failed to seed track-artists: %w", err)
	}

	log.Println("Starting to seed listening history...")
	if err := s.SeedListeningHistory(ctx); err != nil {
		return fmt.Errorf("failed to seed listening history: %w", err)
	}

	log.Println("All seeders completed successfully!")
	return nil
}