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
			URL: "https://q6.itc.cn/images01/20240813/d4b6af09a7d74a05890afa16b9e4dfa8.jpeg",
		}
		req, err := json.Marshal(reqBody)
		assert.Nil(t, err)
		w := c.Post("POST", "/api/task/submit", req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "key")
		time.Sleep(2 * time.Second)
		os.Remove("./d4b6af09a7d74a05890afa16b9e4dfa8.jpeg")
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
