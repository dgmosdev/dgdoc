package pptx

import (
	"strings"
	"testing"
)

func TestSetTextValue(t *testing.T) {
	// Mock PPTX template
	template := &Template{
		files: make(map[string][]byte),
		slides: map[string]*Slide{
			"ppt/slides/slide1.xml": {
				name: "ppt/slides/slide1.xml",
				data: []byte(`<p:sld><p:cSld><p:spTree><p:sp><p:txBody><a:p><a:r><a:t>{title}</a:t></a:r></a:p></p:txBody></p:sp></p:spTree></p:cSld></p:sld>`),
			},
		},
	}

	err := template.SetTextValue("title", "My Presentation")
	if err != nil {
		t.Fatalf("SetTextValue failed: %v", err)
	}

	content := string(template.slides["ppt/slides/slide1.xml"].data)
	if !strings.Contains(content, "My Presentation") {
		t.Errorf("Expected content to contain 'My Presentation', got: %s", content)
	}
	if strings.Contains(content, "{title}") {
		t.Errorf("Placeholder should be replaced, got: %s", content)
	}
}

func TestApply(t *testing.T) {
	template := &Template{
		files: make(map[string][]byte),
		slides: map[string]*Slide{
			"ppt/slides/slide1.xml": {
				name: "ppt/slides/slide1.xml",
				data: []byte(`<p:sld><p:txBody><a:p><a:r><a:t>{name}</a:t></a:r></a:p><a:p><a:r><a:t>{age}</a:t></a:r></a:p></p:txBody></p:sld>`),
			},
		},
	}

	data := map[string]any{
		"name": "Test User",
		"age":  25,
	}

	err := template.Apply(data)
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	content := string(template.slides["ppt/slides/slide1.xml"].data)
	if !strings.Contains(content, "Test User") {
		t.Errorf("Expected 'Test User' in content")
	}
	if !strings.Contains(content, "25") {
		t.Errorf("Expected '25' in content")
	}
}

func TestHTMLConversion(t *testing.T) {
	template := &Template{
		files: make(map[string][]byte),
		slides: map[string]*Slide{
			"ppt/slides/slide1.xml": {
				name: "ppt/slides/slide1.xml",
				data: []byte(`<p:sld><a:p><a:r><a:t>{content}</a:t></a:r></a:p></p:sld>`),
			},
		},
	}

	htmlContent := "<b>Bold</b> and <i>Italic</i> text"
	err := template.SetTextValueHTML("content", htmlContent)
	if err != nil {
		t.Fatalf("SetTextValueHTML failed: %v", err)
	}

	content := string(template.slides["ppt/slides/slide1.xml"].data)

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
		files: make(map[string][]byte),
		slides: map[string]*Slide{
			"ppt/slides/slide1.xml": {
				name: "ppt/slides/slide1.xml",
				data: []byte(`<p:sld><a:p>{#items}<a:r><a:t>{name}</a:t></a:r>{/items}</a:p></p:sld>`),
			},
		},
	}

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

	content := string(template.slides["ppt/slides/slide1.xml"].data)

	// Loop markers should be removed
	if strings.Contains(content, "{#items}") || strings.Contains(content, "{/items}") {
		t.Errorf("Loop markers should be removed, got: %s", content)
	}

	// All items should be present
	if !strings.Contains(content, "Item 1") || !strings.Contains(content, "Item 2") || !strings.Contains(content, "Item 3") {
		t.Errorf("Expected all items to be present, got: %s", content)
	}
}
