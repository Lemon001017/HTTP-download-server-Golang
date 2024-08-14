package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type NotifyingResponseRecorder struct {
	*httptest.ResponseRecorder
	closeNotify chan bool
}

func NewNotifyingResponseRecorder() *NotifyingResponseRecorder {
	return &NotifyingResponseRecorder{
		httptest.NewRecorder(),
		make(chan bool),
	}
}

func (nrr *NotifyingResponseRecorder) CloseNotify() <-chan bool {
	return nrr.closeNotify
}

func (nrr *NotifyingResponseRecorder) CloseConnection() {
	close(nrr.closeNotify)
}

func TestHandleSSE(t *testing.T) {
	_, db := createTestHandlers()
	h := Handlers{
		db:           db,
		eventSources: sync.Map{},
	}

	eventSource := h.createEventSource()

	// 创建gin.Context
	w := NewNotifyingResponseRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/event/"+eventSource.key, nil)

	// 设置key
	c.Params = gin.Params{
		gin.Param{
			Key:   "key",
			Value: eventSource.key,
		},
	}

	// test case: data不为空
	{
		eventSource.eventChan <- "mock message1"
		eventSource.eventChan <- "mock message2"

		go func() {
			time.Sleep(time.Millisecond * 100)
			close(eventSource.eventChan)
		}()

		h.handleSSE(c)
		assert.Equal(t, http.StatusOK, w.Code)
		// 检查header
		assert.Equal(t, "no", w.Header().Get("X-Accel-Buffering"))
	}

	// test case: key no found
	{
		w = NewNotifyingResponseRecorder()
		c, _ = gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/event/"+eventSource.key, nil)
		c.Params = gin.Params{
			gin.Param{
				Key:   "key",
				Value: eventSource.key,
			},
		}
		h.handleSSE(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	}

	// test case: 客户端取消请求
	{
		mockEventSource := h.createEventSource()
		// 创建带取消功能的请求上下文
		reqCtx, reqCancel := context.WithCancel(c.Request.Context())
		c.Request = httptest.NewRequest("GET", "/api/event/"+mockEventSource.key, nil).WithContext(reqCtx)
		c.Params = gin.Params{
			gin.Param{
				Key:   "key",
				Value: mockEventSource.key,
			},
		}
		// 启动处理SSE的goroutine
		go func() {
			h.handleSSE(c)
		}()

		// 模拟客户端在一段时间后取消请求
		time.Sleep(time.Millisecond * 200)
		reqCancel()

		// 检查请求是否被正确取消
		assert.Eventually(t, func() bool {
			return w.Body.String() != ""
		}, time.Second, time.Millisecond*10)

		assert.Contains(t, w.Body.String(), "close", "user actively cancels and disconnects the SSE")
	}

}
