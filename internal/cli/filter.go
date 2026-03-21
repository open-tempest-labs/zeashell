package cli

import (
	"fmt"
	"os"

	"github.com/open-tempest-labs/zeashell/internal/zeaframe"
	"github.com/spf13/cobra"
)

var filterOutput string

var filterCmd = &cobra.Command{
	Use:   "filter [expression]",
	Short: "Filter rows based on an expression",
	Long: `Filter rows based on a boolean expression with nested field support.

Supported operators:
  Comparison: =, !=, >, >=, <, <=
  Array: CONTAINS (supports wildcards)
  Logical: AND, OR

Nested field access:
  Array indexing: field[0], field[1], etc.
  Nested paths: field.subfield.property

Wildcard patterns (with CONTAINS):
  * - Match any characters
  ? - Match single character

Examples of expressions:
  amount > 100
  region = 'west'
  customer != '' AND amount > 100
  orders CONTAINS 1005
  orders[0] > 1000
  address.city = 'SF'
  address.state = 'CA' AND tags CONTAINS 'premium'
  name CONTAINS '*.webshell.*'
  service CONTAINS 'api.*.prod'`,
	Example: `  zea load sales.csv | zea filter "amount > 100"
  zea load data.csv | zea filter "region = 'west' AND amount > 50"
  zea load data.json | zea filter "orders CONTAINS 1005"
  zea load data.json | zea filter "address.city = 'SF'"`,
	Args: cobra.ExactArgs(1),
	RunE: runFilter,
}

func init() {
	filterCmd.Flags().StringVar(&filterOutput, "output", "csv", "Output format: csv, arrow (Arrow IPC)")
}

func runFilter(cmd *cobra.Command, args []string) error {
	// Read from stdin
	zf, err := zeaframe.FromCSV(os.Stdin)
	if err != nil {
		return fmt.Errorf("failed to read from stdin: %w", err)
	}

	// Parse expression
	exprStr := args[0]
	expr, err := zeaframe.ParseExpression(exprStr)
	if err != nil {
		return fmt.Errorf("failed to parse expression: %w", err)
	}

	// Filter
	result, err := zf.Filter(expr)
	if err != nil {
		return fmt.Errorf("failed to filter: %w", err)
	}

	// Write to stdout in specified format
	switch filterOutput {
	case "arrow":
		return result.WriteArrowIPC(os.Stdout)
	case "csv":
		return result.WriteCSV(os.Stdout)
	default:
		return fmt.Errorf("unsupported output format: %s (use 'csv' or 'arrow')", filterOutput)
	}
}
