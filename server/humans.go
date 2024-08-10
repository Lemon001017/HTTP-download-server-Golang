package httpDownloadServer

import "fmt"

func FormatSizeHuman(size float64) string {
	if size <= 0 {
		return "0 B"
	}
	if size < 1024 {
		return fmt.Sprintf("%.0f B", size)
	}
	size = size / 1024
	if size < 1024 {
		return fmt.Sprintf("%.1f KB", size)
	}
	size = size / 1024
	if size < 1024 {
		return fmt.Sprintf("%.1f MB", size)
	}
	size = size / 1024
	return fmt.Sprintf("%.1f GB", size)
}
