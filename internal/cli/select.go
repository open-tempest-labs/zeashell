package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/open-tempest-labs/zeashell/internal/zeaframe"
	"github.com/spf13/cobra"
)

var selectOutput string

var selectCmd = &cobra.Command{
	Use:   "select [columns]",
	Short: "Select specific columns from the input",
	Long: `Select (project) specific columns from the input DataFrame.

Columns should be specified as a comma-separated list.`,
	Example: `  zea load sales.csv | zea select amount,region
  zea load data.csv | zea select customer,total,date`,
	Args: cobra.ExactArgs(1),
	RunE: runSelect,
}

func init() {
	selectCmd.Flags().StringVar(&selectOutput, "output", "csv", "Output format: csv, arrow (Arrow IPC)")
}

func runSelect(cmd *cobra.Command, args []string) error {
	// Read from stdin
	zf, err := zeaframe.FromCSV(os.Stdin)
	if err != nil {
		return fmt.Errorf("failed to read from stdin: %w", err)
	}

	// Parse columns
	columnStr := args[0]
	columns := strings.Split(columnStr, ",")
	for i := range columns {
		columns[i] = strings.TrimSpace(columns[i])
	}

	// Select columns
	result, err := zf.Select(columns...)
	if err != nil {
		return fmt.Errorf("failed to select columns: %w", err)
	}

	// Write to stdout in specified format
	switch selectOutput {
	case "arrow":
		return result.WriteArrowIPC(os.Stdout)
	case "csv":
		return result.WriteCSV(os.Stdout)
	default:
		return fmt.Errorf("unsupported output format: %s (use 'csv' or 'arrow')", selectOutput)
	}
}
