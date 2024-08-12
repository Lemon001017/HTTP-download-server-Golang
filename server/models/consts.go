package models

import (
	httpDownloadServer "HTTP-download-server/server"
	"net/http"

	"gorm.io/gorm"
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

func Migration(db *gorm.DB) error {
	return db.AutoMigrate(
		&Settings{},
		&Task{},
	)
}
