package zeaframe

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsGlobPattern(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"*.csv", true},
		{"data/*.parquet", true},
		{"sales/date=*.csv", true},
		{"data/[abc].csv", true},
		{"data/file?.csv", true},
		{"data.csv", false},
		{"/path/to/file.csv", false},
		{"", false},
	}

	for _, tt := range tests {
		result := IsGlobPattern(tt.path)
		if result != tt.expected {
			t.Errorf("IsGlobPattern(%q) = %v, want %v", tt.path, result, tt.expected)
		}
	}
}

func TestIsDirectory(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	// Create a file
	tmpFile := filepath.Join(tmpDir, "test.csv")
	os.WriteFile(tmpFile, []byte("test"), 0644)

	tests := []struct {
		path     string
		expected bool
	}{
		{tmpDir, true},
		{tmpFile, false},
		{"/nonexistent", false},
	}

	for _, tt := range tests {
		result := IsDirectory(tt.path)
		if result != tt.expected {
			t.Errorf("IsDirectory(%q) = %v, want %v", tt.path, result, tt.expected)
		}
	}
}

func TestGlobPattern(t *testing.T) {
	// Create temp directory with test files
	tmpDir := t.TempDir()

	// Create test files
	files := []string{
		"file1.csv",
		"file2.csv",
		"file3.json",
		"data.parquet",
	}

	for _, file := range files {
		path := filepath.Join(tmpDir, file)
		os.WriteFile(path, []byte("test"), 0644)
	}

	opts := DefaultGlobOptions()

	// Test glob pattern
	pattern := filepath.Join(tmpDir, "*.csv")
	result, err := globPattern(pattern, opts)
	if err != nil {
		t.Fatalf("globPattern failed: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 CSV files, got %d", len(result))
	}
}

func TestGlobDirectory(t *testing.T) {
	// Create temp directory structure
	tmpDir := t.TempDir()

	// Create nested directories
	subDir := filepath.Join(tmpDir, "subdir")
	os.Mkdir(subDir, 0755)

	// Create files
	os.WriteFile(filepath.Join(tmpDir, "file1.csv"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file2.json"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(subDir, "file3.csv"), []byte("test"), 0644)

	opts := DefaultGlobOptions()
	opts.Recursive = true

	result, err := globDirectory(tmpDir, opts)
	if err != nil {
		t.Fatalf("globDirectory failed: %v", err)
	}

	// Should find all 3 files (recursive)
	if len(result) != 3 {
		t.Errorf("Expected 3 files (recursive), got %d", len(result))
	}

	// Test non-recursive
	opts.Recursive = false
	result, err = globDirectory(tmpDir, opts)
	if err != nil {
		t.Fatalf("globDirectory (non-recursive) failed: %v", err)
	}

	// Should find only 2 files (non-recursive)
	if len(result) != 2 {
		t.Errorf("Expected 2 files (non-recursive), got %d", len(result))
	}
}

func TestFilterAndLimit(t *testing.T) {
	files := []string{
		"file1.csv",
		"file2.csv",
		"file3.json",
		"data.parquet",
	}

	// Test format filter
	opts := DefaultGlobOptions()
	opts.FormatFilter = "csv"

	result, err := filterAndLimit(files, opts)
	if err != nil {
		t.Fatalf("filterAndLimit failed: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 CSV files after filter, got %d", len(result))
	}

	// Test max files limit - should succeed with exactly MaxFiles
	opts2 := DefaultGlobOptions()
	opts2.MaxFiles = 4 // Exact match

	result, err = filterAndLimit(files, opts2)
	if err != nil {
		t.Fatalf("filterAndLimit with max files failed: %v", err)
	}

	if len(result) != 4 {
		t.Errorf("Expected 4 files with limit=4, got %d", len(result))
	}

	// Test max files exceeded - should error
	opts3 := DefaultGlobOptions()
	opts3.MaxFiles = 2 // Less than 4 files

	_, err = filterAndLimit(files, opts3)
	if err == nil {
		t.Errorf("Expected error when exceeding max files (4 files > limit of 2)")
	}
}

func TestGetFileFormat(t *testing.T) {
	tests := []struct {
		path     string
		expected FileFormat
	}{
		{"data.csv", FormatCSV},
		{"data.tsv", FormatTSV},
		{"data.json", FormatJSON},
		{"data.jsonl", FormatJSONL},
		{"data.xml", FormatXML},
		{"data.parquet", FormatParquet},
		{"data.CSV", FormatCSV}, // Case insensitive
		{"data.txt", FormatUnknown},
		{"data", FormatUnknown},
	}

	for _, tt := range tests {
		result := GetFileFormat(tt.path)
		if result != tt.expected {
			t.Errorf("GetFileFormat(%q) = %v, want %v", tt.path, result, tt.expected)
		}
	}
}

func TestGlobMultiplePatterns(t *testing.T) {
	// Create temp directory with test files
	tmpDir := t.TempDir()

	// Create test files
	os.WriteFile(filepath.Join(tmpDir, "file1.csv"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file2.csv"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file3.json"), []byte("test"), 0644)

	opts := DefaultGlobOptions()

	// Test comma-separated patterns
	pattern := filepath.Join(tmpDir, "file1.csv") + "," + filepath.Join(tmpDir, "file3.json")
	result, err := globMultiplePatterns(pattern, opts)
	if err != nil {
		t.Fatalf("globMultiplePatterns failed: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 files, got %d", len(result))
	}

	// Verify deduplication
	duplicatePattern := filepath.Join(tmpDir, "file1.csv") + "," + filepath.Join(tmpDir, "file1.csv")
	result, err = globMultiplePatterns(duplicatePattern, opts)
	if err != nil {
		t.Fatalf("globMultiplePatterns with duplicates failed: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("Expected 1 file after deduplication, got %d", len(result))
	}
}

func TestGlobRecursive(t *testing.T) {
	// Create temp directory structure
	tmpDir := t.TempDir()

	// Create nested structure
	subDir1 := filepath.Join(tmpDir, "data", "2026")
	subDir2 := filepath.Join(tmpDir, "data", "2027")
	os.MkdirAll(subDir1, 0755)
	os.MkdirAll(subDir2, 0755)

	// Create files
	os.WriteFile(filepath.Join(tmpDir, "file1.csv"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(subDir1, "file2.csv"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(subDir2, "file3.csv"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(subDir1, "file.json"), []byte("test"), 0644)

	opts := DefaultGlobOptions()

	// Test recursive glob
	pattern := filepath.Join(tmpDir, "**/*.csv")
	result, err := globRecursive(pattern, opts)
	if err != nil {
		t.Fatalf("globRecursive failed: %v", err)
	}

	// Should find all 3 CSV files
	if len(result) != 3 {
		t.Errorf("Expected 3 CSV files, got %d", len(result))
	}
}
