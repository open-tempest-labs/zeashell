package zeaframe

import (
	"testing"
)

func TestPivot(t *testing.T) {
	// Create long format DataFrame
	zf := NewZeaFrame()
	zf.AddColumn("date", StringType, []interface{}{"2026-01-01", "2026-01-01", "2026-01-02", "2026-01-02"})
	zf.AddColumn("region", StringType, []interface{}{"West", "East", "West", "East"})
	zf.AddColumn("amount", Int64Type, []interface{}{int64(100), int64(50), int64(70), int64(60)})
	zf.Rows = 4

	// Pivot: date as index, region as columns, amount as values
	result, err := zf.Pivot([]string{"date"}, "region", "amount")
	if err != nil {
		t.Fatalf("Pivot failed: %v", err)
	}

	// Verify structure
	if result.Rows != 2 {
		t.Errorf("Expected 2 rows, got %d", result.Rows)
	}

	// Should have 3 columns: date, West, East
	if len(result.Columns) != 3 {
		t.Errorf("Expected 3 columns, got %d", len(result.Columns))
	}

	// Verify column names
	colNames := result.ColumnNames()
	if colNames[0] != "date" {
		t.Errorf("Expected first column 'date', got %s", colNames[0])
	}
	if colNames[1] != "West" && colNames[1] != "East" {
		t.Errorf("Unexpected column name: %s", colNames[1])
	}

	// Verify first row values
	dateCol := result.Columns[0]
	if dateCol.Data[0] != "2026-01-01" {
		t.Errorf("Expected date 2026-01-01, got %v", dateCol.Data[0])
	}

	// Find West column and verify value
	westIdx := result.GetColumnIndex("West")
	if westIdx == -1 {
		t.Fatalf("West column not found")
	}
	if result.Columns[westIdx].Data[0] != int64(100) {
		t.Errorf("Expected West amount 100, got %v", result.Columns[westIdx].Data[0])
	}
}

func TestPivotMultipleIndexColumns(t *testing.T) {
	// Create DataFrame with multiple index columns
	zf := NewZeaFrame()
	zf.AddColumn("year", Int64Type, []interface{}{int64(2026), int64(2026), int64(2026), int64(2026)})
	zf.AddColumn("month", Int64Type, []interface{}{int64(1), int64(1), int64(2), int64(2)})
	zf.AddColumn("category", StringType, []interface{}{"A", "B", "A", "B"})
	zf.AddColumn("sales", Int64Type, []interface{}{int64(100), int64(150), int64(120), int64(180)})
	zf.Rows = 4

	// Pivot with multiple index columns
	result, err := zf.Pivot([]string{"year", "month"}, "category", "sales")
	if err != nil {
		t.Fatalf("Pivot failed: %v", err)
	}

	// Should have 2 rows (2 unique year+month combinations)
	if result.Rows != 2 {
		t.Errorf("Expected 2 rows, got %d", result.Rows)
	}

	// Should have 4 columns: year, month, A, B
	if len(result.Columns) != 4 {
		t.Errorf("Expected 4 columns, got %d", len(result.Columns))
	}
}

func TestPivotWithMissingValues(t *testing.T) {
	// Create DataFrame with missing combinations
	zf := NewZeaFrame()
	zf.AddColumn("date", StringType, []interface{}{"2026-01-01", "2026-01-01", "2026-01-02"})
	zf.AddColumn("region", StringType, []interface{}{"West", "East", "West"})
	zf.AddColumn("amount", Int64Type, []interface{}{int64(100), int64(50), int64(70)})
	zf.Rows = 3

	result, err := zf.Pivot([]string{"date"}, "region", "amount")
	if err != nil {
		t.Fatalf("Pivot failed: %v", err)
	}

	// Verify missing value is NULL
	eastIdx := result.GetColumnIndex("East")
	if eastIdx == -1 {
		t.Fatalf("East column not found")
	}

	// Second row should have NULL for East
	if !result.Columns[eastIdx].Nulls[1] {
		t.Errorf("Expected NULL for missing value")
	}
}

func TestUnpivot(t *testing.T) {
	// Create wide format DataFrame
	zf := NewZeaFrame()
	zf.AddColumn("date", StringType, []interface{}{"2026-01-01", "2026-01-02"})
	zf.AddColumn("West", Int64Type, []interface{}{int64(100), int64(70)})
	zf.AddColumn("East", Int64Type, []interface{}{int64(50), int64(60)})
	zf.Rows = 2

	// Unpivot
	result, err := zf.Unpivot([]string{"date"}, []string{"West", "East"}, "region", "amount")
	if err != nil {
		t.Fatalf("Unpivot failed: %v", err)
	}

	// Should have 4 rows (2 dates × 2 regions)
	if result.Rows != 4 {
		t.Errorf("Expected 4 rows, got %d", result.Rows)
	}

	// Should have 3 columns: date, region, amount
	if len(result.Columns) != 3 {
		t.Errorf("Expected 3 columns, got %d", len(result.Columns))
	}

	// Verify column names
	colNames := result.ColumnNames()
	if colNames[0] != "date" || colNames[1] != "region" || colNames[2] != "amount" {
		t.Errorf("Unexpected column names: %v", colNames)
	}

	// Verify first row
	if result.Columns[0].Data[0] != "2026-01-01" {
		t.Errorf("Expected date 2026-01-01, got %v", result.Columns[0].Data[0])
	}
	if result.Columns[1].Data[0] != "West" {
		t.Errorf("Expected region West, got %v", result.Columns[1].Data[0])
	}
	if result.Columns[2].Data[0] != int64(100) {
		t.Errorf("Expected amount 100, got %v", result.Columns[2].Data[0])
	}
}

