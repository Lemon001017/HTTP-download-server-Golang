package handlers

import (
	"HTTP-download-server/server/models"
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/restsend/carrot"
	"github.com/stretchr/testify/assert"
)

func TestSubmit(t *testing.T) {
	r, db := createTestHandlers()
	c := carrot.NewTestClient(r)
	t.Run("get settings error", func(t *testing.T) {
		reqBody := DownloadRequest{
			URL: "https://zy.yunqishi8.com/upload/mp4/202005/1-10.mp4",
		}
		req, err := json.Marshal(reqBody)
		assert.Nil(t, err)
		w := c.Post("POST", "/api/task/submit", req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("submit success", func(t *testing.T) {
		settings := models.Settings{
			UserID:           1,
			DownloadPath:     "./",
			MaxDownloadSpeed: 5,
			MaxTasks:         2,
		}
		db.Create(&settings)

		reqBody := DownloadRequest{
			URL: "https://i1.hdslb.com/bfs/archive/8db3fd38ae6eb0625e0c3b1d274160294d7bd5f5.jpg",
		}
		req, err := json.Marshal(reqBody)
		assert.Nil(t, err)
		w := c.Post("POST", "/api/task/submit", req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "key")
		time.Sleep(2 * time.Second)
		err = os.Remove("./8db3fd38ae6eb0625e0c3b1d274160294d7bd5f5.jpg")
		assert.Nil(t, err)
	})

	t.Run("not exist url", func(t *testing.T) {
		settings := models.Settings{
			UserID:           1,
			DownloadPath:     "./",
			MaxDownloadSpeed: 5,
			MaxTasks:         2,
		}
		db.Create(&settings)

		reqBody := DownloadRequest{
			URL: "https://not_exist",
		}
		req, err := json.Marshal(reqBody)
		assert.Nil(t, err)
		w := c.Post("POST", "/api/task/submit", req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("not parse url", func(t *testing.T) {
		settings := models.Settings{
			UserID:           1,
			DownloadPath:     "./",
			MaxDownloadSpeed: 5,
			MaxTasks:         2,
		}
		db.Create(&settings)

		reqBody := DownloadRequest{
			URL: "https://img2.baidu.com/it/u=652299287,3144977570&fm=253&fmt=auto&app=138&f=JPEG?w=800&h=1372",
		}
		req, err := json.Marshal(reqBody)
		assert.Nil(t, err)
		w := c.Post("POST", "/api/task/submit", req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("bind req error", func(t *testing.T) {
		settings := models.Settings{
			UserID:           1,
			DownloadPath:     "d:/project",
			MaxDownloadSpeed: 5,
			MaxTasks:         2,
		}
		db.Create(&settings)

		reqBody := DownloadRequest{
			URL: "https://zy.yunqishi8.com/upload/mp4/202005/1-10.mp4",
		}
		w := c.Post("POST", "/api/task/submit", []byte(reqBody.URL))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestRestart(t *testing.T) {
	r, db := createTestHandlers()
	c := carrot.NewTestClient(r)
	t.Run("restart success", func(t *testing.T) {
		settings := models.Settings{
			UserID:       1,
			DownloadPath: "./",
		}
		db.Create(&settings)

		task := models.Task{
			ID:              "1",
			Url:             "https://i1.hdslb.com/bfs/archive/8db3fd38ae6eb0625e0c3b1d274160294d7bd5f5.jpg",
			Status:          models.TaskStatusDownloaded,
			TotalDownloaded: 12345,
		}
		db.Create(&task)

		req := []byte(`["1"]`)
		w := c.Post("POST", "/api/task/restart", req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("bing ids error", func(t *testing.T) {
		settings := models.Settings{
			UserID:       1,
			DownloadPath: "./",
		}
		db.Create(&settings)

		task := models.Task{
			ID:              "1",
			Url:             "https://i1.hdslb.com/bfs/archive/8db3fd38ae6eb0625e0c3b1d274160294d7bd5f5.jpg",
			Status:          models.TaskStatusDownloaded,
			TotalDownloaded: 12345,
		}
		db.Create(&task)

		req := []byte("invalid")
		w := c.Post("POST", "/api/task/restart", req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("task status error", func(t *testing.T) {
		settings := models.Settings{
			UserID:       1,
			DownloadPath: "./",
		}
		db.Create(&settings)

		task := models.Task{
			ID:              "1",
			Url:             "https://i1.hdslb.com/bfs/archive/8db3fd38ae6eb0625e0c3b1d274160294d7bd5f5.jpg",
			Status:          models.TaskStatusDownloading,
			TotalDownloaded: 12345,
		}
		db.Create(&task)

		req := []byte(`["1"]`)
		w := c.Post("POST", "/api/task/restart", req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestPause(t *testing.T) {
	r, db := createTestHandlers()
	c := carrot.NewTestClient(r)
	t.Run("pause success", func(t *testing.T) {
		settings := models.Settings{
			UserID:       1,
			DownloadPath: "./",
		}
		db.Create(&settings)

		task := models.Task{
			ID:              "1",
			Url:             "https://i1.hdslb.com/bfs/archive/8db3fd38ae6eb0625e0c3b1d274160294d7bd5f5.jpg",
			Status:          models.TaskStatusDownloading,
			TotalDownloaded: 12345,
		}
		db.Create(&task)

		req := []byte(`["1"]`)
		w := c.Post("POST", "/api/task/pause", req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("bing ids error", func(t *testing.T) {
		settings := models.Settings{
			UserID:       1,
			DownloadPath: "./",
		}
		db.Create(&settings)

		task := models.Task{
			ID:              "1",
			Url:             "https://i1.hdslb.com/bfs/archive/8db3fd38ae6eb0625e0c3b1d274160294d7bd5f5.jpg",
			Status:          models.TaskStatusDownloaded,
			TotalDownloaded: 12345,
		}
		db.Create(&task)

		req := []byte("invalid")
		w := c.Post("POST", "/api/task/pause", req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("task status error", func(t *testing.T) {
		settings := models.Settings{
			UserID:       1,
			DownloadPath: "./",
		}
		db.Create(&settings)

		task := models.Task{
			ID:              "1",
			Url:             "https://i1.hdslb.com/bfs/archive/8db3fd38ae6eb0625e0c3b1d274160294d7bd5f5.jpg",
			Status:          models.TaskFilterDownloaded,
			TotalDownloaded: 12345,
		}
		db.Create(&task)

		req := []byte(`["1"]`)
		w := c.Post("POST", "/api/task/pause", req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
