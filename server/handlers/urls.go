package handlers

import (
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/restsend/carrot/apidocs"
	"gorm.io/gorm"
)

type Handlers struct {
	db           *gorm.DB
	eventSources sync.Map
	wg           sync.WaitGroup
	mu           sync.Mutex
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
	apidocs.RegisterHandler(engine.Group("/api/docs"), h.GetDocs(), nil)
}
