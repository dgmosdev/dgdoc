package odt

import (
	"strings"
	"testing"
)

func TestSetTextValue(t *testing.T) {
	// Mock ODT template
	template := &Template{
		files:   make(map[string][]byte),
		content: []byte(`<?xml version="1.0"?><office:document-content><office:body><office:text><text:p>{title}</text:p></office:text></office:body></office:document-content>`),
	}
	template.files["content.xml"] = template.content

	err := template.SetTextValue("title", "My Document")
	if err != nil {
		t.Fatalf("SetTextValue failed: %v", err)
	}

	content := string(template.content)
	if !strings.Contains(content, "My Document") {
		t.Errorf("Expected content to contain 'My Document', got: %s", content)
	}
	if strings.Contains(content, "{title}") {
		t.Errorf("Placeholder should be replaced, got: %s", content)
	}
}

func TestApply(t *testing.T) {
	template := &Template{
		files:   make(map[string][]byte),
		content: []byte(`<?xml version="1.0"?><office:document-content><office:body><text:p>{name}</text:p><text:p>{age}</text:p></office:body></office:document-content>`),
	}
	template.files["content.xml"] = template.content

	data := map[string]any{
		"name": "Test User",
		"age":  25,
	}

	err := template.Apply(data)
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	content := string(template.content)
	if !strings.Contains(content, "Test User") {
		t.Errorf("Expected 'Test User' in content")
	}
	if !strings.Contains(content, "25") {
		t.Errorf("Expected '25' in content")
	}
}

func TestHTMLConversion(t *testing.T) {
	template := &Template{
		files:   make(map[string][]byte),
		content: []byte(`<?xml version="1.0"?><office:document-content><text:p>{content}</text:p></office:document-content>`),
	}
	template.files["content.xml"] = template.content

	htmlContent := "<b>Bold</b> and <i>Italic</i> text"
	err := template.SetTextValueHTML("content", htmlContent)
	if err != nil {
		t.Fatalf("SetTextValueHTML failed: %v", err)
	}

	content := string(template.content)

	// HTML tags should be stripped
	if strings.Contains(content, "<b>") || strings.Contains(content, "<i>") {
		t.Errorf("HTML tags should be stripped, got: %s", content)
	}

	// Content should be present
	if !strings.Contains(content, "Bold") || !strings.Contains(content, "Italic") {
		t.Errorf("Expected text content to be present, got: %s", content)
	}
}

func TestProcessLoops(t *testing.T) {
	template := &Template{
		files:   make(map[string][]byte),
		content: []byte(`<?xml version="1.0"?><office:document-content><office:body>{#items}<text:p>{name}</text:p>{/items}</office:body></office:document-content>`),
	}
	template.files["content.xml"] = template.content

	data := map[string]any{
		"items": []any{
			map[string]any{"name": "Item 1"},
			map[string]any{"name": "Item 2"},
			map[string]any{"name": "Item 3"},
		},
	}

	err := template.ProcessLoops(data)
	if err != nil {
		t.Fatalf("ProcessLoops failed: %v", err)
	}

	content := string(template.content)

	// Loop markers should be removed
	if strings.Contains(content, "{#items}") || strings.Contains(content, "{/items}") {
		t.Errorf("Loop markers should be removed, got: %s", content)
	}

	// All items should be present
	if !strings.Contains(content, "Item 1") || !strings.Contains(content, "Item 2") || !strings.Contains(content, "Item 3") {
		t.Errorf("Expected all items to be present, got: %s", content)
	}
}
