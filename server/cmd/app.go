package main

import (
	"HTTP-download-server/server/handlers"

	"github.com/gin-gonic/gin"
	"github.com/restsend/carrot"
	"gorm.io/gorm"
)

type downloadServer struct {
	db       *gorm.DB
	handlers *handlers.Handlers
}

func NewDownloadServer(db *gorm.DB) *downloadServer {
	return &downloadServer{
		db:       db,
		handlers: handlers.NewHandlers(db),
	}
}

func (m *downloadServer) Prepare(engine *gin.Engine) error {
	err := carrot.InitCarrot(m.db, engine)
	if err != nil {
		return err
	}
	m.handlers.Register(engine)
	return nil
}
