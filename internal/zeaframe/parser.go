package zeaframe

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

// Expression represents a filter expression
type Expression struct {
	Left       *Expression
	Right      *Expression
	Operator   string
	Value      interface{}
	Column     string
	IsLeaf     bool
	ArrayIndex *int   // For array indexing: orders[0]
	NestedPath string // For nested paths: address.city
}

// ParseExpression parses a filter expression string
func ParseExpression(expr string) (*Expression, error) {
	expr = strings.TrimSpace(expr)
	// Remove shell escape sequences like \!
	expr = strings.ReplaceAll(expr, `\!`, `!`)
	expr = strings.ReplaceAll(expr, `\ `, ` `)
	if expr == "" {
		return nil, fmt.Errorf("empty expression")
	}

	// Handle AND/OR operators (lowest precedence)
	if pos := findOperator(expr, "AND"); pos != -1 {
		left, err := ParseExpression(expr[:pos])
		if err != nil {
			return nil, err
		}
		right, err := ParseExpression(expr[pos+3:])
		if err != nil {
			return nil, err
		}
		return &Expression{
			Left:     left,
			Right:    right,
			Operator: "AND",
			IsLeaf:   false,
		}, nil
	}

	if pos := findOperator(expr, "OR"); pos != -1 {
		left, err := ParseExpression(expr[:pos])
		if err != nil {
			return nil, err
		}
		right, err := ParseExpression(expr[pos+2:])
		if err != nil {
			return nil, err
		}
		return &Expression{
			Left:     left,
			Right:    right,
			Operator: "OR",
			IsLeaf:   false,
		}, nil
	}

	// Handle CONTAINS operator (for arrays)
	if pos := findOperator(expr, "CONTAINS"); pos != -1 {
		column := strings.TrimSpace(expr[:pos])
		valueStr := strings.TrimSpace(expr[pos+8:]) // len("CONTAINS") = 8

		// Parse column (may have array index or nested path)
		baseCol, arrayIdx, nestedPath := parseColumnReference(column)

		// Parse value
		value, err := parseValue(valueStr)
		if err != nil {
			return nil, err
		}

		return &Expression{
			Column:     baseCol,
			Operator:   "CONTAINS",
			Value:      value,
			IsLeaf:     true,
			ArrayIndex: arrayIdx,
			NestedPath: nestedPath,
		}, nil
	}

	// Handle comparison operators
	for _, op := range []string{">=", "<=", "!=", "=", ">", "<"} {
		pos := findOperator(expr, op)
		if pos != -1 {
			column := strings.TrimSpace(expr[:pos])
			valueStr := strings.TrimSpace(expr[pos+len(op):])

			// Parse column (may have array index or nested path)
			baseCol, arrayIdx, nestedPath := parseColumnReference(column)

			// Parse value
			value, err := parseValue(valueStr)
			if err != nil {
				return nil, err
			}

			return &Expression{
				Column:     baseCol,
				Operator:   op,
				Value:      value,
				IsLeaf:     true,
				ArrayIndex: arrayIdx,
				NestedPath: nestedPath,
			}, nil
		}
	}

	return nil, fmt.Errorf("invalid expression: %s", expr)
}

// findOperator finds the position of an operator outside of quotes
func findOperator(expr string, op string) int {
	inQuotes := false
	quoteChar := rune(0)

	for i := 0; i <= len(expr)-len(op); i++ {
		ch := rune(expr[i])

		if ch == '\'' || ch == '"' {
			if inQuotes && ch == quoteChar {
				inQuotes = false
			} else if !inQuotes {
				inQuotes = true
				quoteChar = ch
			}
		}

		if !inQuotes && strings.HasPrefix(expr[i:], op) {
			// For word operators like AND/OR, check boundaries
			if op == "AND" || op == "OR" {
				before := i == 0 || !isAlphaNum(rune(expr[i-1]))
				after := i+len(op) >= len(expr) || !isAlphaNum(rune(expr[i+len(op)]))
				if before && after {
					return i
				}
			} else {
				return i
			}
		}
	}

	return -1
}

// isAlphaNum checks if a character is alphanumeric
func isAlphaNum(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_'
}

