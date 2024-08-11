package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// submit task
func (h *Handlers) handlerSubmit(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}
