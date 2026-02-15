package controller

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"vvinamp/database/postgre"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/fx"
)

// FileServerControllerTarget is parameter object for getting all FileServerController's dependency
type FileServerControllerTarget struct {
	fx.In
	StoragePath string `name:"storage_path"`
	DB          *postgre.Connection
}

// FileServerController handles serving static files from storage directory
type FileServerController struct {
	storagePath string
	db          *postgre.Connection
}

// NewFileServerController is a constructor for FileServerController
func NewFileServerController(target FileServerControllerTarget) Result {
	storagePath := target.StoragePath
	if storagePath == "" {
		// Default path if not configured
		storagePath = "S:\\Project German\\Backend\\Storage\\hls"
	}

	return Result{
		Controller: &FileServerController{
			storagePath: storagePath,
			db:          target.DB,
		},
	}
}

// sanitize applies the same rules as the JavaScript version:
// 1. Replace any char NOT in [a-zA-Z0-9 space - _] with "_"
// 2. Convert to lowercase
// 3. Replace multiple spaces/underscores with a single "_"
func sanitize(input string) string {
	// Step 1: Replace any character not in [a-zA-Z0-9 space - _] with "_"
	reg := regexp.MustCompile(`[^a-zA-Z0-9 -_]`)
	sanitized := reg.ReplaceAllString(input, "_")

	// Step 2: Convert to lowercase
	sanitized = strings.ToLower(sanitized)

	// Step 3: Replace multiple spaces/underscores with a single "_"
	multipleSpacesReg := regexp.MustCompile(`[ _]+`)
	sanitized = multipleSpacesReg.ReplaceAllString(sanitized, "_")

	return sanitized
}

// Artist represents the artist record from the database
type Artist struct {
	ID   uuid.UUID `gorm:"column:id;primaryKey" json:"id"`
	Name string    `gorm:"column:name" json:"name"`
}

// Album represents the album record from the database
type Album struct {
	ID       uuid.UUID `gorm:"column:id;primaryKey" json:"id"`
	Title    string    `gorm:"column:title" json:"title"`
	ArtistID uuid.UUID `gorm:"column:artist_id" json:"artist_id"`
}

// Track represents the track record from the database
type Track struct {
	ID      uuid.UUID `gorm:"column:id;primaryKey" json:"id"`
	Title   string    `gorm:"column:title" json:"title"`
	AlbumID uuid.UUID `gorm:"column:album_id" json:"album_id"`
}

// getArtistNameByID retrieves the artist name by ID from the database
func (f *FileServerController) getArtistNameByID(ctx context.Context, id uuid.UUID) (string, error) {
	var artist Artist
	dbConn, err := f.db.Client()
	if err != nil {
		return "", err
	}

	result := dbConn.WithContext(ctx).Where("id = ?", id).First(&artist)
	if result.Error != nil {
		return "", result.Error
	}
	return artist.Name, nil
}

// getAlbumAndArtistNamesByID retrieves the album title and artist name by album ID from the database
func (f *FileServerController) getAlbumAndArtistNamesByID(ctx context.Context, id uuid.UUID) (string, string, error) {
	var album Album
	var artist Artist

	dbConn, err := f.db.Client()
	if err != nil {
		return "", "", err
	}

	// First get the album
	result := dbConn.WithContext(ctx).Where("id = ?", id).First(&album)
	if result.Error != nil {
		return "", "", result.Error
	}

	// Then get the associated artist
	result = dbConn.WithContext(ctx).Where("id = ?", album.ArtistID).First(&artist)
	if result.Error != nil {
		return "", "", result.Error
	}

	return album.Title, artist.Name, nil
}

