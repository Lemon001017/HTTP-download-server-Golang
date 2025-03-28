package handlers

import (
	"HTTP-download-server/server/models"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
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

// FileRenameRequest represents the request parameters for renaming a file
type FileRenameRequest struct {
	Path    string `json:"path" binding:"required"`    // Current file path
	NewName string `json:"newName" binding:"required"` // New file name
}

// FileDeleteRequest represents the request parameters for deleting a file
type FileDeleteRequest struct {
	Path string `json:"path" binding:"required"` // Path to the file to delete
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

// handleFilePreview handles requests to preview image files
func (h *Handlers) handleFilePreview(c *gin.Context) {
	// Get the file path from the query string
	filePath := c.Query("path")
	if filePath == "" {
		carrot.AbortWithJSONError(c, http.StatusBadRequest, fmt.Errorf("missing file path"))
		return
	}

	// Get settings to obtain the download path
	settings, err := models.GetSettings(h.db, 1) // Using default user ID 1
	if err != nil {
		carrot.AbortWithJSONError(c, http.StatusInternalServerError, err)
		return
	}

	// Determine the full path to the file
	basePath := settings.DownloadPath
	fullPath := filepath.Join(basePath, filePath)

	// Verify that the file exists and is not a directory
	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			carrot.AbortWithJSONError(c, http.StatusNotFound, fmt.Errorf("file not found: %s", fullPath))
			return
		}
		carrot.AbortWithJSONError(c, http.StatusInternalServerError, err)
		return
	}

	if fileInfo.IsDir() {
		carrot.AbortWithJSONError(c, http.StatusBadRequest, fmt.Errorf("cannot preview a directory"))
		return
	}

	// Check if the file is an image based on extension
	ext := strings.ToLower(filepath.Ext(fullPath))
	if !isImageFile(ext) {
		carrot.AbortWithJSONError(c, http.StatusBadRequest, fmt.Errorf("file is not an image: %s", ext))
		return
	}

	// Set appropriate content type based on file extension
	contentType := getContentType(ext)
	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=%s", filepath.Base(fullPath)))

	// Serve the file
	c.File(fullPath)
}

// isImageFile checks if a file extension belongs to an image format
func isImageFile(ext string) bool {
	ext = strings.TrimPrefix(ext, ".")
	return containsString([]string{"jpg", "jpeg", "png", "gif", "webp", "bmp"}, ext)
}

// getContentType returns the appropriate content type for an image based on its extension
func getContentType(ext string) string {
	switch strings.ToLower(strings.TrimPrefix(ext, ".")) {
	case "jpg", "jpeg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "gif":
		return "image/gif"
	case "webp":
		return "image/webp"
	case "bmp":
		return "image/bmp"
	case "mp4":
		return "video/mp4"
	case "webm":
		return "video/webm"
	case "ogg":
		return "video/ogg"
	case "mov":
		return "video/quicktime"
	case "avi":
		return "video/x-msvideo"
	case "mkv":
		return "video/x-matroska"
	case "flv":
		return "video/x-flv"
	default:
		return "application/octet-stream"
	}
}

// isVideoFile checks if a file extension belongs to a video format
func isVideoFile(ext string) bool {
	ext = strings.TrimPrefix(ext, ".")
	return containsString([]string{"mp4", "webm", "ogg", "mov", "avi", "mkv", "flv"}, ext)
}

