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

// handleFileList handles the request to list files in a directory
func (h *Handlers) handleFileList(c *gin.Context) {
	var request FileListRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		carrot.AbortWithJSONError(c, http.StatusBadRequest, err)
		return
	}

	// Get settings to obtain the download path
	settings, err := models.GetSettings(h.db, 1) // Using default user ID 1
	if err != nil {
		carrot.AbortWithJSONError(c, http.StatusInternalServerError, err)
		return
	}

	// Determine the directory to list files from
	basePath := settings.DownloadPath
	fullPath := basePath

	// If a subpath is specified, append it to the base path
	request.Path = strings.TrimPrefix(request.Path, "/")
	if request.Path != "" && request.Path != "/" {
		fullPath = filepath.Join(basePath, request.Path)
	}

	// Check if the directory exists
	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			carrot.AbortWithJSONError(c, http.StatusNotFound, fmt.Errorf("directory not found: %s", fullPath))
			return
		}
		carrot.AbortWithJSONError(c, http.StatusInternalServerError, err)
		return
	}

	// If it's not a directory, return an error
	if !fileInfo.IsDir() {
		carrot.AbortWithJSONError(c, http.StatusBadRequest, fmt.Errorf("not a directory: %s", fullPath))
		return
	}

	// List files in the directory
	files, err := os.ReadDir(fullPath)
	if err != nil {
		carrot.AbortWithJSONError(c, http.StatusInternalServerError, err)
		return
	}

	// Convert file info to our format and apply filters
	var fileList []FileInfo
	for _, file := range files {
		// Skip hidden files that start with a dot
		if strings.HasPrefix(file.Name(), ".") {
			continue
		}

		// Get file info
		info, err := file.Info()
		if err != nil {
			continue
		}

		fileExt := filepath.Ext(file.Name())
		fileType := strings.TrimPrefix(strings.ToLower(fileExt), ".")

		// Check if file matches the requested type filter
		if !file.IsDir() && request.Type != "" && !matchesFileType(fileType, request.Type) {
			continue
		}

		// Create relative path for frontend navigation
		relativePath := request.Path
		if relativePath == "" || relativePath == "/" {
			relativePath = file.Name()
		} else {
			relativePath = filepath.Join(request.Path, file.Name())
		}

		// Create file info object
		fileInfo := FileInfo{
			FileName:    file.Name(),
			FilePath:    relativePath,
			Directory:   file.IsDir(),
			GmtModified: info.ModTime(),
			FileType:    fileType,
		}

		// Set file size (only for files, not directories)
		if !file.IsDir() {
			fileInfo.Size = info.Size()
			fileInfo.FileSize = formatFileSize(info.Size())
		}

		fileList = append(fileList, fileInfo)
	}

	// Sort the file list
	sortFileList(fileList, request.Sort, request.Order)

	c.JSON(http.StatusOK, gin.H{
		"data": fileList,
	})
}
