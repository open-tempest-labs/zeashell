package zeaframe

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
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

// flattenRecord flattens a nested record into path-based keys
func flattenRecord(record map[string]interface{}, prefix string) map[string]interface{} {
	result := make(map[string]interface{})

	for key, value := range record {
		fullPath := key
		if prefix != "" {
			fullPath = prefix + "." + key
		}

		switch v := value.(type) {
		case map[string]interface{}:
			// Recursively flatten nested maps
			nested := flattenRecord(v, fullPath)
			for nestedKey, nestedVal := range nested {
				result[nestedKey] = nestedVal
			}
		case []interface{}:
			// For arrays, check if they contain maps (expand) or primitives (keep as JSON string)
			if len(v) > 0 {
				if _, isMap := v[0].(map[string]interface{}); isMap {
					// Array of objects - keep as JSON string for now
					// (could be expanded into multiple rows in the future)
					jsonBytes, _ := json.Marshal(v)
					result[fullPath] = string(jsonBytes)
				} else {
					// Array of primitives - keep as JSON string
					jsonBytes, _ := json.Marshal(v)
					result[fullPath] = string(jsonBytes)
				}
			} else {
				// Empty array
				jsonBytes, _ := json.Marshal(v)
				result[fullPath] = string(jsonBytes)
			}
		default:
			// Scalar value
			result[fullPath] = value
		}
	}

	return result
}

