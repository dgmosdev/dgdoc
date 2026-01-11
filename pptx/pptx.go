package pptx

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

// Template represents a PPTX document template
type Template struct {
	path    string
	files   map[string][]byte
	zipFile *zip.ReadCloser
	slides  map[string]*Slide
}

// Slide represents a slide in the PPTX file
type Slide struct {
	name string
	data []byte
}

// Open reads a PPTX file and prepares it for manipulation
func Open(path string) (*Template, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open pptx: %w", err)
	}

	t := &Template{
		path:    path,
		files:   make(map[string][]byte),
		zipFile: r,
		slides:  make(map[string]*Slide),
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

		// Track slides (ppt/slides/slide*.xml)
		if strings.HasPrefix(f.Name, "ppt/slides/slide") && strings.HasSuffix(f.Name, ".xml") {
			t.slides[f.Name] = &Slide{
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

// SetTextValue replaces a placeholder in text runs with a value
func (t *Template) SetTextValue(placeholder, value string) error {
	fullPlaceholder := "{" + placeholder + "}"

	for slidePath, slide := range t.slides {
		content := string(slide.data)

		// Replace in text runs
		// PowerPoint text format: <a:t>text</a:t> (in shape text bodies)
		content = replacePlaceholderInTextRuns(content, fullPlaceholder, value)

		t.files[slidePath] = []byte(content)
		slide.data = []byte(content)
	}

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

// Save writes the modified PPTX file
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

// SetTextValueHTML replaces a placeholder with HTML content (basic support)
func (t *Template) SetTextValueHTML(placeholder, htmlValue string) error {
	// Convert HTML to plain text (basic implementation)
	plainText := ConvertHTMLToText(htmlValue)
	return t.SetTextValue(placeholder, plainText)
}

// ProcessLoops handles {#arrayName}...{/arrayName} in slides
// Duplicates slides for each array element
func (t *Template) ProcessLoops(data map[string]any) error {
	// Find loop patterns in slides
	loopPattern := regexp.MustCompile(`\{#([a-zA-Z0-9_]+)\}`)

	for slidePath, slide := range t.slides {
		content := string(slide.data)
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

			// Expand the slide for each array element
			content = expandLoopInSlide(content, arrayName, arr)
		}

		t.files[slidePath] = []byte(content)
		slide.data = []byte(content)
	}

	return nil
}

// expandLoopInSlide expands a loop in the slide content
func expandLoopInSlide(content string, arrayName string, array []any) string {
	// Find markers
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

// replacePlaceholderInTextRuns replaces placeholders in PowerPoint text XML
func replacePlaceholderInTextRuns(content, placeholder, value string) string {
	// PowerPoint stores text in <a:t> tags within text runs
	// Pattern: <a:t>text</a:t>
	pattern := regexp.MustCompile(`(<a:t[^>]*>)([^<]*)` + regexp.QuoteMeta(placeholder) + `([^<]*)(</a:t>)`)

	for pattern.MatchString(content) {
		content = pattern.ReplaceAllString(content, `${1}${2}`+escapeXML(value)+`${3}${4}`)
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
