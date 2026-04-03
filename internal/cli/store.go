package cli

import (
	"fmt"
	"os"

	"github.com/open-tempest-labs/zeashell/zeaframe"
	"github.com/spf13/cobra"
)

var storeCmd = &cobra.Command{
	Use:   "store [file]",
	Short: "Store input data to a file",
	Long: `Store the input DataFrame to a file.

The output format is determined by the file extension:
  .csv     - CSV format
  .tsv     - TSV format
  .json    - JSON format (pretty-printed array of objects)
  .jsonl   - JSON Lines format (one object per line)
  .parquet - Apache Parquet format

If no file is provided, outputs to stdout in CSV format.`,
	Example: `  zea load sales.csv | zea filter "amount > 100" | zea store output.csv
  zea load data.csv | zea select customer,total | zea store summary.json
  zea load data.csv | zea select customer,total | zea store summary.parquet`,
	Args: cobra.MaximumNArgs(1),
	RunE: runStore,
}

func runStore(cmd *cobra.Command, args []string) error {
	// Read from stdin
	zf, err := zeaframe.FromCSV(os.Stdin)
	if err != nil {
		return fmt.Errorf("failed to read from stdin: %w", err)
	}

	if len(args) == 0 {
		// Write to stdout
		return zf.WriteCSV(os.Stdout)
	}

	// Write to file with auto-detection
	filename := args[0]
	err = zf.SaveAuto(filename)
	if err != nil {
		return fmt.Errorf("failed to write to %s: %w", filename, err)
	}

	fmt.Fprintf(os.Stderr, "Wrote %d rows to %s\n", zf.Rows, filename)
	return nil
}
