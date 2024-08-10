package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handlers) handlerSaveSettings(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}
