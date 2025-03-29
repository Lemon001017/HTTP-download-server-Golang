package handlers

import (
	"time"
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

// httpRange specifies the byte range to be sent to the client
type httpRange struct {
	start  int64
	length int64
}
