package handlers

import (
	"HTTP-download-server/server/models"
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/panjf2000/ants"
	"github.com/restsend/carrot"
	"golang.org/x/time/rate"
)

type DownloadRequest struct {
	URL string `json:"url" binding:"required"`
}

type DownloadProgress struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Progress      float64 `json:"progress"`
	Speed         float64 `json:"speed"`
	RemainingTime float64 `json:"remainingTime"`
	Status        string  `json:"status"`
}

// submit task
func (h *Handlers) handleSubmit(c *gin.Context) {
	var request DownloadRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		carrot.AbortWithJSONError(c, http.StatusBadRequest, err)
		return
	}

	eventSource := h.createEventSource()

	task, err := h.initOneTask(request.URL, eventSource.key)
	if err != nil {
		carrot.AbortWithJSONError(c, http.StatusInternalServerError, err)
		return
	}

	// Open a goroutine to handle the download separately
	go func() {
		h.processDownload(eventSource, task, 0)
	}()
	c.JSON(http.StatusOK, EventSourceResult{Key: eventSource.key})
}

func (h *Handlers) processDownload(es *EventSource, task *models.Task, lastTotalDownloaded int64) {
	startTime := time.Now()
	carrot.Info("id:", task.ID, "fileSize:", task.Size, "savePath:", task.SavePath, "chunkSize:", task.ChunkSize, "numChunks:", task.ChunkNum)

	outputFile, err := os.OpenFile(task.SavePath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		carrot.Error("open or create file error", "key:", es.key, "id:", task.ID, "url:", task.Url, "err:", err)
		return
	}

	_, maxDownloadSpeed, _, err := h.getSettingsInfo()
	if err != nil {
		carrot.Error("get settings error", "key:", es.key, "id:", task.ID, "url:", task.Url, "err:", err)
		return
	}

	// Init chunk info
	if task.TotalDownloaded == 0 {
		for i := 0; i < int(task.ChunkNum); i++ {
			start := int64(i) * task.ChunkSize
			end := math.Min(float64(start+task.ChunkSize), float64(task.Size)) - 1
			task.Chunk[i] = models.Chunk{
				TaskID: task.ID,
				Index:  i,
				Start:  int(start),
				End:    int(end),
			}
			models.AddChunk(h.db, &task.Chunk[i])
		}
	}

	task.Status = models.TaskStatusDownloading
	models.UpdateTask(h.db, task)

	// Create a pool of goroutines
	pool, _ := ants.NewPoolWithFunc(int(task.Threads), func(i interface{}) {
		err := h.downloadChunk(es, &task.Chunk[i.(int)], task, outputFile, startTime, maxDownloadSpeed, lastTotalDownloaded)
		if err != nil {
			// Clean all resources
			outputFile.Close()
			h.cleanEventSource(task.ID)

			if !errors.Is(err, context.Canceled) && !errors.Is(err, os.ErrClosed) {
				carrot.Error("download chunk failed", "key:", es.key, "url:", task.Url, "err:", err)
				task.Status = models.TaskStatusFailed
				models.UpdateTask(h.db, task)
				es.Emit(DownloadProgress{ID: task.ID, Name: task.Name, Status: models.TaskStatusFailed})
			}
			return
		}
	})
	defer pool.Release()

	for i := 0; i < int(task.ChunkNum); i++ {
		// Skip completed chunks
		if !task.Chunk[i].Done {
			_ = pool.Invoke(i)
		}
	}
}

