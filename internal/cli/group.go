package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/open-tempest-labs/zeashell/internal/zeaframe"
	"github.com/spf13/cobra"
)

var (
	sumCols   []string
	avgCols   []string
	minCols   []string
	maxCols   []string
	countCols []string
)

var groupCmd = &cobra.Command{
	Use:   "group [columns]",
	Short: "Group by columns and perform aggregations",
	Long: `Group by one or more columns and perform aggregations.

Supported aggregations:
  --sum    Sum of values
  --avg    Average of values
  --min    Minimum value
  --max    Maximum value
  --count  Count of rows (use any column name or '1')`,
	Example: `  zea load sales.csv | zea group region --sum=amount
  zea load data.csv | zea group customer --sum=total --count=1
  zea load sales.csv | zea group region,product --sum=amount --avg=price`,
	Args: cobra.ExactArgs(1),
	RunE: runGroup,
}

func init() {
	groupCmd.Flags().StringSliceVar(&sumCols, "sum", []string{}, "Columns to sum")
	groupCmd.Flags().StringSliceVar(&avgCols, "avg", []string{}, "Columns to average")
	groupCmd.Flags().StringSliceVar(&minCols, "min", []string{}, "Columns to find minimum")
	groupCmd.Flags().StringSliceVar(&maxCols, "max", []string{}, "Columns to find maximum")
	groupCmd.Flags().StringSliceVar(&countCols, "count", []string{}, "Columns to count (use '1' for row count)")
}

func runGroup(cmd *cobra.Command, args []string) error {
	// Read from stdin
	zf, err := zeaframe.FromCSV(os.Stdin)
	if err != nil {
		return fmt.Errorf("failed to read from stdin: %w", err)
	}

	// Parse group by columns
	groupByStr := args[0]
	groupByCols := strings.Split(groupByStr, ",")
	for i := range groupByCols {
		groupByCols[i] = strings.TrimSpace(groupByCols[i])
	}

	// Build aggregations map
	aggregations := make(map[string]string)

	for _, col := range sumCols {
		aggregations[col] = "sum"
	}
	for _, col := range avgCols {
		aggregations[col] = "avg"
	}
	for _, col := range minCols {
		aggregations[col] = "min"
	}
	for _, col := range maxCols {
		aggregations[col] = "max"
	}
	for _, col := range countCols {
		if col == "1" || col == "*" {
			// Use first column for count
			if len(zf.Columns) > 0 {
				aggregations[zf.Columns[0].Name] = "count"
			}
		} else {
			aggregations[col] = "count"
		}
	}

	if len(aggregations) == 0 {
		return fmt.Errorf("no aggregations specified")
	}

	// Group and aggregate
	result, err := zf.GroupBy(groupByCols...).Agg(aggregations)
	if err != nil {
		return fmt.Errorf("failed to group: %w", err)
	}

	// Write to stdout
	return result.WriteCSV(os.Stdout)
}