// parseColumnReference parses a column reference which may include:
// - Simple column: "customer"
// - Array index: "orders[0]"
// - Nested path: "address.city"
// Returns (baseColumn, arrayIndex, nestedPath)
func parseColumnReference(col string) (string, *int, string) {
	col = strings.TrimSpace(col)

	// Check for array indexing: column[index]
	if idx := strings.Index(col, "["); idx != -1 {
		if !strings.HasSuffix(col, "]") {
			return col, nil, "" // Invalid syntax, return as-is
		}
		baseCol := col[:idx]
		indexStr := col[idx+1 : len(col)-1]
		if index, err := strconv.Atoi(indexStr); err == nil {
			return baseCol, &index, ""
		}
		return col, nil, "" // Invalid index, return as-is
	}

	// Check for nested path: column.path.to.field
	if idx := strings.Index(col, "."); idx != -1 {
		baseCol := col[:idx]
		nestedPath := col[idx+1:]
		return baseCol, nil, nestedPath
	}

	// Simple column reference
	return col, nil, ""
}

// parseValue parses a value from a string
func parseValue(s string) (interface{}, error) {
	s = strings.TrimSpace(s)

	// String literal
	if (strings.HasPrefix(s, "'") && strings.HasSuffix(s, "'")) ||
		(strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"")) {
		return s[1 : len(s)-1], nil
	}

	// Boolean
	if s == "true" {
		return true, nil
	}
	if s == "false" {
		return false, nil
	}

	// Try integer
	if intVal, err := strconv.ParseInt(s, 10, 64); err == nil {
		return intVal, nil
	}

	// Try float
	if floatVal, err := strconv.ParseFloat(s, 64); err == nil {
		return floatVal, nil
	}

	return nil, fmt.Errorf("cannot parse value: %s", s)
}

// Evaluate evaluates the expression against a row in a ZeaFrame
func (e *Expression) Evaluate(zf *ZeaFrame, rowIdx int) (bool, error) {
	if !e.IsLeaf {
		// Logical operator
		leftResult, err := e.Left.Evaluate(zf, rowIdx)
		if err != nil {
			return false, err
		}

		rightResult, err := e.Right.Evaluate(zf, rowIdx)
		if err != nil {
			return false, err
		}

		switch e.Operator {
		case "AND":
			return leftResult && rightResult, nil
		case "OR":
			return leftResult || rightResult, nil
		default:
			return false, fmt.Errorf("unknown logical operator: %s", e.Operator)
		}
	}

	// Leaf node - comparison
	col, err := zf.GetColumn(e.Column)
	if err != nil {
		return false, err
	}

	if rowIdx >= len(col.Data) {
		return false, fmt.Errorf("row index out of bounds")
	}

	cellValue := col.Data[rowIdx]

	// Extract the actual value to compare based on array index or nested path
	actualValue, err := extractValue(cellValue, e.ArrayIndex, e.NestedPath)
	if err != nil {
		// If we can't extract the value (e.g., path doesn't exist), treat as not matching
		return false, nil
	}

	// Handle CONTAINS operator specially
	if e.Operator == "CONTAINS" {
		return containsValue(actualValue, e.Value)
	}

	return compareValues(actualValue, e.Value, e.Operator)
}

// compareValues compares two values using the given operator
func compareValues(left, right interface{}, operator string) (bool, error) {
	switch operator {
	case "=":
		return compareEqual(left, right), nil
	case "!=":
		return !compareEqual(left, right), nil
	case ">":
		return compareGreater(left, right)
	case ">=":
		return compareGreaterEqual(left, right)
	case "<":
		return compareLess(left, right)
	case "<=":
		return compareLessEqual(left, right)
	default:
		return false, fmt.Errorf("unknown operator: %s", operator)
	}
}

// compareEqual checks equality
func compareEqual(left, right interface{}) bool {
	// Convert to comparable types
	leftStr := fmt.Sprintf("%v", left)
	rightStr := fmt.Sprintf("%v", right)

	// Try numeric comparison
	leftFloat, leftErr := toFloat64(left)
	rightFloat, rightErr := toFloat64(right)

	if leftErr == nil && rightErr == nil {
		return leftFloat == rightFloat
	}

	// String comparison
	return leftStr == rightStr
}

