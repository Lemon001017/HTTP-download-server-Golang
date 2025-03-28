package handlers

import (
	"HTTP-download-server/server/models"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/restsend/carrot"
	"github.com/stretchr/testify/assert"
)

func TestFileList(t *testing.T) {
	// Create a temporary directory as a test download path
	tempDir, err := ioutil.TempDir("", "file-list-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create some test files and directories
	createTestFiles(t, tempDir)

	r, db := createTestHandlers()
	c := carrot.NewTestClient(r)

	// Create test settings with the temp directory
	settings := &models.Settings{
		UserID:           1,
		DownloadPath:     tempDir,
		MaxDownloadSpeed: 5,
		MaxTasks:         2,
	}
	db.Create(settings)

	t.Run("invalid request", func(t *testing.T) {
		req := []byte(`{"invalid_json":}`)
		w := c.Post("POST", "/api/file/list", req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("list root directory", func(t *testing.T) {
		req := []byte(`{"path":"", "type":"", "sort":"", "order":"down"}`)
		w := c.Post("POST", "/api/file/list", req)
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.Nil(t, err)

		data, ok := response["data"].([]interface{})
		assert.True(t, ok)
		assert.Equal(t, 4, len(data), "Should have 4 files/directories")
	})

	t.Run("filter by type", func(t *testing.T) {
		req := []byte(`{"path":"", "type":"document", "sort":"", "order":"down"}`)
		w := c.Post("POST", "/api/file/list", req)
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.Nil(t, err)

		data, ok := response["data"].([]interface{})
		assert.True(t, ok)

		// Should have the test directory and the txt file
		documentFiles := 0
		for _, item := range data {
			fileInfo := item.(map[string]interface{})
			if fileInfo["directory"].(bool) {
				continue
			}
			if fileInfo["fileType"].(string) == "txt" {
				documentFiles++
			}
		}
		assert.Equal(t, 1, documentFiles, "Should have 1 document file")
	})

	t.Run("sort by size", func(t *testing.T) {
		req := []byte(`{"path":"", "type":"", "sort":"size", "order":"down"}`)
		w := c.Post("POST", "/api/file/list", req)
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.Nil(t, err)

		data, ok := response["data"].([]interface{})
		assert.True(t, ok)

		// Check if directories are first
		firstItem := data[0].(map[string]interface{})
		assert.True(t, firstItem["directory"].(bool), "Directories should be listed first")
	})

	t.Run("list subdirectory", func(t *testing.T) {
		req := []byte(`{"path":"/subdir", "type":"", "sort":"", "order":"down"}`)
		w := c.Post("POST", "/api/file/list", req)
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.Nil(t, err)

		data, ok := response["data"].([]interface{})
		assert.True(t, ok)
		assert.Equal(t, 1, len(data), "Subdirectory should have 1 file")
	})

	t.Run("nonexistent path", func(t *testing.T) {
		req := []byte(`{"path":"/nonexistent", "type":"", "sort":"", "order":"down"}`)
		w := c.Post("POST", "/api/file/list", req)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

// Helper function to create test files and directories
func createTestFiles(t *testing.T, baseDir string) {
	// Create a subdirectory
	subDir := filepath.Join(baseDir, "subdir")
	err := os.Mkdir(subDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	// Create some files
	files := map[string]string{
		filepath.Join(baseDir, "text.txt"):    "This is a text file",
		filepath.Join(baseDir, "image.jpg"):   "Fake image content",
		filepath.Join(baseDir, "archive.zip"): "Fake zip content",
		filepath.Join(subDir, "subfile.txt"):  "This is a file in a subdirectory",
	}

	for path, content := range files {
		err := ioutil.WriteFile(path, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", path, err)
		}
	}
}

// TestFilePreview tests the image preview functionality
func TestFilePreview(t *testing.T) {
	r, db := createTestHandlers()
	c := carrot.NewTestClient(r)

	// Create a temporary directory for test files
	tempDir, err := ioutil.TempDir("", "file-preview-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files
	imageContent := "Fake image data for testing"
	imagePath := filepath.Join(tempDir, "test-image.jpg")
	err = ioutil.WriteFile(imagePath, []byte(imageContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test image: %v", err)
	}

	// Create settings with the temp directory as download path
	settings := &models.Settings{
		UserID:           1,
		DownloadPath:     tempDir,
		MaxDownloadSpeed: 1.0,
		MaxTasks:         10,
	}
	db.Create(settings)

	t.Run("successful image preview", func(t *testing.T) {
		w := c.Get("/api/file/preview?path=test-image.jpg")
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "image/jpeg", w.Header().Get("Content-Type"))
		assert.Equal(t, "inline; filename=test-image.jpg", w.Header().Get("Content-Disposition"))
		assert.Equal(t, imageContent, w.Body.String())
	})

	t.Run("missing path parameter", func(t *testing.T) {
		w := c.Get("/api/file/preview")
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "missing file path")
	})

	t.Run("nonexistent file", func(t *testing.T) {
		w := c.Get("/api/file/preview?path=nonexistent.jpg")
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "file not found")
	})

	t.Run("not an image file", func(t *testing.T) {
		// Create a non-image file
		textPath := filepath.Join(tempDir, "test.txt")
		err = ioutil.WriteFile(textPath, []byte("This is a text file"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test text file: %v", err)
		}

		w := c.Get("/api/file/preview?path=test.txt")
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "file is not an image")
	})
}

// TestFileStream tests the video streaming functionality
func TestFileStream(t *testing.T) {
	r, db := createTestHandlers()
	c := carrot.NewTestClient(r)

	// Create a temporary directory for test files
	tempDir, err := ioutil.TempDir("", "file-stream-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test video file
	videoContent := "Fake video data for testing MP4 format"
	videoPath := filepath.Join(tempDir, "test-video.mp4")
	err = ioutil.WriteFile(videoPath, []byte(videoContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test video: %v", err)
	}

	// Create settings with the temp directory as download path
	settings := &models.Settings{
		UserID:           1,
		DownloadPath:     tempDir,
		MaxDownloadSpeed: 1.0,
		MaxTasks:         10,
	}
	db.Create(settings)

	t.Run("successful video stream", func(t *testing.T) {
		w := c.Get("/api/file/stream?path=test-video.mp4")
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "video/mp4", w.Header().Get("Content-Type"))
		assert.Equal(t, fmt.Sprintf("%d", len(videoContent)), w.Header().Get("Content-Length"))
		assert.Equal(t, "bytes", w.Header().Get("Accept-Ranges"))
		assert.Equal(t, videoContent, w.Body.String())
	})

	t.Run("partial content request", func(t *testing.T) {
		// Create a request with a range header
		req, _ := http.NewRequest("GET", "/api/file/stream?path=test-video.mp4", nil)
		req.Header.Set("Range", "bytes=0-9") // Request first 10 bytes

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusPartialContent, w.Code)
		assert.Equal(t, "video/mp4", w.Header().Get("Content-Type"))
		assert.Equal(t, "10", w.Header().Get("Content-Length"))
		assert.Equal(t, fmt.Sprintf("bytes 0-9/%d", len(videoContent)), w.Header().Get("Content-Range"))
		assert.Equal(t, videoContent[:10], w.Body.String())
	})

	t.Run("missing path parameter", func(t *testing.T) {
		w := c.Get("/api/file/stream")
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "missing file path")
	})

	t.Run("nonexistent file", func(t *testing.T) {
		w := c.Get("/api/file/stream?path=nonexistent.mp4")
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "file not found")
	})

	t.Run("not a video file", func(t *testing.T) {
		// Create a non-video file
		textPath := filepath.Join(tempDir, "test.txt")
		err = ioutil.WriteFile(textPath, []byte("This is a text file"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test text file: %v", err)
		}

		w := c.Get("/api/file/stream?path=test.txt")
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "file is not a video")
	})

	t.Run("invalid range request", func(t *testing.T) {
		// Create a request with an invalid range header
		req, _ := http.NewRequest("GET", "/api/file/stream?path=test-video.mp4", nil)
		req.Header.Set("Range", "bytes=invalid-range")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusRequestedRangeNotSatisfiable, w.Code)
		assert.Contains(t, w.Body.String(), "invalid range header format")
	})

	t.Run("range start beyond file size", func(t *testing.T) {
		// Create a request with a range start beyond the file size
		req, _ := http.NewRequest("GET", "/api/file/stream?path=test-video.mp4", nil)
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", len(videoContent)+100))

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusRequestedRangeNotSatisfiable, w.Code)
		assert.Contains(t, w.Body.String(), "invalid range header format")
	})
}

// TestFileRename tests the file renaming functionality
func TestFileRename(t *testing.T) {
	r, db := createTestHandlers()
	c := carrot.NewTestClient(r)

	// Create a temporary directory for test files
	tempDir, err := ioutil.TempDir("", "file-rename-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test file
	testFilePath := filepath.Join(tempDir, "test-file.txt")
	err = ioutil.WriteFile(testFilePath, []byte("Test file content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create test subdirectory
	subDir := filepath.Join(tempDir, "subdir")
	err = os.Mkdir(subDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test subdirectory: %v", err)
	}

	// Create a file in the subdirectory
	subFilePath := filepath.Join(subDir, "sub-test-file.txt")
	err = ioutil.WriteFile(subFilePath, []byte("Subdirectory test file content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file in subdirectory: %v", err)
	}

	// Create settings with the temp directory as download path
	settings := &models.Settings{
		UserID:           1,
		DownloadPath:     tempDir,
		MaxDownloadSpeed: 1.0,
		MaxTasks:         10,
	}
	db.Create(settings)

	t.Run("successful rename", func(t *testing.T) {
		// Create JSON request body
		requestBody := FileRenameRequest{
			Path:    "test-file.txt",
			NewName: "renamed-file.txt",
		}
		body, err := json.Marshal(requestBody)
		assert.NoError(t, err)

		// Send the request
		w := c.Post("POST", "/api/file/rename", body)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "File renamed successfully")

		// Verify the file has been renamed
		_, err = os.Stat(filepath.Join(tempDir, "renamed-file.txt"))
		assert.NoError(t, err)
		_, err = os.Stat(filepath.Join(tempDir, "test-file.txt"))
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("rename file in subdirectory", func(t *testing.T) {
		// Create JSON request body
		requestBody := FileRenameRequest{
			Path:    filepath.Join("subdir", "sub-test-file.txt"),
			NewName: "renamed-subfile.txt",
		}
		body, err := json.Marshal(requestBody)
		assert.NoError(t, err)

		// Send the request
		w := c.Post("POST", "/api/file/rename", body)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "File renamed successfully")

		// Verify the file has been renamed
		_, err = os.Stat(filepath.Join(subDir, "renamed-subfile.txt"))
		assert.NoError(t, err)
		_, err = os.Stat(filepath.Join(subDir, "sub-test-file.txt"))
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("file not found", func(t *testing.T) {
		// Create JSON request body
		requestBody := FileRenameRequest{
			Path:    "nonexistent-file.txt",
			NewName: "new-name.txt",
		}
		body, err := json.Marshal(requestBody)
		assert.NoError(t, err)

		// Send the request
		w := c.Post("POST", "/api/file/rename", body)
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "file not found")
	})

	t.Run("invalid request", func(t *testing.T) {
		// Invalid JSON
		w := c.Post("POST", "/api/file/rename", []byte(`{"invalid_json":`))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// TestFileDelete tests the file deletion functionality
func TestFileDelete(t *testing.T) {
	r, db := createTestHandlers()
	c := carrot.NewTestClient(r)

	// Create a temporary directory for test files
	tempDir, err := ioutil.TempDir("", "file-delete-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test file
	testFilePath := filepath.Join(tempDir, "test-file.txt")
	err = ioutil.WriteFile(testFilePath, []byte("Test file content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create empty test directory
	emptyDir := filepath.Join(tempDir, "empty-dir")
	err = os.Mkdir(emptyDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create empty test directory: %v", err)
	}

	// Create non-empty test directory
	nonEmptyDir := filepath.Join(tempDir, "non-empty-dir")
	err = os.Mkdir(nonEmptyDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create non-empty test directory: %v", err)
	}

	// Create a file in the non-empty directory
	nonEmptyFilePath := filepath.Join(nonEmptyDir, "file.txt")
	err = ioutil.WriteFile(nonEmptyFilePath, []byte("Non-empty directory test file"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file in non-empty directory: %v", err)
	}

	// Create settings with the temp directory as download path
	settings := &models.Settings{
		UserID:           1,
		DownloadPath:     tempDir,
		MaxDownloadSpeed: 1.0,
		MaxTasks:         10,
	}
	db.Create(settings)

	t.Run("successful file delete", func(t *testing.T) {
		// Create JSON request body
		requestBody := FileDeleteRequest{
			Path: "test-file.txt",
		}
		body, err := json.Marshal(requestBody)
		assert.NoError(t, err)

		// Send the request
		w := c.Post("POST", "/api/file/delete", body)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "File deleted successfully")

		// Verify the file has been deleted
		_, err = os.Stat(testFilePath)
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("successful empty directory delete", func(t *testing.T) {
		// Create JSON request body
		requestBody := FileDeleteRequest{
			Path: "empty-dir",
		}
		body, err := json.Marshal(requestBody)
		assert.NoError(t, err)

		// Send the request
		w := c.Post("POST", "/api/file/delete", body)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "File deleted successfully")

		// Verify the directory has been deleted
		_, err = os.Stat(emptyDir)
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("non-empty directory delete error", func(t *testing.T) {
		// Create JSON request body
		requestBody := FileDeleteRequest{
			Path: "non-empty-dir",
		}
		body, err := json.Marshal(requestBody)
		assert.NoError(t, err)

		// Send the request
		w := c.Post("POST", "/api/file/delete", body)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "directory is not empty")

		// Verify the directory still exists
		_, err = os.Stat(nonEmptyDir)
		assert.NoError(t, err)
	})

	t.Run("file not found", func(t *testing.T) {
		// Create JSON request body
		requestBody := FileDeleteRequest{
			Path: "nonexistent-file.txt",
		}
		body, err := json.Marshal(requestBody)
		assert.NoError(t, err)

		// Send the request
		w := c.Post("POST", "/api/file/delete", body)
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "file not found")
	})

	t.Run("invalid request", func(t *testing.T) {
		// Invalid JSON
		w := c.Post("POST", "/api/file/delete", []byte(`{"invalid_json":`))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
