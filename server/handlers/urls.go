package handlers

import (
	"sync"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handlers struct {
	db *gorm.DB
	eventSources sync.Map
}

func NewHandlers(db *gorm.DB) *Handlers {
	return &Handlers{
		db: db,
	}
}

func (h *Handlers) Register(engine *gin.Engine) {
	r := engine.Group("/api")
	r.POST("/settings/:userId", h.handlerSaveSettings)
	r.GET("/settings/:userId", h.handlerGetSettings)
	r.POST("/task/submit", h.handlerSubmit)
	r.GET("/event/:key", h.handlerSSE)
}
