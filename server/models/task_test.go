package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInsertTask(t *testing.T) {
	db := createTestDB()
	task := &Task{
		ID:       "123",
		Name:     "test",
		Url:      "http://test.com",
		SavePath: "/tmp/test",
		Threads:  1,
		Status:   TaskStatusPending,
		Size:     1024,
		Speed:    6.6,
		Progress: 0.5,
		RemainingTime: 10,
		FileType: ".zip",
	}
	err := AddTask(db, task)
	assert.Nil(t, err)
}
