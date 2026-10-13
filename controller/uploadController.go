package controller

import (
    "fmt"
    "net/http"
    "os"
    "path/filepath"
    "strings"
    "time"

    "github.com/gin-gonic/gin"
    "go.uber.org/fx"
)

// UploadControllerTarget contains constructor dependencies for UploadController.
type UploadControllerTarget struct {
    fx.In
    StoragePath string `name:"storage_path"`
}

// UploadController handles upload endpoints for the application.
type UploadController struct {
    storagePath string
}

// NewUploadController creates a new UploadController.
func NewUploadController(target UploadControllerTarget) Result {
    return Result{
        Controller: &UploadController{
            storagePath: target.StoragePath,
        },
    }
}

func (u *UploadController) getBaseStoragePath() string {
    if u.storagePath != "" {
        return u.storagePath
    }
    if p := os.Getenv("STORAGE_PATH"); p != "" {
        return p
    }
    if p := os.Getenv("BASE_STORAGE_PATH"); p != "" {
        return p
    }
    return filepath.Join(".", "Storage")
}

func (u *UploadController) Register(r *gin.Engine) {
    r.OPTIONS("/v1/upload/playlist-cover", CORSMiddleware())
    r.POST("/v1/upload/playlist-cover", CORSMiddleware(), u.UploadPlaylistCover)
    r.DELETE("/v1/upload/playlist-cover", CORSMiddleware(), u.DeletePlaylistCover)
    // Serve static storage with StaticMediaCORSMiddleware so images are accessible from allowed origins
    storageGroup := r.Group("/storage")
    storageGroup.Use(StaticMediaCORSMiddleware())
    // Ensure OPTIONS preflight is handled for any storage path
    storageGroup.OPTIONS("/*filepath", StaticMediaCORSMiddleware())
    storageGroup.Static("/", u.getBaseStoragePath())
}

func (u *UploadController) UploadPlaylistCover(c *gin.Context) {
    playlistID := c.PostForm("playlistId")
    if playlistID == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "playlistId is required"})
        return
    }

    file, err := c.FormFile("file")
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
        return
    }

    ext := strings.ToLower(filepath.Ext(file.Filename))
    if ext == "" || !(ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".webp") {
        ext = ".jpg"
    }

    basePath := u.getBaseStoragePath()
    destDir := filepath.Join(basePath, "Mobile", "Playlist", playlistID)
    if err := os.MkdirAll(destDir, 0755); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create target directory"})
        return
    }

    destFileName := "playlist" + ext
    destFilePath := filepath.Join(destDir, destFileName)

    if err := c.SaveUploadedFile(file, destFilePath); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save uploaded file"})
        return
    }

    // Add cache-busting parameter with current timestamp
    timestamp := time.Now().Unix()
    url := fmt.Sprintf("https://cirio.my.id/storage/Mobile/Playlist/%s/%s?v=%d", playlistID, destFileName, timestamp)

    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "url":     url,
        "path":    destFilePath,
    })
}

func (u *UploadController) DeletePlaylistCover(c *gin.Context) {
    playlistID := c.Query("playlistId")
    if playlistID == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "playlistId is required"})
        return
    }

    basePath := u.getBaseStoragePath()
    dirPath := filepath.Join(basePath, "Mobile", "Playlist", playlistID)

    possibleFiles := []string{"playlist.jpg", "playlist.jpeg", "playlist.png", "playlist.webp"}
    deleted := false

    for _, fileName := range possibleFiles {
        filePath := filepath.Join(dirPath, fileName)
        if _, err := os.Stat(filePath); err == nil {
            if err := os.Remove(filePath); err == nil {
                deleted = true
                break
            }
        }
    }

    if !deleted {
        c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"success": true})
}
