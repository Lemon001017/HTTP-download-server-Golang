package handlers

import (
	"HTTP-download-server/server/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/restsend/carrot"
)

func (h *Handlers) handlerSaveSettings(c *gin.Context) {
	userId, err := strconv.ParseUint(c.Param("userId"), 10, 64)
	if err != nil {
		carrot.AbortWithJSONError(c, http.StatusBadRequest, err)
		return
	}

	var settings models.Settings
	err = c.ShouldBindJSON(&settings)
	if err != nil {
		carrot.AbortWithJSONError(c, http.StatusBadRequest, models.ErrInputParam)
		return
	}

	err = models.UpdateSettings(h.db, &settings, uint(userId))
	if err != nil {
		carrot.AbortWithJSONError(c, http.StatusInternalServerError, models.ErrSaveFailed)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "保存成功"})
}

func (h *Handlers) handlerGetSettings(c *gin.Context) {
	userId, err := strconv.ParseUint(c.Param("userId"), 10, 64)
	if err != nil {
		carrot.AbortWithJSONError(c, http.StatusBadRequest, err)
		return
	}

	settings, err := models.GetSettings(h.db, uint(userId))
	if err != nil {
		carrot.AbortWithJSONError(c, http.StatusInternalServerError, models.ErrGetSettings)
		return
	}

	c.JSON(http.StatusOK, settings)
}