// getTrackAlbumAndArtistNamesByID retrieves the track title, album title, and artist name by track ID from the database
func (f *FileServerController) getTrackAlbumAndArtistNamesByID(ctx context.Context, id uuid.UUID) (string, string, string, error) {
	var track Track
	var album Album
	var artist Artist

	dbConn, err := f.db.Client()
	if err != nil {
		return "", "", "", err
	}

	// Get the track
	result := dbConn.WithContext(ctx).Where("id = ?", id).First(&track)
	if result.Error != nil {
		return "", "", "", result.Error
	}

	// Get the album
	result = dbConn.WithContext(ctx).Where("id = ?", track.AlbumID).First(&album)
	if result.Error != nil {
		return "", "", "", result.Error
	}

	// Get the artist
	result = dbConn.WithContext(ctx).Where("id = ?", album.ArtistID).First(&artist)
	if result.Error != nil {
		return "", "", "", result.Error
	}

	return track.Title, album.Title, artist.Name, nil
}

// Register is function to register all controller's endpoint handler
func (f *FileServerController) Register(r *gin.Engine) {
	// Serve static files from the storage directory
	// This will serve files at routes like /storage/images/, /storage/tracks/, etc.
	r.Static("/hls", f.storagePath)

	// Media API endpoints
	r.GET("/artist/:artistId/:fileName", f.ArtistMediaHandler)
	r.GET("/album/:albumId/:fileName", f.AlbumMediaHandler)
	r.GET("/track/:trackId/:fileName", f.TrackMediaHandler)

	// You can also create specific routes for different file types
	// These will serve from specific subdirectories in storage

	// r.Static("/hls", filepath.Join(f.storagePath, "hls"))
	// r.Static("/images", filepath.Join(f.storagePath, "images"))
	// r.Static("/tracks", filepath.Join(f.storagePath, "tracks"))
	// r.Static("/files", filepath.Join(f.storagePath, "files"))
	// r.Static("/artist", filepath.Join(f.storagePath, "artist"))
	// r.Static("/vvinamp", filepath.Join(f.storagePath, "/vvinamp/artist/michael_jackson")) //this works
	// r.Static("/vvinamp", filepath.Join(f.storagePath, "/vvinamp/artist"))//this doesnt works
}

// ArtistMediaHandler handles requests for artist media files
func (f *FileServerController) ArtistMediaHandler(c *gin.Context) {
	artistIDStr := c.Param("artistId")
	fileName := c.Param("fileName")

	// Validate and parse the artist ID
	artistID, err := uuid.Parse(artistIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid artist ID"})
		return
	}

	// Clean the filename to prevent directory traversal
	cleanFileName := filepath.Clean(fileName)

	// Check if the cleaned filename still contains the original filename
	// This prevents cases where someone tries to access parent directories
	if cleanFileName != fileName || strings.Contains(cleanFileName, "..") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Invalid file path"})
		return
	}

	// Get artist name from database
	ctx := c.Request.Context()
	artistName, err := f.getArtistNameByID(ctx, artistID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Artist not found"})
		return
	}

	// Sanitize the artist name
	sanitizedArtistName := sanitize(artistName)

	// Construct the file path
	baseStoragePath := os.Getenv("BASE_STORAGE_PATH")
	if baseStoragePath == "" {
		baseStoragePath = "S:\\Project German\\Backend\\Storage" // Default path
	}

	filePath := filepath.Join(baseStoragePath, "artist", sanitizedArtistName, cleanFileName)

	// Log the constructed path for debugging
	fmt.Printf("Constructed path for artist media: %s\n", filePath)
	fmt.Printf("Sanitized artist name: %s\n", sanitizedArtistName)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	// Serve the file
	c.File(filePath)
}

