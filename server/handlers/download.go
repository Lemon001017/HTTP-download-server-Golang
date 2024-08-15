package handlers

import (
	"HTTP-download-server/server/models"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/restsend/carrot"
)

type DownloadRequest struct {
	URL string `json:"url" binding:"required"`
}

type DownloadProgress struct {
	Progress float64 `json:"progress"`
	Speed    float64 `json:"speed"`
}

// submit task
func (h *Handlers) handleSubmit(c *gin.Context) {
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
		return
	}

	eventSource := h.createEventSource()

	// Open a goroutine to handle the download separately
	go func() {
		h.processDownload(task, eventSource)
	}()
	c.JSON(http.StatusOK, EventSourceResult{Key: eventSource.key})
}

func (h *Handlers) processDownload(task *models.Task, eventSource *EventSource) {
	fileSize := int64(task.Size)
	chunkSize := task.ChunkSize
	numChunks := task.ChunkNum
	carrot.Info("fileSize:", fileSize, "savePath:", task.SavePath, "chunkSize:", chunkSize, "numChunks:", numChunks)

	outputFile, err := os.Create(task.SavePath)
	if err != nil {
		carrot.Error("create tempFile error", "key:", eventSource.key, "id:", task.ID, "url:", task.Url)
		return
	}
	defer outputFile.Close()

	// Set the size of the temporary file to be the same as the destination file
	if err := outputFile.Truncate(fileSize); err != nil {
		return
	}

	for i := 0; i < int(numChunks); i++ {
		start := int64(i) * chunkSize
		end := start + chunkSize - 1
		if end >= fileSize {
			end = fileSize - 1
		}
		task.Chunk[i].Url = task.Url
		task.Chunk[i].Index = i
		task.Chunk[i].Start = uint(start)
		task.Chunk[i].End = int(end)
		task.Chunk[i].Done = false
	}

	for i := 0; i < int(numChunks); i++ {
		h.wg.Add(1)
		go func() {
			err := h.downloadChunk(&task.Chunk[i], outputFile, eventSource)
			if err != nil {
				carrot.Error("download the", i, "Chunk error", "key:", eventSource.key, "id:", task.ID, "url:", task.Url, "err:", err)
			}
		}()
	}
	h.wg.Wait()

	task.Status = models.TaskStatusDownloaded
	err = models.UpdateTask(h.db, task)
	if err != nil {
		carrot.Error("update task error", "key:", eventSource.key, "id:", task.ID, "url:", task.Url, "err:", err)
		return
	}

	carrot.Info("Download complete", "key:", eventSource.key, "id:", task.ID, "url:", task.Url)
}

func (h *Handlers) downloadChunk(chunk *models.Chunk, outputFile *os.File, es *EventSource) error {
	defer h.wg.Done()
	req, err := http.NewRequest(http.MethodGet, chunk.Url, nil)
	if err != nil {
		carrot.Error("Failed to create HTTP request", "key:", es.key, "url:", chunk.Url)
		return err
	}

	// Sets the request header, specifying the range of bytes to download
	req.Header.Set("Range", fmt.Sprintf("bytes=%v-%v", chunk.Start, chunk.End))
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.159 Safari/537.36")
	req.Header.Set("Accept", "*/*")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		carrot.Error("Failed to send HTTP request", "key:", es.key, "url:", chunk.Url, "err:", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusPartialContent {
		carrot.Error("Failed to download file", "key:", es.key, "url:", chunk.Url, "status:", resp.StatusCode)
		return err
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if _, err := outputFile.Seek(int64(chunk.Start), 0); err != nil {
		carrot.Error("seek error", "key:", es.key, "url:", chunk.Url, "err:", err)
		return err
	}

	_, err = io.Copy(outputFile, resp.Body)
	if err != nil {
		carrot.Error("Failed to copy HTTP response body", "key:", es.key, "url:", chunk.Url, "err:", err)
		return err
	}

	chunk.Done = true

	carrot.Info("the", chunk.Index, "part of the file has been downloaded", chunk.Done)
	return nil
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

	chunkSize, numChunks := h.getChunkInfo(fileSize)

	task := models.Task{
		ID:        carrot.RandText(4),
		Name:      fileName,
		Url:       url,
		Size:      float64(fileSize),
		SavePath:  outputPath,
		FileType:  filepath.Ext(fileName),
		Status:    models.TaskStatusDownloading,
		Threads:   4,
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
