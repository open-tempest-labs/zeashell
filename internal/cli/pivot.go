package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/open-tempest-labs/zeashell/zeaframe"
	"github.com/spf13/cobra"
)

var (
	pivotIndex  string
	pivotColumn string
	pivotValues string
)

var pivotCmd = &cobra.Command{
	Use:   "pivot",
	Short: "Transform long format to wide format",
	Long: `Pivot transforms long format data to wide format.

Takes rows and converts unique values from one column into multiple columns,
filling cells with values from another column.

Example transformation:
  Input (long):                Output (wide):
  date,region,amount          date,west,east
  2026-01-01,west,100         2026-01-01,100,50
  2026-01-01,east,50          2026-01-02,70,
  2026-01-02,west,70`,
	Example: `  # Simple pivot
  zea load sales_long.csv \
    | zea pivot --index=date --column=region --values=amount

  # Multiple index columns
  zea load data.csv \
    | zea pivot --index=year,month --column=category --values=sales`,
	Args: cobra.NoArgs,
	RunE: runPivot,
}

func init() {
	pivotCmd.Flags().StringVar(&pivotIndex, "index", "", "Index column(s), comma-separated (required)")
	pivotCmd.MarkFlagRequired("index")
	pivotCmd.Flags().StringVar(&pivotColumn, "column", "", "Column whose values become new columns (required)")
	pivotCmd.MarkFlagRequired("column")
	pivotCmd.Flags().StringVar(&pivotValues, "values", "", "Column whose values fill the new columns (required)")
	pivotCmd.MarkFlagRequired("values")
}

func runPivot(cmd *cobra.Command, args []string) error {
	// Parse index columns
	if pivotIndex == "" {
		return fmt.Errorf("--index flag is required")
	}
	indexCols := strings.Split(pivotIndex, ",")
	for i := range indexCols {
		indexCols[i] = strings.TrimSpace(indexCols[i])
	}

	if pivotColumn == "" {
		return fmt.Errorf("--column flag is required")
	}
	if pivotValues == "" {
		return fmt.Errorf("--values flag is required")
	}

	// Read from stdin
	zf, err := zeaframe.FromCSV(os.Stdin)
	if err != nil {
		return fmt.Errorf("failed to read from stdin: %w", err)
	}

	// Perform pivot
	result, err := zf.Pivot(indexCols, pivotColumn, pivotValues)
	if err != nil {
		return fmt.Errorf("failed to pivot: %w", err)
	}

	// Write result to stdout
	return result.WriteCSV(os.Stdout)
}
