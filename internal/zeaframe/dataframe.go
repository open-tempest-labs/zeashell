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

// Join performs an inner join with another ZeaFrame
func (zf *ZeaFrame) Join(other *ZeaFrame, onColumn string) (*ZeaFrame, error) {
	// Build index on join column in other ZeaFrame
	otherCol, err := other.GetColumn(onColumn)
	if err != nil {
		return nil, err
	}

	otherIndex := make(map[interface{}][]int)
	for i := 0; i < other.Rows; i++ {
		key := otherCol.Data[i]
		otherIndex[key] = append(otherIndex[key], i)
	}

	// Create result ZeaFrame
	result := NewZeaFrame()

	// Add all columns from left ZeaFrame
	for _, col := range zf.Columns {
		newCol := &Column{
			Name:  col.Name,
			Type:  col.Type,
			Data:  make([]interface{}, 0),
			Nulls: make([]bool, 0),
		}
		result.Columns = append(result.Columns, newCol)
	}

	// Add columns from right ZeaFrame (except join key)
	for _, col := range other.Columns {
		if col.Name == onColumn {
			continue
		}
		newCol := &Column{
			Name:  col.Name,
			Type:  col.Type,
			Data:  make([]interface{}, 0),
			Nulls: make([]bool, 0),
		}
		result.Columns = append(result.Columns, newCol)
	}

	// Perform join
	leftCol, _ := zf.GetColumn(onColumn)
	for leftIdx := 0; leftIdx < zf.Rows; leftIdx++ {
		key := leftCol.Data[leftIdx]
		rightIndices, found := otherIndex[key]

		if found {
			for _, rightIdx := range rightIndices {
				// Add left row
				for i, col := range zf.Columns {
					result.Columns[i].Data = append(result.Columns[i].Data, col.Data[leftIdx])
					result.Columns[i].Nulls = append(result.Columns[i].Nulls, col.Nulls[leftIdx])
				}

				// Add right row (skip join key)
				resultColIdx := len(zf.Columns)
				for _, col := range other.Columns {
					if col.Name == onColumn {
						continue
					}
					result.Columns[resultColIdx].Data = append(result.Columns[resultColIdx].Data, col.Data[rightIdx])
					result.Columns[resultColIdx].Nulls = append(result.Columns[resultColIdx].Nulls, col.Nulls[rightIdx])
					resultColIdx++
				}

				result.Rows++
			}
		}
	}

	return result, nil
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
