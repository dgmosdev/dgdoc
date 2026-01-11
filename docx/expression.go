package docx

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// evaluateCondition evaluates a conditional expression against the provided data
// Supports: ==, !=, <, >, <=, >=, &&, ||, !
// Examples: "age > 18", "status == 'active'", "name", "!disabled"
func evaluateCondition(expr string, data map[string]any) (bool, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return false, nil
	}

	// Handle logical OR (||)
	if strings.Contains(expr, "||") {
		parts := strings.Split(expr, "||")
		for _, part := range parts {
			result, err := evaluateCondition(strings.TrimSpace(part), data)
			if err != nil {
				return false, err
			}
			if result {
				return true, nil
			}
		}
		return false, nil
	}

	// Handle logical AND (&&)
	if strings.Contains(expr, "&&") {
		parts := strings.Split(expr, "&&")
		for _, part := range parts {
			result, err := evaluateCondition(strings.TrimSpace(part), data)
			if err != nil {
				return false, err
			}
			if !result {
				return false, nil
			}
		}
		return true, nil
	}

	// Handle NOT (!)
	if strings.HasPrefix(expr, "!") {
		result, err := evaluateCondition(strings.TrimSpace(expr[1:]), data)
		if err != nil {
			return false, err
		}
		return !result, nil
	}

	// Handle comparison operators
	operators := []string{"==", "!=", "<=", ">=", "<", ">"}
	for _, op := range operators {
		if strings.Contains(expr, op) {
			parts := strings.SplitN(expr, op, 2)
			if len(parts) != 2 {
				continue
			}

			left := strings.TrimSpace(parts[0])
			right := strings.TrimSpace(parts[1])

			leftVal, err := resolveValue(left, data)
			if err != nil {
				return false, err
			}

			rightVal, err := resolveValue(right, data)
			if err != nil {
				return false, err
			}

			return compareValues(leftVal, rightVal, op)
		}
	}

	// Simple truthiness check - check if variable exists and is truthy
	val, err := resolveValue(expr, data)
	if err != nil {
		return false, nil // Variable doesn't exist = false
	}

	return isTruthy(val), nil
}

// resolveValue resolves a value from the expression
// Handles: variable names, string literals, number literals, path navigation
func resolveValue(expr string, data map[string]any) (any, error) {
	expr = strings.TrimSpace(expr)

	// String literal (single or double quotes)
	if (strings.HasPrefix(expr, "'") && strings.HasSuffix(expr, "'")) ||
		(strings.HasPrefix(expr, `"`) && strings.HasSuffix(expr, `"`)) {
		return expr[1 : len(expr)-1], nil
	}

	// Number literal
	if num, err := strconv.ParseFloat(expr, 64); err == nil {
		return num, nil
	}

	// Boolean literal
	if expr == "true" {
		return true, nil
	}
	if expr == "false" {
		return false, nil
	}

	// Variable with path navigation (e.g., "user.name", "items[0].price")
	return getNestedValue(expr, data)
}

// getNestedValue retrieves a nested value from data using path notation
// Supports: "user.name", "items[0]", "config.settings.enabled"
func getNestedValue(path string, data map[string]any) (any, error) {
	parts := strings.Split(path, ".")
	var current any = data

	for _, part := range parts {
		// Handle array indexing: "items[0]"
		if strings.Contains(part, "[") && strings.Contains(part, "]") {
			arrName := part[:strings.Index(part, "[")]
			indexStr := part[strings.Index(part, "[")+1 : strings.Index(part, "]")]
			index, err := strconv.Atoi(indexStr)
			if err != nil {
				return nil, fmt.Errorf("invalid array index: %s", indexStr)
			}

			// Get array
			if arrName != "" {
				val, ok := current.(map[string]any)
				if !ok {
					return nil, fmt.Errorf("cannot navigate: %s is not an object", arrName)
				}
				current = val[arrName]
			}

			// Get index
			arr, ok := current.([]any)
			if !ok {
				return nil, fmt.Errorf("not an array: %s", arrName)
			}
			if index < 0 || index >= len(arr) {
				return nil, fmt.Errorf("index out of bounds: %d", index)
			}
			current = arr[index]
			continue
		}

		// Regular object property access
		m, ok := current.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("cannot access property '%s' on non-object", part)
		}

		val, exists := m[part]
		if !exists {
			return nil, fmt.Errorf("property '%s' does not exist", part)
		}
		current = val
	}

	return current, nil
}

// compareValues compares two values using the given operator
func compareValues(left, right any, op string) (bool, error) {
	switch op {
	case "==":
		return equals(left, right), nil
	case "!=":
		return !equals(left, right), nil
	case "<", ">", "<=", ">=":
		return numericCompare(left, right, op)
	default:
		return false, fmt.Errorf("unknown operator: %s", op)
	}
}

// equals checks if two values are equal
func equals(left, right any) bool {
	// Handle nil cases
	if left == nil && right == nil {
		return true
	}
	if left == nil || right == nil {
		return false
	}

	// Try direct comparison
	if left == right {
		return true
	}

	// Try string comparison
	leftStr := fmt.Sprintf("%v", left)
	rightStr := fmt.Sprintf("%v", right)
	if leftStr == rightStr {
		return true
	}

	// Try numeric comparison
	leftNum, leftOk := toFloat64(left)
	rightNum, rightOk := toFloat64(right)
	if leftOk && rightOk {
		return leftNum == rightNum
	}

	return false
}

// numericCompare performs numeric comparison
func numericCompare(left, right any, op string) (bool, error) {
	leftNum, leftOk := toFloat64(left)
	rightNum, rightOk := toFloat64(right)

	if !leftOk || !rightOk {
		return false, fmt.Errorf("cannot compare non-numeric values with %s", op)
	}

	switch op {
	case "<":
		return leftNum < rightNum, nil
	case ">":
		return leftNum > rightNum, nil
	case "<=":
		return leftNum <= rightNum, nil
	case ">=":
		return leftNum >= rightNum, nil
	default:
		return false, fmt.Errorf("invalid numeric operator: %s", op)
	}
}

// toFloat64 converts a value to float64 if possible
func toFloat64(val any) (float64, bool) {
	switch v := val.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case int32:
		return float64(v), true
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f, true
		}
	}
	return 0, false
}

// isTruthy checks if a value is truthy
func isTruthy(val any) bool {
	if val == nil {
		return false
	}

	switch v := val.(type) {
	case bool:
		return v
	case string:
		return v != ""
	case int, int64, int32, float64, float32:
		num, _ := toFloat64(v)
		return num != 0
	case []any:
		return len(v) > 0
	case map[string]any:
		return len(v) > 0
	default:
		// Use reflection for other types
		rv := reflect.ValueOf(v)
		switch rv.Kind() {
		case reflect.Slice, reflect.Map, reflect.Array:
			return rv.Len() > 0
		case reflect.Ptr, reflect.Interface:
			return !rv.IsNil()
		default:
			return true // Unknown types are considered truthy if they exist
		}
	}
}
