package zeaframe

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// FromCSV creates a ZeaFrame from a CSV reader
func FromCSV(reader io.Reader) (*ZeaFrame, error) {
	return fromDelimitedReader(reader, ',')
}

// FromTSV creates a ZeaFrame from a TSV reader
func FromTSV(reader io.Reader) (*ZeaFrame, error) {
	return fromDelimitedReader(reader, '\t')
}

// fromDelimitedReader creates a ZeaFrame from a delimited reader
func fromDelimitedReader(reader io.Reader, delimiter rune) (*ZeaFrame, error) {
	csvReader := csv.NewReader(reader)
	csvReader.Comma = delimiter
	csvReader.TrimLeadingSpace = true
	csvReader.LazyQuotes = true

	// Read header
	headers, err := csvReader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	// Initialize columns
	zf := NewZeaFrame()
	columnData := make([][]interface{}, len(headers))
	for i := range columnData {
		columnData[i] = make([]interface{}, 0)
	}

	// Read all rows
	rowCount := 0
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read row %d: %w", rowCount+1, err)
		}

		// Ensure record has same number of fields as headers
		for len(record) < len(headers) {
			record = append(record, "")
		}

		for i, value := range record {
			if i < len(columnData) {
				columnData[i] = append(columnData[i], value)
			}
		}
		rowCount++
	}

	// Infer types and create columns
	for i, header := range headers {
		colType, typedData := inferType(columnData[i])
		err := zf.AddColumn(header, colType, typedData)
		if err != nil {
			return nil, err
		}
	}

	return zf, nil
}

// FromJSONL creates a ZeaFrame from a JSONL (JSON Lines) reader
func FromJSONL(reader io.Reader) (*ZeaFrame, error) {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 1024*1024), 10*1024*1024) // 10MB max line

	var records []map[string]interface{}
	columnNames := make(map[string]bool)

	// Read all records and collect column names
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var record map[string]interface{}
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			return nil, fmt.Errorf("failed to parse JSON on line %d: %w", lineNum, err)
		}

		records = append(records, record)
		for key := range record {
			columnNames[key] = true
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading JSONL: %w", err)
	}

	// Create ordered list of column names
	columns := make([]string, 0, len(columnNames))
	for name := range columnNames {
		columns = append(columns, name)
	}

	// Build column data
	zf := NewZeaFrame()
	columnData := make(map[string][]interface{})
	for _, col := range columns {
		columnData[col] = make([]interface{}, 0, len(records))
	}

	for _, record := range records {
		for _, col := range columns {
			val, exists := record[col]
			if !exists {
				val = ""
			}
			columnData[col] = append(columnData[col], val)
		}
	}

	// Infer types and add columns
	for _, col := range columns {
		colType, typedData := inferType(columnData[col])
		err := zf.AddColumn(col, colType, typedData)
		if err != nil {
			return nil, err
		}
	}

	return zf, nil
}

// WriteCSV writes the ZeaFrame to a CSV writer
func (zf *ZeaFrame) WriteCSV(writer io.Writer) error {
	return zf.writeDelimited(writer, ',')
}

// WriteTSV writes the ZeaFrame to a TSV writer
func (zf *ZeaFrame) WriteTSV(writer io.Writer) error {
	return zf.writeDelimited(writer, '\t')
}

// writeDelimited writes the ZeaFrame to a delimited writer
func (zf *ZeaFrame) writeDelimited(writer io.Writer, delimiter rune) error {
	csvWriter := csv.NewWriter(writer)
	csvWriter.Comma = delimiter

	// Write header
	headers := make([]string, len(zf.Columns))
	for i, col := range zf.Columns {
		headers[i] = col.Name
	}
	if err := csvWriter.Write(headers); err != nil {
		return err
	}

	// Write rows
	for rowIdx := 0; rowIdx < zf.Rows; rowIdx++ {
		record := make([]string, len(zf.Columns))
		for i, col := range zf.Columns {
			if rowIdx < len(col.Data) {
				record[i] = fmt.Sprintf("%v", col.Data[rowIdx])
			} else {
				record[i] = ""
			}
		}
		if err := csvWriter.Write(record); err != nil {
			return err
		}
	}

	csvWriter.Flush()
	return csvWriter.Error()
}

// WriteJSONL writes the ZeaFrame to a JSONL writer
func (zf *ZeaFrame) WriteJSONL(writer io.Writer) error {
	for rowIdx := 0; rowIdx < zf.Rows; rowIdx++ {
		record := make(map[string]interface{})
		for _, col := range zf.Columns {
			if rowIdx < len(col.Data) {
				record[col.Name] = col.Data[rowIdx]
			}
		}

		data, err := json.Marshal(record)
		if err != nil {
			return err
		}

		if _, err := writer.Write(data); err != nil {
			return err
		}
		if _, err := writer.Write([]byte("\n")); err != nil {
			return err
		}
	}

	return nil
}

