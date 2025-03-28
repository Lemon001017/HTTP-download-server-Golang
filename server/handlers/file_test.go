package handlers

import (
	"HTTP-download-server/server/models"
	"encoding/json"
	"io/ioutil"
	"net/http"
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