// AlbumMediaHandler handles requests for album media files
func (f *FileServerController) AlbumMediaHandler(c *gin.Context) {
	albumIDStr := c.Param("albumId")
	fileName := c.Param("fileName")

	// Validate and parse the album ID
	albumID, err := uuid.Parse(albumIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid album ID"})
		return
	}

	// Clean the filename to prevent directory traversal
	cleanFileName := filepath.Clean(fileName)

	// Check if the cleaned filename still contains the original filename
	if cleanFileName != fileName || strings.Contains(cleanFileName, "..") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Invalid file path"})
		return
	}

	// Get album and artist names from database
	ctx := c.Request.Context()
	albumTitle, artistName, err := f.getAlbumAndArtistNamesByID(ctx, albumID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Album not found"})
		return
	}

	// Sanitize the names
	sanitizedArtistName := sanitize(artistName)
	sanitizedAlbumTitle := sanitize(albumTitle)

	// Construct the file path
	baseStoragePath := os.Getenv("BASE_STORAGE_PATH")
	if baseStoragePath == "" {
		baseStoragePath = "S:\\Project German\\Backend\\Storage" // Default path
	}

	filePath := filepath.Join(baseStoragePath, "artist", sanitizedArtistName, "album", sanitizedAlbumTitle, cleanFileName)

	// Log the constructed path for debugging
	fmt.Printf("Constructed path for album media: %s\n", filePath)
	fmt.Printf("Sanitized artist name: %s, Sanitized album title: %s\n", sanitizedArtistName, sanitizedAlbumTitle)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	// Serve the file
	c.File(filePath)
}

// TrackMediaHandler handles requests for track media files
func (f *FileServerController) TrackMediaHandler(c *gin.Context) {
	trackIDStr := c.Param("trackId")
	fileName := c.Param("fileName")

	// Validate and parse the track ID
	trackID, err := uuid.Parse(trackIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid track ID"})
		return
	}

	// Clean the filename to prevent directory traversal
	cleanFileName := filepath.Clean(fileName)

	// Check if the cleaned filename still contains the original filename
	if cleanFileName != fileName || strings.Contains(cleanFileName, "..") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Invalid file path"})
		return
	}

	// Get track, album, and artist names from database
	ctx := c.Request.Context()
	_, albumTitle, artistName, err := f.getTrackAlbumAndArtistNamesByID(ctx, trackID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Track not found"})
		return
	}

	// Sanitize the names
	sanitizedArtistName := sanitize(artistName)
	sanitizedAlbumTitle := sanitize(albumTitle)

	// Construct the file path
	baseStoragePath := os.Getenv("BASE_STORAGE_PATH")
	if baseStoragePath == "" {
		baseStoragePath = "S:\\Project German\\Backend\\Storage" // Default path
	}

	filePath := filepath.Join(baseStoragePath, "artist", sanitizedArtistName, "album", sanitizedAlbumTitle, "tracks", cleanFileName)

	// Log the constructed path for debugging
	fmt.Printf("Constructed path for track media: %s\n", filePath)
	fmt.Printf("Sanitized artist name: %s, Sanitized album title: %s\n", sanitizedArtistName, sanitizedAlbumTitle)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	// Serve the file
	c.File(filePath)
}

func (f *FileServerController) ConvertToHLS(ctx context.Context, mp3RelativePath, trackId string) (string, error) {
	mp3Path := filepath.Join(f.storagePath, mp3RelativePath) // e.g. "mp3/track123.mp3"

	hlsDir := filepath.Join(f.storagePath, "hls", trackId)
	if err := os.MkdirAll(hlsDir, 0755); err != nil {
		return "", fmt.Errorf("mkdir %s: %w", hlsDir, err)
	}

	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-i", mp3Path,
		"-c:a", "aac", "-b:a", "128k",
		"-f", "hls",
		"-hls_time", "10",
		"-hls_list_size", "0",
		"-hls_segment_filename", filepath.Join(hlsDir, "seg_%03d.ts"),
		filepath.Join(hlsDir, "playlist.m3u8"))

	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("ffmpeg failed: %v\noutput: %s", err, output)
	}

	// Return public URL (adjust baseURL from config/resolver)
	return fmt.Sprintf("/hls/%s/playlist.m3u8", trackId), nil
}

// Bonus: Check if HLS exists (avoid reconvert)
func (f *FileServerController) HLSExists(trackId string) bool {
	return fileExists(filepath.Join(f.storagePath, "hls", trackId, "playlist.m3u8"))
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
