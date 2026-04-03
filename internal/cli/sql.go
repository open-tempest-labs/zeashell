package cli

import (
	"fmt"

	"github.com/open-tempest-labs/zeashell/duckdb"
	"github.com/spf13/cobra"
)

var (
	sqlUseArrow bool
	sqlUseFile  bool
)

func init() {
	sqlCmd.Flags().BoolVar(&sqlUseArrow, "arrow", false, "Force Arrow-native execution (fastest, zero-copy)")
	sqlCmd.Flags().BoolVar(&sqlUseFile, "file", false, "Force file-based execution (reliable fallback)")
}

var sqlCmd = &cobra.Command{
	Use:   "sql [query]",
	Short: "Execute SQL query with DuckDB",
	Long: `Execute SQL queries using DuckDB on data piped from stdin.

Execution Modes:
  Auto (default)  - Automatically choose best execution path
  Arrow (--arrow) - Force Arrow-native path (fastest, zero-copy)
  File (--file)   - Force file-based path (reliable, debuggable)

Performance Hierarchy:
  1. Arrow-native (fastest, zero-copy, in-memory)
  2. Parquet temp files (good fallback, type-safe)
  3. CSV temp files (universal compatibility)

The sql command automatically selects the best execution mode based on input.
Use --arrow to force the fastest path or --file for maximum reliability.

The sql command reads CSV data from stdin and makes it available as the "stdin" table.
You can then run SQL queries using DuckDB's powerful analytics engine.

Stdin data is automatically registered as a table named "stdin" that you can query.

Features:
  - Full DuckDB SQL support (GROUP BY, JOIN, window functions, etc.)
  - Stdin data available as "stdin" table
  - Results output as CSV to stdout (pipeable to other zea commands)
  - In-memory processing (no data persistence)

Examples:
  # Simple aggregation
  zea load sales.csv | zea sql "SELECT region, SUM(amount) FROM stdin GROUP BY region"

  # Filtering with SQL
  zea load sales.csv | zea sql "SELECT * FROM stdin WHERE amount > 1000"

  # Complex query with window functions
  zea load sales.csv | zea sql "SELECT *, ROW_NUMBER() OVER (PARTITION BY region ORDER BY amount DESC) as rank FROM stdin"

  # Self-join on stdin table
  zea load sales.csv | zea sql "SELECT s1.customer, s1.amount FROM stdin s1 JOIN stdin s2 ON s1.region = s2.region"

  # Combine with other zea commands
  zea load sales.csv | zea sql "SELECT * FROM stdin WHERE amount > 1000" | zea view

  # Query without stdin (just compute)
  zea sql "SELECT 1 as x, 2 as y, 3 as z" | zea view`,
	Example: `  # Aggregate sales by region
  zea load sales.csv | zea sql "SELECT region, SUM(amount) as total FROM stdin GROUP BY region"

  # Filter and transform
  zea load data.csv | zea sql "SELECT customer, ROUND(amount * 1.1, 2) as adjusted FROM stdin WHERE status = 'active'"

  # Complex analytics
  zea load sales.csv | zea sql "SELECT region, AVG(amount) OVER (PARTITION BY region) as avg_amount FROM stdin"`,
	Args: cobra.ExactArgs(1),
	RunE: runSQL,
}

func runSQL(cmd *cobra.Command, args []string) error {
	query := args[0]

	// Validate query
	if err := duckdb.ValidateQuery(query); err != nil {
		return fmt.Errorf("invalid query: %w", err)
	}

	// Determine execution mode
	mode := duckdb.ModeAuto
	if sqlUseArrow && sqlUseFile {
		return fmt.Errorf("cannot specify both --arrow and --file")
	}
	if sqlUseArrow {
		mode = duckdb.ModeArrow
	} else if sqlUseFile {
		mode = duckdb.ModeFile
	}

	// Execute SQL query with options
	opts := duckdb.RunSQLOptions{Mode: mode}
	if err := duckdb.RunSQLWithOptions(query, opts); err != nil {
		return fmt.Errorf("failed to execute SQL: %w", err)
	}

	return nil
}
