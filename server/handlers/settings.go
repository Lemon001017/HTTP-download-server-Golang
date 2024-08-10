package handlers

import (
	"HTTP-download-server/server/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handlers) handlerSaveSettings(c *gin.Context) {
	var settings models.Settings
	err := c.ShouldBindJSON(&settings)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrInputParam)
		return
	}
	
	err = models.UpdateSettings(h.db, &settings)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrSaveFailed)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "保存成功"})
}

func (h *Handlers) handlerGetSettings(c *gin.Context) {
	settings, err := models.GetSettings(h.db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrGetSettings)
		return
	}
	c.JSON(http.StatusOK, settings)
}
