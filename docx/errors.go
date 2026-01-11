package docx

import (
	"fmt"
)

// ValidationError represents a template validation error
type ValidationError struct {
	Message  string
	Location string // e.g., "placeholder", "loop", "conditional"
	Details  string
}

func (e *ValidationError) Error() string {
	if e.Location != "" {
		return fmt.Sprintf("%s in %s: %s", e.Message, e.Location, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Message, e.Details)
}

// TemplateError wraps errors with context
type TemplateError struct {
	Op  string // Operation that failed
	Err error  // Original error
}

func (e *TemplateError) Error() string {
	return fmt.Sprintf("template %s failed: %v", e.Op, e.Err)
}

func (e *TemplateError) Unwrap() error {
	return e.Err
}

// NewTemplateError creates a new template error with context
func NewTemplateError(op string, err error) error {
	if err == nil {
		return nil
	}
	return &TemplateError{Op: op, Err: err}
}

// ValidatePlaceholders checks if placeholders are properly formatted
func ValidatePlaceholders(content string) []ValidationError {
	var errors []ValidationError

	// Check for unmatched braces
	openCount := 0
	for i, char := range content {
		switch char {
		case '{':
			openCount++
		case '}':
			openCount--
			if openCount < 0 {
				errors = append(errors, ValidationError{
					Message:  "Unmatched closing brace",
					Location: "placeholder",
					Details:  fmt.Sprintf("at position %d", i),
				})
			}
		}
	}

	if openCount > 0 {
		errors = append(errors, ValidationError{
			Message:  "Unclosed braces",
			Location: "placeholder",
			Details:  fmt.Sprintf("%d opening brace(s) without closing", openCount),
		})
	}

	return errors
}
