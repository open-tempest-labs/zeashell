package zeaframe

import (
	"fmt"
	"strings"
)

// Pivot transforms long format to wide format
// indexCols: columns to use as row identifiers
// columnCol: column whose values become new column names
// valueCol: column whose values populate the new columns
func (zf *ZeaFrame) Pivot(indexCols []string, columnCol string, valueCol string) (*ZeaFrame, error) {
	// Validate inputs
	if len(indexCols) == 0 {
		return nil, fmt.Errorf("at least one index column required")
	}

	for _, col := range indexCols {
		if zf.GetColumnIndex(col) == -1 {
			return nil, fmt.Errorf("index column %s not found", col)
		}
	}

	if zf.GetColumnIndex(columnCol) == -1 {
		return nil, fmt.Errorf("column column %s not found", columnCol)
	}

	if zf.GetColumnIndex(valueCol) == -1 {
		return nil, fmt.Errorf("value column %s not found", valueCol)
	}

	// Get column indices
	columnColIdx := zf.GetColumnIndex(columnCol)
	valueColIdx := zf.GetColumnIndex(valueCol)

	// Collect unique values from columnCol (these become new column names)
	uniqueColumns := make(map[string]bool)
	columnOrder := make([]string, 0)
	for rowIdx := 0; rowIdx < zf.Rows; rowIdx++ {
		colName := fmt.Sprintf("%v", zf.Columns[columnColIdx].Data[rowIdx])
		if !uniqueColumns[colName] {
			uniqueColumns[colName] = true
			columnOrder = append(columnOrder, colName)
		}
	}

	// Build index of index column combinations to row indices
	indexMap := make(map[string]int)
	indexRows := make([][]interface{}, 0)

	for rowIdx := 0; rowIdx < zf.Rows; rowIdx++ {
		// Build composite index key
		indexKey := make([]string, len(indexCols))
		indexVals := make([]interface{}, len(indexCols))
		for i, idxCol := range indexCols {
			colIdx := zf.GetColumnIndex(idxCol)
			val := zf.Columns[colIdx].Data[rowIdx]
			indexKey[i] = fmt.Sprintf("%v", val)
			indexVals[i] = val
		}
		key := strings.Join(indexKey, "\x00")

		if _, exists := indexMap[key]; !exists {
			indexMap[key] = len(indexRows)
			indexRows = append(indexRows, indexVals)
		}
	}

	// Create result DataFrame
	result := NewZeaFrame()

	// Add index columns
	for _, idxCol := range indexCols {
		srcCol := zf.Columns[zf.GetColumnIndex(idxCol)]
		newCol := &Column{
			Name:  srcCol.Name,
			Type:  srcCol.Type,
			Data:  make([]interface{}, len(indexRows)),
			Nulls: make([]bool, len(indexRows)),
		}
		result.Columns = append(result.Columns, newCol)
	}

	// Add pivoted columns
	for _, colName := range columnOrder {
		newCol := &Column{
			Name:  colName,
			Type:  zf.Columns[valueColIdx].Type,
			Data:  make([]interface{}, len(indexRows)),
			Nulls: make([]bool, len(indexRows)),
		}
		// Initialize with empty values
		for i := range newCol.Data {
			newCol.Data[i] = ""
			newCol.Nulls[i] = true
		}
		result.Columns = append(result.Columns, newCol)
	}

	result.Rows = len(indexRows)

	// Fill in index column values
	for i, indexVals := range indexRows {
		for j, val := range indexVals {
			result.Columns[j].Data[i] = val
			result.Columns[j].Nulls[i] = false
		}
	}

	// Fill in pivoted values
	for rowIdx := 0; rowIdx < zf.Rows; rowIdx++ {
		// Find which index row this belongs to
		indexKey := make([]string, len(indexCols))
		for i, idxCol := range indexCols {
			colIdx := zf.GetColumnIndex(idxCol)
			indexKey[i] = fmt.Sprintf("%v", zf.Columns[colIdx].Data[rowIdx])
		}
		key := strings.Join(indexKey, "\x00")
		resultRowIdx := indexMap[key]

		// Find which pivot column this value belongs to
		colName := fmt.Sprintf("%v", zf.Columns[columnColIdx].Data[rowIdx])
		pivotColIdx := -1
		for i := len(indexCols); i < len(result.Columns); i++ {
			if result.Columns[i].Name == colName {
				pivotColIdx = i
				break
			}
		}

		if pivotColIdx != -1 {
			// Set the value
			result.Columns[pivotColIdx].Data[resultRowIdx] = zf.Columns[valueColIdx].Data[rowIdx]
			result.Columns[pivotColIdx].Nulls[resultRowIdx] = false
		}
	}

	return result, nil
}

