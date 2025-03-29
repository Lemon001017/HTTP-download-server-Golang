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

// handleFileRename handles the request to rename a file or directory
func (h *Handlers) handleFileRename(c *gin.Context) {
	var request FileRenameRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		carrot.AbortWithJSONError(c, http.StatusBadRequest, err)
		return
	}

	// Validate inputs
	if request.Path == "" {
		carrot.AbortWithJSONError(c, http.StatusBadRequest, fmt.Errorf("missing path parameter"))
		return
	}

	if request.NewName == "" {
		carrot.AbortWithJSONError(c, http.StatusBadRequest, fmt.Errorf("missing new name parameter"))
		return
	}

	// Get settings to obtain the download path
	settings, err := models.GetSettings(h.db, 1) // Using default user ID 1
	if err != nil {
		carrot.AbortWithJSONError(c, http.StatusInternalServerError, err)
		return
	}

	// Ensure paths are relative and safe
	request.Path = strings.TrimPrefix(request.Path, "/")

	// Ensure the new name doesn't contain path separators or other invalid characters
	if strings.Contains(request.NewName, "/") || strings.Contains(request.NewName, "\\") {
		carrot.AbortWithJSONError(c, http.StatusBadRequest, fmt.Errorf("new name contains invalid characters"))
		return
	}

	// Build the full paths
	fullPath := filepath.Join(settings.DownloadPath, request.Path)

	// Get the directory of the current file or directory
	dir := filepath.Dir(fullPath)

	// Create the new path by joining the directory with the new name
	newPath := filepath.Join(dir, request.NewName)

	// Check if the file exists
	_, err = os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			carrot.AbortWithJSONError(c, http.StatusNotFound, fmt.Errorf("file not found: %s", request.Path))
			return
		}
		carrot.AbortWithJSONError(c, http.StatusInternalServerError, err)
		return
	}

	// Check if the destination already exists
	_, err = os.Stat(newPath)
	if err == nil {
		carrot.AbortWithJSONError(c, http.StatusConflict, fmt.Errorf("destination already exists: %s", request.NewName))
		return
	}

	// Rename the file
	err = os.Rename(fullPath, newPath)
	if err != nil {
		carrot.AbortWithJSONError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "File renamed successfully",
	})
}

// handleFileDelete handles the request to delete a file or directory
func (h *Handlers) handleFileDelete(c *gin.Context) {
	var request FileDeleteRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		carrot.AbortWithJSONError(c, http.StatusBadRequest, err)
		return
	}

	// Validate inputs
	if request.Path == "" {
		carrot.AbortWithJSONError(c, http.StatusBadRequest, fmt.Errorf("missing path parameter"))
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

	// Build the full path
	fullPath := filepath.Join(settings.DownloadPath, request.Path)

	// Check if the file exists
	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			carrot.AbortWithJSONError(c, http.StatusNotFound, fmt.Errorf("file not found: %s", request.Path))
			return
		}
		carrot.AbortWithJSONError(c, http.StatusInternalServerError, err)
		return
	}

	// If it's a directory, ensure it's empty
	if fileInfo.IsDir() {
		entries, err := os.ReadDir(fullPath)
		if err != nil {
			carrot.AbortWithJSONError(c, http.StatusInternalServerError, err)
			return
		}

		if len(entries) > 0 {
			carrot.AbortWithJSONError(c, http.StatusBadRequest, fmt.Errorf("directory is not empty: %s", request.Path))
			return
		}

		// Remove the empty directory
		err = os.Remove(fullPath)
		if err != nil {
			carrot.AbortWithJSONError(c, http.StatusInternalServerError, err)
			return
		}
	} else {
		// Remove the file
		err = os.Remove(fullPath)
		if err != nil {
			carrot.AbortWithJSONError(c, http.StatusInternalServerError, err)
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "File deleted successfully",
	})
}
