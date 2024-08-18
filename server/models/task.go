package models

import (
	"time"

	"gorm.io/gorm"
)

type Task struct {
	ID              string    `json:"id" gorm:"primaryKey" comment:"任务ID"`
	Name            string    `json:"name" gorm:"size:200" comment:"文件名"`
	FileType        string    `json:"type" gorm:"size:20" comment:"文件类型"`
	Size            int64     `json:"size" gorm:"size:20" comment:"文件大小"`
	TotalDownloaded int64     `json:"totalDownloaded" gorm:"size:200" comment:"已下载字节数"`
	Url             string    `json:"url" gorm:"size:200" comment:"下载地址"`
	SavePath        string    `json:"savePath" gorm:"size:200" comment:"保存路径"`
	Status          string    `json:"status" gorm:"size:20" comment:"任务状态"`
	Md5             string    `json:"md5" gorm:"size:200" comment:"校验码"`
	Threads         uint      `json:"threads" gorm:"size:20" comment:"线程数"`
	Speed           float64   `json:"speed" gorm:"size:20" comment:"下载速度"`
	Progress        float64   `json:"progress" gorm:"size:20" comment:"下载进度"`
	RemainingTime   float64   `json:"remainingTime" gorm:"size:20" comment:"剩余时间"`
	CreatedAt       time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt       time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
	ChunkNum        int64     `json:"chunkNum" gorm:"size:200" comment:"分片数量"`
	ChunkSize       int64     `json:"chunkSize" gorm:"size:200" comment:"分片大小"`
	Chunk           []Chunk   `json:"doneChunk" gorm:"-" comment:"已完成分片"`
}

type Chunk struct {
	Url   string
	Index int
	Start int
	End   int
	Done  bool // is finished
}

func AddTask(db *gorm.DB, task *Task) error {
	return db.Create(task).Error
}

func UpdateTask(db *gorm.DB, task *Task) error {
	return db.Model(task).Where("id = ?", task.ID).Updates(map[string]interface{}{
		"status":           task.Status,
		"progress":         task.Progress,
		"remaining_time":   task.RemainingTime,
		"total_downloaded": task.TotalDownloaded,
		"speed":            task.Speed,
		"updated_at":       time.Now(),
	}).Error
}

func GetTaskById(db *gorm.DB, id string) (*Task, error) {
	var task Task
	err := db.Where("id = ?", id).First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}
