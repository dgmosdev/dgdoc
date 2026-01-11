package docx

import (
	"strings"
	"testing"
)

func TestProcessConditionals_Simple(t *testing.T) {
	// Simple mock XML with conditional
	content := `<w:p><w:r><w:t>{#if show}Hello World{/if}</w:t></w:r></w:p>`

	tests := []struct {
		name        string
		data        map[string]any
		contains    string
		notContains string
	}{
		{
			name:     "condition true",
			data:     map[string]any{"show": true},
			contains: "Hello World",
		},
		{
			name:        "condition false",
			data:        map[string]any{"show": false},
			notContains: "Hello World",
		},
		{
			name:        "condition missing",
			data:        map[string]any{},
			notContains: "Hello World",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := processConditionals(content, tt.data)
			if err != nil {
				t.Fatalf("processConditionals failed: %v", err)
			}

			if tt.contains != "" && !strings.Contains(extractTextContent(result), tt.contains) {
				t.Errorf("Expected result to contain %q, got: %s", tt.contains, extractTextContent(result))
			}
			if tt.notContains != "" && strings.Contains(extractTextContent(result), tt.notContains) {
				t.Errorf("Expected result to NOT contain %q, got: %s", tt.notContains, extractTextContent(result))
			}
		})
	}
}

func TestProcessConditionals_WithElse(t *testing.T) {
	content := `<w:p><w:r><w:t>{#if premium}Premium User{#else}Free User{/if}</w:t></w:r></w:p>`

	tests := []struct {
		name     string
		data     map[string]any
		contains string
	}{
		{
			name:     "true shows if block",
			data:     map[string]any{"premium": true},
			contains: "Premium User",
		},
		{
			name:     "false shows else block",
			data:     map[string]any{"premium": false},
			contains: "Free User",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := processConditionals(content, tt.data)
			if err != nil {
				t.Fatalf("processConditionals failed: %v", err)
			}

			text := extractTextContent(result)
			if !strings.Contains(text, tt.contains) {
				t.Errorf("Expected result to contain %q, got: %s", tt.contains, text)
			}
		})
	}
}

func TestProcessLoops_Simple(t *testing.T) {
	content := `<w:p><w:r><w:t>{#items}{name}{/items}</w:t></w:r></w:p>`

	data := map[string]any{
		"items": []any{
			map[string]any{"name": "Item1"},
			map[string]any{"name": "Item2"},
			map[string]any{"name": "Item3"},
		},
	}

	result, err := processLoops(content, data)
	if err != nil {
		t.Fatalf("processLoops failed: %v", err)
	}

	text := extractTextContent(result)
	if !strings.Contains(text, "Item1") {
		t.Errorf("Expected result to contain Item1, got: %s", text)
	}
	if !strings.Contains(text, "Item2") {
		t.Errorf("Expected result to contain Item2, got: %s", text)
	}
	if !strings.Contains(text, "Item3") {
		t.Errorf("Expected result to contain Item3, got: %s", text)
	}
}

func TestProcessLoops_WithMetadata(t *testing.T) {
	content := `<w:p><w:r><w:t>{#items}{@index}. {name}{/items}</w:t></w:r></w:p>`

	data := map[string]any{
		"items": []any{
			map[string]any{"name": "First"},
			map[string]any{"name": "Second"},
		},
	}

	result, err := processLoops(content, data)
	if err != nil {
		t.Fatalf("processLoops failed: %v", err)
	}

	text := extractTextContent(result)
	if !strings.Contains(text, "0") || !strings.Contains(text, "First") {
		t.Errorf("Expected result to contain index 0 and First, got: %s", text)
	}
	if !strings.Contains(text, "1") || !strings.Contains(text, "Second") {
		t.Errorf("Expected result to contain index 1 and Second, got: %s", text)
	}
}

func TestPreprocessTemplate_Combined(t *testing.T) {
	// Template with both conditionals and loops
	content := `<w:p><w:r><w:t>{#if show_items}{#items}- {name}{/items}{/if}</w:t></w:r></w:p>`

	data := map[string]any{
		"show_items": true,
		"items": []any{
			map[string]any{"name": "A"},
			map[string]any{"name": "B"},
		},
	}

	result, err := PreprocessTemplate(content, data)
	if err != nil {
		t.Fatalf("PreprocessTemplate failed: %v", err)
	}

	text := extractTextContent(result)
	if !strings.Contains(text, "A") || !strings.Contains(text, "B") {
		t.Errorf("Expected result to contain items A and B, got: %s", text)
	}
}

func TestPreprocessTemplate_NoDirectives(t *testing.T) {
	// Template without any directives should remain unchanged
	content := `<w:p><w:r><w:t>Hello {name}!</w:t></w:r></w:p>`

	data := map[string]any{
		"name": "World",
	}

	result, err := PreprocessTemplate(content, data)
	if err != nil {
		t.Fatalf("PreprocessTemplate failed: %v", err)
	}

	// Should preserve the original structure
	if !strings.Contains(result, "{name}") {
		t.Errorf("Expected placeholder to be preserved, got: %s", result)
	}
}

func TestExtractTextContent(t *testing.T) {
	xml := `<w:p><w:r><w:t>Hello</w:t></w:r><w:r><w:t> </w:t></w:r><w:r><w:t>World</w:t></w:r></w:p>`

	result := extractTextContent(xml)
	expected := "Hello World"

	if result != expected {
		t.Errorf("extractTextContent() = %q, want %q", result, expected)
	}
}
