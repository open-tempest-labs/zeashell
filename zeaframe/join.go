package zeaframe

import (
	"fmt"
	"strings"
)

// JoinType represents the type of join operation
type JoinType int

const (
	JoinInner JoinType = iota
	JoinLeft
	JoinRight
	JoinFull
)

// String returns the string representation of a JoinType
func (jt JoinType) String() string {
	switch jt {
	case JoinInner:
		return "inner"
	case JoinLeft:
		return "left"
	case JoinRight:
		return "right"
	case JoinFull:
		return "full"
	default:
		return "unknown"
	}
}

// Join performs a relational join between two DataFrames
func (zf *ZeaFrame) Join(other *ZeaFrame, keys []string, joinType JoinType) (*ZeaFrame, error) {
	if len(keys) == 0 {
		return nil, fmt.Errorf("at least one join key required")
	}

	// Validate that all keys exist in both DataFrames
	for _, key := range keys {
		if zf.GetColumnIndex(key) == -1 {
			return nil, fmt.Errorf("join key %s not found in left DataFrame", key)
		}
		if other.GetColumnIndex(key) == -1 {
			return nil, fmt.Errorf("join key %s not found in right DataFrame", key)
		}
	}

	// Build hash table on right side (other)
	hashTable := buildHashTable(other, keys)

	// Create result DataFrame
	result := NewZeaFrame()

	// Add columns from left side
	leftColNames := make([]string, len(zf.Columns))
	for i, col := range zf.Columns {
		leftColNames[i] = col.Name
		newCol := &Column{
			Name:  col.Name,
			Type:  col.Type,
			Data:  make([]interface{}, 0),
			Nulls: make([]bool, 0),
		}
		result.Columns = append(result.Columns, newCol)
	}

	// Add columns from right side (excluding join keys to avoid duplication)
	rightColNames := make([]string, 0)
	rightColIndices := make([]int, 0)
	for i, col := range other.Columns {
		// Skip columns that are join keys (already in left)
		if contains(keys, col.Name) {
			continue
		}

		// Handle column name collisions
		colName := col.Name
		if zf.GetColumnIndex(colName) != -1 {
			colName = colName + "_right"
		}

		rightColNames = append(rightColNames, colName)
		rightColIndices = append(rightColIndices, i)
		newCol := &Column{
			Name:  colName,
			Type:  col.Type,
			Data:  make([]interface{}, 0),
			Nulls: make([]bool, 0),
		}
		result.Columns = append(result.Columns, newCol)
	}

	// Track matched right rows for full/right joins
	matchedRight := make(map[int]bool)

	// Probe hash table with left side rows
	for leftRowIdx := 0; leftRowIdx < zf.Rows; leftRowIdx++ {
		// Build key for this left row
		keyVals := make([]string, len(keys))
		for i, key := range keys {
			colIdx := zf.GetColumnIndex(key)
			keyVals[i] = fmt.Sprintf("%v", zf.Columns[colIdx].Data[leftRowIdx])
		}
		hashKey := strings.Join(keyVals, "\x00")

		// Look up matching right rows
		rightRows, found := hashTable[hashKey]

		if found && len(rightRows) > 0 {
			// Match found - emit rows for each matching right row
			for _, rightRowIdx := range rightRows {
				matchedRight[rightRowIdx] = true
				emitJoinedRow(result, zf, other, leftRowIdx, rightRowIdx, rightColIndices)
			}
		} else if joinType == JoinLeft || joinType == JoinFull {
			// No match, but left join - emit left row with NULLs for right
			emitLeftOnlyRow(result, zf, leftRowIdx, len(rightColNames))
		}
		// For inner join, skip unmatched left rows
	}

	// For right/full joins, emit unmatched right rows
	if joinType == JoinRight || joinType == JoinFull {
		for rightRowIdx := 0; rightRowIdx < other.Rows; rightRowIdx++ {
			if !matchedRight[rightRowIdx] {
				emitRightOnlyRow(result, zf, other, rightRowIdx, rightColIndices)
			}
		}
	}

	return result, nil
}