func TestUnpivotNoIDColumns(t *testing.T) {
	// Create DataFrame with no ID columns
	zf := NewZeaFrame()
	zf.AddColumn("Q1", Int64Type, []interface{}{int64(100)})
	zf.AddColumn("Q2", Int64Type, []interface{}{int64(150)})
	zf.AddColumn("Q3", Int64Type, []interface{}{int64(200)})
	zf.Rows = 1

	// Unpivot without ID columns
	result, err := zf.Unpivot([]string{}, []string{"Q1", "Q2", "Q3"}, "quarter", "sales")
	if err != nil {
		t.Fatalf("Unpivot failed: %v", err)
	}

	// Should have 3 rows (3 quarters)
	if result.Rows != 3 {
		t.Errorf("Expected 3 rows, got %d", result.Rows)
	}

	// Should have 2 columns: quarter, sales
	if len(result.Columns) != 2 {
		t.Errorf("Expected 2 columns, got %d", len(result.Columns))
	}
}

func TestUnpivotMultipleIDColumns(t *testing.T) {
	// Create DataFrame with multiple ID columns
	zf := NewZeaFrame()
	zf.AddColumn("year", Int64Type, []interface{}{int64(2026), int64(2026)})
	zf.AddColumn("month", Int64Type, []interface{}{int64(1), int64(2)})
	zf.AddColumn("A", Int64Type, []interface{}{int64(100), int64(120)})
	zf.AddColumn("B", Int64Type, []interface{}{int64(150), int64(180)})
	zf.Rows = 2

	// Unpivot with multiple ID columns
	result, err := zf.Unpivot([]string{"year", "month"}, []string{"A", "B"}, "category", "sales")
	if err != nil {
		t.Fatalf("Unpivot failed: %v", err)
	}

	// Should have 4 rows (2 months × 2 categories)
	if result.Rows != 4 {
		t.Errorf("Expected 4 rows, got %d", result.Rows)
	}

	// Should have 4 columns: year, month, category, sales
	if len(result.Columns) != 4 {
		t.Errorf("Expected 4 columns, got %d", len(result.Columns))
	}

	// Verify ID columns are repeated correctly
	yearCol := result.Columns[0]
	if yearCol.Data[0] != int64(2026) || yearCol.Data[1] != int64(2026) {
		t.Errorf("Year should be repeated for each category")
	}
}

func TestPivotUnpivotRoundTrip(t *testing.T) {
	// Create long format data
	original := NewZeaFrame()
	original.AddColumn("date", StringType, []interface{}{"2026-01-01", "2026-01-01", "2026-01-02", "2026-01-02"})
	original.AddColumn("region", StringType, []interface{}{"West", "East", "West", "East"})
	original.AddColumn("amount", Int64Type, []interface{}{int64(100), int64(50), int64(70), int64(60)})
	original.Rows = 4

	// Pivot to wide
	pivoted, err := original.Pivot([]string{"date"}, "region", "amount")
	if err != nil {
		t.Fatalf("Pivot failed: %v", err)
	}

	// Unpivot back to long
	unpivoted, err := pivoted.Unpivot([]string{"date"}, []string{"West", "East"}, "region", "amount")
	if err != nil {
		t.Fatalf("Unpivot failed: %v", err)
	}

	// Should have same number of rows as original
	if unpivoted.Rows != original.Rows {
		t.Errorf("Expected %d rows after round trip, got %d", original.Rows, unpivoted.Rows)
	}

	// Should have same column structure
	if len(unpivoted.Columns) != len(original.Columns) {
		t.Errorf("Expected %d columns after round trip, got %d", len(original.Columns), len(unpivoted.Columns))
	}
}

func TestPivotInvalidColumn(t *testing.T) {
	zf := NewZeaFrame()
	zf.AddColumn("date", StringType, []interface{}{"2026-01-01"})
	zf.AddColumn("amount", Int64Type, []interface{}{int64(100)})
	zf.Rows = 1

	// Try to pivot with non-existent column
	_, err := zf.Pivot([]string{"date"}, "nonexistent", "amount")
	if err == nil {
		t.Errorf("Expected error for non-existent column")
	}
}

func TestUnpivotInvalidColumn(t *testing.T) {
	zf := NewZeaFrame()
	zf.AddColumn("date", StringType, []interface{}{"2026-01-01"})
	zf.AddColumn("amount", Int64Type, []interface{}{int64(100)})
	zf.Rows = 1

	// Try to unpivot with non-existent column
	_, err := zf.Unpivot([]string{"date"}, []string{"nonexistent"}, "variable", "value")
	if err == nil {
		t.Errorf("Expected error for non-existent column")
	}
}
