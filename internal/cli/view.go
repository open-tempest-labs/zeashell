package cli

import (
	"os"

	"github.com/open-tempest-labs/zeashell/tui"
	"github.com/spf13/cobra"
)

var viewCmd = &cobra.Command{
	Use:   "view [file|url|pattern|directory]",
	Short: "Interactive terminal viewer for data exploration",
	Long: `Launch an interactive terminal UI to explore tabular data.

The viewer provides:
  - Scrollable table view with keyboard navigation
  - Sort by column (press 's')
  - Filter with expressions (press 'f')
  - Graph/chart view for columns (press 'g')
  - Export current view to file (press 'e')
  - Reset filters and sorts (press 'r')
  - Help overlay (press '?')

Supports all data sources:
  - Single files (CSV, TSV, JSON, JSONL, XML, Parquet)
  - Glob patterns ("*.csv", "sales/**/*.parquet")
  - Directories (recursive loading)
  - URLs (HTTP/HTTPS)
  - stdin (use '-' or no argument)

Navigation:
  ↑↓←→        Move cursor
  PgUp/PgDn   Page up/down
  Home/End    First/last row

Operations:
  s           Sort by current column
  f           Filter with expression
  g           Show graph for column
  e           Export to file
  r           Reset filters/sorts
  ?           Show help
  q           Quit

Examples:
  zea view sales.csv
  zea view sales.parquet
  zea view "sales/**/*.csv"
  zea view "https://example.com/data.csv"
  cat data.csv | zea view -
  zea load sales.csv | zea filter "amount > 100" | zea view -`,
	Args: cobra.MaximumNArgs(1),
	RunE: runView,
}

func runView(cmd *cobra.Command, args []string) error {
	var source string
	if len(args) > 0 {
		source = args[0]
	} else {
		// No argument, read from stdin
		source = "-"
	}

	return tui.RunViewFromSource(source, os.Stdin)
}
