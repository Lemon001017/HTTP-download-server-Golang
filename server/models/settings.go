package models

import (
	"time"

	"gorm.io/gorm"
)

type Settings struct {
	UserID           uint      `json:"userId" gorm:"primaryKey" comment:"用户ID"`
	DownloadPath     string    `json:"downloadPath" gorm:"size:200" comment:"下载路径"`
	MaxTasks         uint      `json:"maxTasks,string" gorm:"size:20" comment:"最大任务数"`
	MaxDownloadSpeed float64   `json:"maxDownloadSpeed,string" gorm:"size:20" comment:"最大下载速度"`
	CreatedAt        time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt        time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

func UpdateSettings(db *gorm.DB, settings *Settings, userId uint) error {
	res := db.Model(&Settings{}).Where("user_id = ?", userId).Save(map[string]interface{}{
		"user_id":            userId,
		"download_path":      settings.DownloadPath,
		"max_tasks":          settings.MaxTasks,
		"max_download_speed": settings.MaxDownloadSpeed,
	})
	return res.Error
}

func GetSettings(db *gorm.DB, userId uint) (*Settings, error) {
	var settings Settings
	err := db.Model(&Settings{}).Where("user_id = ?", userId).First(&settings).Error
	return &settings, err
}