// Unpivot transforms wide format to long format
// idCols: columns to preserve as-is
// valueCols: columns to unpivot into rows
// nameCol: name for the new column containing the original column names
// valueCol: name for the new column containing the values
func (zf *ZeaFrame) Unpivot(idCols []string, valueCols []string, nameCol string, valueCol string) (*ZeaFrame, error) {
	// Validate inputs
	if len(valueCols) == 0 {
		return nil, fmt.Errorf("at least one value column required")
	}

	for _, col := range idCols {
		if zf.GetColumnIndex(col) == -1 {
			return nil, fmt.Errorf("id column %s not found", col)
		}
	}

	for _, col := range valueCols {
		if zf.GetColumnIndex(col) == -1 {
			return nil, fmt.Errorf("value column %s not found", col)
		}
	}

	// Create result DataFrame
	result := NewZeaFrame()

	// Add ID columns
	for _, idCol := range idCols {
		srcCol := zf.Columns[zf.GetColumnIndex(idCol)]
		newCol := &Column{
			Name:  srcCol.Name,
			Type:  srcCol.Type,
			Data:  make([]interface{}, 0),
			Nulls: make([]bool, 0),
		}
		result.Columns = append(result.Columns, newCol)
	}

	// Add name column (will contain original column names)
	nameColumn := &Column{
		Name:  nameCol,
		Type:  StringType,
		Data:  make([]interface{}, 0),
		Nulls: make([]bool, 0),
	}
	result.Columns = append(result.Columns, nameColumn)

	// Add value column
	// Determine type from first value column
	firstValueColIdx := zf.GetColumnIndex(valueCols[0])
	valueColumn := &Column{
		Name:  valueCol,
		Type:  zf.Columns[firstValueColIdx].Type,
		Data:  make([]interface{}, 0),
		Nulls: make([]bool, 0),
	}
	result.Columns = append(result.Columns, valueColumn)

	// Process each row in source DataFrame
	for rowIdx := 0; rowIdx < zf.Rows; rowIdx++ {
		// For each value column, create a new row in result
		for _, valCol := range valueCols {
			valColIdx := zf.GetColumnIndex(valCol)

			// Copy ID column values
			for _, idCol := range idCols {
				idColIdx := zf.GetColumnIndex(idCol)
				resultColIdx := result.GetColumnIndex(idCol)
				result.Columns[resultColIdx].Data = append(
					result.Columns[resultColIdx].Data,
					zf.Columns[idColIdx].Data[rowIdx],
				)
				result.Columns[resultColIdx].Nulls = append(
					result.Columns[resultColIdx].Nulls,
					zf.Columns[idColIdx].Nulls[rowIdx],
				)
			}

			// Add column name to nameCol
			nameColIdx := result.GetColumnIndex(nameCol)
			result.Columns[nameColIdx].Data = append(
				result.Columns[nameColIdx].Data,
				valCol,
			)
			result.Columns[nameColIdx].Nulls = append(
				result.Columns[nameColIdx].Nulls,
				false,
			)

			// Add value to valueCol
			valueColIdx := result.GetColumnIndex(valueCol)
			result.Columns[valueColIdx].Data = append(
				result.Columns[valueColIdx].Data,
				zf.Columns[valColIdx].Data[rowIdx],
			)
			result.Columns[valueColIdx].Nulls = append(
				result.Columns[valueColIdx].Nulls,
				zf.Columns[valColIdx].Nulls[rowIdx],
			)

			result.Rows++
		}
	}

	return result, nil
}
