package xlsx

import (
	"strings"
	"testing"
)

func TestSetCellValue(t *testing.T) {
	// Mock XLSX template
	template := &Template{
		files: make(map[string][]byte),
		sheets: map[string]*Sheet{
			"xl/worksheets/sheet1.xml": {
				name: "xl/worksheets/sheet1.xml",
				data: []byte(`<worksheet><sheetData><row><c t="inlineStr"><is><t>{name}</t></is></c></row></sheetData></worksheet>`),
			},
		},
	}

	err := template.SetCellValue("name", "Ahmet")
	if err != nil {
		t.Fatalf("SetCellValue failed: %v", err)
	}

	content := string(template.sheets["xl/worksheets/sheet1.xml"].data)
	if !strings.Contains(content, "Ahmet") {
		t.Errorf("Expected content to contain 'Ahmet', got: %s", content)
	}
	if strings.Contains(content, "{name}") {
		t.Errorf("Placeholder should be replaced, got: %s", content)
	}
}

func TestApply(t *testing.T) {
	template := &Template{
		files: make(map[string][]byte),
		sheets: map[string]*Sheet{
			"xl/worksheets/sheet1.xml": {
				name: "xl/worksheets/sheet1.xml",
				data: []byte(`<worksheet><sheetData><row><c><is><t>{name}</t></is></c><c><v>{age}</v></c></row></sheetData></worksheet>`),
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

	content := string(template.sheets["xl/worksheets/sheet1.xml"].data)
	if !strings.Contains(content, "Test User") {
		t.Errorf("Expected 'Test User' in content")
	}
	if !strings.Contains(content, "25") {
		t.Errorf("Expected '25' in content")
	}
}

func TestFormulaPreservation(t *testing.T) {
	template := &Template{
		files: make(map[string][]byte),
		sheets: map[string]*Sheet{
			"xl/worksheets/sheet1.xml": {
				name: "xl/worksheets/sheet1.xml",
				// Cell A1 has placeholder, Cell B1 has formula
				data: []byte(`<worksheet><sheetData><row><c r="A1" t="inlineStr"><is><t>{value}</t></is></c><c r="B1"><f>SUM(A1:A10)</f><v>100</v></c></row></sheetData></worksheet>`),
			},
		},
	}

	err := template.SetCellValue("value", "Test")
	if err != nil {
		t.Fatalf("SetCellValue failed: %v", err)
	}

	content := string(template.sheets["xl/worksheets/sheet1.xml"].data)

	// Placeholder should be replaced
	if !strings.Contains(content, "Test") {
		t.Errorf("Expected 'Test' in content")
	}

	// Formula should be preserved
	if !strings.Contains(content, "<f>SUM(A1:A10)</f>") {
		t.Errorf("Formula should be preserved, got: %s", content)
	}
}

func TestHTMLConversion(t *testing.T) {
	template := &Template{
		files: make(map[string][]byte),
		sheets: map[string]*Sheet{
			"xl/worksheets/sheet1.xml": {
				name: "xl/worksheets/sheet1.xml",
				data: []byte(`<worksheet><sheetData><row><c t="inlineStr"><is><t>{content}</t></is></c></row></sheetData></worksheet>`),
			},
		},
	}

	htmlContent := "<b>Bold</b> and <i>Italic</i> text"
	err := template.SetCellValueHTML("content", htmlContent)
	if err != nil {
		t.Fatalf("SetCellValueHTML failed: %v", err)
	}

	content := string(template.sheets["xl/worksheets/sheet1.xml"].data)

	// HTML tags should be stripped
	if strings.Contains(content, "<b>") || strings.Contains(content, "<i>") {
		t.Errorf("HTML tags should be stripped, got: %s", content)
	}

	// Content should be present
	if !strings.Contains(content, "Bold") || !strings.Contains(content, "Italic") {
		t.Errorf("Expected text content to be present, got: %s", content)
	}
}
