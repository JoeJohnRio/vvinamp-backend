package controller

import (
	"path/filepath"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

// FileServerControllerTarget is parameter object for getting all FileServerController's dependency
type FileServerControllerTarget struct {
	fx.In
	StoragePath string `name:"storage_path"`
}

// FileServerController handles serving static files from storage directory
type FileServerController struct {
	storagePath string
}

// NewFileServerController is a constructor for FileServerController
func NewFileServerController(target FileServerControllerTarget) Result {
	storagePath := target.StoragePath
	if storagePath == "" {
		// Default path if not configured
		storagePath = "S:\\Project German\\Backend\\Storage"
	}

	return Result{
		Controller: &FileServerController{
			storagePath: storagePath,
		},
	}
}

// Register is function to register all controller's endpoint handler
func (f *FileServerController) Register(r *gin.Engine) {
	// Serve static files from the storage directory
	// This will serve files at routes like /storage/images/, /storage/tracks/, etc.
	r.Static("/storage", f.storagePath)

	// You can also create specific routes for different file types
	// These will serve from specific subdirectories in storage
	r.Static("/images", filepath.Join(f.storagePath, "images"))
	r.Static("/tracks", filepath.Join(f.storagePath, "tracks"))
	r.Static("/files", filepath.Join(f.storagePath, "files"))
}

