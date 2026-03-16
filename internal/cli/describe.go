package cli

import (
	"fmt"
	"os"

	"github.com/open-tempest-labs/zeashell/internal/zeaframe"
	"github.com/spf13/cobra"
)

var describeCmd = &cobra.Command{
	Use:   "describe",
	Short: "Show schema and preview of the data",
	Long: `Display the schema (column names and types) and a preview of the data.

Shows the first 10 rows by default.`,
	Example: `  zea load sales.csv | zea describe
  zea load data.csv | zea filter "amount > 100" | zea describe`,
	RunE: runDescribe,
}

func runDescribe(cmd *cobra.Command, args []string) error {
	// Read from stdin
	zf, err := zeaframe.FromCSV(os.Stdin)
	if err != nil {
		return fmt.Errorf("failed to read from stdin: %w", err)
	}

	// Print schema
	fmt.Fprintf(os.Stderr, "Schema:\n")
	fmt.Fprintf(os.Stderr, "-------\n")
	for _, col := range zf.Columns {
		typeName := getTypeName(col.Type)
		fmt.Fprintf(os.Stderr, "  %-20s %s\n", col.Name, typeName)
	}
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "Rows: %d\n", zf.Rows)
	fmt.Fprintf(os.Stderr, "Columns: %d\n", len(zf.Columns))
	fmt.Fprintf(os.Stderr, "\n")

	// Print preview (first 10 rows)
	fmt.Fprintf(os.Stderr, "Preview (first 10 rows):\n")
	fmt.Fprintf(os.Stderr, "------------------------\n")

	previewRows := 10
	if zf.Rows < previewRows {
		previewRows = zf.Rows
	}

	// Create a preview ZeaFrame
	preview := zeaframe.NewZeaFrame()
	for _, col := range zf.Columns {
		previewData := make([]interface{}, previewRows)
		copy(previewData, col.Data[:previewRows])
		preview.AddColumn(col.Name, col.Type, previewData)
	}

	// Write preview to stderr as CSV
	return preview.WriteCSV(os.Stderr)
}

func getTypeName(colType zeaframe.ColumnType) string {
	switch colType {
	case zeaframe.StringType:
		return "string"
	case zeaframe.Int64Type:
		return "int64"
	case zeaframe.Float64Type:
		return "float64"
	case zeaframe.BoolType:
		return "bool"
	case zeaframe.MultiType:
		return "multi"
	default:
		return "unknown"
	}
}
