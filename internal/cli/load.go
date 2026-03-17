package cli

import (
	"fmt"
	"os"

	"github.com/open-tempest-labs/zeashell/internal/zeaframe"
	"github.com/spf13/cobra"
)

var loadCmd = &cobra.Command{
	Use:   "load [file|url]",
	Short: "Load CSV/TSV/JSON/JSONL/XML/Parquet file or URL and output to stdout",
	Long: `Load a CSV, TSV, JSON, JSONL, XML, or Parquet file (or URL) and stream it to stdout in CSV format.

The file format is automatically detected from the file extension:
  .csv     - CSV format
  .tsv     - TSV format
  .json    - JSON format (array of objects, preserves nested structures)
  .jsonl   - JSON Lines format (one object per line)
  .xml     - XML format (flattened to path-based columns)
  .parquet - Apache Parquet format

Supports loading from:
  - Local files: ./data.csv, /path/to/file.json
  - HTTP/HTTPS URLs: http://example.com/data.csv
  - stdin: cat sales.csv | zea load

If no file is provided, reads from stdin.`,
	Example: `  zea load sales.csv
  zea load data.tsv
  zea load data.json
  zea load events.jsonl
  zea load topology.xml
  zea load sales.parquet
  zea load http://example.com/data.csv
  zea load https://example.com/data.json
  cat sales.csv | zea load`,
	RunE: runLoad,
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
		// Read from file with auto-detection
		filename := args[0]
		zf, err = zeaframe.LoadAuto(filename)
		if err != nil {
			return fmt.Errorf("failed to load %s: %w", filename, err)
		}
	}

	// Write to stdout
	return zf.WriteCSV(os.Stdout)
}