// compareGreater checks if left > right
func compareGreater(left, right interface{}) (bool, error) {
	leftFloat, err1 := toFloat64(left)
	rightFloat, err2 := toFloat64(right)

	if err1 != nil || err2 != nil {
		// Check if we're trying to compare JSON structures with numbers
		leftStr := fmt.Sprintf("%v", left)
		rightStr := fmt.Sprintf("%v", right)

		// If left looks like JSON array and right is numeric, check if ANY element > right
		if strings.HasPrefix(leftStr, "[") && err2 == nil {
			var arr []interface{}
			if err := json.Unmarshal([]byte(leftStr), &arr); err == nil {
				for _, item := range arr {
					itemFloat, itemErr := toFloat64(item)
					if itemErr == nil && itemFloat > rightFloat {
						return true, nil
					}
				}
				return false, nil
			}
		}

		// If left is object, can't compare with number
		if strings.HasPrefix(leftStr, "{") && err2 == nil {
			return false, nil
		}

		// If right is JSON, can't compare
		if (strings.HasPrefix(rightStr, "[") || strings.HasPrefix(rightStr, "{")) && err1 == nil {
			return false, nil
		}

		// Otherwise fall back to string comparison
		return leftStr > rightStr, nil
	}

	return leftFloat > rightFloat, nil
}

// compareGreaterEqual checks if left >= right
func compareGreaterEqual(left, right interface{}) (bool, error) {
	leftFloat, err1 := toFloat64(left)
	rightFloat, err2 := toFloat64(right)

	if err1 != nil || err2 != nil {
		leftStr := fmt.Sprintf("%v", left)
		rightStr := fmt.Sprintf("%v", right)

		// If left looks like JSON array and right is numeric, check if ANY element >= right
		if strings.HasPrefix(leftStr, "[") && err2 == nil {
			var arr []interface{}
			if err := json.Unmarshal([]byte(leftStr), &arr); err == nil {
				for _, item := range arr {
					itemFloat, itemErr := toFloat64(item)
					if itemErr == nil && itemFloat >= rightFloat {
						return true, nil
					}
				}
				return false, nil
			}
		}

		// If left is object, can't compare with number
		if strings.HasPrefix(leftStr, "{") && err2 == nil {
			return false, nil
		}

		// If right is JSON, can't compare
		if (strings.HasPrefix(rightStr, "[") || strings.HasPrefix(rightStr, "{")) && err1 == nil {
			return false, nil
		}

		return leftStr >= rightStr, nil
	}

	return leftFloat >= rightFloat, nil
}

// compareLess checks if left < right
func compareLess(left, right interface{}) (bool, error) {
	leftFloat, err1 := toFloat64(left)
	rightFloat, err2 := toFloat64(right)

	if err1 != nil || err2 != nil {
		leftStr := fmt.Sprintf("%v", left)
		rightStr := fmt.Sprintf("%v", right)

		// If left looks like JSON array and right is numeric, check if ANY element < right
		if strings.HasPrefix(leftStr, "[") && err2 == nil {
			var arr []interface{}
			if err := json.Unmarshal([]byte(leftStr), &arr); err == nil {
				for _, item := range arr {
					itemFloat, itemErr := toFloat64(item)
					if itemErr == nil && itemFloat < rightFloat {
						return true, nil
					}
				}
				return false, nil
			}
		}

		// If left is object, can't compare with number
		if strings.HasPrefix(leftStr, "{") && err2 == nil {
			return false, nil
		}

		// If right is JSON, can't compare
		if (strings.HasPrefix(rightStr, "[") || strings.HasPrefix(rightStr, "{")) && err1 == nil {
			return false, nil
		}

		return leftStr < rightStr, nil
	}

	return leftFloat < rightFloat, nil
}

