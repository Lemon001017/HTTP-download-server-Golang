package handlers

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

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

// isImageFile checks if a file extension belongs to an image format
func isImageFile(ext string) bool {
	ext = strings.TrimPrefix(ext, ".")
	return containsString([]string{"jpg", "jpeg", "png", "gif", "webp", "bmp"}, ext)
}

// isVideoFile checks if a file extension belongs to a video format
func isVideoFile(ext string) bool {
	ext = strings.TrimPrefix(ext, ".")
	return containsString([]string{"mp4", "webm", "ogg", "mov", "avi", "mkv", "flv"}, ext)
}

// getContentType returns the appropriate content type based on file extension
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

// isFileMatch checks if a filename matches the search query
func isFileMatch(filename, query string, exactMatch bool) bool {
	filename = strings.ToLower(filename)

	if exactMatch {
		return filename == query
	}

	return strings.Contains(filename, query)
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
