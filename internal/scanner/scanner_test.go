package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestGetTopEntries(t *testing.T) {
	entry := &Entry{
		Name:     "root",
		IsDir:    true,
		Children: []*Entry{},
	}

	for i := 5; i > 0; i-- {
		entry.Children = append(entry.Children, &Entry{
			Name:     fmt.Sprintf("file%d", i),
			Size:     int64(i * 100),
			IsDir:    false,
			FileType: "Test",
		})
	}

	top5 := GetTopEntries(entry, 5)
	if len(top5) != 5 {
		t.Errorf("GetTopEntries returned %d entries, want 5", len(top5))
	}

	top3 := GetTopEntries(entry, 3)
	if len(top3) != 3 {
		t.Errorf("GetTopEntries returned %d entries, want 3", len(top3))
	}

	if top3[0].Size != 500 || top3[0].Name != "file5" {
		t.Errorf("Expected first entry to be file5 with size 500, got %s with size %d", top3[0].Name, top3[0].Size)
	}
}

func TestResetStats(t *testing.T) {
	categorySizes["Test"] = 1000
	categoryCounts["Test"] = 5

	ResetStats()

	if len(categorySizes) != 0 || len(categoryCounts) != 0 {
		t.Error("Expected stats to be reset")
	}
}

func TestGetFileTypeStats(t *testing.T) {
	ResetStats()

	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(file1, []byte("hello"), 0644)
	file2 := filepath.Join(tmpDir, "test.go")
	os.WriteFile(file2, []byte("package main"), 0644)

	Scan(tmpDir)
	stats := GetFileTypeStats()

	if len(stats) == 0 {
		t.Error("Expected file type stats")
	}

	for _, stat := range stats {
		if stat.Category == "Documents" && stat.FileCount != 1 {
			t.Errorf("Expected 1 document file, got %d", stat.FileCount)
		}
		if stat.Category == "Source Code" && stat.FileCount != 1 {
			t.Errorf("Expected 1 source code file, got %d", stat.FileCount)
		}
	}
}

func TestScanEmptyDir(t *testing.T) {
	tmpDir := t.TempDir()

	entry, err := Scan(tmpDir)
	if err != nil {
		t.Fatalf("Scan error: %v", err)
	}

	if !entry.IsDir {
		t.Error("Expected entry to be a directory")
	}

	if entry.Size != 0 {
		t.Errorf("Expected empty dir size to be 0, got %d", entry.Size)
	}
}

func TestScanNestedDirs(t *testing.T) {
	tmpDir := t.TempDir()

	sub1 := filepath.Join(tmpDir, "sub1")
	sub2 := filepath.Join(sub1, "sub2")
	os.MkdirAll(sub2, 0755)

	os.WriteFile(filepath.Join(tmpDir, "a.txt"), []byte("aaa"), 0644)
	os.WriteFile(filepath.Join(sub1, "b.txt"), []byte("bbbbb"), 0644)
	os.WriteFile(filepath.Join(sub2, "c.txt"), []byte("cccccccccc"), 0644)

	entry, err := Scan(tmpDir)
	if err != nil {
		t.Fatalf("Scan error: %v", err)
	}

	if entry.Size != 18 {
		t.Errorf("Expected total size 18, got %d", entry.Size)
	}
}

func TestScanFile(t *testing.T) {
	tmpDir := t.TempDir()
	file := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(file, []byte("hello"), 0644)

	entry, err := Scan(file)
	if err != nil {
		t.Fatalf("Scan error: %v", err)
	}

	if entry.IsDir {
		t.Error("Expected entry to be a file, not a directory")
	}

	if entry.Name != "test.txt" {
		t.Errorf("Expected name 'test.txt', got '%s'", entry.Name)
	}
}

func TestGetCategory(t *testing.T) {
	tests := []struct {
		ext      string
		expected string
	}{
		{".go", "Source Code"},
		{".py", "Source Code"},
		{".jpg", "Images"},
		{".png", "Images"},
		{".pdf", "Documents"},
		{".txt", "Documents"},
		{".zip", "Archives"},
		{".tar", "Archives"},
		{".mp3", "Media"},
		{".mp4", "Media"},
		{".json", "Config"},
		{".yaml", "Config"},
		{".xyz", "Other"},
		{"", "Other"},
	}

	for _, tt := range tests {
		result := GetCategory(tt.ext)
		if result != tt.expected {
			t.Errorf("GetCategory(%s) = %s; want %s", tt.ext, result, tt.expected)
		}
	}
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		size     int64
		expected string
	}{
		{500, "< 1 KB"},
		{1024, "1.00 KB"},
		{1536, "1.50 KB"},
		{1048576, "1.00 MB"},
		{1572864, "1.50 MB"},
		{1073741824, "1.00 GB"},
	}

	for _, tt := range tests {
		result := FormatSize(tt.size)
		if result != tt.expected {
			t.Errorf("FormatSize(%d) = %s; want %s", tt.size, result, tt.expected)
		}
	}
}

func TestScan(t *testing.T) {
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(file1, []byte("hello"), 0644)

	subDir := filepath.Join(tmpDir, "subdir")
	os.Mkdir(subDir, 0755)

	file2 := filepath.Join(subDir, "test.go")
	os.WriteFile(file2, []byte("package main"), 0644)

	entry, err := Scan(tmpDir)
	if err != nil {
		t.Fatalf("Scan error: %v", err)
	}

	if !entry.IsDir {
		t.Error("Expected entry to be a directory")
	}

	if entry.Size == 0 {
		t.Error("Expected non-zero size")
	}
}
