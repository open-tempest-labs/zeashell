package zeaframe

import (
	"fmt"
	"sort"
	"strings"
)

// ColumnType represents the type of data in a column
type ColumnType int

const (
	StringType ColumnType = iota
	Int64Type
	Float64Type
	BoolType
	MultiType
)

// Column represents a typed column of data
type Column struct {
	Name   string
	Type   ColumnType
	Data   []interface{}
	Nulls  []bool
}

// ZeaFrame is a columnar data structure
type ZeaFrame struct {
	Columns []*Column
	Rows    int
}

// NewZeaFrame creates a new empty ZeaFrame
func NewZeaFrame() *ZeaFrame {
	return &ZeaFrame{
		Columns: make([]*Column, 0),
		Rows:    0,
	}
}

// AddColumn adds a new column to the ZeaFrame
func (zf *ZeaFrame) AddColumn(name string, colType ColumnType, data []interface{}) error {
	if len(data) > 0 && zf.Rows > 0 && len(data) != zf.Rows {
		return fmt.Errorf("column length %d does not match zeaframe rows %d", len(data), zf.Rows)
	}

	col := &Column{
		Name:  name,
		Type:  colType,
		Data:  data,
		Nulls: make([]bool, len(data)),
	}

	zf.Columns = append(zf.Columns, col)
	if len(data) > zf.Rows {
		zf.Rows = len(data)
	}

	return nil
}

// GetColumn returns a column by name
func (zf *ZeaFrame) GetColumn(name string) (*Column, error) {
	for _, col := range zf.Columns {
		if col.Name == name {
			return col, nil
		}
	}
	return nil, fmt.Errorf("column %s not found", name)
}

// GetColumnIndex returns the index of a column by name
func (zf *ZeaFrame) GetColumnIndex(name string) int {
	for i, col := range zf.Columns {
		if col.Name == name {
			return i
		}
	}
	return -1
}

// Select returns a new ZeaFrame with only the specified columns
func (zf *ZeaFrame) Select(columns ...string) (*ZeaFrame, error) {
	result := NewZeaFrame()
	result.Rows = zf.Rows

	for _, colName := range columns {
		col, err := zf.GetColumn(colName)
		if err != nil {
			return nil, err
		}

		// Deep copy the column
		newData := make([]interface{}, len(col.Data))
		copy(newData, col.Data)
		newNulls := make([]bool, len(col.Nulls))
		copy(newNulls, col.Nulls)

		newCol := &Column{
			Name:  col.Name,
			Type:  col.Type,
			Data:  newData,
			Nulls: newNulls,
		}
		result.Columns = append(result.Columns, newCol)
	}

	return result, nil
}

// Filter returns a new ZeaFrame with rows that match the filter expression
func (zf *ZeaFrame) Filter(expr *Expression) (*ZeaFrame, error) {
	result := NewZeaFrame()

	// Initialize columns
	for _, col := range zf.Columns {
		newCol := &Column{
			Name:  col.Name,
			Type:  col.Type,
			Data:  make([]interface{}, 0),
			Nulls: make([]bool, 0),
		}
		result.Columns = append(result.Columns, newCol)
	}

	// Evaluate expression for each row
	for rowIdx := 0; rowIdx < zf.Rows; rowIdx++ {
		match, err := expr.Evaluate(zf, rowIdx)
		if err != nil {
			return nil, err
		}

		if match {
			// Copy this row to result
			for i, col := range zf.Columns {
				result.Columns[i].Data = append(result.Columns[i].Data, col.Data[rowIdx])
				result.Columns[i].Nulls = append(result.Columns[i].Nulls, col.Nulls[rowIdx])
			}
			result.Rows++
		}
	}

	return result, nil
}

// SortOrder specifies sort direction
type SortOrder int

const (
	Ascending SortOrder = iota
	Descending
)

// SortColumn specifies a column and its sort order
type SortColumn struct {
	Name  string
	Order SortOrder
}

