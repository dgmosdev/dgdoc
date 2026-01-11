package docx

import (
	"strings"
	"testing"
)

func TestSetMetadata(t *testing.T) {
	template := &Template{
		files: make(map[string][]byte),
	}

	meta := Metadata{
		Title:       "Test Document",
		Author:      "Test Author",
		Subject:     "Testing",
		Keywords:    "test, docx, metadata",
		Description: "A test document",
		Category:    "Testing",
	}

	err := template.SetMetadata(meta)
	if err != nil {
		t.Fatalf("SetMetadata failed: %v", err)
	}

	// Verify core.xml was created
	coreXML, exists := template.files["docProps/core.xml"]
	if !exists {
		t.Fatal("core.xml was not created")
	}

	content := string(coreXML)

	// Verify all metadata fields are present
	if !strings.Contains(content, "Test Document") {
		t.Error("Title not found in core.xml")
	}
	if !strings.Contains(content, "Test Author") {
		t.Error("Author not found in core.xml")
	}
	if !strings.Contains(content, "Testing") {
		t.Error("Subject not found in core.xml")
	}
	if !strings.Contains(content, "test, docx, metadata") {
		t.Error("Keywords not found in core.xml")
	}
}

func TestSetPageSetup(t *testing.T) {
	template := &Template{
		files: make(map[string][]byte),
	}
	template.files["word/document.xml"] = []byte(`<?xml version="1.0"?>
<w:document>
	<w:body>
		<w:p><w:r><w:t>Content</w:t></w:r></w:p>
	</w:body>
</w:document>`)

	setup := PageSetup{
		PageWidth:    11906, // A4 width
		PageHeight:   16838, // A4 height
		MarginTop:    1440,
		MarginBottom: 1440,
		MarginLeft:   1440,
		MarginRight:  1440,
	}

	err := template.SetPageSetup(setup)
	if err != nil {
		t.Fatalf("SetPageSetup failed: %v", err)
	}

	content := string(template.files["word/document.xml"])

	// Verify sectPr was added
	if !strings.Contains(content, "<w:sectPr>") {
		t.Error("Section properties not found")
	}
	if !strings.Contains(content, "<w:pgSz") {
		t.Error("Page size not found")
	}
	if !strings.Contains(content, "<w:pgMar") {
		t.Error("Page margins not found")
	}
}

func TestGenerateSectPr(t *testing.T) {
	setup := PageSetup{
		PageWidth:  12240,
		PageHeight: 15840,
	}

	sectPr := generateSectPr(setup)

	if !strings.Contains(sectPr, "w:pgSz") {
		t.Error("Page size tag not found")
	}
	if !strings.Contains(sectPr, "w:pgMar") {
		t.Error("Page margin tag not found")
	}
	if !strings.Contains(sectPr, `w:w="12240"`) {
		t.Error("Page width not correct")
	}
}

func TestReplaceMetadataTag(t *testing.T) {
	content := `<dc:title>Old Title</dc:title>`
	result := replaceMetadataTag(content, "dc:title", "New Title")

	if !strings.Contains(result, "New Title") {
		t.Error("Tag replacement failed")
	}
	if strings.Contains(result, "Old Title") {
		t.Error("Old value still present")
	}
}
