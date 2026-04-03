package cli

import (
	"fmt"
	"os"

	"github.com/open-tempest-labs/zeashell/zeaframe"
	"github.com/spf13/cobra"
)

var (
	loadMaxFiles      int
	loadParallel      int
	loadFormat        string
	loadSchemaPreview bool
	loadOutput        string
)

var loadCmd = &cobra.Command{
	Use:   "load [file|url|pattern|directory]",
	Short: "Load files with glob patterns, directories, or URLs",
	Long: `Load data from files, glob patterns, directories, or URLs and stream to stdout in CSV format.

The file format is automatically detected from the file extension:
  .csv     - CSV format
  .tsv     - TSV format
  .json    - JSON format (array of objects, preserves nested structures)
  .jsonl   - JSON Lines format (one object per line)
  .xml     - XML format (flattened to path-based columns)
  .parquet - Apache Parquet format

Supports loading from:
  - Single files: ./data.csv, /path/to/file.json
  - Glob patterns: "*.csv", "sales/date=*.parquet"
  - Multiple globs: "file1.csv,file2.parquet"
  - Directories: "sales/" (recursive by default)
  - Recursive globs: "sales/**/*.parquet"
  - HTTP/HTTPS URLs: http://example.com/data.csv
  - stdin: cat sales.csv | zea load

Multi-file loading:
  - Files are loaded in parallel for performance
  - Schema is inferred from first 3 files
  - All files are unioned into a single table
  - Missing columns are filled with NULLs

If no file is provided, reads from stdin.`,
	Example: `  # Single file
  zea load sales.csv

  # Glob pattern
  zea load "sales/*.csv"
  zea load "sales/date=2026-03-*.parquet"

  # Directory (recursive)
  zea load "sales/"
  zea load "sales/**/*.parquet"

  # Multiple files
  zea load "file1.csv,file2.csv"

  # With filters
  zea load "sales/*.parquet" --format=parquet --max-files=100

  # Schema preview
  zea load "sales/" --schema-preview

  # URLs
  zea load http://example.com/data.csv

  # stdin
  cat sales.csv | zea load`,
	RunE: runLoad,
}

func init() {
	loadCmd.Flags().IntVar(&loadMaxFiles, "max-files", 0, "Maximum files to load (0 = unlimited)")
	loadCmd.Flags().IntVar(&loadParallel, "parallel", 8, "Number of parallel workers for multi-file loading")
	loadCmd.Flags().StringVar(&loadFormat, "format", "", "Filter by format (csv, parquet, json, etc.)")
	loadCmd.Flags().BoolVar(&loadSchemaPreview, "schema-preview", false, "Show inferred schema without loading data")
	loadCmd.Flags().StringVar(&loadOutput, "output", "csv", "Output format: csv, arrow (Arrow IPC)")
}

func runLoad(cmd *cobra.Command, args []string) error {
	var zf *zeaframe.ZeaFrame
	var err error

	if len(args) == 0 {
		// Read from stdin (assume CSV)
		zf, err = zeaframe.FromCSV(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read from stdin: %w", err)
		}
	} else {
		source := args[0]

		// Create glob options
		opts := &zeaframe.GlobOptions{
			MaxFiles:      loadMaxFiles,
			Parallel:      loadParallel,
			FormatFilter:  loadFormat,
			SchemaPreview: loadSchemaPreview,
			Recursive:     true,
		}

		// Check if source needs globbing/multi-file loading
		if zeaframe.IsGlobPattern(source) || zeaframe.IsDirectory(source) || contains(source, ",") {
			// Resolve files from glob/directory
			files, err := zeaframe.GlobFiles(source, opts)
			if err != nil {
				return fmt.Errorf("failed to resolve files: %w", err)
			}

			// Schema preview mode
			if loadSchemaPreview {
				schema, err := zeaframe.PreviewSchema(files)
				if err != nil {
					return fmt.Errorf("failed to preview schema: %w", err)
				}
				fmt.Fprintf(os.Stderr, "Found %d files\n", len(files))
				fmt.Fprintf(os.Stderr, "\nInferred schema:\n")
				for name, colType := range schema {
					fmt.Fprintf(os.Stderr, "  %s: %s\n", name, colType)
				}
				return nil
			}

			// Load multiple files
			fmt.Fprintf(os.Stderr, "Loading %d files in parallel...\n", len(files))
			zf, err = zeaframe.LoadMultipleFiles(files, opts)
			if err != nil {
				return fmt.Errorf("failed to load files: %w", err)
			}
			fmt.Fprintf(os.Stderr, "Loaded %d rows from %d files\n", zf.Rows, len(files))
		} else {
			// Single file or URL
			zf, err = zeaframe.LoadAuto(source)
			if err != nil {
				return fmt.Errorf("failed to load %s: %w", source, err)
			}
		}
	}

	// Write to stdout in specified format
	switch loadOutput {
	case "arrow":
		return zf.WriteArrowIPC(os.Stdout)
	case "csv":
		return zf.WriteCSV(os.Stdout)
	default:
		return fmt.Errorf("unsupported output format: %s (use 'csv' or 'arrow')", loadOutput)
	}
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && indexOf(s, substr) >= 0
}

// indexOf returns the index of substr in s, or -1
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