// FromJSON creates a ZeaFrame from a JSON reader
// Supports both array of objects: [{"a":1},{"a":2}]
// and object with array values: {"col1":[1,2], "col2":[3,4]}
func FromJSON(reader io.Reader) (*ZeaFrame, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read JSON: %w", err)
	}

	// Try to parse as array of objects first (most common)
	var records []map[string]interface{}
	if err := json.Unmarshal(data, &records); err == nil {
		return fromJSONRecords(records)
	}

	// Try to parse as object with array values (columnar JSON)
	var columnarData map[string][]interface{}
	if err := json.Unmarshal(data, &columnarData); err == nil {
		return fromJSONColumnar(columnarData)
	}

	// Try as single object (convert to single-row frame)
	var singleRecord map[string]interface{}
	if err := json.Unmarshal(data, &singleRecord); err == nil {
		return fromJSONRecords([]map[string]interface{}{singleRecord})
	}

	return nil, fmt.Errorf("unsupported JSON structure")
}

// fromJSONRecords creates a ZeaFrame from array of records
func fromJSONRecords(records []map[string]interface{}) (*ZeaFrame, error) {
	if len(records) == 0 {
		return NewZeaFrame(), nil
	}

	// Collect all column names
	columnNames := make(map[string]bool)
	for _, record := range records {
		for key := range record {
			columnNames[key] = true
		}
	}

	// Create ordered list of column names
	columns := make([]string, 0, len(columnNames))
	for name := range columnNames {
		columns = append(columns, name)
	}

	// Build column data
	zf := NewZeaFrame()
	columnData := make(map[string][]interface{})
	for _, col := range columns {
		columnData[col] = make([]interface{}, 0, len(records))
	}

	for _, record := range records {
		for _, col := range columns {
			val, exists := record[col]
			if !exists {
				val = ""
			}
			// Handle nested structures by converting to string
			if isNested(val) {
				jsonBytes, _ := json.Marshal(val)
				val = string(jsonBytes)
			}
			columnData[col] = append(columnData[col], val)
		}
	}

	// Infer types and add columns
	for _, col := range columns {
		colType, typedData := inferType(columnData[col])
		err := zf.AddColumn(col, colType, typedData)
		if err != nil {
			return nil, err
		}
	}

	return zf, nil
}

// fromJSONColumnar creates a ZeaFrame from columnar JSON
func fromJSONColumnar(data map[string][]interface{}) (*ZeaFrame, error) {
	zf := NewZeaFrame()

	for colName, colData := range data {
		colType, typedData := inferType(colData)
		err := zf.AddColumn(colName, colType, typedData)
		if err != nil {
			return nil, err
		}
	}

	return zf, nil
}

// isNested checks if a value is a nested structure
func isNested(val interface{}) bool {
	switch val.(type) {
	case map[string]interface{}, []interface{}:
		return true
	default:
		return false
	}
}

// WriteJSON writes the ZeaFrame as a JSON array of objects
func (zf *ZeaFrame) WriteJSON(writer io.Writer) error {
	records := make([]map[string]interface{}, 0, zf.Rows)

	for rowIdx := 0; rowIdx < zf.Rows; rowIdx++ {
		record := make(map[string]interface{})
		for _, col := range zf.Columns {
			if rowIdx < len(col.Data) {
				record[col.Name] = col.Data[rowIdx]
			}
		}
		records = append(records, record)
	}

	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(records)
}

// inferType infers the column type from the data
func inferType(data []interface{}) (ColumnType, []interface{}) {
	if len(data) == 0 {
		return StringType, data
	}

	// Try to determine type by sampling
	hasFloat := false
	hasInt := false
	hasBool := false
	allNumeric := true
	allBool := true

	for _, val := range data {
		str, ok := val.(string)
		if !ok {
			str = fmt.Sprintf("%v", val)
		}

		str = strings.TrimSpace(str)
		if str == "" {
			continue
		}

		// Check bool
		if str == "true" || str == "false" {
			hasBool = true
		} else {
			allBool = false
		}

		// Check int
		if _, err := strconv.ParseInt(str, 10, 64); err == nil {
			hasInt = true
			continue
		}

		// Check float
		if _, err := strconv.ParseFloat(str, 64); err == nil {
			hasFloat = true
			continue
		}

		allNumeric = false
	}

	// Determine type based on analysis
	var colType ColumnType
	if allBool && hasBool {
		colType = BoolType
	} else if allNumeric && hasFloat {
		colType = Float64Type
	} else if allNumeric && hasInt {
		colType = Int64Type
	} else {
		colType = StringType
	}

	// Convert data to appropriate type
	typedData := make([]interface{}, len(data))
	for i, val := range data {
		str, ok := val.(string)
		if !ok {
			str = fmt.Sprintf("%v", val)
		}

		str = strings.TrimSpace(str)
		if str == "" {
			typedData[i] = getZeroValue(colType)
			continue
		}

		switch colType {
		case BoolType:
			typedData[i] = str == "true"
		case Int64Type:
			if intVal, err := strconv.ParseInt(str, 10, 64); err == nil {
				typedData[i] = intVal
			} else {
				typedData[i] = int64(0)
			}
		case Float64Type:
			if floatVal, err := strconv.ParseFloat(str, 64); err == nil {
				typedData[i] = floatVal
			} else {
				typedData[i] = 0.0
			}
		default:
			typedData[i] = str
		}
	}

	return colType, typedData
}

// getZeroValue returns the zero value for a column type
func getZeroValue(colType ColumnType) interface{} {
	switch colType {
	case BoolType:
		return false
	case Int64Type:
		return int64(0)
	case Float64Type:
		return 0.0
	default:
		return ""
	}
}
