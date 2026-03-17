package zeaframe

import (
	"testing"
)

func TestJoinInner(t *testing.T) {
	// Create left DataFrame
	left := NewZeaFrame()
	left.AddColumn("id", Int64Type, []interface{}{int64(1), int64(2), int64(3)})
	left.AddColumn("name", StringType, []interface{}{"Alice", "Bob", "Charlie"})
	left.Rows = 3

	// Create right DataFrame
	right := NewZeaFrame()
	right.AddColumn("id", Int64Type, []interface{}{int64(1), int64(2), int64(4)})
	right.AddColumn("amount", Int64Type, []interface{}{int64(100), int64(200), int64(400)})
	right.Rows = 3

	// Perform inner join
	result, err := left.Join(right, []string{"id"}, JoinInner)
	if err != nil {
		t.Fatalf("Join failed: %v", err)
	}

	// Verify result
	if result.Rows != 2 {
		t.Errorf("Expected 2 rows, got %d", result.Rows)
	}

	if len(result.Columns) != 3 { // id, name, amount
		t.Errorf("Expected 3 columns, got %d", len(result.Columns))
	}

	// Verify first row
	if result.Columns[0].Data[0] != int64(1) {
		t.Errorf("Expected id=1, got %v", result.Columns[0].Data[0])
	}
	if result.Columns[1].Data[0] != "Alice" {
		t.Errorf("Expected name=Alice, got %v", result.Columns[1].Data[0])
	}
	if result.Columns[2].Data[0] != int64(100) {
		t.Errorf("Expected amount=100, got %v", result.Columns[2].Data[0])
	}
}

func TestJoinLeft(t *testing.T) {
	// Create left DataFrame
	left := NewZeaFrame()
	left.AddColumn("id", Int64Type, []interface{}{int64(1), int64(2), int64(3)})
	left.AddColumn("name", StringType, []interface{}{"Alice", "Bob", "Charlie"})
	left.Rows = 3

	// Create right DataFrame
	right := NewZeaFrame()
	right.AddColumn("id", Int64Type, []interface{}{int64(1), int64(2)})
	right.AddColumn("amount", Int64Type, []interface{}{int64(100), int64(200)})
	right.Rows = 2

	// Perform left join
	result, err := left.Join(right, []string{"id"}, JoinLeft)
	if err != nil {
		t.Fatalf("Join failed: %v", err)
	}

	// Verify result - should have all 3 left rows
	if result.Rows != 3 {
		t.Errorf("Expected 3 rows, got %d", result.Rows)
	}

	// Verify third row has NULL for amount
	if !result.Columns[2].Nulls[2] {
		t.Errorf("Expected NULL for unmatched row")
	}
}

func TestJoinRight(t *testing.T) {
	// Create left DataFrame
	left := NewZeaFrame()
	left.AddColumn("id", Int64Type, []interface{}{int64(1), int64(2)})
	left.AddColumn("name", StringType, []interface{}{"Alice", "Bob"})
	left.Rows = 2

	// Create right DataFrame
	right := NewZeaFrame()
	right.AddColumn("id", Int64Type, []interface{}{int64(1), int64(2), int64(3)})
	right.AddColumn("amount", Int64Type, []interface{}{int64(100), int64(200), int64(300)})
	right.Rows = 3

	// Perform right join
	result, err := left.Join(right, []string{"id"}, JoinRight)
	if err != nil {
		t.Fatalf("Join failed: %v", err)
	}

	// Verify result - should have all 3 right rows
	if result.Rows != 3 {
		t.Errorf("Expected 3 rows, got %d", result.Rows)
	}

	// Verify third row has NULL for name
	if !result.Columns[1].Nulls[2] {
		t.Errorf("Expected NULL for unmatched row")
	}
}

