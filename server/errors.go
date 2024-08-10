package httpDownloadServer

import "fmt"

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e Error) StatusCode() int {
	return e.Code
}

func (e Error) Error() string {
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}
