package cli

import (
	"fmt"
	"os"

	"github.com/open-tempest-labs/zeashell/internal/zeaframe"
	"github.com/spf13/cobra"
)

var loadCmd = &cobra.Command{
	Use:   "load [file]",
	Short: "Load CSV/TSV/JSON/JSONL/Parquet file and output to stdout",
	Long: `Load a CSV, TSV, JSON, JSONL, or Parquet file and stream it to stdout in CSV format.

The file format is automatically detected from the file extension:
  .csv     - CSV format
  .tsv     - TSV format
  .json    - JSON format (array of objects, preserves nested structures)
  .jsonl   - JSON Lines format (one object per line)
  .parquet - Apache Parquet format

If no file is provided, reads from stdin.`,
	Example: `  zea load sales.csv
  zea load data.tsv
  zea load data.json
  zea load events.jsonl
  zea load sales.parquet
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
