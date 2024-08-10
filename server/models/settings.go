package models

import "gorm.io/gorm"

type Settings struct {
	DownloadPath     string  `json:"downloadPath" gorm:"size:200" comment:"下载路径"`
	MaxTasks         uint    `json:"maxTasks" gorm:"size:20" comment:"最大任务数"`
	MaxDownloadSpeed float64 `json:"maxDownloadSpeed" gorm:"size:20" comment:"最大下载速度"`
}

func UpdateSettings(db *gorm.DB, settings *Settings) error {
	return db.Model(&Settings{}).Where("download_path = ?", settings.DownloadPath).Updates(map[string]interface{}{
		"max_tasks":          settings.MaxTasks,
		"max_download_speed": settings.MaxDownloadSpeed,
	}).Error
}

func GetSettings(db *gorm.DB) (*Settings, error) {
	var settings Settings
	err := db.Model(&Settings{}).First(&settings).Error
	return &settings, err
}
