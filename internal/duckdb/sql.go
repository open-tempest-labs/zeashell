package duckdb

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/marcboeker/go-duckdb"
)

// RunSQLFromStdin reads CSV from stdin, executes SQL query, and writes results as CSV to stdout
func RunSQLFromStdin(query string, stdin io.Reader, stdout io.Writer) error {
	// Create temp file for stdin data
	tempDir := os.TempDir()
	tempFile := filepath.Join(tempDir, fmt.Sprintf("zea_stdin_%d.csv", os.Getpid()))
	defer os.Remove(tempFile)

	// Write stdin to temp file
	f, err := os.Create(tempFile)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	written, err := io.Copy(f, stdin)
	f.Close()
	if err != nil {
		return fmt.Errorf("failed to write stdin to temp file: %w", err)
	}

	// Connect to DuckDB in-memory database
	db, err := sql.Open("duckdb", ":memory:")
	if err != nil {
		return fmt.Errorf("failed to open duckdb: %w", err)
	}
	defer db.Close()

	// If we have stdin data, create stdin table from the CSV file
	if written > 0 {
		createTableSQL := fmt.Sprintf("CREATE TABLE stdin AS SELECT * FROM read_csv_auto('%s')", tempFile)
		if _, err := db.Exec(createTableSQL); err != nil {
			return fmt.Errorf("failed to create stdin table: %w", err)
		}
	}

	// Execute the user's query
	rows, err := db.Query(query)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("failed to get columns: %w", err)
	}

	// Create CSV writer for output
	writer := csv.NewWriter(stdout)
	defer writer.Flush()

	// Write header
	if err := writer.Write(columns); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Prepare value holders
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	// Write rows
	for rows.Next() {
		if err := rows.Scan(valuePtrs...); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}

		record := make([]string, len(columns))
		for i, v := range values {
			if v == nil {
				record[i] = ""
			} else {
				record[i] = fmt.Sprintf("%v", v)
			}
		}

		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write row: %w", err)
		}
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating rows: %w", err)
	}

	return nil
}

// ExecutionMode determines how SQL queries are executed
type ExecutionMode int

const (
	ModeAuto ExecutionMode = iota // Auto-detect based on input
	ModeArrow                      // Force Arrow-native path
	ModeFile                       // Force file-based path (CSV/Parquet)
)

// RunSQLOptions holds options for SQL execution
type RunSQLOptions struct {
	Mode ExecutionMode
}

// RunSQL executes a SQL query with optional stdin table using auto-detected mode
// TODO: Support -c "md:database" for MotherDuck connections
// TODO: Support zea sql -i for REPL mode
// TODO: Support zea dbt-sync for dbt model generation
func RunSQL(query string) error {
	return RunSQLWithOptions(query, RunSQLOptions{Mode: ModeAuto})
}

// RunSQLWithOptions executes SQL with specific execution mode
func RunSQLWithOptions(query string, opts RunSQLOptions) error {
	mode := opts.Mode

	// Auto-detection: try Arrow first, fallback to CSV
	if mode == ModeAuto {
		// For now, detect based on first bytes of stdin
		// In future, could check for Arrow magic bytes
		mode = ModeFile // Default to file mode for reliability
	}

	switch mode {
	case ModeArrow:
		// Try Arrow-native path
		// Note: Arrow path expects Arrow IPC format stdin
		// If stdin is CSV (current zeashell default), this will fail
		// TODO: Add Arrow IPC output format to zeashell commands
		if err := RunSQLArrowNative(query, os.Stdin, os.Stdout); err != nil {
			// Fallback to file mode on Arrow errors
			if os.Getenv("ZEA_DEBUG") != "" {
				fmt.Fprintf(os.Stderr, "Debug: Arrow path failed (input may be CSV), falling back to file mode: %v\n", err)
			}
			// Cannot fallback here because stdin has been consumed
			return fmt.Errorf("Arrow mode failed (use --file for CSV input): %w", err)
		}
		return nil
	case ModeFile:
		return RunSQLFromStdin(query, os.Stdin, os.Stdout)
	default:
		return RunSQLFromStdin(query, os.Stdin, os.Stdout)
	}
}

// ValidateQuery performs basic SQL query validation
func ValidateQuery(query string) error {
	if strings.TrimSpace(query) == "" {
		return fmt.Errorf("query cannot be empty")
	}
	return nil
}
