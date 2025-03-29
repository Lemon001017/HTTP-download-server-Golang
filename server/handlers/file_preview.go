package handlers

import (
	"HTTP-download-server/server/models"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/restsend/carrot"
)

// handleFilePreview handles the request to preview a file's contents
func (h *Handlers) handleFilePreview(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		carrot.AbortWithJSONError(c, http.StatusBadRequest, fmt.Errorf("missing file path"))
		return
	}

	// Get settings to obtain the download path
	settings, err := models.GetSettings(h.db, 1) // Using default user ID 1
	if err != nil {
		carrot.AbortWithJSONError(c, http.StatusInternalServerError, err)
		return
	}

	// Ensure path is relative and safe
	path = filepath.Clean(path)
	path = strings.TrimPrefix(path, "/")

	// Build the full path to the file
	fullPath := filepath.Join(settings.DownloadPath, path)

	// Check if the file exists
	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			carrot.AbortWithJSONError(c, http.StatusNotFound, fmt.Errorf("file not found: %s", path))
			return
		}
		carrot.AbortWithJSONError(c, http.StatusInternalServerError, err)
		return
	}

	// Don't allow directory previews
	if fileInfo.IsDir() {
		carrot.AbortWithJSONError(c, http.StatusBadRequest, fmt.Errorf("cannot preview directory: %s", path))
		return
	}

	// Get file extension
	fileExt := strings.ToLower(filepath.Ext(fullPath))

	// Only allow image files for preview
	if !isImageFile(fileExt) {
		carrot.AbortWithJSONError(c, http.StatusBadRequest, fmt.Errorf("file is not an image: %s", path))
		return
	}

	// Handle image file
	file, err := os.Open(fullPath)
	if err != nil {
		carrot.AbortWithJSONError(c, http.StatusInternalServerError, err)
		return
	}
	defer file.Close()

	// Set appropriate content headers
	contentType := getContentType(fileExt)
	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=%s", filepath.Base(path)))

	// Stream image content directly
	http.ServeContent(c.Writer, c.Request, filepath.Base(path), fileInfo.ModTime(), file)
}

// handleFileStream handles streaming a file (like video) for preview
func (h *Handlers) handleFileStream(c *gin.Context) {
	// Get file path from path parameter or query parameter
	path := c.Param("path")
	if path == "" {
		path = c.Query("path")
		if path == "" {
			carrot.AbortWithJSONError(c, http.StatusBadRequest, fmt.Errorf("missing file path"))
			return
		}
	}

	// Get settings to obtain the download path
	settings, err := models.GetSettings(h.db, 1) // Using default user ID 1
	if err != nil {
		carrot.AbortWithJSONError(c, http.StatusInternalServerError, err)
		return
	}

	// Ensure path is relative and safe
	path = strings.TrimPrefix(path, "/")
	fullPath := filepath.Join(settings.DownloadPath, path)

	// Check if the file exists
	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			carrot.AbortWithJSONError(c, http.StatusNotFound, fmt.Errorf("file not found: %s", path))
			return
		}
		carrot.AbortWithJSONError(c, http.StatusInternalServerError, err)
		return
	}

	// Don't allow directory streaming
	if fileInfo.IsDir() {
		carrot.AbortWithJSONError(c, http.StatusBadRequest, fmt.Errorf("cannot stream directory: %s", path))
		return
	}

	fileExt := strings.ToLower(filepath.Ext(fullPath))
	contentType := getContentType(fileExt)

	// Check if it's a video or image file
	if !isVideoFile(fileExt) && !isImageFile(fileExt) {
		carrot.AbortWithJSONError(c, http.StatusBadRequest, fmt.Errorf("file is not a video or image: %s", path))
		return
	}

	file, err := os.Open(fullPath)
	if err != nil {
		carrot.AbortWithJSONError(c, http.StatusInternalServerError, err)
		return
	}
	defer file.Close()

	// Get the file size
	fileSize := fileInfo.Size()

	// Set the Content-Type and Content-Length headers
	c.Header("Content-Type", contentType)
	c.Header("Content-Length", strconv.FormatInt(fileSize, 10))
	c.Header("Accept-Ranges", "bytes")

	// Handle Range header for video seeking
	rangeHeader := c.GetHeader("Range")
	if rangeHeader != "" {
		ranges, err := parseRange(rangeHeader, fileSize)
		if err != nil {
			// Set error message before status
			c.String(http.StatusRequestedRangeNotSatisfiable, "invalid range header format: %v", err)
			return
		}

		if len(ranges) > 1 {
			// Multiple ranges not supported for simplicity
			c.String(http.StatusRequestedRangeNotSatisfiable, "invalid range header format: multiple ranges not supported")
			return
		}

		// Handle single range
		if len(ranges) == 1 {
			r := ranges[0]
			_, err = file.Seek(r.start, io.SeekStart)
			if err != nil {
				carrot.AbortWithJSONError(c, http.StatusInternalServerError, err)
				return
			}

			// Calculate content length for the range
			contentLength := r.length
			c.Header("Content-Length", strconv.FormatInt(contentLength, 10))
			c.Header("Content-Range", fmt.Sprintf("bytes %d-%d/%d", r.start, r.start+contentLength-1, fileSize))
			c.Status(http.StatusPartialContent)

			// Stream the range
			_, err = io.CopyN(c.Writer, file, contentLength)
			if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
				// Connection may have been closed by client, which is normal
				return
			}
			return
		}
	}

	// If no Range header or ranges, stream the entire file
	c.Status(http.StatusOK)
	_, err = io.Copy(c.Writer, file)
	if err != nil && err != io.EOF {
		// Connection may have been closed by client, which is normal
		return
	}
}
