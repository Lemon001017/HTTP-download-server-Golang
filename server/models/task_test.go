package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInsertTask(t *testing.T) {
	db := createTestDB()
	task := &Task{
		ID:            "123",
		Name:          "test",
		Url:           "http://test.com",
		SavePath:      "/tmp/test",
		Threads:       1,
		Status:        TaskStatusPending,
		Size:          1024,
		Speed:         6.6,
		Progress:      0.5,
		RemainingTime: 10,
		FileType:      ".zip",
	}
	err := AddTask(db, task)
	assert.Nil(t, err)
}

func TestUpdateTask(t *testing.T) {
	db := createTestDB()
	task := &Task{
		ID:            "123",
		Name:          "test",
		Url:           "http://test.com",
		SavePath:      "/tmp/test",
		Threads:       1,
		Status:        TaskStatusPending,
		Size:          1024,
		Speed:         6.6,
		Progress:      0.5,
		RemainingTime: 10,
		FileType:      ".zip",
	}
	err := AddTask(db, task)
	assert.Nil(t, err)

	task.Status = TaskStatusDownloading
	err = UpdateTask(db, task)
	assert.Nil(t, err)
}

func TestGetTaskById(t *testing.T) {
	db := createTestDB()
	task := &Task{
		ID:            "123",
		Name:          "test",
		Url:           "http://test.com",
		SavePath:      "/tmp/test",
		Threads:       1,
		Status:        TaskStatusPending,
		Size:          1024,
		Speed:         6.6,
		Progress:      0.5,
		RemainingTime: 10,
		FileType:      ".zip",
	}
	db.Create(task)
	result, err := GetTaskById(db, "123")
	assert.Nil(t, err)
	assert.Equal(t, result.ID, "123")
	assert.Equal(t, result.Name, "test")
	assert.Equal(t, result.Url, "http://test.com")

	_, err = GetTaskById(db, "456")
	assert.NotNil(t, err)
}
