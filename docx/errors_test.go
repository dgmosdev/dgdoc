package docx

import (
	"strings"
	"testing"
)

func TestValidationError(t *testing.T) {
	err := &ValidationError{
		Message:  "Invalid placeholder",
		Location: "cell A1",
		Details:  "missing closing brace",
	}

	expected := "Invalid placeholder in cell A1: missing closing brace"
	if err.Error() != expected {
		t.Errorf("Expected %q, got %q", expected, err.Error())
	}
}

func TestTemplateError(t *testing.T) {
	originalErr := &ValidationError{
		Message: "Invalid syntax",
		Details: "test",
	}

	err := NewTemplateError("apply", originalErr)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if !strings.Contains(err.Error(), "apply") {
		t.Errorf("Error should contain operation name: %v", err)
	}
}

func TestValidatePlaceholders(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		hasError bool
	}{
		{
			name:     "Valid placeholder",
			content:  "Hello {name}",
			hasError: false,
		},
		{
			name:     "Unmatched closing brace",
			content:  "Hello }name{",
			hasError: true,
		},
		{
			name:     "Unclosed brace",
			content:  "Hello {name",
			hasError: true,
		},
		{
			name:     "Multiple valid placeholders",
			content:  "{first} {last}",
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidatePlaceholders(tt.content)
			hasError := len(errors) > 0

			if hasError != tt.hasError {
				t.Errorf("Expected hasError=%v, got %v (errors: %v)",
					tt.hasError, hasError, errors)
			}
		})
	}
}
