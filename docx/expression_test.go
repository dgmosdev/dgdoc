package docx

import (
	"testing"
)

func TestEvaluateCondition_Simple(t *testing.T) {
	data := map[string]any{
		"name":   "Ahmet",
		"age":    25,
		"active": true,
		"empty":  "",
	}

	tests := []struct {
		expr     string
		expected bool
	}{
		// Truthiness
		{"name", true},
		{"age", true},
		{"active", true},
		{"empty", false},
		{"missing", false},

		// Equality
		{"name == 'Ahmet'", true},
		{"name == 'Mehmet'", false},
		{"age == 25", true},
		{"age == 30", false},
		{"active == true", true},
		{"active == false", false},

		// Inequality
		{"name != 'Mehmet'", true},
		{"name != 'Ahmet'", false},
		{"age != 30", true},
		{"age != 25", false},

		// Numeric comparison
		{"age > 20", true},
		{"age > 30", false},
		{"age < 30", true},
		{"age < 20", false},
		{"age >= 25", true},
		{"age >= 26", false},
		{"age <= 25", true},
		{"age <= 24", false},

		// Logical NOT
		{"!empty", true},
		{"!name", false},
		{"!missing", true},
		{"!active", false},

		// Logical AND
		{"age > 20 && active", true},
		{"age > 30 && active", false},
		{"name == 'Ahmet' && age == 25", true},
		{"name == 'Ahmet' && age == 30", false},

		// Logical OR
		{"age > 30 || active", true},
		{"age > 30 || empty", false},
		{"name == 'Mehmet' || age == 25", true},
		{"name == 'Mehmet' || age == 30", false},
	}

	for _, tt := range tests {
		result, err := evaluateCondition(tt.expr, data)
		if err != nil {
			t.Errorf("evaluateCondition(%q) returned error: %v", tt.expr, err)
			continue
		}
		if result != tt.expected {
			t.Errorf("evaluateCondition(%q) = %v, want %v", tt.expr, result, tt.expected)
		}
	}
}

func TestEvaluateCondition_NestedPath(t *testing.T) {
	data := map[string]any{
		"user": map[string]any{
			"name": "Ahmet",
			"age":  25,
			"address": map[string]any{
				"city": "Istanbul",
			},
		},
		"items": []any{
			map[string]any{"name": "Item1", "price": 100},
			map[string]any{"name": "Item2", "price": 200},
		},
	}

	tests := []struct {
		expr     string
		expected bool
	}{
		{"user.name == 'Ahmet'", true},
		{"user.age > 20", true},
		{"user.address.city == 'Istanbul'", true},
		{"user.address.city == 'Ankara'", false},
		{"items[0].name == 'Item1'", true},
		{"items[1].price == 200", true},
		{"items[0].price > 50", true},
		{"items[1].price < 100", false},
	}

	for _, tt := range tests {
		result, err := evaluateCondition(tt.expr, data)
		if err != nil {
			t.Errorf("evaluateCondition(%q) returned error: %v", tt.expr, err)
			continue
		}
		if result != tt.expected {
			t.Errorf("evaluateCondition(%q) = %v, want %v", tt.expr, result, tt.expected)
		}
	}
}

func TestEvaluateCondition_ComplexLogic(t *testing.T) {
	data := map[string]any{
		"age":    25,
		"active": true,
		"status": "premium",
	}

	tests := []struct {
		expr     string
		expected bool
	}{
		{"age > 18 && active", true},
		{"age > 30 && active", false},
		{"age < 18 || status == 'premium'", true},
		{"!active || age > 30", false},
		// Note: Parentheses not yet supported
		// && has higher precedence than ||, so this works as expected:
		{"active && age > 20 || status == 'basic'", true},
	}

	for _, tt := range tests {
		result, err := evaluateCondition(tt.expr, data)
		if err != nil {
			t.Errorf("evaluateCondition(%q) returned error: %v", tt.expr, err)
			continue
		}
		if result != tt.expected {
			t.Errorf("evaluateCondition(%q) = %v, want %v", tt.expr, result, tt.expected)
		}
	}
}

func TestIsTruthy(t *testing.T) {
	tests := []struct {
		val      any
		expected bool
	}{
		{nil, false},
		{true, true},
		{false, false},
		{"", false},
		{"hello", true},
		{0, false},
		{1, true},
		{-1, true},
		{0.0, false},
		{1.5, true},
		{[]any{}, false},
		{[]any{1}, true},
		{map[string]any{}, false},
		{map[string]any{"key": "val"}, true},
	}

	for _, tt := range tests {
		result := isTruthy(tt.val)
		if result != tt.expected {
			t.Errorf("isTruthy(%v) = %v, want %v", tt.val, result, tt.expected)
		}
	}
}

func TestGetNestedValue(t *testing.T) {
	data := map[string]any{
		"user": map[string]any{
			"name": "Ahmet",
			"contacts": map[string]any{
				"email": "test@example.com",
			},
		},
		"items": []any{
			map[string]any{"id": 1},
			map[string]any{"id": 2},
		},
	}

	tests := []struct {
		path     string
		expected any
		hasError bool
	}{
		{"user.name", "Ahmet", false},
		{"user.contacts.email", "test@example.com", false},
		{"items[0].id", float64(1), false}, // JSON numbers are float64
		{"items[1].id", float64(2), false},
		{"user.missing", nil, true},
		{"items[5]", nil, true}, // Out of bounds
	}

	for _, tt := range tests {
		result, err := getNestedValue(tt.path, data)
		if tt.hasError {
			if err == nil {
				t.Errorf("getNestedValue(%q) expected error but got none", tt.path)
			}
		} else {
			if err != nil {
				t.Errorf("getNestedValue(%q) unexpected error: %v", tt.path, err)
			}
			// For numbers, convert to float64 for comparison
			if expectedFloat, ok := tt.expected.(float64); ok {
				if resultFloat, ok := result.(int); ok {
					result = float64(resultFloat)
				}
				if result != expectedFloat {
					t.Errorf("getNestedValue(%q) = %v, want %v", tt.path, result, tt.expected)
				}
			} else if result != tt.expected {
				t.Errorf("getNestedValue(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		}
	}
}
