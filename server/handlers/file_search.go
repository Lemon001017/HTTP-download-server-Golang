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

// handleFileSearch handles searching for files and directories
func (h *Handlers) handleFileSearch(c *gin.Context) {
	var request FileSearchRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		carrot.AbortWithJSONError(c, http.StatusBadRequest, err)
		return
	}

	// Validate query - check if empty or contains only whitespace
	if strings.TrimSpace(request.Query) == "" {
		carrot.AbortWithJSONError(c, http.StatusBadRequest, fmt.Errorf("search query cannot be empty"))
		return
	}

	// Get settings to obtain the download path
	settings, err := models.GetSettings(h.db, 1) // Using default user ID 1
	if err != nil {
		carrot.AbortWithJSONError(c, http.StatusInternalServerError, err)
		return
	}

	// Determine the base search path
	basePath := settings.DownloadPath
	searchPath := basePath

	// If a specific path is provided, use it as the search root
	if request.Path != "" {
		request.Path = strings.TrimPrefix(request.Path, "/")
		searchPath = filepath.Join(basePath, request.Path)

		// Check if the search path exists
		searchPathInfo, err := os.Stat(searchPath)
		if err != nil {
			if os.IsNotExist(err) {
				carrot.AbortWithJSONError(c, http.StatusNotFound, fmt.Errorf("search path not found: %s", searchPath))
				return
			}
			carrot.AbortWithJSONError(c, http.StatusInternalServerError, err)
			return
		}

		// Ensure the search path is a directory
		if !searchPathInfo.IsDir() {
			carrot.AbortWithJSONError(c, http.StatusBadRequest, fmt.Errorf("search path is not a directory: %s", request.Path))
			return
		}
	}

	// Prepare results
	var results []FileInfo

	// Define a walk function to process each file
	walkFunc := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Skip any errors and continue
			return nil
		}

		// Skip hidden files and directories
		if strings.HasPrefix(filepath.Base(path), ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if the file or directory name matches the query
		if isFileMatch(info.Name(), request.Query, request.Exact) {
			// Create file info directly without using relPath
			fileInfo := createFileInfo(path, info, basePath)
			results = append(results, fileInfo)
		}

		return nil
	}

	// Either search recursively or just the top level
	if request.Recursive {
		// Walk the directory tree recursively
		err = filepath.Walk(searchPath, walkFunc)
		if err != nil {
			carrot.AbortWithJSONError(c, http.StatusInternalServerError, err)
			return
		}
	} else {
		// Just search the top level directory
		files, err := os.ReadDir(searchPath)
		if err != nil {
			carrot.AbortWithJSONError(c, http.StatusInternalServerError, err)
			return
		}

		for _, file := range files {
			// Skip hidden files
			if strings.HasPrefix(file.Name(), ".") {
				continue
			}

			info, err := file.Info()
			if err != nil {
				continue
			}

			// Check if the file or directory name matches the query
			if isFileMatch(file.Name(), request.Query, request.Exact) {
				// Get the full path
				fullPath := filepath.Join(searchPath, file.Name())

				// Create file info directly using fullPath
				fileInfo := createFileInfo(fullPath, info, basePath)
				results = append(results, fileInfo)
			}
		}
	}

	// Return empty array instead of null if no results found
	if results == nil {
		results = []FileInfo{}
	}

	// Return the search results
	c.JSON(http.StatusOK, gin.H{
		"data": results,
	})
}
