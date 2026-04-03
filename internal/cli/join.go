package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/open-tempest-labs/zeashell/zeaframe"
	"github.com/spf13/cobra"
)

var (
	joinOn   string
	joinType string
)

var joinCmd = &cobra.Command{
	Use:   "join [left-source] [right-source]",
	Short: "Join two datasets on specified keys",
	Long: `Join two DataFrames on one or more key columns.

If left-source is omitted, reads from stdin.
Right-source is required (file path or URL).

Join types:
  inner - Only rows with matches in both datasets (default)
  left  - All left rows, NULLs for unmatched right
  right - All right rows, NULLs for unmatched left
  full  - All rows from both, NULLs where no match

Column name collisions are resolved by adding '_right' suffix.`,
	Example: `  # Inner join on single key
  zea join customers.csv orders.csv --on=cust_id

  # Left join on multiple keys
  zea join customers.csv orders.csv --on=id,date --type=left

  # Join stdin with file
  zea load customers.csv | zea join orders.csv --on=cust_id

  # Join and filter
  zea join customers.csv orders.csv --on=cust_id --type=left \
    | zea filter "order_id IS NULL"`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runJoin,
}

func init() {
	joinCmd.Flags().StringVar(&joinOn, "on", "", "Join key column(s), comma-separated (required)")
	joinCmd.MarkFlagRequired("on")
	joinCmd.Flags().StringVar(&joinType, "type", "inner", "Join type: inner, left, right, full")
}

func runJoin(cmd *cobra.Command, args []string) error {
	// Parse join keys
	if joinOn == "" {
		return fmt.Errorf("--on flag is required")
	}
	keys := strings.Split(joinOn, ",")
	for i := range keys {
		keys[i] = strings.TrimSpace(keys[i])
	}

	// Parse join type
	var jt zeaframe.JoinType
	switch strings.ToLower(joinType) {
	case "inner":
		jt = zeaframe.JoinInner
	case "left":
		jt = zeaframe.JoinLeft
	case "right":
		jt = zeaframe.JoinRight
	case "full":
		jt = zeaframe.JoinFull
	default:
		return fmt.Errorf("invalid join type: %s (must be inner, left, right, or full)", joinType)
	}

	// Determine left and right sources
	var leftSource, rightSource string
	if len(args) == 1 {
		// Read left from stdin
		leftSource = ""
		rightSource = args[0]
	} else {
		// Both sources specified
		leftSource = args[0]
		rightSource = args[1]
	}

	// Load left DataFrame
	var left *zeaframe.ZeaFrame
	var err error
	if leftSource == "" {
		// Read from stdin
		left, err = zeaframe.FromCSV(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read left from stdin: %w", err)
		}
	} else {
		left, err = zeaframe.LoadAuto(leftSource)
		if err != nil {
			return fmt.Errorf("failed to load left source %s: %w", leftSource, err)
		}
	}

	// Load right DataFrame
	right, err := zeaframe.LoadAuto(rightSource)
	if err != nil {
		return fmt.Errorf("failed to load right source %s: %w", rightSource, err)
	}

	// Perform join
	result, err := left.Join(right, keys, jt)
	if err != nil {
		return fmt.Errorf("failed to join: %w", err)
	}

	// Write result to stdout
	return result.WriteCSV(os.Stdout)
}
