package models

import "time"

type Task struct {
	ID            uint      `json:"id" gorm:"primaryKey" comment:"任务ID"`
	FileType      string    `json:"type" gorm:"size:20" comment:"文件类型"`
	Url           string    `json:"url" gorm:"size:200" comment:"下载地址"`
	Name          string    `json:"name" gorm:"size:200" comment:"文件名"`
	Threads       uint      `json:"threads" gorm:"size:20" comment:"线程数"`
	Size          float64   `json:"size" gorm:"size:20" comment:"文件大小"`
	Status        string    `json:"status" gorm:"size:20" comment:"任务状态"`
	Speed         float64   `json:"speed" gorm:"size:20" comment:"下载速度"`
	Progress      float64   `json:"progress" gorm:"size:20" comment:"下载进度"`
	RemainingTime float64   `json:"remainingTime" gorm:"size:20" comment:"剩余时间"`
	CreatedAt     time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt     time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}