// buildHashTable creates a hash table mapping join key values to row indices
func buildHashTable(df *ZeaFrame, keys []string) map[string][]int {
	hashTable := make(map[string][]int)

	for rowIdx := 0; rowIdx < df.Rows; rowIdx++ {
		// Build composite key
		keyVals := make([]string, len(keys))
		for i, key := range keys {
			colIdx := df.GetColumnIndex(key)
			keyVals[i] = fmt.Sprintf("%v", df.Columns[colIdx].Data[rowIdx])
		}
		hashKey := strings.Join(keyVals, "\x00")

		// Add row index to hash table
		hashTable[hashKey] = append(hashTable[hashKey], rowIdx)
	}

	return hashTable
}

// emitJoinedRow emits a row with data from both left and right DataFrames
func emitJoinedRow(result, left, right *ZeaFrame, leftRowIdx, rightRowIdx int, rightColIndices []int) {
	// Add left columns
	for _, col := range left.Columns {
		resultCol := result.Columns[result.GetColumnIndex(col.Name)]
		resultCol.Data = append(resultCol.Data, col.Data[leftRowIdx])
		resultCol.Nulls = append(resultCol.Nulls, col.Nulls[leftRowIdx])
	}

	// Add right columns (excluding join keys)
	for _, rightColIdx := range rightColIndices {
		rightCol := right.Columns[rightColIdx]

		// Find corresponding result column (may have _right suffix)
		var resultCol *Column
		for _, rc := range result.Columns {
			if rc.Name == rightCol.Name || rc.Name == rightCol.Name+"_right" {
				if result.GetColumnIndex(rc.Name) >= len(left.Columns) {
					resultCol = rc
					break
				}
			}
		}

		if resultCol != nil {
			resultCol.Data = append(resultCol.Data, rightCol.Data[rightRowIdx])
			resultCol.Nulls = append(resultCol.Nulls, rightCol.Nulls[rightRowIdx])
		}
	}

	result.Rows++
}

// emitLeftOnlyRow emits a row with left data and NULLs for right columns
func emitLeftOnlyRow(result, left *ZeaFrame, leftRowIdx int, numRightCols int) {
	// Add left columns
	for _, col := range left.Columns {
		resultCol := result.Columns[result.GetColumnIndex(col.Name)]
		resultCol.Data = append(resultCol.Data, col.Data[leftRowIdx])
		resultCol.Nulls = append(resultCol.Nulls, col.Nulls[leftRowIdx])
	}

	// Add NULL values for right columns
	for i := len(left.Columns); i < len(result.Columns); i++ {
		resultCol := result.Columns[i]
		resultCol.Data = append(resultCol.Data, "")
		resultCol.Nulls = append(resultCol.Nulls, true)
	}

	result.Rows++
}

// emitRightOnlyRow emits a row with NULLs for left data and right data
func emitRightOnlyRow(result, left, right *ZeaFrame, rightRowIdx int, rightColIndices []int) {
	// Add NULL values for left columns
	for _, col := range left.Columns {
		resultCol := result.Columns[result.GetColumnIndex(col.Name)]
		resultCol.Data = append(resultCol.Data, "")
		resultCol.Nulls = append(resultCol.Nulls, true)
	}

	// Add right columns
	for _, rightColIdx := range rightColIndices {
		rightCol := right.Columns[rightColIdx]

		// Find corresponding result column
		var resultCol *Column
		for _, rc := range result.Columns {
			if rc.Name == rightCol.Name || rc.Name == rightCol.Name+"_right" {
				if result.GetColumnIndex(rc.Name) >= len(left.Columns) {
					resultCol = rc
					break
				}
			}
		}

		if resultCol != nil {
			resultCol.Data = append(resultCol.Data, rightCol.Data[rightRowIdx])
			resultCol.Nulls = append(resultCol.Nulls, rightCol.Nulls[rightRowIdx])
		}
	}

	result.Rows++
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
