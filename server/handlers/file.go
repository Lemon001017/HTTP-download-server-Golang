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

// FileMkdirRequest represents the request parameters for creating a directory
type FileMkdirRequest struct {
	Path    string `json:"path" binding:"required"`    // Parent directory path
	DirName string `json:"dirName" binding:"required"` // Name of the new directory
}

// FileSearchRequest represents the request parameters for searching files
type FileSearchRequest struct {
	Path      string `json:"path"`                     // Directory path to search in
	Query     string `json:"query" binding:"required"` // Search query
	Exact     bool   `json:"exact"`                    // Whether to perform exact match or fuzzy search
	Recursive bool   `json:"recursive"`                // Whether to search recursively
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

// handleFileMkdir handles the request to create a new directory
func (h *Handlers) handleFileMkdir(c *gin.Context) {
	var request FileMkdirRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		carrot.AbortWithJSONError(c, http.StatusBadRequest, err)
		return
	}

	// Get settings to obtain the base download path
	settings, err := models.GetSettings(h.db, 1) // Using default user ID 1
	if err != nil {
		carrot.AbortWithJSONError(c, http.StatusInternalServerError, err)
		return
	}

	// Clean and validate the path
	basePath := settings.DownloadPath
	parentPath := strings.TrimPrefix(request.Path, "/")
	if parentPath == "" {
		parentPath = "."
	}

	// Construct the full path for the new directory
	fullParentPath := filepath.Join(basePath, parentPath)

	// Check if parent directory exists
	parentInfo, err := os.Stat(fullParentPath)
	if err != nil {
		if os.IsNotExist(err) {
			carrot.AbortWithJSONError(c, http.StatusNotFound, fmt.Errorf("parent directory not found: %s", fullParentPath))
			return
		}
		carrot.AbortWithJSONError(c, http.StatusInternalServerError, err)
		return
	}

	// Make sure the parent path is indeed a directory
	if !parentInfo.IsDir() {
		carrot.AbortWithJSONError(c, http.StatusBadRequest, fmt.Errorf("not a directory: %s", fullParentPath))
		return
	}

	// Create the new directory
	newDirName := request.DirName
	if newDirName == "" {
		carrot.AbortWithJSONError(c, http.StatusBadRequest, fmt.Errorf("directory name cannot be empty"))
		return
	}

	// Check for invalid characters in directory name
	if strings.ContainsAny(newDirName, "\\/:*?\"<>|") {
		carrot.AbortWithJSONError(c, http.StatusBadRequest, fmt.Errorf("directory name contains invalid characters"))
		return
	}

	// Create the full path for the new directory
	newDirPath := filepath.Join(fullParentPath, newDirName)

	// Check if the directory already exists
	if _, err := os.Stat(newDirPath); err == nil {
		carrot.AbortWithJSONError(c, http.StatusConflict, fmt.Errorf("directory already exists: %s", newDirName))
		return
	}

	// Create the directory
	if err := os.Mkdir(newDirPath, 0755); err != nil {
		carrot.AbortWithJSONError(c, http.StatusInternalServerError, err)
		return
	}

	// Calculate the relative path for the response
	relativePath := filepath.Join(parentPath, newDirName)
	if parentPath == "." {
		relativePath = newDirName
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"message": "Directory created successfully",
		"path":    relativePath,
	})
}

// handleFileSearch handles file search requests
func (h *Handlers) handleFileSearch(c *gin.Context) {
	var request FileSearchRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		carrot.AbortWithJSONError(c, http.StatusBadRequest, err)
		return
	}

	// Get settings to obtain the base download path
	settings, err := models.GetSettings(h.db, 1) // Using default user ID 1
	if err != nil {
		carrot.AbortWithJSONError(c, http.StatusInternalServerError, err)
		return
	}

	// Clean and validate the path
	basePath := settings.DownloadPath
	searchPath := strings.TrimPrefix(request.Path, "/")
	fullSearchPath := basePath
	if searchPath != "" {
		fullSearchPath = filepath.Join(basePath, searchPath)
	}

	// Ensure the search path exists and is a directory
	pathInfo, err := os.Stat(fullSearchPath)
	if err != nil {
		if os.IsNotExist(err) {
			carrot.AbortWithJSONError(c, http.StatusNotFound, fmt.Errorf("search path not found: %s", fullSearchPath))
			return
		}
		carrot.AbortWithJSONError(c, http.StatusInternalServerError, err)
		return
	}

	if !pathInfo.IsDir() {
		carrot.AbortWithJSONError(c, http.StatusBadRequest, fmt.Errorf("search path is not a directory: %s", fullSearchPath))
		return
	}

	// Normalize the search query
	query := strings.TrimSpace(strings.ToLower(request.Query))
	if query == "" {
		carrot.AbortWithJSONError(c, http.StatusBadRequest, fmt.Errorf("search query cannot be empty"))
		return
	}

	// Perform the search
	results := []FileInfo{} // 初始化为空数组而不是nil
	if request.Recursive {
		err = filepath.Walk(fullSearchPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // Skip files with errors
			}

			// Skip hidden files
			if strings.HasPrefix(filepath.Base(path), ".") {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			// Check if this file matches the search criteria
			if isFileMatch(info.Name(), query, request.Exact) {
				fileInfo := createFileInfo(path, info, basePath)
				results = append(results, fileInfo)
			}

			return nil
		})
		if err != nil {
			carrot.AbortWithJSONError(c, http.StatusInternalServerError, err)
			return
		}
	} else {
		// Non-recursive search only looks at immediate files in the directory
		files, err := os.ReadDir(fullSearchPath)
		if err != nil {
			carrot.AbortWithJSONError(c, http.StatusInternalServerError, err)
			return
		}

		for _, file := range files {
			// Skip hidden files
			if strings.HasPrefix(file.Name(), ".") {
				continue
			}

			// Check if this file matches the search criteria
			if isFileMatch(file.Name(), query, request.Exact) {
				info, err := file.Info()
				if err != nil {
					continue
				}
				filePath := filepath.Join(fullSearchPath, file.Name())
				fileInfo := createFileInfo(filePath, info, basePath)
				results = append(results, fileInfo)
			}
		}
	}

	// Return the search results
	c.JSON(http.StatusOK, gin.H{
		"data": results,
	})
}

// isFileMatch checks if a filename matches the search query
func isFileMatch(filename, query string, exactMatch bool) bool {
	filename = strings.ToLower(filename)

	if exactMatch {
		return filename == query
	}

	return strings.Contains(filename, query)
}

// createFileInfo creates a FileInfo struct from file information
func createFileInfo(filePath string, info os.FileInfo, basePath string) FileInfo {
	// Get the relative path from the base path
	relPath, _ := filepath.Rel(basePath, filePath)
	// Normalize slashes for web interface
	relPath = filepath.ToSlash(relPath)

	// Create file info
	fileInfo := FileInfo{
		FileName:    info.Name(),
		FilePath:    relPath,
		Directory:   info.IsDir(),
		GmtModified: info.ModTime(),
	}

	// Set file type and size for non-directories
	if !info.IsDir() {
		fileInfo.Size = info.Size()
		fileInfo.FileSize = formatFileSize(info.Size())
		fileInfo.FileType = strings.TrimPrefix(strings.ToLower(filepath.Ext(info.Name())), ".")
	}

	return fileInfo
}