func (h *Handlers) downloadChunk(es *EventSource, chunk *models.Chunk, task *models.Task,
	outputFile *os.File, startTime time.Time, maxDownloadSpeed float64, lastTotalDownloaded int64) error {
	req, err := http.NewRequest(http.MethodGet, task.Url, nil)
	if err != nil {
		carrot.Error("Failed to create HTTP request", "key:", es.key, "url:", task.Url)
		return err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/127.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Range", fmt.Sprintf("bytes=%v-%v", chunk.Start, chunk.End))
	req = req.WithContext(es.ctx)

	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusPartialContent {
		carrot.Error("key:", es.key, "url:", task.Url, "status:", resp.StatusCode)
		return err
	}

	buf := make([]byte, 2048)

	h.mu.Lock()

	_, err = outputFile.Seek(int64(chunk.Start), 0)
	if err != nil {
		h.mu.Unlock()
		return err
	}

	// Create a rate limiter
	maxDownloadSpeedInBytes := maxDownloadSpeed * 1000 * 1000
	limiter := rate.NewLimiter(rate.Limit(maxDownloadSpeedInBytes), int(maxDownloadSpeedInBytes))

	for {
		n, err := resp.Body.Read(buf)
		if err != nil && err != io.EOF {
			h.mu.Unlock()
			return err
		}

		if n == 0 {
			break
		}

		err = limiter.WaitN(es.ctx, n)
		if err != nil {
			h.mu.Unlock()
			return err
		}

		_, err = outputFile.Write(buf[:n])
		if err != nil {
			h.mu.Unlock()
			return err
		}

		task.TotalDownloaded += int64(n)

		speed, progress, remainingTime := h.calculateDownloadData(task, startTime, lastTotalDownloaded)
		es.Emit(DownloadProgress{
			ID:            task.ID,
			Name:          task.Name,
			Progress:      progress,
			Speed:         speed,
			RemainingTime: remainingTime,
			Status:        task.Status,
		})
	}

	chunk.Done = true
	models.UpdateChunk(h.db, chunk)
	// download complete
	if task.TotalDownloaded == task.Size {
		task.Status = models.TaskStatusDownloaded
		models.UpdateTask(h.db, task)

		es.Emit(DownloadProgress{
			ID:     task.ID,
			Name:   task.Name,
			Status: task.Status,
		})

		carrot.Info("Download complete", "key:", es.key, "id:", task.ID, "url:", task.Url)

		models.DeleteChunks(h.db, task.ID)
		outputFile.Close()
		close(es.eventChan)
	}

	models.UpdateTask(h.db, task)

	h.mu.Unlock()
	return nil
}

func (h *Handlers) initOneTask(url, key string) (*models.Task, error) {
	outputDir, _, _, err := h.getSettingsInfo()
	if err != nil {
		return nil, err
	}

	fileSize, outputPath, fileName, err := h.getFileInfo(url, outputDir)
	if err != nil {
		return nil, err
	}

	chunkSize, numChunks := h.getChunkInfo(fileSize)

	task := models.Task{
		ID:        key,
		Name:      fileName,
		Url:       url,
		Size:      fileSize,
		SavePath:  outputPath,
		FileType:  filepath.Ext(fileName),
		Threads:   models.DefaultThreads,
		Status:    models.TaskStatusPending,
		ChunkNum:  numChunks,
		ChunkSize: chunkSize,
		Chunk:     make([]models.Chunk, numChunks),
	}

	err = models.AddTask(h.db, &task)
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// Pause download
func (h *Handlers) handlePause(c *gin.Context) {
	var ids []string
	err := c.ShouldBindJSON(&ids)
	if err != nil {
		carrot.AbortWithJSONError(c, http.StatusBadRequest, err)
		return
	}

	tasks, err := models.GetTaskByIds(h.db, ids)
	if err != nil {
		carrot.AbortWithJSONError(c, http.StatusInternalServerError, err)
		return
	}

	for _, task := range tasks {
		if task.Status == models.TaskStatusDownloading {
			h.cleanEventSource(task.ID)
			task.Status = models.TaskStatusCanceled
			models.UpdateTask(h.db, &task)
		} else {
			carrot.AbortWithJSONError(c, http.StatusBadRequest, models.ErrStatusNotDownloading)
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"ids": ids})
}

// Resume download
func (h *Handlers) handleResume(c *gin.Context) {
	var ids []string
	err := c.ShouldBindJSON(&ids)
	if err != nil {
		carrot.AbortWithJSONError(c, http.StatusBadRequest, err)
		return
	}

	tasks, err := models.GetTaskByIds(h.db, ids)
	if err != nil {
		carrot.AbortWithJSONError(c, http.StatusInternalServerError, err)
		return
	}

	for _, task := range tasks {
		if task.Status != models.TaskStatusCanceled {
			carrot.AbortWithJSONError(c, http.StatusBadRequest, models.ErrStatusNotCanceled)
			return
		}

		es := h.createEventSourceWithKey(task.ID)

		task.Status = models.TaskStatusPending
		task.Speed = 0
		task.Chunk = models.GetChunksByTaskId(h.db, task.ID)
		models.UpdateTask(h.db, &task)

		lastTotalDownloaded := task.TotalDownloaded

		go func() {
			h.processDownload(es, &task, lastTotalDownloaded)
		}()
	}
	c.JSON(http.StatusOK, gin.H{"ids": ids})
}

// Re download
func (h *Handlers) handleRestart(c *gin.Context) {
	var ids []string
	err := c.ShouldBindJSON(&ids)
	if err != nil {
		carrot.AbortWithJSONError(c, http.StatusBadRequest, err)
		return
	}

	tasks, err := models.GetTaskByIds(h.db, ids)
	if err != nil {
		carrot.AbortWithJSONError(c, http.StatusInternalServerError, err)
		return
	}

	for _, task := range tasks {
		if task.Status != models.TaskStatusDownloaded {
			carrot.AbortWithJSONError(c, http.StatusBadRequest, models.ErrStatusNotDownloaded)
			return
		}
		es := h.createEventSourceWithKey(task.ID)

		task.Status = models.TaskStatusPending
		task.TotalDownloaded = 0
		task.Progress = 0
		task.Speed = 0
		task.Chunk = make([]models.Chunk, task.ChunkNum)
		models.UpdateTask(h.db, &task)

		go func() {
			h.processDownload(es, &task, 0)
		}()
	}
	c.JSON(http.StatusOK, gin.H{"ids": ids})
}

func extractFileName(resp *http.Response, downloadURL string) string {
	if contentDisposition := resp.Header.Get("Content-Disposition"); contentDisposition != "" {
		_, params, err := mime.ParseMediaType(contentDisposition)
		if err == nil && params["filename"] != "" {
			return params["filename"]
		}
	}

	parsedURL, err := url.Parse(downloadURL)
	if err != nil {
		parsedURL.Path = "/unknown"
	}

	re := regexp.MustCompile(`[^\/]+\.[a-zA-Z0-9]+$`)
	fileName := re.FindString(parsedURL.Path)
	if fileName == "" {
		fileName = "unknown_file"
	}
	return fileName
}

func (h *Handlers) getSettingsInfo() (string, float64, uint, error) {
	settings, err := models.GetSettings(h.db, 1)
	if err != nil {
		return "", 0, 0, err
	}

	outputDir := settings.DownloadPath
	if outputDir == "" {
		outputDir, err = os.Getwd()
		if err != nil {
			return "", 0, 0, err
		}
	}

	maxDownloadSpeed := settings.MaxDownloadSpeed
	if maxDownloadSpeed == 0 {
		maxDownloadSpeed = 1e9
	}

	maxTasks := settings.MaxTasks
	if maxTasks == 0 {
		maxTasks = 4
	}
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
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, "", "", err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.159 Safari/537.36")
	req.Header.Set("Accept", "*/*")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, "", "", err
	}

	fileSize := resp.ContentLength
	fileName := extractFileName(resp, url)
	outputPath := filepath.Join(outputDir, fileName)

	return fileSize, outputPath, fileName, nil
}

func (h *Handlers) calculateDownloadData(task *models.Task, startTime time.Time, lastTotalDownloaded int64) (float64, float64, float64) {
	elapsedTime := time.Since(startTime).Seconds()
	speed := math.Round((float64(task.TotalDownloaded-lastTotalDownloaded)/elapsedTime/1024/1024)*10) / 10 // MB/s
	progress := math.Round((float64(task.TotalDownloaded)/float64(task.Size)*100)*10) / 10
	remainingTime := math.Round((float64((task.Size-task.TotalDownloaded)/1024/1024)/speed)*10) / 10

	task.Progress = progress
	task.Speed = speed
	task.RemainingTime = remainingTime
	return speed, progress, remainingTime
}
