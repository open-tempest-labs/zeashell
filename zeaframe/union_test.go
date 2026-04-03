package zeaframe

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUnionFrames(t *testing.T) {
	// Create two frames with same schema
	frame1 := NewZeaFrame()
	frame1.AddColumn("id", Int64Type, []interface{}{int64(1), int64(2)})
	frame1.AddColumn("name", StringType, []interface{}{"Alice", "Bob"})
	frame1.Rows = 2

	frame2 := NewZeaFrame()
	frame2.AddColumn("id", Int64Type, []interface{}{int64(3), int64(4)})
	frame2.AddColumn("name", StringType, []interface{}{"Charlie", "Diana"})
	frame2.Rows = 2

	schema := map[string]ColumnType{
		"id":   Int64Type,
		"name": StringType,
	}

	result, err := UnionFrames([]*ZeaFrame{frame1, frame2}, schema)
	if err != nil {
		t.Fatalf("UnionFrames failed: %v", err)
	}

	if result.Rows != 4 {
		t.Errorf("Expected 4 rows, got %d", result.Rows)
	}

	if len(result.Columns) != 2 {
		t.Errorf("Expected 2 columns, got %d", len(result.Columns))
	}
}

func TestUnionFramesSchemaEvolution(t *testing.T) {
	// Frame 1 has columns A, B
	frame1 := NewZeaFrame()
	frame1.AddColumn("A", StringType, []interface{}{"a1"})
	frame1.AddColumn("B", StringType, []interface{}{"b1"})
	frame1.Rows = 1

	// Frame 2 has columns B, C
	frame2 := NewZeaFrame()
	frame2.AddColumn("B", StringType, []interface{}{"b2"})
	frame2.AddColumn("C", StringType, []interface{}{"c2"})
	frame2.Rows = 1

	schema := map[string]ColumnType{
		"A": StringType,
		"B": StringType,
		"C": StringType,
	}

	result, err := UnionFrames([]*ZeaFrame{frame1, frame2}, schema)
	if err != nil {
		t.Fatalf("UnionFrames with schema evolution failed: %v", err)
	}

	if result.Rows != 2 {
		t.Errorf("Expected 2 rows, got %d", result.Rows)
	}

	if len(result.Columns) != 3 {
		t.Errorf("Expected 3 columns, got %d", len(result.Columns))
	}

	// Verify NULLs for missing columns
	// Frame 1 should have NULL for C
	// Frame 2 should have NULL for A
	var colC *Column
	for _, col := range result.Columns {
		if col.Name == "C" {
			colC = col
			break
		}
	}

	if colC == nil {
		t.Fatalf("Column C not found")
	}

	// First row (from frame1) should have NULL for C
	if !colC.Nulls[0] {
		t.Errorf("Expected NULL for C in first row")
	}
}

func TestPromoteType(t *testing.T) {
	tests := []struct {
		t1       ColumnType
		t2       ColumnType
		expected ColumnType
	}{
		{StringType, Int64Type, StringType},
		{Int64Type, StringType, StringType},
		{Int64Type, Float64Type, Float64Type},
		{Float64Type, Int64Type, Float64Type},
		{BoolType, Int64Type, Int64Type},
		{Int64Type, Int64Type, Int64Type},
		{StringType, StringType, StringType},
	}

	for _, tt := range tests {
		result := promoteType(tt.t1, tt.t2)
		if result != tt.expected {
			t.Errorf("promoteType(%v, %v) = %v, want %v", tt.t1, tt.t2, result, tt.expected)
		}
	}
}

func TestInferSchemaFromFiles(t *testing.T) {
	// Create temp directory with test files
	tmpDir := t.TempDir()

	// Create CSV files with different schemas
	file1 := filepath.Join(tmpDir, "file1.csv")
	os.WriteFile(file1, []byte("id,name\n1,Alice\n2,Bob\n"), 0644)

	file2 := filepath.Join(tmpDir, "file2.csv")
	os.WriteFile(file2, []byte("id,name,age\n3,Charlie,30\n4,Diana,25\n"), 0644)

	files := []string{file1, file2}

	schema, err := inferSchemaFromFiles(files, 2)
	if err != nil {
		t.Fatalf("inferSchemaFromFiles failed: %v", err)
	}

	// Should have all 3 columns
	if len(schema) != 3 {
		t.Errorf("Expected schema with 3 columns, got %d", len(schema))
	}

	// Check column names
	expectedCols := map[string]bool{"id": true, "name": true, "age": true}
	for col := range schema {
		if !expectedCols[col] {
			t.Errorf("Unexpected column in schema: %s", col)
		}
	}
}

func TestLoadMultipleFiles(t *testing.T) {
	// Create temp directory with test files
	tmpDir := t.TempDir()

	// Create CSV files
	file1 := filepath.Join(tmpDir, "file1.csv")
	os.WriteFile(file1, []byte("id,name\n1,Alice\n2,Bob\n"), 0644)

	file2 := filepath.Join(tmpDir, "file2.csv")
	os.WriteFile(file2, []byte("id,name\n3,Charlie\n4,Diana\n"), 0644)

	files := []string{file1, file2}

	opts := DefaultGlobOptions()
	opts.Parallel = 2

	result, err := LoadMultipleFiles(files, opts)
	if err != nil {
		t.Fatalf("LoadMultipleFiles failed: %v", err)
	}

	// Should have 4 rows total
	if result.Rows != 4 {
		t.Errorf("Expected 4 rows, got %d", result.Rows)
	}

	// Should have 2 columns
	if len(result.Columns) != 2 {
		t.Errorf("Expected 2 columns, got %d", len(result.Columns))
	}
}

func TestLoadMultipleFilesSingleFile(t *testing.T) {
	// Create temp directory with single file
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "file1.csv")
	os.WriteFile(file1, []byte("id,name\n1,Alice\n"), 0644)

	files := []string{file1}

	opts := DefaultGlobOptions()

	result, err := LoadMultipleFiles(files, opts)
	if err != nil {
		t.Fatalf("LoadMultipleFiles with single file failed: %v", err)
	}

	if result.Rows != 1 {
		t.Errorf("Expected 1 row, got %d", result.Rows)
	}
}

func TestLoadMultipleFilesEmpty(t *testing.T) {
	files := []string{}

	opts := DefaultGlobOptions()

	_, err := LoadMultipleFiles(files, opts)
	if err == nil {
		t.Errorf("Expected error for empty file list")
	}
}

func TestPreviewSchema(t *testing.T) {
	// Create temp directory with test files
	tmpDir := t.TempDir()

	// Create CSV files
	file1 := filepath.Join(tmpDir, "file1.csv")
	os.WriteFile(file1, []byte("id,name\n1,Alice\n"), 0644)

	file2 := filepath.Join(tmpDir, "file2.csv")
	os.WriteFile(file2, []byte("id,age\n1,30\n"), 0644)

	files := []string{file1, file2}

	schema, err := PreviewSchema(files)
	if err != nil {
		t.Fatalf("PreviewSchema failed: %v", err)
	}

	// Should have 3 columns (id, name, age)
	if len(schema) != 3 {
		t.Errorf("Expected 3 columns in schema, got %d", len(schema))
	}

	// Verify column types
	expectedTypes := map[string]ColumnType{
		"id":   Int64Type,
		"name": StringType,
		"age":  Int64Type,
	}

	for name, expectedType := range expectedTypes {
		actualType, exists := schema[name]
		if !exists {
			t.Errorf("Expected column %s not found in schema", name)
			continue
		}
		if actualType != expectedType {
			t.Errorf("Column %s: expected type %v, got %v", name, expectedType, actualType)
		}
	}
}
