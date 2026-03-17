package zeaframe

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// FileFormat represents supported file formats
type FileFormat int

const (
	FormatUnknown FileFormat = iota
	FormatCSV
	FormatTSV
	FormatJSON
	FormatJSONL
	FormatXML
	FormatParquet
)

// DetectFormat detects the file format from filename or content
func DetectFormat(filename string) FileFormat {
	ext := strings.ToLower(filepath.Ext(filename))

	switch ext {
	case ".csv":
		return FormatCSV
	case ".tsv":
		return FormatTSV
	case ".json":
		return FormatJSON
	case ".jsonl", ".ndjson":
		return FormatJSONL
	case ".xml":
		return FormatXML
	case ".parquet":
		return FormatParquet
	default:
		return FormatUnknown
	}
}

// isURL checks if a path is an HTTP/HTTPS URL
func isURL(path string) bool {
	return strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://")
}

// fetchURL fetches content from an HTTP/HTTPS URL
func fetchURL(url string) (io.ReadCloser, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("HTTP error: %s", resp.Status)
	}

	return resp.Body, nil
}

// LoadAuto automatically detects format and loads the file or URL
func LoadAuto(path string) (*ZeaFrame, error) {
	format := DetectFormat(path)

	// Check if it's a URL
	if isURL(path) {
		return loadFromURL(path, format)
	}

	// Local file loading
	switch format {
	case FormatCSV:
		return loadCSVFile(path)
	case FormatTSV:
		return loadTSVFile(path)
	case FormatJSON:
		return loadJSONFile(path)
	case FormatJSONL:
		return loadJSONLFile(path)
	case FormatXML:
		return loadXMLFile(path)
	case FormatParquet:
		return FromParquet(path)
	default:
		// Try CSV as default
		return loadCSVFile(path)
	}
}

// loadFromURL loads data from an HTTP/HTTPS URL
func loadFromURL(url string, format FileFormat) (*ZeaFrame, error) {
	reader, err := fetchURL(url)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	// For Parquet, we need to download to a temp file since it requires seekable access
	if format == FormatParquet {
		return loadParquetFromURL(url)
	}

	return LoadAutoFromReader(reader, format)
}

// loadParquetFromURL downloads a Parquet file to a temp location and loads it
func loadParquetFromURL(url string) (*ZeaFrame, error) {
	reader, err := fetchURL(url)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	// Create temp file
	tmpFile, err := os.CreateTemp("", "zeashell-*.parquet")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	// Download to temp file
	_, err = io.Copy(tmpFile, reader)
	tmpFile.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to download parquet file: %w", err)
	}

	// Load from temp file
	return FromParquet(tmpPath)
}

// LoadAutoFromReader loads data from a reader with format hint
func LoadAutoFromReader(reader io.Reader, formatHint FileFormat) (*ZeaFrame, error) {
	switch formatHint {
	case FormatCSV, FormatUnknown:
		return FromCSV(reader)
	case FormatTSV:
		return FromTSV(reader)
	case FormatJSON:
		return FromJSON(reader)
	case FormatJSONL:
		return FromJSONL(reader)
	case FormatXML:
		return FromXML(reader)
	default:
		return FromCSV(reader)
	}
}

// SaveAuto automatically detects format and saves the file
func (zf *ZeaFrame) SaveAuto(path string) error {
	format := DetectFormat(path)

	switch format {
	case FormatCSV, FormatUnknown:
		return zf.saveCSVFile(path)
	case FormatTSV:
		return zf.saveTSVFile(path)
	case FormatJSON:
		return zf.saveJSONFile(path)
	case FormatJSONL:
		return zf.saveJSONLFile(path)
	case FormatXML:
		return zf.saveXMLFile(path)
	case FormatParquet:
		return zf.WriteParquet(path)
	default:
		return zf.saveCSVFile(path)
	}
}

// Helper functions to load from files
func loadCSVFile(path string) (*ZeaFrame, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	return FromCSV(file)
}

func loadTSVFile(path string) (*ZeaFrame, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	return FromTSV(file)
}

func loadJSONLFile(path string) (*ZeaFrame, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	return FromJSONL(file)
}

// Helper functions to save to files
func (zf *ZeaFrame) saveCSVFile(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	return zf.WriteCSV(file)
}

func (zf *ZeaFrame) saveTSVFile(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	return zf.WriteTSV(file)
}

func (zf *ZeaFrame) saveJSONLFile(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	return zf.WriteJSONL(file)
}

func loadJSONFile(path string) (*ZeaFrame, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	return FromJSON(file)
}

func (zf *ZeaFrame) saveJSONFile(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	return zf.WriteJSON(file)
}

func loadXMLFile(path string) (*ZeaFrame, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	return FromXML(file)
}

func (zf *ZeaFrame) saveXMLFile(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Use default root/record names
	return zf.WriteXML(file, "root", "record")
}

// FormatName returns the human-readable name of a format
func (f FileFormat) String() string {
	switch f {
	case FormatCSV:
		return "CSV"
	case FormatTSV:
		return "TSV"
	case FormatJSON:
		return "JSON"
	case FormatJSONL:
		return "JSONL"
	case FormatXML:
		return "XML"
	case FormatParquet:
		return "Parquet"
	default:
		return "Unknown"
	}
}
