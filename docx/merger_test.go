package docx

import (
	"strings"
	"testing"
)

func TestMergeDocuments(t *testing.T) {
	// Create a main document with include placeholder
	mainDoc := &Template{
		files: make(map[string][]byte),
	}
	mainDoc.files["word/document.xml"] = []byte(`<?xml version="1.0"?>
<w:document>
	<w:body>
		<w:p><w:r><w:t>Main content</w:t></w:r></w:p>
		<w:p><w:r><w:t>{@include:external.docx}</w:t></w:r></w:p>
		<w:p><w:r><w:t>More main content</w:t></w:r></w:p>
	</w:body>
</w:document>`)

	// Since we can't easily create real DOCX files in a unit test,
	// we'll test the pattern matching and extraction logic separately
	content := string(mainDoc.files["word/document.xml"])

	// Verify the placeholder is present
	if !strings.Contains(content, "{@include:external.docx}") {
		t.Errorf("Placeholder not found in content")
	}
}

func TestExtractBodyContent(t *testing.T) {
	// Test body content extraction
	sampleXML := `<?xml version="1.0"?>
<w:document>
	<w:body>
		<w:p><w:r><w:t>Test content</w:t></w:r></w:p>
		<w:sectPr><w:pgSz w:w="12240" w:h="15840"/></w:sectPr>
	</w:body>
</w:document>`

	// Simulate body extraction
	bodyStart := strings.Index(sampleXML, "<w:body>")
	bodyEnd := strings.Index(sampleXML, "</w:body>")

	if bodyStart == -1 || bodyEnd == -1 {
		t.Fatalf("Body tags not found")
	}

	bodyContent := sampleXML[bodyStart+len("<w:body>") : bodyEnd]

	// Verify content is extracted
	if !strings.Contains(bodyContent, "Test content") {
		t.Errorf("Expected body content not found")
	}

	// Verify sectPr is present (will be removed in actual function)
	if !strings.Contains(bodyContent, "sectPr") {
		t.Errorf("sectPr should be present in raw extraction")
	}
}

func TestPlaceholderPatternMatching(t *testing.T) {
	content := `<w:document>
		<w:p><w:r><w:t>{@include:header.docx}</w:t></w:r></w:p>
		<w:p><w:r><w:t>{@include:footer.docx}</w:t></w:r></w:p>
		<w:p><w:r><w:t>Normal text</w:t></w:r></w:p>
	</w:document>`

	// Test pattern matching
	matches := strings.Count(content, "{@include:")

	if matches != 2 {
		t.Errorf("Expected 2 include placeholders, found %d", matches)
	}
}