// Sort sorts the ZeaFrame by one or more columns
// Returns a new sorted ZeaFrame
func (zf *ZeaFrame) Sort(sortCols ...SortColumn) (*ZeaFrame, error) {
	if len(sortCols) == 0 {
		return nil, fmt.Errorf("at least one sort column required")
	}

	// Validate columns exist
	colIndices := make([]int, len(sortCols))
	for i, sc := range sortCols {
		idx := zf.GetColumnIndex(sc.Name)
		if idx == -1 {
			return nil, fmt.Errorf("column %s not found", sc.Name)
		}
		colIndices[i] = idx
	}

	// Create index array
	indices := make([]int, zf.Rows)
	for i := range indices {
		indices[i] = i
	}

	// Sort indices based on column values
	sort.SliceStable(indices, func(i, j int) bool {
		row1 := indices[i]
		row2 := indices[j]

		// Compare by each sort column in order
		for k, sc := range sortCols {
			colIdx := colIndices[k]
			col := zf.Columns[colIdx]

			val1 := col.Data[row1]
			val2 := col.Data[row2]

			cmp := compareForSort(val1, val2)

			if cmp != 0 {
				if sc.Order == Descending {
					return cmp > 0
				}
				return cmp < 0
			}
			// If equal, continue to next sort column
		}
		return false // All columns equal
	})

	// Build result ZeaFrame with sorted rows
	result := NewZeaFrame()
	for _, col := range zf.Columns {
		newCol := &Column{
			Name:  col.Name,
			Type:  col.Type,
			Data:  make([]interface{}, zf.Rows),
			Nulls: make([]bool, zf.Rows),
		}
		result.Columns = append(result.Columns, newCol)
	}
	result.Rows = zf.Rows

	// Copy data in sorted order
	for newIdx, oldIdx := range indices {
		for i, col := range zf.Columns {
			result.Columns[i].Data[newIdx] = col.Data[oldIdx]
			result.Columns[i].Nulls[newIdx] = col.Nulls[oldIdx]
		}
	}

	return result, nil
}

// compareForSort compares two values for sorting
// Returns: -1 if val1 < val2, 0 if equal, 1 if val1 > val2
func compareForSort(val1, val2 interface{}) int {
	// Try numeric comparison first
	f1, err1 := toFloat64(val1)
	f2, err2 := toFloat64(val2)

	if err1 == nil && err2 == nil {
		if f1 < f2 {
			return -1
		}
		if f1 > f2 {
			return 1
		}
		return 0
	}

	// Fall back to string comparison
	s1 := fmt.Sprintf("%v", val1)
	s2 := fmt.Sprintf("%v", val2)

	if s1 < s2 {
		return -1
	}
	if s1 > s2 {
		return 1
	}
	return 0
}

// GroupByResult represents the result of a groupby operation
type GroupByResult struct {
	zf      *ZeaFrame
	groupBy []string
}

// GroupBy groups the ZeaFrame by the specified columns
func (zf *ZeaFrame) GroupBy(columns ...string) *GroupByResult {
	return &GroupByResult{
		zf:      zf,
		groupBy: columns,
	}
}

