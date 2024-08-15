package models

import (
	httpDownloadServer "HTTP-download-server/server"
	"net/http"

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
	TaskOptionStart   = "start"
	TaskOptionPause   = "pause"
	TaskOptionDelete  = "delete"
	TaskOptionRefresh = "refresh"
	TaskOptionCancel  = "cancel"
)

var ErrInputParam = httpDownloadServer.Error{Code: http.StatusBadRequest, Message: "输入参数有误"}
var ErrSaveFailed = httpDownloadServer.Error{Code: http.StatusInternalServerError, Message: "保存失败"}
var ErrGetSettings = httpDownloadServer.Error{Code: http.StatusInternalServerError, Message: "获取设置失败"}

var ErrInputUrl = httpDownloadServer.Error{Code: http.StatusBadRequest, Message: "输入地址有误"}
var ErrExceedMaxRetries = httpDownloadServer.Error{Code: http.StatusInternalServerError, Message: "下载失败，已达到最大重试次数"}

func Migration(db *gorm.DB) error {
	return db.AutoMigrate(
		&Settings{},
		&Task{},
	)
}
