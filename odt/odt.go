package odt

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

// Template represents an ODT document template
type Template struct {
	path    string
	files   map[string][]byte
	zipFile *zip.ReadCloser
	content []byte // content.xml contains the main document
}

// Open reads an ODT file and prepares it for manipulation
func Open(path string) (*Template, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open odt: %w", err)
	}

	t := &Template{
		path:    path,
		files:   make(map[string][]byte),
		zipFile: r,
	}

	// Read all files into memory
	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			_ = r.Close()
			return nil, fmt.Errorf("failed to read file %s: %w", f.Name, err)
		}

		fileContent, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			_ = r.Close()
			return nil, fmt.Errorf("failed to read content of %s: %w", f.Name, err)
		}

		t.files[f.Name] = fileContent

		// Track main content.xml
		if f.Name == "content.xml" {
			t.content = fileContent
		}
	}

	if t.content == nil {
		_ = r.Close()
		return nil, fmt.Errorf("content.xml not found in ODT file")
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

// SetTextValue replaces a placeholder in text with a value
func (t *Template) SetTextValue(placeholder, value string) error {
	fullPlaceholder := "{" + placeholder + "}"

	content := string(t.content)

	// Replace in text runs
	// ODT text format: <text:span>text</text:span> or <text:p>text</text:p>
	content = replacePlaceholderInText(content, fullPlaceholder, value)

	t.content = []byte(content)
	t.files["content.xml"] = t.content

	return nil
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

		if err := t.SetTextValue(placeholder, strValue); err != nil {
			return fmt.Errorf("failed to apply %s: %w", placeholder, err)
		}
	}

	return nil
}

// Save writes the modified ODT file
func (t *Template) Save(path string) error {
	outFile, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer func() { _ = outFile.Close() }()

	w := zip.NewWriter(outFile)
	defer func() { _ = w.Close() }()

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

// SetTextValueHTML replaces a placeholder with HTML content (basic support)
func (t *Template) SetTextValueHTML(placeholder, htmlValue string) error {
	// Convert HTML to plain text (basic implementation)
	plainText := ConvertHTMLToText(htmlValue)
	return t.SetTextValue(placeholder, plainText)
}

// ProcessLoops handles {#arrayName}...{/arrayName} in content
func (t *Template) ProcessLoops(data map[string]any) error {
	content := string(t.content)

	// Find loop patterns
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

		// Expand the loop
		content = expandLoopInContent(content, arrayName, arr)
	}

	t.content = []byte(content)
	t.files["content.xml"] = t.content

	return nil
}

// expandLoopInContent expands a loop in the content
func expandLoopInContent(content string, arrayName string, array []any) string {
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

	// Extract the template section
	template := content[startIdx : endIdx+len(endMarker)]

	// Remove markers from template
	template = strings.ReplaceAll(template, startMarker, "")
	template = strings.ReplaceAll(template, endMarker, "")

	// Expand for each item
	var expanded strings.Builder
	for _, item := range array {
		itemContent := template

		// Replace placeholders in the item
		if itemMap, ok := item.(map[string]any); ok {
			for key, val := range itemMap {
				placeholder := "{" + key + "}"
				itemContent = strings.ReplaceAll(itemContent, placeholder, fmt.Sprintf("%v", val))
			}
		}

		expanded.WriteString(itemContent)
	}

	// Replace the loop block with expanded content
	result := content[:startIdx] + expanded.String() + content[endIdx+len(endMarker):]
	return result
}

// replacePlaceholderInText replaces placeholders in ODT text XML
func replacePlaceholderInText(content, placeholder, value string) string {
	// ODT uses text:span and text:p for text content
	// We'll use a simple approach: replace in any text content

	// Try multiple patterns for ODT text elements
	patterns := []string{
		`(<text:span[^>]*>)([^<]*)` + regexp.QuoteMeta(placeholder) + `([^<]*)(</text:span>)`,
		`(<text:p[^>]*>)([^<]*)` + regexp.QuoteMeta(placeholder) + `([^<]*)(</text:p>)`,
		`(<text:h[^>]*>)([^<]*)` + regexp.QuoteMeta(placeholder) + `([^<]*)(</text:h>)`,
	}

	for _, patternStr := range patterns {
		pattern := regexp.MustCompile(patternStr)
		for pattern.MatchString(content) {
			content = pattern.ReplaceAllString(content, `${1}${2}`+escapeXML(value)+`${3}${4}`)
		}
	}

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