// handleFileStream handles requests to stream video files
func (h *Handlers) handleFileStream(c *gin.Context) {
	// Get the file path from the query string
	filePath := c.Query("path")
	if filePath == "" {
		carrot.AbortWithJSONError(c, http.StatusBadRequest, fmt.Errorf("missing file path"))
		return
	}

	// Get settings to obtain the download path
	settings, err := models.GetSettings(h.db, 1) // Using default user ID 1
	if err != nil {
		carrot.AbortWithJSONError(c, http.StatusInternalServerError, err)
		return
	}

	// Determine the full path to the file
	basePath := settings.DownloadPath
	fullPath := filepath.Join(basePath, filePath)

	// Verify that the file exists and is not a directory
	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			carrot.AbortWithJSONError(c, http.StatusNotFound, fmt.Errorf("file not found: %s", fullPath))
			return
		}
		carrot.AbortWithJSONError(c, http.StatusInternalServerError, err)
		return
	}

	if fileInfo.IsDir() {
		carrot.AbortWithJSONError(c, http.StatusBadRequest, fmt.Errorf("cannot stream a directory"))
		return
	}

	// Check if the file is a video based on extension
	ext := strings.ToLower(filepath.Ext(fullPath))
	if !isVideoFile(ext) {
		carrot.AbortWithJSONError(c, http.StatusBadRequest, fmt.Errorf("file is not a video: %s", ext))
		return
	}

	// Set appropriate content type based on file extension
	contentType := getContentType(ext)
	c.Header("Content-Type", contentType)

	// Support for range requests (important for video streaming)
	file, err := os.Open(fullPath)
	if err != nil {
		carrot.AbortWithJSONError(c, http.StatusInternalServerError, err)
		return
	}
	defer file.Close()

	// Get file size
	fileSize := fileInfo.Size()

	// Parse range header
	rangeHeader := c.GetHeader("Range")
	if rangeHeader != "" {
		// Parse the range header
		ranges, err := parseRange(rangeHeader, fileSize)
		if err != nil {
			c.Header("Content-Range", fmt.Sprintf("bytes */%d", fileSize))
			carrot.AbortWithJSONError(c, http.StatusRequestedRangeNotSatisfiable, err)
			return
		}

		// We only support a single range for now
		if len(ranges) > 1 {
			c.Header("Content-Range", fmt.Sprintf("bytes */%d", fileSize))
			carrot.AbortWithJSONError(c, http.StatusRequestedRangeNotSatisfiable, fmt.Errorf("multiple ranges not supported"))
			return
		}

		// Get the range
		ra := ranges[0]

		// Set content length
		c.Header("Content-Length", fmt.Sprintf("%d", ra.length))
		c.Header("Content-Range", fmt.Sprintf("bytes %d-%d/%d", ra.start, ra.start+ra.length-1, fileSize))
		c.Header("Accept-Ranges", "bytes")
		c.Status(http.StatusPartialContent)

		// Seek to the start position
		_, err = file.Seek(ra.start, io.SeekStart)
		if err != nil {
			carrot.AbortWithJSONError(c, http.StatusInternalServerError, err)
			return
		}

		// Copy the specified range to the response
		_, err = io.CopyN(c.Writer, file, ra.length)
		if err != nil {
			carrot.AbortWithJSONError(c, http.StatusInternalServerError, err)
			return
		}
	} else {
		// No range requested, send the entire file
		c.Header("Content-Length", fmt.Sprintf("%d", fileSize))
		c.Header("Accept-Ranges", "bytes")
		c.Status(http.StatusOK)
		_, err = io.Copy(c.Writer, file)
		if err != nil {
			carrot.AbortWithJSONError(c, http.StatusInternalServerError, err)
			return
		}
	}
}

// httpRange specifies the byte range to be sent to the client
type httpRange struct {
	start  int64
	length int64
}

