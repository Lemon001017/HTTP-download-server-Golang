package handlers

import (
	"HTTP-download-server/server/models"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/restsend/carrot"
)

// handleFileMkdir handles the request to create a new directory
func (h *Handlers) handleFileMkdir(c *gin.Context) {
	var request FileMkdirRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		carrot.AbortWithJSONError(c, http.StatusBadRequest, err)
		return
	}

	// Validate inputs
	if request.DirName == "" {
		carrot.AbortWithJSONError(c, http.StatusBadRequest, fmt.Errorf("directory name is required"))
		return
	}

	// Ensure directory name doesn't contain path separators or other invalid characters
	if strings.Contains(request.DirName, "/") || strings.Contains(request.DirName, "\\") {
		carrot.AbortWithJSONError(c, http.StatusBadRequest, fmt.Errorf("directory name contains invalid characters"))
		return
	}

	// Get settings to obtain the download path
	settings, err := models.GetSettings(h.db, 1) // Using default user ID 1
	if err != nil {
		carrot.AbortWithJSONError(c, http.StatusInternalServerError, err)
		return
	}

	// Ensure path is relative and safe
	request.Path = strings.TrimPrefix(request.Path, "/")

	// Build the full path for the parent directory
	parentPath := settings.DownloadPath
	if request.Path != "" {
		parentPath = filepath.Join(settings.DownloadPath, request.Path)
	}

	// Check if the parent directory exists
	parentInfo, err := os.Stat(parentPath)
	if err != nil {
		if os.IsNotExist(err) {
			carrot.AbortWithJSONError(c, http.StatusNotFound, fmt.Errorf("parent directory not found: %s", parentPath))
			return
		}
		carrot.AbortWithJSONError(c, http.StatusInternalServerError, err)
		return
	}

	// Ensure the parent path is a directory
	if !parentInfo.IsDir() {
		carrot.AbortWithJSONError(c, http.StatusBadRequest, fmt.Errorf("parent path is not a directory: %s", parentPath))
		return
	}

	// Create the full path for the new directory
	newDirPath := filepath.Join(parentPath, request.DirName)

	// Check if the directory already exists
	_, err = os.Stat(newDirPath)
	if err == nil {
		carrot.AbortWithJSONError(c, http.StatusConflict, fmt.Errorf("directory already exists: %s", request.DirName))
		return
	}

	// Create the directory
	err = os.Mkdir(newDirPath, 0755)
	if err != nil {
		carrot.AbortWithJSONError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Directory created successfully",
	})
}
