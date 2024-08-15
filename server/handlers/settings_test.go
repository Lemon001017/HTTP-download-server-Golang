package handlers

import (
	"HTTP-download-server/server/models"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/restsend/carrot"
	"github.com/stretchr/testify/assert"
)

func TestSaveSettings(t *testing.T) {
	r, db := createTestHandlers()
	c := carrot.NewTestClient(r)
	t.Run("successfulUpdate", func(t *testing.T) {
		settings := &models.Settings{
			UserID:           1,
			DownloadPath:     "test",
			MaxDownloadSpeed: 100,
			MaxTasks:         10,
		}
		db.Create(settings)

		settings1 := &models.Settings{
			UserID:           1,
			DownloadPath:     "test1",
			MaxDownloadSpeed: 66.6,
			MaxTasks:         5,
		}
		req, err := json.Marshal(settings1)
		assert.Nil(t, err)
		w := c.Post("POST", "/api/settings/1", req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "保存成功")
	})

	t.Run("input err", func(t *testing.T) {
		req := []byte(`{"invalid_json":}`)
		w := c.Post("POST", "/api/settings/1", req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "输入参数有误")
	})

	t.Run("parse err", func(t *testing.T) {
		req := []byte(`{"invalid_json":}`)
		w := c.Post("POST", "/api/settings/err", req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestGetSettings(t *testing.T) {
	r, db := createTestHandlers()
	c := carrot.NewTestClient(r)
	t.Run("parse err", func(t *testing.T) {
		w := c.Get("/api/settings/err")
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("successfulGetSettings", func(t *testing.T) {
		settings := &models.Settings{
			UserID:           1,
			DownloadPath:     "/test",
			MaxDownloadSpeed: 2.2,
			MaxTasks:         10,
		}
		db.Create(settings)
		w := c.Get("/api/settings/1")
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("db get err", func(t *testing.T) {
		settings := &models.Settings{
			UserID:           1,
			DownloadPath:     "/test",
			MaxDownloadSpeed: 2.2,
			MaxTasks:         10,
		}
		db.Create(settings)
		w := c.Get("/api/settings/2")
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
