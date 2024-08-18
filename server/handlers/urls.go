package handlers

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/restsend/carrot/apidocs"
	"gorm.io/gorm"
)

type Handlers struct {
	db           *gorm.DB
	eventSources sync.Map
	wg           sync.WaitGroup
	mu           sync.Mutex
	client       *http.Client
}

func NewHandlers(db *gorm.DB) *Handlers {
	return &Handlers{
		db: db,
		client: &http.Client{
			Timeout: 60 * time.Minute,
		},
	}
}

func (h *Handlers) Register(engine *gin.Engine) {
	r := engine.Group("/api")
	r.POST("/settings/:userId", h.handleSaveSettings)
	r.GET("/settings/:userId", h.handleGetSettings)
	r.POST("/task/submit", h.handleSubmit)
	r.POST("/task/pause/:key", h.handlePause)
	r.GET("/event/:key", h.handleSSE)
	apidocs.RegisterHandler(engine.Group("/api/docs"), h.GetDocs(), nil)
}