// compareLessEqual checks if left <= right
func compareLessEqual(left, right interface{}) (bool, error) {
	leftFloat, err1 := toFloat64(left)
	rightFloat, err2 := toFloat64(right)

	if err1 != nil || err2 != nil {
		leftStr := fmt.Sprintf("%v", left)
		rightStr := fmt.Sprintf("%v", right)

		// If left looks like JSON array and right is numeric, check if ANY element <= right
		if strings.HasPrefix(leftStr, "[") && err2 == nil {
			var arr []interface{}
			if err := json.Unmarshal([]byte(leftStr), &arr); err == nil {
				for _, item := range arr {
					itemFloat, itemErr := toFloat64(item)
					if itemErr == nil && itemFloat <= rightFloat {
						return true, nil
					}
				}
				return false, nil
			}
		}

		// If left is object, can't compare with number
		if strings.HasPrefix(leftStr, "{") && err2 == nil {
			return false, nil
		}

		// If right is JSON, can't compare
		if (strings.HasPrefix(rightStr, "[") || strings.HasPrefix(rightStr, "{")) && err1 == nil {
			return false, nil
		}

		return leftStr <= rightStr, nil
	}

	return leftFloat <= rightFloat, nil
}

// extractValue extracts a value from a cell, handling JSON arrays/objects
// if arrayIndex or nestedPath are specified
func extractValue(cellValue interface{}, arrayIndex *int, nestedPath string) (interface{}, error) {
	// If no special extraction needed, return as-is
	if arrayIndex == nil && nestedPath == "" {
		return cellValue, nil
	}

	// Convert cell value to string (may be JSON)
	cellStr := fmt.Sprintf("%v", cellValue)

	// Try to parse as JSON
	var jsonData interface{}
	if err := json.Unmarshal([]byte(cellStr), &jsonData); err != nil {
		// Not valid JSON, return original value
		return cellValue, nil
	}

	// Handle array indexing
	if arrayIndex != nil {
		arr, ok := jsonData.([]interface{})
		if !ok {
			return nil, fmt.Errorf("value is not an array")
		}
		if *arrayIndex < 0 || *arrayIndex >= len(arr) {
			return nil, fmt.Errorf("array index out of bounds: %d", *arrayIndex)
		}
		return arr[*arrayIndex], nil
	}

	// Handle nested path
	if nestedPath != "" {
		return extractNestedPath(jsonData, nestedPath)
	}

	return jsonData, nil
}

// extractNestedPath navigates a nested path like "address.city" in a JSON object
func extractNestedPath(data interface{}, path string) (interface{}, error) {
	parts := strings.Split(path, ".")
	current := data

	for _, part := range parts {
		obj, ok := current.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("cannot navigate path '%s': not an object", part)
		}
		val, exists := obj[part]
		if !exists {
			return nil, fmt.Errorf("path not found: %s", part)
		}
		current = val
	}

	return current, nil
}

// containsValue checks if a value (which may be an array) contains the target value
func containsValue(cellValue interface{}, targetValue interface{}) (bool, error) {
	// Convert target to string to check for wildcard patterns
	targetStr := fmt.Sprintf("%v", targetValue)
	isWildcard := strings.Contains(targetStr, "*") || strings.Contains(targetStr, "?")

	// If cellValue is a string that looks like JSON, try to parse it
	cellStr := fmt.Sprintf("%v", cellValue)

	var jsonData interface{}
	if err := json.Unmarshal([]byte(cellStr), &jsonData); err == nil {
		cellValue = jsonData
	}

	// Check if it's an array
	arr, ok := cellValue.([]interface{})
	if !ok {
		// Not an array, do direct comparison or pattern match
		if isWildcard {
			return matchesPattern(cellValue, targetStr), nil
		}
		return compareEqual(cellValue, targetValue), nil
	}

	// Search array for target value (with wildcard support)
	for _, item := range arr {
		if isWildcard {
			if matchesPattern(item, targetStr) {
				return true, nil
			}
		} else {
			if compareEqual(item, targetValue) {
				return true, nil
			}
		}
	}

	return false, nil
}

// matchesPattern checks if a value matches a wildcard pattern
// Supports * (any characters) and ? (single character)
func matchesPattern(value interface{}, pattern string) bool {
	valueStr := fmt.Sprintf("%v", value)

	// Use filepath.Match for glob-style pattern matching
	matched, err := filepath.Match(pattern, valueStr)
	if err != nil {
		// If pattern is invalid, fall back to exact match
		return valueStr == pattern
	}

	return matched
}
