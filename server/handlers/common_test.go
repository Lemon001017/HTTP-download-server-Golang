package handlers

import (
	"HTTP-download-server/server/models"

	"github.com/gin-gonic/gin"
	"github.com/restsend/carrot"
	"gorm.io/gorm"
)

func createTestHandlers() (*gin.Engine, *gorm.DB) {
	db, err := carrot.InitDatabase(nil, "", "")
	if err != nil {
		panic(err)
	}
	gin.SetMode(gin.ReleaseMode)
	m := NewHandlers(db)
	r := gin.Default()
	carrot.InitCarrot(db, r)
	models.Migration(db)
	m.Register(r)
	return r, db
}
