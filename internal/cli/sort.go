package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/open-tempest-labs/zeashell/zeaframe"
	"github.com/spf13/cobra"
)

var sortCmd = &cobra.Command{
	Use:   "sort [columns]",
	Short: "Sort rows by one or more columns",
	Long: `Sort rows by one or more columns with optional order specification.

Column format: column[:asc|:desc]
  - column      : Sort ascending (default)
  - column:asc  : Sort ascending (explicit)
  - column:desc : Sort descending

Multiple columns are applied in order (stable sort).

Examples:
  zea load data.csv | zea sort amount              # Sort by amount ascending
  zea load data.csv | zea sort amount:desc         # Sort by amount descending
  zea load data.csv | zea sort region,amount       # Sort by region, then amount
  zea load data.csv | zea sort region:asc,amount:desc  # Mixed order`,
	Example: `  zea load sales.csv | zea sort amount
  zea load sales.csv | zea sort customer,date:desc
  zea load sales.csv | zea sort region,amount:desc`,
	Args: cobra.MinimumNArgs(1),
	RunE: runSort,
}

func runSort(cmd *cobra.Command, args []string) error {
	// Read from stdin
	zf, err := zeaframe.FromCSV(os.Stdin)
	if err != nil {
		return fmt.Errorf("failed to read from stdin: %w", err)
	}

	// Parse sort columns from args
	sortCols, err := parseSortColumns(args)
	if err != nil {
		return err
	}

	// Sort
	result, err := zf.Sort(sortCols...)
	if err != nil {
		return fmt.Errorf("failed to sort: %w", err)
	}

	// Write to stdout
	return result.WriteCSV(os.Stdout)
}

// parseSortColumns parses column specifications from command args
// Format: "col1,col2:desc,col3:asc" or separate args
func parseSortColumns(args []string) ([]zeaframe.SortColumn, error) {
	var sortCols []zeaframe.SortColumn

	// Join all args and split by comma
	joined := strings.Join(args, ",")
	specs := strings.Split(joined, ",")

	for _, spec := range specs {
		spec = strings.TrimSpace(spec)
		if spec == "" {
			continue
		}

		// Parse column:order format
		parts := strings.Split(spec, ":")
		colName := strings.TrimSpace(parts[0])
		order := zeaframe.Ascending

		if len(parts) > 1 {
			orderStr := strings.ToLower(strings.TrimSpace(parts[1]))
			switch orderStr {
			case "desc", "descending", "d":
				order = zeaframe.Descending
			case "asc", "ascending", "a":
				order = zeaframe.Ascending
			default:
				return nil, fmt.Errorf("invalid sort order '%s' for column '%s' (use 'asc' or 'desc')", orderStr, colName)
			}
		}

		sortCols = append(sortCols, zeaframe.SortColumn{
			Name:  colName,
			Order: order,
		})
	}

	if len(sortCols) == 0 {
		return nil, fmt.Errorf("no sort columns specified")
	}

	return sortCols, nil
}
