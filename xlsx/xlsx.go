package xlsx

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

// Template represents an XLSX document template
type Template struct {
	path    string
	files   map[string][]byte
	zipFile *zip.ReadCloser
	sheets  map[string]*Sheet
}

// Sheet represents a worksheet in the XLSX file
type Sheet struct {
	name string
	data []byte
}

// CellStyle represents formatting options for a cell
type CellStyle struct {
	BackgroundColor string // Hex color like "FFFF00" for yellow
	FontColor       string // Hex color like "FF0000" for red
	FontSize        int    // Font size in points
	Bold            bool
	Border          bool // Simple border on/off
}

// Open reads an XLSX file and prepares it for manipulation
func Open(path string) (*Template, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open xlsx: %w", err)
	}

	t := &Template{
		path:    path,
		files:   make(map[string][]byte),
		zipFile: r,
		sheets:  make(map[string]*Sheet),
	}

	// Read all files into memory
	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			r.Close()
			return nil, fmt.Errorf("failed to read file %s: %w", f.Name, err)
		}

		content, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			r.Close()
			return nil, fmt.Errorf("failed to read content of %s: %w", f.Name, err)
		}

		t.files[f.Name] = content

		// Track sheets (xl/worksheets/sheet*.xml)
		if strings.HasPrefix(f.Name, "xl/worksheets/sheet") && strings.HasSuffix(f.Name, ".xml") {
			t.sheets[f.Name] = &Sheet{
				name: f.Name,
				data: content,
			}
		}
	}

	return t, nil
}

// Close releases resources
func (t *Template) Close() error {
	if t.zipFile != nil {
		return t.zipFile.Close()
	}
	return nil
}

// SetCellValue replaces a placeholder in cells with a value
func (t *Template) SetCellValue(placeholder, value string) error {
	fullPlaceholder := "{" + placeholder + "}"

	for sheetPath, sheet := range t.sheets {
		content := string(sheet.data)

		// Replace in cell values
		// Excel cell format: <c><v>value</v></c> or <c t="inlineStr"><is><t>text</t></is></c>
		content = replacePlaceholderInCells(content, fullPlaceholder, value)

		t.files[sheetPath] = []byte(content)
		sheet.data = []byte(content)
	}

	return nil
}

// SetCellValueHTML replaces a placeholder with HTML content (basic support)
func (t *Template) SetCellValueHTML(placeholder, htmlValue string) error {
	// Convert HTML to plain text (basic implementation)
	plainText := ConvertHTMLToRichText(htmlValue)
	return t.SetCellValue(placeholder, plainText)
}

// Apply replaces multiple placeholders
func (t *Template) Apply(data map[string]any) error {
	for placeholder, val := range data {
		// Clean placeholder name
		placeholder = strings.TrimPrefix(placeholder, "{")
		placeholder = strings.TrimSuffix(placeholder, "}")

		var strValue string
		if s, ok := val.(string); ok {
			strValue = s
		} else {
			strValue = fmt.Sprintf("%v", val)
		}

		if err := t.SetCellValue(placeholder, strValue); err != nil {
			return fmt.Errorf("failed to apply %s: %w", placeholder, err)
		}
	}

	return nil
}

// Save writes the modified XLSX file
func (t *Template) Save(path string) error {
	outFile, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	w := zip.NewWriter(outFile)
	defer w.Close()

	// Write all files
	for name, content := range t.files {
		writer, err := w.Create(name)
		if err != nil {
			return fmt.Errorf("failed to create file %s in zip: %w", name, err)
		}

		_, err = io.Copy(writer, bytes.NewReader(content))
		if err != nil {
			return fmt.Errorf("failed to write file %s: %w", name, err)
		}
	}

	return nil
}

