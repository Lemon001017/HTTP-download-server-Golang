package handlers

import (
	"HTTP-download-server/server/models"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/restsend/carrot"
)

type EventSource struct {
	lastTime  time.Time
	key       string
	eventChan chan any
	ctx       context.Context
	cancel    context.CancelFunc
}

type EventSourceResult struct {
	Key string `json:"key"`
}

type DownloadRequest struct {
	URL string `json:"url" binding:"required"`
}

type DownloadProgress struct {
	Progress float64 `json:"progress"`
	Speed    float64 `json:"speed"`
}

func (es *EventSource) Emit(event any) {
	if es.eventChan == nil {
		return
	}
	select {
	case es.eventChan <- event:
	default:
	}
}

// submit task
func (h *Handlers) handlerSubmit(c *gin.Context) {
	var request DownloadRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		carrot.AbortWithJSONError(c, http.StatusBadRequest, err)
		return
	}

	url := request.URL
	task, err := h.initTask(url)
	if err != nil {
		carrot.AbortWithJSONError(c, http.StatusInternalServerError, err)
	}

	eventSource := h.createEventSource()

	// Open a goroutine to handle the download separately
	go func() {
		h.processDownload(eventSource, task)
	}()
	c.JSON(http.StatusOK, EventSourceResult{Key: eventSource.key})
}

func (h *Handlers) processDownload(eventSource *EventSource, task *models.Task) {
	fileSize := int64(task.Size)
	chunkSize, numChunks := h.getChunkInfo(fileSize)
	carrot.Info("fileSize:", fileSize, "savePath:", task.SavePath, "chunkSize:", chunkSize, "numChunks:", numChunks)

	outputFile, err := os.Create(task.SavePath)
	if err != nil {
		carrot.Error("create tempFile error", "key:", eventSource.key, "url:", task.Url)
		return
	}
	defer outputFile.Close()

	// Set the size of the temporary file to be the same as the destination file
	if err := outputFile.Truncate(fileSize); err != nil {
		carrot.Error("set fileSize error", "key:", eventSource.key, "url:", task.Url)
		return
	}

	h.wg.Add(int(numChunks))

	for i := 0; i < int(numChunks); i++ {
		start := int64(i) * chunkSize
		end := start + chunkSize - 1
		if end >= fileSize {
			end = fileSize - 1
		}
		// carrot.Info("key:", eventSource.key, "start:", start, "end:", end)

		go func() {
			h.downloadChunk(i, start, end, outputFile, task.Url, eventSource)
		}()
	}
	h.wg.Wait()
	carrot.Info("Download complete", "key:", eventSource.key, "url:", task.Url)
}

func (h *Handlers) downloadChunk(i int, start, end int64, outputFile *os.File, url string, es *EventSource) {
	defer h.wg.Done()
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		carrot.Error("Failed to create HTTP request", "key:", es.key, "url:", url)
		return
	}

	// Sets the request header, specifying the range of bytes to download
	req.Header.Set("Range", fmt.Sprintf("bytes=%v-%v", start, end))
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.159 Safari/537.36")
	req.Header.Set("Accept", "*/*")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		carrot.Error("Failed to send HTTP request", "key:", es.key, "url:", url, "err:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusPartialContent {
		carrot.Error("Failed to download file", "key:", es.key, "url:", url, "status:", resp.StatusCode)
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if _, err := outputFile.Seek(start, 0); err != nil {
		carrot.Error("seek error", "key:", es.key, "url:", url, "err:", err)
		return
	}

	_, err = io.Copy(outputFile, resp.Body)
	if err != nil {
		carrot.Error("Failed to copy HTTP response body", "key:", es.key, "url:", url, "err:", err)
		return
	}
	carrot.Info("the", i, "part of the file has been downloaded")
}

func (h *Handlers) handlerSSE(c *gin.Context) {
	key := c.Param("key")
	v, ok := h.eventSources.LoadAndDelete(key)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event source not found"})
		return
	}
	defer h.cleanEventSource(key)
	eventSource := v.(*EventSource)

	c.Stream(func(w io.Writer) bool {
		select {
		case <-c.Request.Context().Done():
			c.SSEvent("close", "user cancel download")
			return false
		case data, ok := <-eventSource.eventChan:
			if data == nil || !ok {
				c.SSEvent("close", "download complete")
				return false
			}
			eventSource.lastTime = time.Now()
			byteData, _ := json.Marshal(data)
			c.SSEvent("message", string(byteData))
			return true
		}
	})
}

func (h *Handlers) createEventSource() *EventSource {
	ctx, cancel := context.WithCancel(context.Background())
	key := carrot.RandText(8)

	eventSource := &EventSource{
		lastTime:  time.Now(),
		cancel:    cancel,
		ctx:       ctx,
		eventChan: make(chan any, 100),
		key:       key,
	}
	h.eventSources.Store(key, eventSource)

	go func() {
		defer h.cleanEventSource(key)
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(1 * time.Minute):
				if time.Since(eventSource.lastTime) > 10*time.Minute {
					return
				}
			}
		}
	}()
	return eventSource
}

func (h *Handlers) cleanEventSource(key string) {
	v, ok := h.eventSources.LoadAndDelete(key)
	if !ok {
		return
	}

	eventSource, ok := v.(*EventSource)
	if !ok {
		return
	}
	eventSource.cancel()
	if eventSource.eventChan != nil {
		close(eventSource.eventChan)
		eventSource.eventChan = nil
	}
}

func (h *Handlers) getSettingsInfo() (string, float64, uint, error) {
	settings, err := models.GetSettings(h.db, 1)
	if err != nil {
		return "", 0, 0, err
	}

	outputDir := settings.DownloadPath
	maxDownloadSpeed := settings.MaxDownloadSpeed
	maxTasks := settings.MaxTasks
	return outputDir, maxDownloadSpeed, maxTasks, nil
}

func (h *Handlers) getChunkInfo(fileSize int64) (int64, int64) {
	var chunkSize int64
	switch {
	case fileSize <= 10*1024*1024:
		chunkSize = models.MinChunkSize
	case fileSize <= 100*1024*1024:
		chunkSize = models.MidChunkSize
	default:
		chunkSize = models.MaxChunkSize
	}
	numChunks := (fileSize + chunkSize - 1) / chunkSize
	return chunkSize, numChunks
}

func (h *Handlers) getFileInfo(url string, outputDir string) (int64, string, string, error) {
	resp, err := http.Head(url)
	if err != nil {
		return 0, "", "", err
	}
	defer resp.Body.Close()

	resp.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.159 Safari/537.36")
	resp.Header.Set("Accept", "*/*")

	if resp.StatusCode != http.StatusOK {
		return 0, "", "", err
	}

	fileSize := resp.ContentLength
	segments := strings.Split(url, "/")

	fileName := segments[len(segments)-1]
	outputPath := filepath.Join(outputDir, fileName)

	return fileSize, outputPath, fileName, nil
}

func (h *Handlers) initTask(url string) (*models.Task, error) {
	outputDir, _, _, err := h.getSettingsInfo()
	if err != nil {
		return nil, err
	}

	fileSize, outputPath, fileName, err := h.getFileInfo(url, outputDir)
	if err != nil {
		return nil, err
	}
	task := models.Task{
		ID:       carrot.RandText(4),
		Name:     fileName,
		Url:      url,
		Size:     float64(fileSize),
		SavePath: outputPath,
		FileType: filepath.Ext(fileName),
		Threads:  4,
	}
	err = models.AddTask(h.db, &task)
	if err != nil {
		return nil, err
	}

	return &task, nil
}