// fromJSONRecords creates a ZeaFrame from array of records
func fromJSONRecords(records []map[string]interface{}) (*ZeaFrame, error) {
	if len(records) == 0 {
		return NewZeaFrame(), nil
	}

	// Flatten all records and collect all column names
	flatRecords := make([]map[string]interface{}, len(records))
	columnNames := make(map[string]bool)

	for i, record := range records {
		flatRecords[i] = flattenRecord(record, "")
		for key := range flatRecords[i] {
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
		columnData[col] = make([]interface{}, 0, len(flatRecords))
	}

	for _, record := range flatRecords {
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
	emptyCount := 0
	nonEmptyCount := 0

	for _, val := range data {
		str, ok := val.(string)
		if !ok {
			str = fmt.Sprintf("%v", val)
		}

		str = strings.TrimSpace(str)
		if str == "" {
			emptyCount++
			continue
		}

		nonEmptyCount++

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
	// If more than 50% of values are empty, treat as string to avoid confusing zero values
	var colType ColumnType
	if emptyCount > nonEmptyCount {
		colType = StringType
	} else if allBool && hasBool {
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

// FromXML creates a ZeaFrame from an XML reader
// Supports various XML structures:
// - Array of elements: <root><item>...</item><item>...</item></root>
// - Single object: <root>...</root>
// Attributes and nested elements are preserved
func FromXML(reader io.Reader) (*ZeaFrame, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read XML: %w", err)
	}

	// Parse the XML into a generic map structure
	decoder := xml.NewDecoder(strings.NewReader(string(data)))
	xmlMap, err := parseXMLToMap(decoder)
	if err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}

	// Find the array of records in the XML structure
	records, err := extractXMLRecords(xmlMap)
	if err != nil {
		return nil, err
	}

	// Convert interface{} slice to map slice
	mapRecords := make([]map[string]interface{}, len(records))
	for i, rec := range records {
		if m, ok := rec.(map[string]interface{}); ok {
			mapRecords[i] = m
		} else {
			return nil, fmt.Errorf("invalid record type at index %d", i)
		}
	}

	// Convert records to ZeaFrame using same logic as JSON
	return fromJSONRecords(mapRecords)
}

// parseXMLToMap converts XML to a map[string]interface{} structure
// This is a simplified parser that treats XML more like JSON
func parseXMLToMap(decoder *xml.Decoder) (map[string]interface{}, error) {
	type stackItem struct {
		data map[string]interface{}
		text *strings.Builder
	}

	root := stackItem{data: make(map[string]interface{}), text: &strings.Builder{}}
	stack := []*stackItem{&root}
	var keyStack []string

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		switch elem := token.(type) {
		case xml.StartElement:
			// New element
			key := elem.Name.Local
			newItem := &stackItem{data: make(map[string]interface{}), text: &strings.Builder{}}

			// Add attributes as fields
			for _, attr := range elem.Attr {
				newItem.data[attr.Name.Local] = attr.Value
			}

			keyStack = append(keyStack, key)
			stack = append(stack, newItem)

		case xml.EndElement:
			// Pop from stack
			if len(stack) > 1 {
				completed := stack[len(stack)-1]
				stack = stack[:len(stack)-1]

				key := keyStack[len(keyStack)-1]
				keyStack = keyStack[:len(keyStack)-1]

				parent := stack[len(stack)-1]

				// Determine the value to store
				var value interface{}
				text := strings.TrimSpace(completed.text.String())

				if len(completed.data) == 0 && text != "" {
					// Text-only element
					value = text
				} else if len(completed.data) > 0 && text == "" {
					// Element with children
					value = completed.data
				} else if text != "" {
					// Mixed content - store text as special field
					completed.data["_text"] = text
					value = completed.data
				} else {
					// Empty element
					value = ""
				}

				// If key already exists, convert to array
				if existing, exists := parent.data[key]; exists {
					switch v := existing.(type) {
					case []interface{}:
						parent.data[key] = append(v, value)
					default:
						parent.data[key] = []interface{}{v, value}
					}
				} else {
					parent.data[key] = value
				}
			}

		case xml.CharData:
			// Accumulate text content
			if len(stack) > 0 {
				stack[len(stack)-1].text.WriteString(string(elem))
			}
		}
	}

	return root.data, nil
}

// extractXMLRecords finds the array of records from parsed XML
func extractXMLRecords(xmlMap map[string]interface{}) ([]interface{}, error) {
	// If there's only one key, drill down
	if len(xmlMap) == 1 {
		for _, v := range xmlMap {
			switch val := v.(type) {
			case []interface{}:
				// Found array of records
				return val, nil
			case map[string]interface{}:
				// Check if this contains multiple arrays (peer structures)
				// or a mix of single elements and arrays
				return extractPeerStructures(val)
			}
		}
	}

	// Multiple top-level elements - treat as peer structures
	return extractPeerStructures(xmlMap)
}

// extractPeerStructures handles XML with peer elements at the same level
// For example: <root><gateway>...</gateway><service>...</service><service>...</service></root>
// This recursively expands all arrays to create flat records with path-based columns
func extractPeerStructures(xmlMap map[string]interface{}) ([]interface{}, error) {
	// Recursively expand all arrays at all levels
	expandedRecords := expandAllArrays("", xmlMap)
	if len(expandedRecords) > 0 {
		return expandedRecords, nil
	}

	// No arrays found, treat as single record
	return []interface{}{flattenXMLRecord(xmlMap, "")}, nil
}

// expandAllArrays recursively expands arrays in the XML structure
// Returns records with path-based column names (e.g., "topology.gateway.provider.role")
func expandAllArrays(parentPath string, data interface{}) []interface{} {
	var records []interface{}

	switch val := data.(type) {
	case map[string]interface{}:
		// First, recursively process all non-array children to expand nested structures
		processedMap := make(map[string]interface{})
		var arrayKeys []string
		var expandedArrayRecords [][]interface{}

		for k, v := range val {
			elementPath := k
			if parentPath != "" {
				elementPath = parentPath + "." + k
			}

			if arr, ok := v.([]interface{}); ok {
				// This is an array - expand each element
				arrayKeys = append(arrayKeys, k)
				var arrayRecords []interface{}
				for _, item := range arr {
					expanded := expandAllArrays(elementPath, item)
					arrayRecords = append(arrayRecords, expanded...)
				}
				expandedArrayRecords = append(expandedArrayRecords, arrayRecords)
			} else if nestedMap, ok := v.(map[string]interface{}); ok {
				// Nested map - recursively expand it (it might contain arrays)
				nestedExpanded := expandAllArrays(elementPath, nestedMap)
				// If the nested map expands to records, we need to handle them
				// For now, if it's a single record, merge its fields into processedMap
				if len(nestedExpanded) == 1 {
					if recMap, ok := nestedExpanded[0].(map[string]interface{}); ok {
						for fk, fv := range recMap {
							processedMap[fk] = fv
						}
					}
				} else if len(nestedExpanded) > 1 {
					// Multiple records from nested expansion - treat as array expansion
					arrayKeys = append(arrayKeys, k)
					expandedArrayRecords = append(expandedArrayRecords, nestedExpanded)
				}
			} else {
				// Scalar value - add to processed map with full path
				processedMap[elementPath] = flattenXMLValue(v)
			}
		}

		if len(arrayKeys) == 0 {
			// No arrays - return the flattened map as a single record
			if len(processedMap) > 0 {
				return []interface{}{processedMap}
			}
			return []interface{}{}
		}

		// We have arrays - need to create one record per array element
		// Each array element record gets merged with the base scalar fields
		for _, arrayRecords := range expandedArrayRecords {
			for _, arrayRec := range arrayRecords {
				mergedRecord := make(map[string]interface{})
				// Add base scalar fields
				for k, v := range processedMap {
					mergedRecord[k] = v
				}
				// Merge in the array element fields
				if recMap, ok := arrayRec.(map[string]interface{}); ok {
					for k, v := range recMap {
						mergedRecord[k] = v
					}
				}
				records = append(records, mergedRecord)
			}
		}

	case []interface{}:
		// Array at top level
		for _, item := range val {
			expanded := expandAllArrays(parentPath, item)
			records = append(records, expanded...)
		}

	default:
		// Scalar value
		record := make(map[string]interface{})
		if parentPath != "" {
			record[parentPath] = flattenXMLValue(val)
		} else {
			record["value"] = flattenXMLValue(val)
		}
		records = append(records, record)
	}

	return records
}

// flattenXMLRecord flattens an XML map into path-based column names
func flattenXMLRecord(m map[string]interface{}, prefix string) map[string]interface{} {
	result := make(map[string]interface{})

	for k, v := range m {
		if k == "__text" {
			// Handle text content
			if prefix != "" {
				result[prefix] = flattenXMLValue(v)
			} else {
				result["__text"] = flattenXMLValue(v)
			}
			continue
		}

		fullPath := k
		if prefix != "" {
			fullPath = prefix + "." + k
		}

		switch val := v.(type) {
		case map[string]interface{}:
			// Recursively flatten nested maps
			nested := flattenXMLRecord(val, fullPath)
			for nestedKey, nestedVal := range nested {
				result[nestedKey] = nestedVal
			}
		case []interface{}:
			// Arrays are kept as JSON strings (consistent with JSON handling)
			jsonBytes, _ := json.Marshal(val)
			result[fullPath] = string(jsonBytes)
		default:
			result[fullPath] = flattenXMLValue(v)
		}
	}

	return result
}

// flattenXMLMap converts XML map to flat structure with JSON for nested parts
func flattenXMLMap(m map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for k, v := range m {
		if k == "__text" {
			continue // Skip text-only marker
		}

		result[k] = flattenXMLValue(v)
	}

	return result
}

// flattenXMLValue recursively flattens XML values
func flattenXMLValue(v interface{}) interface{} {
	switch val := v.(type) {
	case map[string]interface{}:
		// Check if it's a simple text-only element
		if text, ok := val["__text"]; ok && len(val) == 1 {
			return text
		}
		// Flatten nested map
		flattened := make(map[string]interface{})
		for k, v := range val {
			if k != "__text" {
				flattened[k] = flattenXMLValue(v)
			}
		}
		// If still has content, convert to JSON
		if len(flattened) > 0 {
			jsonBytes, _ := json.Marshal(flattened)
			return string(jsonBytes)
		}
		return ""

	case []interface{}:
		// Flatten array elements
		flattened := make([]interface{}, len(val))
		for i, item := range val {
			flattened[i] = flattenXMLValue(item)
		}
		// Convert to JSON array
		jsonBytes, _ := json.Marshal(flattened)
		return string(jsonBytes)

	default:
		return v
	}
}

// WriteXML writes the ZeaFrame to an XML writer
// Creates structure: <root><record>...</record><record>...</record></root>
func (zf *ZeaFrame) WriteXML(writer io.Writer, rootName, recordName string) error {
	if rootName == "" {
		rootName = "root"
	}
	if recordName == "" {
		recordName = "record"
	}

	// Write XML header
	if _, err := fmt.Fprintf(writer, "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n"); err != nil {
		return err
	}

	// Write root start tag
	if _, err := fmt.Fprintf(writer, "<%s>\n", rootName); err != nil {
		return err
	}

	// Write each record
	for i := 0; i < zf.Rows; i++ {
		if _, err := fmt.Fprintf(writer, "  <%s>\n", recordName); err != nil {
			return err
		}

		for _, col := range zf.Columns {
			if i < len(col.Data) {
				value := col.Data[i]

				// Check if value is nested JSON (array or object)
				valueStr := fmt.Sprintf("%v", value)
				if strings.HasPrefix(valueStr, "[") || strings.HasPrefix(valueStr, "{") {
					// Parse JSON and convert to nested XML
					var jsonData interface{}
					if err := json.Unmarshal([]byte(valueStr), &jsonData); err == nil {
						if err := writeXMLValue(writer, col.Name, jsonData, 4); err != nil {
							return err
						}
						continue
					}
				}

				// Simple value
				xmlValue := escapeXML(fmt.Sprintf("%v", value))
				if _, err := fmt.Fprintf(writer, "    <%s>%s</%s>\n", col.Name, xmlValue, col.Name); err != nil {
					return err
				}
			}
		}

		if _, err := fmt.Fprintf(writer, "  </%s>\n", recordName); err != nil {
			return err
		}
	}

	// Write root end tag
	if _, err := fmt.Fprintf(writer, "</%s>\n", rootName); err != nil {
		return err
	}

	return nil
}

// writeXMLValue writes a JSON value as XML
func writeXMLValue(writer io.Writer, name string, value interface{}, indent int) error {
	spaces := strings.Repeat(" ", indent)

	switch v := value.(type) {
	case map[string]interface{}:
		// Object - write as nested element
		if _, err := fmt.Fprintf(writer, "%s<%s>\n", spaces, name); err != nil {
			return err
		}
		for k, val := range v {
			if err := writeXMLValue(writer, k, val, indent+2); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintf(writer, "%s</%s>\n", spaces, name); err != nil {
			return err
		}

	case []interface{}:
		// Array - write each element with same tag name
		for _, item := range v {
			if err := writeXMLValue(writer, name, item, indent); err != nil {
				return err
			}
		}

	default:
		// Scalar value
		xmlValue := escapeXML(fmt.Sprintf("%v", v))
		if _, err := fmt.Fprintf(writer, "%s<%s>%s</%s>\n", spaces, name, xmlValue, name); err != nil {
			return err
		}
	}

	return nil
}

// escapeXML escapes special XML characters
func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}
