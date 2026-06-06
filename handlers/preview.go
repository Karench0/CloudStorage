package handlers

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

const maxTextPreviewBytes = 512 * 1024

func PreviewKind(ext string) string {
	ext = strings.ToLower(ext)
	switch ext {
	case ".pdf":
		return "pdf"
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg", ".bmp":
		return "image"
	case ".txt", ".md", ".markdown", ".json", ".xml", ".html", ".htm", ".css",
		".js", ".ts", ".go", ".py", ".java", ".c", ".cpp", ".h", ".yaml", ".yml",
		".ini", ".cfg", ".conf", ".log", ".csv", ".sql", ".sh", ".env":
		return "text"
	default:
		return ""
	}
}

func IsPreviewable(ext string) bool {
	return PreviewKind(ext) != ""
}

func isTextPreview(ext string) bool {
	return PreviewKind(ext) == "text"
}

func previewContentType(fileName string) string {
	ext := strings.ToLower(filepath.Ext(fileName))
	switch ext {
	case ".pdf":
		return "application/pdf"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".svg":
		return "image/svg+xml"
	case ".bmp":
		return "image/bmp"
	case ".html", ".htm":
		return "text/html; charset=utf-8"
	case ".css":
		return "text/css; charset=utf-8"
	case ".json":
		return "application/json; charset=utf-8"
	default:
		return "text/plain; charset=utf-8"
	}
}

func limitedTextReader(r io.Reader, maxBytes int64) io.Reader {
	return io.LimitReader(r, maxBytes)
}

func FormatBytes(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}
