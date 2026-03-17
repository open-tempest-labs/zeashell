package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/open-tempest-labs/zeashell/internal/zeaframe"
	"github.com/spf13/cobra"
)

var (
	unpivotID     string
	unpivotValues string
	unpivotName   string
	unpivotValue  string
)

var unpivotCmd = &cobra.Command{
	Use:   "unpivot",
	Short: "Transform wide format to long format",
	Long: `Unpivot transforms wide format data to long format.

Takes multiple columns and stacks them into rows, creating two new columns:
one for the original column names and one for the values.

Example transformation:
  Input (wide):                 Output (long):
  date,west,east               date,region,amount
  2026-01-01,100,50            2026-01-01,west,100
  2026-01-02,70,               2026-01-01,east,50
                               2026-01-02,west,70

ID columns are preserved in every output row.`,
	Example: `  # Simple unpivot
  zea load sales_wide.csv \
    | zea unpivot --id=date --values=west,east --name=region --value=amount

  # Multiple ID columns
  zea load data.csv \
    | zea unpivot --id=year,month --values=q1,q2,q3,q4 --name=quarter --value=sales`,
	Args: cobra.NoArgs,
	RunE: runUnpivot,
}

func init() {
	unpivotCmd.Flags().StringVar(&unpivotID, "id", "", "ID column(s) to preserve, comma-separated (optional)")
	unpivotCmd.Flags().StringVar(&unpivotValues, "values", "", "Columns to unpivot into rows, comma-separated (required)")
	unpivotCmd.MarkFlagRequired("values")
	unpivotCmd.Flags().StringVar(&unpivotName, "name", "variable", "Name for column containing original column names (default: variable)")
	unpivotCmd.Flags().StringVar(&unpivotValue, "value", "value", "Name for column containing values (default: value)")
}

func runUnpivot(cmd *cobra.Command, args []string) error {
	// Parse ID columns
	var idCols []string
	if unpivotID != "" {
		idCols = strings.Split(unpivotID, ",")
		for i := range idCols {
			idCols[i] = strings.TrimSpace(idCols[i])
		}
	}

	// Parse value columns
	if unpivotValues == "" {
		return fmt.Errorf("--values flag is required")
	}
	valueCols := strings.Split(unpivotValues, ",")
	for i := range valueCols {
		valueCols[i] = strings.TrimSpace(valueCols[i])
	}

	// Defaults for name and value columns
	if unpivotName == "" {
		unpivotName = "variable"
	}
	if unpivotValue == "" {
		unpivotValue = "value"
	}

	// Read from stdin
	zf, err := zeaframe.FromCSV(os.Stdin)
	if err != nil {
		return fmt.Errorf("failed to read from stdin: %w", err)
	}

	// Perform unpivot
	result, err := zf.Unpivot(idCols, valueCols, unpivotName, unpivotValue)
	if err != nil {
		return fmt.Errorf("failed to unpivot: %w", err)
	}

	// Write result to stdout
	return result.WriteCSV(os.Stdout)
}
