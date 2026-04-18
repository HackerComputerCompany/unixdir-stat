package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync/atomic"
)

type ProgressCallback func(scannedFiles, scannedDirs int, totalSize int64)

var progressCallback ProgressCallback
var scannedFiles int64
var scannedDirs int64

type Entry struct {
	Path     string
	Name     string
	Size     int64
	IsDir    bool
	Children []*Entry
	FileType string
}

type FileTypeStats struct {
	Extension string
	Category  string
	TotalSize int64
	FileCount int
}

var extensionCategories = map[string]string{
	".go":    "Source Code",
	".rs":    "Source Code",
	".py":    "Source Code",
	".js":    "Source Code",
	".ts":    "Source Code",
	".java":  "Source Code",
	".c":     "Source Code",
	".cpp":   "Source Code",
	".h":     "Source Code",
	".hpp":   "Source Code",
	".rb":    "Source Code",
	".php":   "Source Code",
	".swift": "Source Code",
	".kt":    "Source Code",
	".scala": "Source Code",

	".jpg":  "Images",
	".jpeg": "Images",
	".png":  "Images",
	".gif":  "Images",
	".bmp":  "Images",
	".svg":  "Images",
	".ico":  "Images",
	".webp": "Images",
	".tiff": "Images",

	".pdf":  "Documents",
	".doc":  "Documents",
	".docx": "Documents",
	".txt":  "Documents",
	".md":   "Documents",
	".rtf":  "Documents",
	".odt":  "Documents",
	".xls":  "Documents",
	".xlsx": "Documents",
	".ppt":  "Documents",
	".pptx": "Documents",

	".zip": "Archives",
	".tar": "Archives",
	".gz":  "Archives",
	".bz2": "Archives",
	".xz":  "Archives",
	".7z":  "Archives",
	".rar": "Archives",

	".mp3":  "Media",
	".wav":  "Media",
	".flac": "Media",
	".ogg":  "Media",
	".mp4":  "Media",
	".avi":  "Media",
	".mkv":  "Media",
	".mov":  "Media",
	".webm": "Media",

	".json": "Config",
	".yaml": "Config",
	".yml":  "Config",
	".toml": "Config",
	".xml":  "Config",
	".ini":  "Config",
	".conf": "Config",
	".cfg":  "Config",
}

var categorySizes = map[string]int64{}
var categoryCounts = map[string]int{}

func GetCategory(ext string) string {
	if cat, ok := extensionCategories[ext]; ok {
		return cat
	}
	return "Other"
}

func Scan(path string) (*Entry, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return nil, err
	}

	atomic.StoreInt64(&scannedFiles, 0)
	atomic.StoreInt64(&scannedDirs, 0)
	atomic.StoreInt64(&totalSize, 0)

	if !info.IsDir() {
		atomic.AddInt64(&scannedFiles, 1)
		atomic.AddInt64(&totalSize, info.Size())
		return &Entry{
			Path:     absPath,
			Name:     info.Name(),
			Size:     info.Size(),
			IsDir:    false,
			FileType: GetCategory(strings.ToLower(filepath.Ext(info.Name()))),
		}, nil
	}

	return scanDir(absPath, info.Name())
}

func scanDir(path, name string) (*Entry, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	atomic.AddInt64(&scannedDirs, 1)

	entry := &Entry{
		Path:     path,
		Name:     name,
		IsDir:    true,
		Children: []*Entry{},
	}

	for _, e := range entries {
		childPath := filepath.Join(path, e.Name())

		if e.IsDir() {
			child, err := scanDir(childPath, e.Name())
			if err != nil {
				continue
			}
			entry.Size += child.Size
			entry.Children = append(entry.Children, child)
		} else {
			info, err := e.Info()
			if err != nil {
				continue
			}
			ext := strings.ToLower(filepath.Ext(e.Name()))
			category := GetCategory(ext)

			child := &Entry{
				Path:     childPath,
				Name:     e.Name(),
				Size:     info.Size(),
				IsDir:    false,
				FileType: category,
			}
			entry.Size += child.Size
			entry.Children = append(entry.Children, child)

			atomic.AddInt64(&scannedFiles, 1)
			atomic.AddInt64(&totalSize, info.Size())
			categorySizes[category] += info.Size()
			categoryCounts[category]++

			if progressCallback != nil {
				progressCallback(
					int(atomic.LoadInt64(&scannedFiles)),
					int(atomic.LoadInt64(&scannedDirs)),
					atomic.LoadInt64(&totalSize),
				)
			}
		}
	}

	sort.Slice(entry.Children, func(i, j int) bool {
		return entry.Children[i].Size > entry.Children[j].Size
	})

	return entry, nil
}

func GetFileTypeStats() []FileTypeStats {
	stats := make([]FileTypeStats, 0, len(categorySizes))

	for category, size := range categorySizes {
		stats = append(stats, FileTypeStats{
			Category:  category,
			TotalSize: size,
			FileCount: categoryCounts[category],
		})
	}

	sort.Slice(stats, func(i, j int) bool {
		return stats[i].TotalSize > stats[j].TotalSize
	})

	return stats
}

func ResetStats() {
	categorySizes = map[string]int64{}
	categoryCounts = map[string]int{}
	atomic.StoreInt64(&scannedFiles, 0)
	atomic.StoreInt64(&scannedDirs, 0)
}

func SetProgressCallback(cb ProgressCallback) {
	progressCallback = cb
}

func GetProgress() (files, dirs int, size int64) {
	return int(atomic.LoadInt64(&scannedFiles)), int(atomic.LoadInt64(&scannedDirs)), atomic.LoadInt64(&totalSize)
}

var totalSize int64

func init() {
	ResetStats()
}

func GetTopEntries(entry *Entry, n int) []*Entry {
	if len(entry.Children) > n {
		return entry.Children[:n]
	}
	return entry.Children
}

func FormatSize(size int64) string {
	if size < 1024 {
		return "< 1 KB"
	}

	sizes := []string{"KB", "MB", "GB", "TB", "PB"}
	div := int64(1024)
	exp := 0
	for size >= div*1024 && exp < len(sizes)-1 {
		div *= 1024
		exp++
	}

	f := float64(size) / float64(div)
	if f < 10 {
		return sprintf("%.2f %s", f, sizes[exp])
	} else if f < 100 {
		return sprintf("%.1f %s", f, sizes[exp])
	}
	return sprintf("%.0f %s", f, sizes[exp])
}

func sprintf(format string, a ...interface{}) string {
	return fmt.Sprintf(format, a...)
}