// replacePlaceholderInCells replaces placeholders in Excel cell XML
// while preserving formulas
func replacePlaceholderInCells(content, placeholder, value string) string {
	// Excel stores text in <t> tags and formulas in <f> tags
	// Pattern: <t>text</t> or <t xml:space="preserve">text</t>

	// First, we need to avoid replacing placeholders in cells that contain formulas
	// Excel cell with formula: <c><f>SUM(A1:A10)</f><v>result</v></c>
	// We want to skip these cells

	// Find all cell elements
	cellPattern := regexp.MustCompile(`<c[^>]*>.*?</c>`)

	content = cellPattern.ReplaceAllStringFunc(content, func(cell string) string {
		// Skip cells with formulas
		if strings.Contains(cell, "<f>") || strings.Contains(cell, "<f ") {
			return cell
		}

		// Only replace in text cells
		textPattern := regexp.MustCompile(`(<t[^>]*>)([^<]*)` + regexp.QuoteMeta(placeholder) + `([^<]*)(</t>)`)
		if textPattern.MatchString(cell) {
			cell = textPattern.ReplaceAllString(cell, `${1}${2}`+escapeXML(value)+`${3}${4}`)
		}

		// Also replace in value cells (<v> tags for numeric values)
		valuePattern := regexp.MustCompile(`(<v[^>]*>)([^<]*)` + regexp.QuoteMeta(placeholder) + `([^<]*)(</v>)`)
		if valuePattern.MatchString(cell) {
			cell = valuePattern.ReplaceAllString(cell, `${1}${2}`+escapeXML(value)+`${3}${4}`)
		}

		return cell
	})

	return content
}

// escapeXML escapes special XML characters
func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	return s
}

// Row represents a row with dynamic data for loops
type Row struct {
	XMLNode xml.Name
	Cells   []Cell
}

// Cell represents a cell in Excel
type Cell struct {
	Ref   string        `xml:"r,attr,omitempty"`
	Type  string        `xml:"t,attr,omitempty"`
	Value string        `xml:"v,omitempty"`
	IS    *InlineString `xml:"is,omitempty"`
}

// InlineString for text cells
type InlineString struct {
	T string `xml:"t"`
}

// ProcessLoops handles {#arrayName}...{/arrayName} in rows
func (t *Template) ProcessLoops(data map[string]any) error {
	for sheetPath, sheet := range t.sheets {
		content := string(sheet.data)

		// Find loop patterns in the sheet
		// Simplified: look for {#items} in a cell, duplicate that row
		loopPattern := regexp.MustCompile(`\{#([a-zA-Z0-9_]+)\}`)
		matches := loopPattern.FindAllStringSubmatch(content, -1)

		for _, match := range matches {
			if len(match) < 2 {
				continue
			}
			arrayName := match[1]

			// Get array from data
			arrayData, ok := data[arrayName]
			if !ok {
				continue
			}

			arr, ok := arrayData.([]any)
			if !ok {
				continue
			}

			// Find the row containing this loop marker
			// This is simplified - in a full implementation, we'd parse the XML properly
			content = expandLoopInSheet(content, arrayName, arr)
		}

		t.files[sheetPath] = []byte(content)
		sheet.data = []byte(content)
	}

	return nil
}

// expandLoopInSheet expands a loop in the sheet content
// This is a simplified implementation
func expandLoopInSheet(content string, arrayName string, array []any) string {
	// Find row with {#arrayName}
	startMarker := "{#" + arrayName + "}"
	endMarker := "{/" + arrayName + "}"

	startIdx := strings.Index(content, startMarker)
	if startIdx == -1 {
		return content
	}

	endIdx := strings.Index(content[startIdx:], endMarker)
	if endIdx == -1 {
		return content
	}
	endIdx += startIdx

	// Extract the row template (simplified)
	template := content[startIdx : endIdx+len(endMarker)]

	// Remove markers from template
	template = strings.ReplaceAll(template, startMarker, "")
	template = strings.ReplaceAll(template, endMarker, "")

	// Expand for each item
	var expanded strings.Builder
	for _, item := range array {
		rowContent := template

		// Replace placeholders in the row
		if itemMap, ok := item.(map[string]any); ok {
			for key, val := range itemMap {
				placeholder := "{" + key + "}"
				rowContent = strings.ReplaceAll(rowContent, placeholder, fmt.Sprintf("%v", val))
			}
		}

		expanded.WriteString(rowContent)
	}

	// Replace the loop block with expanded content
	result := content[:startIdx] + expanded.String() + content[endIdx+len(endMarker):]
	return result
}
