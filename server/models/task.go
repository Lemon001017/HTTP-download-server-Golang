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
	TaskID string `json:"id" gorm:"index" comment:"任务ID"`
	Index  int    `json:"index" gorm:"size:200" comment:"分片索引"`
	Start  int    `json:"start" gorm:"size:200" comment:"分片开始位置"`
	End    int    `json:"end" gorm:"size:200" comment:"分片结束位置"`
	Done   bool   `json:"done" gorm:"size:20" comment:"是否下载完成"`
}

// Insert a task
func AddTask(db *gorm.DB, task *Task) error {
	return db.Create(task).Error
}

// Update a task
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

// Get one task by ids
func GetTaskByIds(db *gorm.DB, ids []string) ([]Task, error) {
	var tasks []Task
	err := db.Model(&Task{}).Where("id IN (?)", ids).Find(&tasks).Error
	return tasks, err
}

// Get the task list according to the task status
func GetTasksByStatus(db *gorm.DB, status string) []Task {
	var tasks []Task
	if status == "" || status == "all" {
		db.Find(&tasks)
	} else {
		db.Where("status = ?", status).Find(&tasks)
	}
	return tasks
}

// Remove tasks according to ids
func DeleteTasksByIds(db *gorm.DB, ids []string) error {
	return db.Where("id IN (?)", ids).Delete(&Task{}).Error
}

// Insert a chunk
func AddChunk(db *gorm.DB, chunk *Chunk) error {
	return db.Create(chunk).Error
}

// Update a chunk
func UpdateChunk(db *gorm.DB, chunk *Chunk) error {
	return db.Model(chunk).Where("task_id = ? AND `index` = ?", chunk.TaskID, chunk.Index).Updates(map[string]interface{}{
		"done": chunk.Done,
	}).Error
}

// Delete all chunks by task id
func DeleteChunks(db *gorm.DB, taskId string) error {
	return db.Where("task_id = ?", taskId).Delete(&Chunk{}).Error
}

// Get all chunks by task id
func GetChunksByTaskId(db *gorm.DB, taskId string) ([]Chunk, error) {
	var chunks []Chunk
	err := db.Where("task_id = ?", taskId).Find(&chunks).Error
	return chunks, err
}