// Agg performs aggregation on grouped data
func (gbr *GroupByResult) Agg(aggregations map[string]string) (*ZeaFrame, error) {
	zf := gbr.zf

	// Build group keys
	groups := make(map[string][]int) // string key -> row indices

	for rowIdx := 0; rowIdx < zf.Rows; rowIdx++ {
		keyParts := make([]string, 0)
		for _, colName := range gbr.groupBy {
			col, _ := zf.GetColumn(colName)
			keyParts = append(keyParts, fmt.Sprintf("%v", col.Data[rowIdx]))
		}
		key := strings.Join(keyParts, "|")
		groups[key] = append(groups[key], rowIdx)
	}

	// Create result ZeaFrame
	result := NewZeaFrame()

	// Add groupby columns
	for _, colName := range gbr.groupBy {
		srcCol, _ := zf.GetColumn(colName)
		newCol := &Column{
			Name:  colName,
			Type:  srcCol.Type,
			Data:  make([]interface{}, 0),
			Nulls: make([]bool, 0),
		}
		result.Columns = append(result.Columns, newCol)
	}

	// Add aggregation columns
	for colName, aggFunc := range aggregations {
		var colType ColumnType
		switch aggFunc {
		case "sum", "avg", "min", "max":
			colType = Float64Type
		case "count":
			colType = Int64Type
		default:
			colType = StringType
		}

		newCol := &Column{
			Name:  fmt.Sprintf("%s_%s", colName, aggFunc),
			Type:  colType,
			Data:  make([]interface{}, 0),
			Nulls: make([]bool, 0),
		}
		result.Columns = append(result.Columns, newCol)
	}

	// Sort group keys for deterministic output
	groupKeys := make([]string, 0, len(groups))
	for key := range groups {
		groupKeys = append(groupKeys, key)
	}
	sort.Strings(groupKeys)

	// Process each group
	for _, key := range groupKeys {
		rowIndices := groups[key]

		// Add groupby column values
		for i, colName := range gbr.groupBy {
			col, _ := zf.GetColumn(colName)
			result.Columns[i].Data = append(result.Columns[i].Data, col.Data[rowIndices[0]])
			result.Columns[i].Nulls = append(result.Columns[i].Nulls, false)
		}

		// Compute aggregations
		aggIdx := len(gbr.groupBy)
		for colName, aggFunc := range aggregations {
			col, err := zf.GetColumn(colName)
			if err != nil && aggFunc != "count" {
				return nil, err
			}

			aggValue, err := computeAggregation(col, rowIndices, aggFunc)
			if err != nil {
				return nil, err
			}

			result.Columns[aggIdx].Data = append(result.Columns[aggIdx].Data, aggValue)
			result.Columns[aggIdx].Nulls = append(result.Columns[aggIdx].Nulls, false)
			aggIdx++
		}

		result.Rows++
	}

	return result, nil
}

// computeAggregation computes an aggregation function on a column for specific rows
func computeAggregation(col *Column, rowIndices []int, aggFunc string) (interface{}, error) {
	switch aggFunc {
	case "count":
		return int64(len(rowIndices)), nil

	case "sum":
		var sum float64
		for _, idx := range rowIndices {
			val, err := toFloat64(col.Data[idx])
			if err == nil {
				sum += val
			}
		}
		return sum, nil

	case "avg":
		var sum float64
		count := 0
		for _, idx := range rowIndices {
			val, err := toFloat64(col.Data[idx])
			if err == nil {
				sum += val
				count++
			}
		}
		if count == 0 {
			return 0.0, nil
		}
		return sum / float64(count), nil

	case "min":
		var min float64
		initialized := false
		for _, idx := range rowIndices {
			val, err := toFloat64(col.Data[idx])
			if err == nil {
				if !initialized || val < min {
					min = val
					initialized = true
				}
			}
		}
		return min, nil

	case "max":
		var max float64
		initialized := false
		for _, idx := range rowIndices {
			val, err := toFloat64(col.Data[idx])
			if err == nil {
				if !initialized || val > max {
					max = val
					initialized = true
				}
			}
		}
		return max, nil

	default:
		return nil, fmt.Errorf("unsupported aggregation function: %s", aggFunc)
	}
}

// ColumnNames returns the names of all columns
func (zf *ZeaFrame) ColumnNames() []string {
	names := make([]string, len(zf.Columns))
	for i, col := range zf.Columns {
		names[i] = col.Name
	}
	return names
}

// Helper function to convert values to float64
func toFloat64(val interface{}) (float64, error) {
	switch v := val.(type) {
	case float64:
		return v, nil
	case int64:
		return float64(v), nil
	case int:
		return float64(v), nil
	case string:
		var f float64
		_, err := fmt.Sscanf(v, "%f", &f)
		return f, err
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", val)
	}
}
