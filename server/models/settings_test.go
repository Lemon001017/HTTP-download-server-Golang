package models

import (
	"testing"

	"github.com/restsend/carrot"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func createTestDB() *gorm.DB {
	db, err := carrot.InitDatabase(nil, "", "")
	if err != nil {
		panic(err)
	}
	Migration(db)
	return db
}

func TestUpdateSettings(t *testing.T) {
	db := createTestDB()
	settings1 := Settings{
		DownloadPath:     "/test1",
		MaxTasks:         123,
		MaxDownloadSpeed: 123.456,
	}
	db.Create(&settings1)
	settings2 := Settings{
		DownloadPath:     "/test2",
		MaxTasks:         666,
		MaxDownloadSpeed: 999.666,
	}
	err := UpdateSettings(db, &settings2)
	assert.Nil(t, err)
}

func TestGetSettings(t *testing.T) {
	db := createTestDB()
	settings := Settings{
		DownloadPath:     "/test1",
		MaxTasks:         123,
		MaxDownloadSpeed: 123.456,
	}
	db.Create(&settings)
	s, err := GetSettings(db)
	assert.Nil(t, err)
	assert.Equal(t, settings.DownloadPath, s.DownloadPath)
	assert.Equal(t, settings.MaxTasks, s.MaxTasks)
	assert.Equal(t, settings.MaxDownloadSpeed, s.MaxDownloadSpeed)
}