func TestJoinFull(t *testing.T) {
	// Create left DataFrame
	left := NewZeaFrame()
	left.AddColumn("id", Int64Type, []interface{}{int64(1), int64(2), int64(3)})
	left.AddColumn("name", StringType, []interface{}{"Alice", "Bob", "Charlie"})
	left.Rows = 3

	// Create right DataFrame
	right := NewZeaFrame()
	right.AddColumn("id", Int64Type, []interface{}{int64(2), int64(3), int64(4)})
	right.AddColumn("amount", Int64Type, []interface{}{int64(200), int64(300), int64(400)})
	right.Rows = 3

	// Perform full join
	result, err := left.Join(right, []string{"id"}, JoinFull)
	if err != nil {
		t.Fatalf("Join failed: %v", err)
	}

	// Verify result - should have rows for 1,2,3,4
	if result.Rows != 4 {
		t.Errorf("Expected 4 rows, got %d", result.Rows)
	}
}

func TestJoinMultipleKeys(t *testing.T) {
	// Create left DataFrame
	left := NewZeaFrame()
	left.AddColumn("year", Int64Type, []interface{}{int64(2026), int64(2026)})
	left.AddColumn("month", Int64Type, []interface{}{int64(1), int64(2)})
	left.AddColumn("sales", Int64Type, []interface{}{int64(100), int64(150)})
	left.Rows = 2

	// Create right DataFrame
	right := NewZeaFrame()
	right.AddColumn("year", Int64Type, []interface{}{int64(2026), int64(2026)})
	right.AddColumn("month", Int64Type, []interface{}{int64(1), int64(2)})
	right.AddColumn("expenses", Int64Type, []interface{}{int64(80), int64(90)})
	right.Rows = 2

	// Perform join on multiple keys
	result, err := left.Join(right, []string{"year", "month"}, JoinInner)
	if err != nil {
		t.Fatalf("Join failed: %v", err)
	}

	// Verify result
	if result.Rows != 2 {
		t.Errorf("Expected 2 rows, got %d", result.Rows)
	}

	if len(result.Columns) != 4 { // year, month, sales, expenses
		t.Errorf("Expected 4 columns, got %d", len(result.Columns))
	}
}

func TestJoinColumnCollision(t *testing.T) {
	// Create left DataFrame
	left := NewZeaFrame()
	left.AddColumn("id", Int64Type, []interface{}{int64(1)})
	left.AddColumn("value", Int64Type, []interface{}{int64(100)})
	left.Rows = 1

	// Create right DataFrame with same column name
	right := NewZeaFrame()
	right.AddColumn("id", Int64Type, []interface{}{int64(1)})
	right.AddColumn("value", Int64Type, []interface{}{int64(200)})
	right.Rows = 1

	// Perform join
	result, err := left.Join(right, []string{"id"}, JoinInner)
	if err != nil {
		t.Fatalf("Join failed: %v", err)
	}

	// Verify right column was renamed
	foundRightValue := false
	for _, col := range result.Columns {
		if col.Name == "value_right" {
			foundRightValue = true
			break
		}
	}
	if !foundRightValue {
		t.Errorf("Expected column name collision to be resolved with _right suffix")
	}
}

func TestJoinOneToMany(t *testing.T) {
	// Create left DataFrame (one customer)
	left := NewZeaFrame()
	left.AddColumn("cust_id", Int64Type, []interface{}{int64(1)})
	left.AddColumn("name", StringType, []interface{}{"Alice"})
	left.Rows = 1

	// Create right DataFrame (multiple orders for same customer)
	right := NewZeaFrame()
	right.AddColumn("cust_id", Int64Type, []interface{}{int64(1), int64(1), int64(1)})
	right.AddColumn("order_id", Int64Type, []interface{}{int64(101), int64(102), int64(103)})
	right.Rows = 3

	// Perform join
	result, err := left.Join(right, []string{"cust_id"}, JoinInner)
	if err != nil {
		t.Fatalf("Join failed: %v", err)
	}

	// Should create 3 rows (one for each order)
	if result.Rows != 3 {
		t.Errorf("Expected 3 rows for 1:N join, got %d", result.Rows)
	}

	// All rows should have the same customer name
	nameCol := result.Columns[1]
	for i := 0; i < result.Rows; i++ {
		if nameCol.Data[i] != "Alice" {
			t.Errorf("Expected all rows to have name=Alice, got %v at row %d", nameCol.Data[i], i)
		}
	}
}
