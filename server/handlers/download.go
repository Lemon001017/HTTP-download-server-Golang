package handlers

import (
	"HTTP-download-server/server/models"
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

	url := request.URL
	task, err := h.initTask(url, eventSource.key)
	if err != nil {
		carrot.AbortWithJSONError(c, http.StatusInternalServerError, err)
		return
	}

	// Open a goroutine to handle the download separately
	go func() {
		h.processDownload(task, eventSource)
	}()
	c.JSON(http.StatusOK, EventSourceResult{Key: eventSource.key})
}

func (h *Handlers) processDownload(task *models.Task, es *EventSource) {
	fileSize := int64(task.Size)
	chunkSize := task.ChunkSize
	numChunks := task.ChunkNum
	startTime := time.Now()
	carrot.Info("fileSize:", fileSize, "savePath:", task.SavePath, "chunkSize:", chunkSize, "numChunks:", numChunks)

	outputFile, err := os.Create(task.SavePath)
	if err != nil {
		carrot.Error("create tempFile error", "key:", es.key, "id:", task.ID, "url:", task.Url, "err:", err)
		return
	}
	defer outputFile.Close()

	for i := 0; i < int(numChunks); i++ {
		start := int64(i) * chunkSize
		end := math.Min(float64(start+chunkSize), float64(fileSize)) - 1
		task.Chunk[i] = models.Chunk{
			Index:    i,
			Url:      task.Url,
			Start:    int(start),
			End:      int(end),
			Done:     false,
		}
	}

	// Create a pool of goroutines
	pool, _ := ants.NewPoolWithFunc(int(task.Threads), func(i interface{}) {
		h.downloadChunk(&task.Chunk[i.(int)], outputFile, es, startTime, task)
	})
	defer pool.Release()

	for i := 0; i < int(numChunks); i++ {
		h.wg.Add(1)
		_ = pool.Invoke(i)
	}
	h.wg.Wait()

	if task.TotalDownloaded == fileSize {
		task.Status = models.TaskStatusDownloaded
		carrot.Info("Download complete", "key:", es.key, "id:", task.ID, "url:", task.Url)
	} else {
		task.Status = models.TaskStatusFailed
		carrot.Error("Download failed", "key:", es.key, "id:", task.ID, "url:", task.Url, "err:", models.ErrIncompleteFile)
	}

	err = models.UpdateTask(h.db, task)
	if err != nil {
		carrot.Error("update task error", "key:", es.key, "id:", task.ID, "url:", task.Url, "err:", err)
	}
}

func (h *Handlers) downloadChunk(chunk *models.Chunk, outputFile *os.File, es *EventSource, startTime time.Time, task *models.Task) {
	defer h.wg.Done()

	req, err := http.NewRequest(http.MethodGet, chunk.Url, nil)
	if err != nil {
		carrot.Error("Failed to create HTTP request", "key:", es.key, "url:", chunk.Url)
		return
	}
	req = req.WithContext(es.ctx)

	req.Header.Set("Range", fmt.Sprintf("bytes=%v-%v", chunk.Start, chunk.End))
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/127.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "*/*")

	resp, err := h.client.Do(req)
	if err != nil {
		carrot.Error("Failed to send HTTP request", "key:", es.key, "url:", chunk.Url, "err:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusPartialContent {
		carrot.Error("Failed to download file", "key:", es.key, "url:", chunk.Url, "status:", resp.StatusCode)
		return
	}

	buf := make([]byte, 2048)

	h.mu.Lock()
	_, err = outputFile.Seek(int64(chunk.Start), 0)
	if err != nil {
		carrot.Error("seek error", "key:", es.key, "url:", chunk.Url, "err:", err)
		return
	}

	n, err := io.CopyBuffer(outputFile, resp.Body, buf)
	if err != nil {
		carrot.Error("Failed to copy HTTP response body", "key:", es.key, "url:", chunk.Url, "err:", err)
		return
	}

	chunk.Done = true
	task.TotalDownloaded += n
	h.mu.Unlock()

	speed, progress, remainingTime := h.calculateDownloadData(task, startTime)
	carrot.Info("speed", speed, "MB/s", "progress", progress, "remainingTime", remainingTime, "s")

	es.Emit(DownloadProgress{
		ID:            task.ID,
		Name:          task.Name,
		Progress:      progress,
		Speed:         speed,
		RemainingTime: remainingTime},
	)
}

func (h *Handlers) initTask(url, key string) (*models.Task, error) {
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
		ID:              key,
		Name:            fileName,
		Url:             url,
		Size:            fileSize,
		SavePath:        outputPath,
		FileType:        filepath.Ext(fileName),
		Status:          models.TaskStatusDownloading,
		Threads:         4,
		ChunkNum:        numChunks,
		ChunkSize:       chunkSize,
		Chunk:           make([]models.Chunk, numChunks),
		TotalDownloaded: 0,
	}

	err = models.AddTask(h.db, &task)
	if err != nil {
		return nil, err
	}

	return &task, nil
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

func (h *Handlers) calculateDownloadData(task *models.Task, startTime time.Time) (float64, float64, float64) {
	elapsedTime := time.Since(startTime).Seconds()
	speed := math.Round((float64(task.TotalDownloaded)/elapsedTime/1024/1024)*10) / 10 // MB/s
	progress := math.Round((float64(task.TotalDownloaded)/float64(task.Size)*100)*10) / 10
	remainingTime := math.Round((float64((task.Size-task.TotalDownloaded)/1024/1024)/speed)*10) / 10

	task.Progress = progress
	task.Speed = speed
	task.RemainingTime = remainingTime
	err := models.UpdateTask(h.db, task)
	if err != nil {
		carrot.Error("update task error", "id:", task.ID, "url:", task.Url, "err:", err)
	}
	return speed, progress, remainingTime
}

func (h *Handlers) handlePause(c *gin.Context) {
	key := c.Param("key")

	task, err := models.GetTaskById(h.db, key)
	if err != nil {
		carrot.AbortWithJSONError(c, http.StatusInternalServerError, err)
		return
	}

	task.Status = models.TaskStatusPending
	err = models.UpdateTask(h.db, task)
	if err != nil {
		carrot.AbortWithJSONError(c, http.StatusInternalServerError, err)
		return
	}
	h.cleanEventSource(key)
	c.JSON(http.StatusOK, gin.H{"message": "pause download"})
}
