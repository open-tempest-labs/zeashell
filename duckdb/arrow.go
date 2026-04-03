package duckdb

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/apache/arrow-go/v18/arrow/ipc"
	"github.com/apache/arrow-go/v18/arrow/memory"
	"github.com/apache/arrow-go/v18/parquet/pqarrow"
)

// RunSQLArrowNative executes SQL using Arrow IPC for stdin, CSV for stdout
func RunSQLArrowNative(query string, stdin io.Reader, stdout io.Writer) error {
	// Create DuckDB connection
	db, err := sql.Open("duckdb", ":memory:")
	if err != nil {
		return fmt.Errorf("failed to open duckdb: %w", err)
	}
	defer db.Close()

	// Read stdin as Arrow IPC, convert to temp Parquet for DuckDB
	tempFile, err := arrowStdinToParquet(stdin)
	if err != nil {
		return fmt.Errorf("failed to process stdin Arrow: %w", err)
	}
	defer os.Remove(tempFile)

	// Register stdin table
	if tempFile != "" {
		createTableSQL := fmt.Sprintf("CREATE TABLE stdin AS SELECT * FROM read_parquet('%s')", tempFile)
		if _, err := db.Exec(createTableSQL); err != nil {
			return fmt.Errorf("failed to create stdin table: %w", err)
		}
	}

	// Execute query
	rows, err := db.Query(query)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	// Convert results to CSV (for human-readable output)
	if err := rowsToCSV(rows, stdout); err != nil {
		return fmt.Errorf("failed to write CSV: %w", err)
	}

	return nil
}

// rowsToCSV converts SQL result rows to CSV format
func rowsToCSV(rows *sql.Rows, w io.Writer) error {
	columns, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("failed to get columns: %w", err)
	}

	// Create CSV writer
	writer := csv.NewWriter(w)
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

// arrowStdinToParquet reads Arrow IPC from stdin and writes to temp Parquet file
func arrowStdinToParquet(stdin io.Reader) (string, error) {
	// Check if stdin has data
	firstByte := make([]byte, 1)
	n, err := stdin.Read(firstByte)
	if err == io.EOF || n == 0 {
		return "", nil // No stdin data
	}
	if err != nil {
		return "", err
	}

	// Create temp file
	tempDir := os.TempDir()
	tempFile := filepath.Join(tempDir, fmt.Sprintf("zea_arrow_%d.parquet", os.Getpid()))

	// Create a multi-reader with the first byte we read
	reader := io.MultiReader(io.NopCloser(io.NopCloser(io.NopCloser(&firstByteReader{firstByte, 0}))), stdin)

	// Read Arrow IPC stream
	mem := memory.NewGoAllocator()
	arrowReader, err := ipc.NewReader(reader, ipc.WithAllocator(mem))
	if err != nil {
		return "", fmt.Errorf("failed to create Arrow IPC reader: %w", err)
	}
	defer arrowReader.Release()

	// Read all records into a table
	var records []arrow.Record
	for arrowReader.Next() {
		rec := arrowReader.Record()
		rec.Retain()
		records = append(records, rec)
	}

	if err := arrowReader.Err(); err != nil {
		return "", fmt.Errorf("error reading Arrow stream: %w", err)
	}

	if len(records) == 0 {
		return "", nil // Empty input
	}

	// Create Arrow table from records
	schema := records[0].Schema()
	table := array.NewTableFromRecords(schema, records)
	defer table.Release()

	// Write to Parquet file
	f, err := os.Create(tempFile)
	if err != nil {
		return "", fmt.Errorf("failed to create temp parquet file: %w", err)
	}
	defer f.Close()

	if err := pqarrow.WriteTable(table, f, table.NumRows(), nil, pqarrow.NewArrowWriterProperties(pqarrow.WithAllocator(mem))); err != nil {
		os.Remove(tempFile)
		return "", fmt.Errorf("failed to write parquet: %w", err)
	}

	return tempFile, nil
}

// rowsToArrowIPC converts SQL result rows to Arrow IPC format
func rowsToArrowIPC(rows *sql.Rows, w io.Writer) error {
	columns, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("failed to get columns: %w", err)
	}

	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return fmt.Errorf("failed to get column types: %w", err)
	}

	// Build Arrow schema from SQL column types
	fields := make([]arrow.Field, len(columns))
	for i, colType := range columnTypes {
		fields[i] = arrow.Field{
			Name:     columns[i],
			Type:     sqlTypeToArrowType(colType.DatabaseTypeName()),
			Nullable: true,
		}
	}
	schema := arrow.NewSchema(fields, nil)

	// Collect all rows first
	var allRows [][]interface{}
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	for rows.Next() {
		if err := rows.Scan(valuePtrs...); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}
		rowCopy := make([]interface{}, len(values))
		copy(rowCopy, values)
		allRows = append(allRows, rowCopy)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating rows: %w", err)
	}

	// Create Arrow record from rows
	mem := memory.NewGoAllocator()
	builders := make([]array.Builder, len(fields))
	for i, field := range fields {
		builders[i] = array.NewBuilder(mem, field.Type)
	}
	defer func() {
		for _, b := range builders {
			b.Release()
		}
	}()

	// Fill builders with data
	for _, row := range allRows {
		for i, val := range row {
			appendToBuilder(builders[i], val)
		}
	}

	// Build arrays
	arrays := make([]arrow.Array, len(builders))
	for i, builder := range builders {
		arrays[i] = builder.NewArray()
		defer arrays[i].Release()
	}

	// Create record
	record := array.NewRecord(schema, arrays, int64(len(allRows)))
	defer record.Release()

	// Write as Arrow IPC
	writer := ipc.NewWriter(w, ipc.WithSchema(schema), ipc.WithAllocator(mem))
	defer writer.Close()

	if err := writer.Write(record); err != nil {
		return fmt.Errorf("failed to write Arrow record: %w", err)
	}

	return nil
}

// sqlTypeToArrowType converts SQL type to Arrow type
func sqlTypeToArrowType(sqlType string) arrow.DataType {
	switch sqlType {
	case "BIGINT", "INTEGER", "INT":
		return arrow.PrimitiveTypes.Int64
	case "DOUBLE", "FLOAT", "REAL":
		return arrow.PrimitiveTypes.Float64
	case "BOOLEAN":
		return arrow.FixedWidthTypes.Boolean
	case "DATE":
		return arrow.FixedWidthTypes.Date32
	case "TIMESTAMP":
		return arrow.FixedWidthTypes.Timestamp_us
	default:
		return arrow.BinaryTypes.String
	}
}

// appendToBuilder appends a value to an Arrow builder
func appendToBuilder(builder array.Builder, val interface{}) {
	if val == nil {
		builder.AppendNull()
		return
	}

	switch b := builder.(type) {
	case *array.Int64Builder:
		switch v := val.(type) {
		case int64:
			b.Append(v)
		case int:
			b.Append(int64(v))
		case int32:
			b.Append(int64(v))
		default:
			b.AppendNull()
		}
	case *array.Float64Builder:
		switch v := val.(type) {
		case float64:
			b.Append(v)
		case float32:
			b.Append(float64(v))
		default:
			b.AppendNull()
		}
	case *array.BooleanBuilder:
		if v, ok := val.(bool); ok {
			b.Append(v)
		} else {
			b.AppendNull()
		}
	case *array.StringBuilder:
		b.Append(fmt.Sprintf("%v", val))
	default:
		// Fallback: append null
		builder.AppendNull()
	}
}

// firstByteReader implements io.Reader for a single byte
type firstByteReader struct {
	data []byte
	pos  int
}

func (r *firstByteReader) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n = copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}
