package handlers

import (
	"HTTP-download-server/server/models"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/restsend/carrot"
)

// FileInfo represents the file or directory information
type FileInfo struct {
	FileName    string    `json:"fileName"`
	FilePath    string    `json:"filePath"`
	FileSize    string    `json:"fileSize"`
	Size        int64     `json:"size"`
	Directory   bool      `json:"directory"`
	FileType    string    `json:"fileType"`
	GmtModified time.Time `json:"gmtModified"`
}

// FileListRequest represents the request parameters for listing files
type FileListRequest struct {
	Path  string `json:"path"`
	Type  string `json:"type"`
	Sort  string `json:"sort"`
	Order string `json:"order"`
}

// HandleFileList handles the request to list files in a directory
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

// matchesFileType checks if a file extension matches the requested type filter
func matchesFileType(fileExt, requestType string) bool {
	switch strings.ToLower(requestType) {
	case "video":
		return containsString([]string{"mp4", "mov", "avi", "mkv", "wmv", "flv"}, fileExt)
	case "jpg", "photo":
		return containsString([]string{"jpg", "jpeg", "png", "gif", "bmp", "webp"}, fileExt)
	case "archive":
		return containsString([]string{"zip", "rar", "tar", "gz", "7z"}, fileExt)
	case "document":
		return containsString([]string{"pptx", "docx", "xlsx", "pdf", "txt", "doc", "xls", "ppt"}, fileExt)
	default:
		return true // No filter specified, include all files
	}
}

// containsString checks if a string is in a slice of strings
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// formatFileSize formats a file size in bytes to a human-readable string
func formatFileSize(size int64) string {
	const (
		B  = 1
		KB = B * 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	if size < KB {
		return fmt.Sprintf("%d B", size)
	} else if size < MB {
		return fmt.Sprintf("%.2f KB", float64(size)/float64(KB))
	} else if size < GB {
		return fmt.Sprintf("%.2f MB", float64(size)/float64(MB))
	} else if size < TB {
		return fmt.Sprintf("%.2f GB", float64(size)/float64(GB))
	} else {
		return fmt.Sprintf("%.2f TB", float64(size)/float64(TB))
	}
}

// sortFileList sorts a list of files based on specified criteria
func sortFileList(files []FileInfo, sortBy, order string) {
	// Always put directories first
	sort.SliceStable(files, func(i, j int) bool {
		// Directories go first
		if files[i].Directory != files[j].Directory {
			return files[i].Directory
		}

		// Then sort by the specified field
		switch strings.ToLower(sortBy) {
		case "size":
			if order == "down" {
				return files[i].Size > files[j].Size
			}
			return files[i].Size < files[j].Size
		case "gmtmodified", "gmtcreated":
			if order == "down" {
				return files[i].GmtModified.After(files[j].GmtModified)
			}
			return files[i].GmtModified.Before(files[j].GmtModified)
		case "name", "filename":
			if order == "down" {
				return strings.ToLower(files[i].FileName) > strings.ToLower(files[j].FileName)
			}
			return strings.ToLower(files[i].FileName) < strings.ToLower(files[j].FileName)
		default: // Default sort by name
			if order == "down" {
				return strings.ToLower(files[i].FileName) > strings.ToLower(files[j].FileName)
			}
			return strings.ToLower(files[i].FileName) < strings.ToLower(files[j].FileName)
		}
	})
}