// parseRange parses a Range header string as per RFC 7233
func parseRange(s string, size int64) ([]httpRange, error) {
	// Format: "bytes=0-499" or "bytes=500-"
	const b = "bytes="
	if !strings.HasPrefix(s, b) {
		return nil, fmt.Errorf("invalid range header format: %s", s)
	}

	ranges := []httpRange{}
	rangeSpecs := strings.Split(s[len(b):], ",")

	for _, rangeSpec := range rangeSpecs {
		rangeSpec = strings.TrimSpace(rangeSpec)
		if rangeSpec == "" {
			continue
		}

		i := strings.Index(rangeSpec, "-")
		if i < 0 {
			return nil, fmt.Errorf("invalid range header format: %s", s)
		}

		start, end := strings.TrimSpace(rangeSpec[:i]), strings.TrimSpace(rangeSpec[i+1:])

		var r httpRange

		// Parse start
		if start == "" {
			// No start specified, e.g. "-500" means the last 500 bytes
			i, err := strconv.ParseInt(end, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid range header format: %s", s)
			}
			if i > size {
				i = size
			}
			r.start = size - i
			r.length = size - r.start
		} else {
			// Start specified, e.g. "500-" or "500-999"
			i, err := strconv.ParseInt(start, 10, 64)
			if err != nil || i >= size {
				return nil, fmt.Errorf("invalid range header format: %s", s)
			}
			r.start = i
			if end == "" {
				// No end specified, e.g. "500-", means to the end of the file
				r.length = size - r.start
			} else {
				// End specified, e.g. "500-999"
				i, err := strconv.ParseInt(end, 10, 64)
				if err != nil || r.start > i {
					return nil, fmt.Errorf("invalid range header format: %s", s)
				}
				if i >= size {
					i = size - 1
				}
				r.length = i - r.start + 1
			}
		}

		ranges = append(ranges, r)
	}

	if len(ranges) == 0 {
		return nil, fmt.Errorf("invalid range header format: %s", s)
	}

	return ranges, nil
}

// handleFileRename handles the request to rename a file
func (h *Handlers) handleFileRename(c *gin.Context) {
	var request FileRenameRequest
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

	// Determine the full path of the file to be renamed
	basePath := settings.DownloadPath
	sourcePath := filepath.Join(basePath, request.Path)

	// Check if the file exists
	_, err = os.Stat(sourcePath)
	if err != nil {
		if os.IsNotExist(err) {
			carrot.AbortWithJSONError(c, http.StatusNotFound, fmt.Errorf("file not found: %s", sourcePath))
			return
		}
		carrot.AbortWithJSONError(c, http.StatusInternalServerError, err)
		return
	}

	// Get directory of the file
	dir := filepath.Dir(request.Path)

	// Determine the destination path with the new filename
	destPath := filepath.Join(basePath, dir, request.NewName)

	// Check if the destination already exists
	_, err = os.Stat(destPath)
	if err == nil {
		carrot.AbortWithJSONError(c, http.StatusBadRequest, fmt.Errorf("destination already exists: %s", destPath))
		return
	}

	// Rename the file
	err = os.Rename(sourcePath, destPath)
	if err != nil {
		carrot.AbortWithJSONError(c, http.StatusInternalServerError, fmt.Errorf("failed to rename file: %v", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "File renamed successfully",
		"newPath": filepath.Join(dir, request.NewName),
	})
}

// handleFileDelete handles the request to delete a file
func (h *Handlers) handleFileDelete(c *gin.Context) {
	var request FileDeleteRequest
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

	// Determine the full path of the file to delete
	basePath := settings.DownloadPath
	fullPath := filepath.Join(basePath, request.Path)

	// Get file info to check if it's a directory
	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			carrot.AbortWithJSONError(c, http.StatusNotFound, fmt.Errorf("file not found: %s", fullPath))
			return
		}
		carrot.AbortWithJSONError(c, http.StatusInternalServerError, err)
		return
	}

	// Delete the file or directory
	var deleteErr error
	if fileInfo.IsDir() {
		// Check if directory is empty
		entries, err := os.ReadDir(fullPath)
		if err != nil {
			carrot.AbortWithJSONError(c, http.StatusInternalServerError, err)
			return
		}

		if len(entries) > 0 {
			carrot.AbortWithJSONError(c, http.StatusBadRequest, fmt.Errorf("directory is not empty: %s", fullPath))
			return
		}

		deleteErr = os.Remove(fullPath) // Remove empty directory
	} else {
		deleteErr = os.Remove(fullPath) // Remove file
	}

	if deleteErr != nil {
		carrot.AbortWithJSONError(c, http.StatusInternalServerError, fmt.Errorf("failed to delete: %v", deleteErr))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "File deleted successfully",
	})
}
