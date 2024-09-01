package models

import (
	httpDownloadServer "HTTP-download-server/server"
	"net/http"
	"time"

	"gorm.io/gorm"
)

const (
	MinChunkSize = 32 * 1024
	MidChunkSize = 1024 * 1024
	MaxChunkSize = 10 * 1024 * 1024
)

const (
	TaskStatusDownloading = "downloading"
	TaskStatusDownloaded  = "downloaded"
	TaskStatusPending     = "pending"
	TaskStatusFailed      = "failed"
	TaskStatusCanceled    = "canceled"
)

const (
	TaskFilterAll        = "all"
	TaskFilterPending    = "pending"
	TaskFilterFailed     = "failed"
	TaskFilterCanceled   = "canceled"
	TaskFilterDownloaded = "downloaded"
)

const DefaultThreads = 4
const MessageInterval = 200 * time.Millisecond

var ErrInputParam = httpDownloadServer.Error{Code: http.StatusBadRequest, Message: "Input is invalid"}
var ErrSaveFailed = httpDownloadServer.Error{Code: http.StatusInternalServerError, Message: "Save settigns failed"}
var ErrGetSettings = httpDownloadServer.Error{Code: http.StatusInternalServerError, Message: "Get settings failed"}

var ErrInputUrl = httpDownloadServer.Error{Code: http.StatusBadRequest, Message: "Input url is invalid"}
var ErrIncompleteFile = httpDownloadServer.Error{Code: http.StatusInternalServerError, Message: "File is incomplete"}
var ErrStatusNotDownloading = httpDownloadServer.Error{Code: http.StatusBadRequest, Message: "Task is not downloading"}
var ErrStatusNotDownloaded = httpDownloadServer.Error{Code: http.StatusBadRequest, Message: "Task is not downloaded"}
var ErrStatusNotCanceled = httpDownloadServer.Error{Code: http.StatusBadRequest, Message: "Task is not canceled"}
var ErrExpectedFileSize = httpDownloadServer.Error{Code: http.StatusBadRequest, Message: "File size is not expected"}

func Migration(db *gorm.DB) error {
	return db.AutoMigrate(
		&Settings{},
		&Task{},
		&Chunk{},
	)
}
