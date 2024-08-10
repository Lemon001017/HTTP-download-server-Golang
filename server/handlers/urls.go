package handlers

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handlers struct {
	db *gorm.DB
}

func NewHandlers(db *gorm.DB) *Handlers {
	return &Handlers{
		db: db,
	}
}

func (h *Handlers) Register(engine *gin.Engine) {
	r := engine.Group("/api")
	r.POST("/settings", h.handlerSaveSettings)
	r.GET("/settings", h.handlerGetSettings)
}