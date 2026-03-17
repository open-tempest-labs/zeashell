package zeaframe

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// GlobOptions configures glob behavior
type GlobOptions struct {
	MaxFiles      int      // Maximum files to load (0 = unlimited)
	Parallel      int      // Number of parallel workers (0 = auto)
	FormatFilter  string   // Filter by format (csv, parquet, json, etc.)
	SchemaPreview bool     // Only preview schema without loading
	Recursive     bool     // Enable recursive directory traversal
}

// DefaultGlobOptions returns sensible defaults
func DefaultGlobOptions() *GlobOptions {
	return &GlobOptions{
		MaxFiles:      0,    // No limit
		Parallel:      8,    // 8 parallel workers
		FormatFilter:  "",   // All formats
		SchemaPreview: false,
		Recursive:     true, // Recursive by default
	}
}

// IsGlobPattern checks if a path contains glob wildcards
func IsGlobPattern(path string) bool {
	return strings.ContainsAny(path, "*?[")
}

// IsDirectory checks if a path is a directory
func IsDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// GlobFiles resolves glob patterns and directories to concrete file paths
func GlobFiles(pattern string, opts *GlobOptions) ([]string, error) {
	if opts == nil {
		opts = DefaultGlobOptions()
	}

	// Handle comma-separated patterns
	if strings.Contains(pattern, ",") {
		return globMultiplePatterns(pattern, opts)
	}

	// Handle directory
	if IsDirectory(pattern) {
		return globDirectory(pattern, opts)
	}

	// Handle glob pattern
	if IsGlobPattern(pattern) {
		return globPattern(pattern, opts)
	}

	// Single file
	return []string{pattern}, nil
}

// globMultiplePatterns handles comma-separated glob patterns
func globMultiplePatterns(patterns string, opts *GlobOptions) ([]string, error) {
	parts := strings.Split(patterns, ",")
	allFiles := make([]string, 0)
	seen := make(map[string]bool)

	for _, pattern := range parts {
		pattern = strings.TrimSpace(pattern)
		files, err := GlobFiles(pattern, opts)
		if err != nil {
			return nil, fmt.Errorf("pattern %s: %w", pattern, err)
		}

		// Deduplicate
		for _, file := range files {
			if !seen[file] {
				seen[file] = true
				allFiles = append(allFiles, file)
			}
		}
	}

	return filterAndLimit(allFiles, opts)
}

// globPattern handles glob patterns like "*.csv" or "sales/date=*.parquet"
func globPattern(pattern string, opts *GlobOptions) ([]string, error) {
	// Check for recursive glob (**/*)
	if strings.Contains(pattern, "**") {
		return globRecursive(pattern, opts)
	}

	// Standard glob
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid glob pattern: %w", err)
	}

	// Filter out directories
	files := make([]string, 0, len(matches))
	for _, match := range matches {
		if !IsDirectory(match) {
			files = append(files, match)
		}
	}

	return filterAndLimit(files, opts)
}

// globRecursive handles recursive glob patterns like "sales/**/*.parquet"
func globRecursive(pattern string, opts *GlobOptions) ([]string, error) {
	// Split pattern into base and suffix
	parts := strings.Split(pattern, "**")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid recursive glob pattern: %s", pattern)
	}

	baseDir := parts[0]
	if baseDir == "" {
		baseDir = "."
	} else {
		baseDir = strings.TrimSuffix(baseDir, "/")
	}

	suffix := strings.TrimPrefix(parts[1], "/")

	files := make([]string, 0)
	err := filepath.WalkDir(baseDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		// Match suffix pattern
		matched, err := filepath.Match(suffix, filepath.Base(path))
		if err != nil {
			return err
		}

		if matched || suffix == "" {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("recursive glob failed: %w", err)
	}

	return filterAndLimit(files, opts)
}

// globDirectory walks a directory and collects all supported files
func globDirectory(dirPath string, opts *GlobOptions) ([]string, error) {
	files := make([]string, 0)

	walkFunc := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			if !opts.Recursive && path != dirPath {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if file has supported extension
		if isSupportedFile(path) {
			files = append(files, path)
		}

		return nil
	}

	err := filepath.WalkDir(dirPath, walkFunc)
	if err != nil {
		return nil, fmt.Errorf("directory walk failed: %w", err)
	}

	return filterAndLimit(files, opts)
}

// isSupportedFile checks if a file has a supported extension
func isSupportedFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	supported := map[string]bool{
		".csv":     true,
		".tsv":     true,
		".json":    true,
		".jsonl":   true,
		".xml":     true,
		".parquet": true,
	}
	return supported[ext]
}

// filterAndLimit applies format filter and max files limit
func filterAndLimit(files []string, opts *GlobOptions) ([]string, error) {
	// Apply format filter
	if opts.FormatFilter != "" {
		filtered := make([]string, 0)
		filterExt := "." + strings.ToLower(opts.FormatFilter)
		for _, file := range files {
			if strings.HasSuffix(strings.ToLower(file), filterExt) {
				filtered = append(filtered, file)
			}
		}
		files = filtered
	}

	// Sort for deterministic ordering
	sort.Strings(files)

	// Apply max files limit
	if opts.MaxFiles > 0 && len(files) > opts.MaxFiles {
		return nil, fmt.Errorf("glob matched %d files, exceeds limit of %d (use --max-files to increase)", len(files), opts.MaxFiles)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no matching files found")
	}

	return files, nil
}

// GetFileFormat detects format from file extension
func GetFileFormat(path string) FileFormat {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".csv":
		return FormatCSV
	case ".tsv":
		return FormatTSV
	case ".json":
		return FormatJSON
	case ".jsonl":
		return FormatJSONL
	case ".xml":
		return FormatXML
	case ".parquet":
		return FormatParquet
	default:
		return FormatUnknown
	}
}
