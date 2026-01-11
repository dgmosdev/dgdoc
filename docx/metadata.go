package docx

import (
	"fmt"
	"regexp"
	"strings"
)

// Metadata represents document metadata properties
type Metadata struct {
	Title       string
	Subject     string
	Author      string
	Keywords    string
	Description string
	Category    string
}

// SetMetadata sets document properties in core.xml
func (t *Template) SetMetadata(meta Metadata) error {
	// Get or create core.xml
	coreXML, exists := t.files["docProps/core.xml"]
	if !exists {
		// Create basic core.xml structure
		coreXML = []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<cp:coreProperties xmlns:cp="http://schemas.openxmlformats.org/package/2006/metadata/core-properties" 
                   xmlns:dc="http://purl.org/dc/elements/1.1/" 
                   xmlns:dcterms="http://purl.org/dc/terms/" 
                   xmlns:dcmitype="http://purl.org/dc/dcmitype/" 
                   xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
	<dc:title></dc:title>
	<dc:subject></dc:subject>
	<dc:creator></dc:creator>
	<cp:keywords></cp:keywords>
	<dc:description></dc:description>
	<cp:category></cp:category>
</cp:coreProperties>`)
	}

	content := string(coreXML)

	// Replace metadata values
	if meta.Title != "" {
		content = replaceMetadataTag(content, "dc:title", meta.Title)
	}
	if meta.Subject != "" {
		content = replaceMetadataTag(content, "dc:subject", meta.Subject)
	}
	if meta.Author != "" {
		content = replaceMetadataTag(content, "dc:creator", meta.Author)
	}
	if meta.Keywords != "" {
		content = replaceMetadataTag(content, "cp:keywords", meta.Keywords)
	}
	if meta.Description != "" {
		content = replaceMetadataTag(content, "dc:description", meta.Description)
	}
	if meta.Category != "" {
		content = replaceMetadataTag(content, "cp:category", meta.Category)
	}

	t.files["docProps/core.xml"] = []byte(content)
	return nil
}

// replaceMetadataTag replaces content within an XML tag
func replaceMetadataTag(content, tag, value string) string {
	pattern := regexp.MustCompile(fmt.Sprintf(`<%s[^>]*>.*?</%s>`, tag, tag))
	replacement := fmt.Sprintf(`<%s>%s</%s>`, tag, escapeXML(value), tag)
	return pattern.ReplaceAllString(content, replacement)
}

// AddWatermark adds a text watermark to the document
// This is a simplified implementation
func (t *Template) AddWatermark(text string) error {
	// Watermarks in Word are typically added to headers
	// For simplicity, we'll create a basic watermark structure
	// A full implementation would create proper header files with watermark styling

	// This would require:
	// 1. Creating or modifying header files
	// 2. Adding watermark WordArt with proper rotation and transparency
	// 3. Updating document.xml.rels and document.xml to reference the header

	return fmt.Errorf("watermark feature requires header creation - use IncludeHeader for now")
}

// PageSetup represents page setup options
type PageSetup struct {
	PageWidth    int // in twips (1/1440 inch)
	PageHeight   int
	MarginTop    int
	MarginBottom int
	MarginLeft   int
	MarginRight  int
	Orientation  string // "portrait" or "landscape"
}

// SetPageSetup configures page margins and size
func (t *Template) SetPageSetup(setup PageSetup) error {
	content := string(t.files["word/document.xml"])

	// Find sectPr (section properties)
	sectPrPattern := regexp.MustCompile(`<w:sectPr>.*?</w:sectPr>`)

	if !sectPrPattern.MatchString(content) {
		// No sectPr found, add one at the end of body
		bodyEndIdx := strings.LastIndex(content, "</w:body>")
		if bodyEndIdx == -1 {
			return fmt.Errorf("document body not found")
		}

		sectPr := generateSectPr(setup)
		content = content[:bodyEndIdx] + sectPr + content[bodyEndIdx:]
	} else {
		// Replace existing sectPr
		sectPr := generateSectPr(setup)
		content = sectPrPattern.ReplaceAllString(content, sectPr)
	}

	t.files["word/document.xml"] = []byte(content)
	return nil
}

// generateSectPr creates section properties XML
func generateSectPr(setup PageSetup) string {
	// Default values if not specified
	if setup.PageWidth == 0 {
		setup.PageWidth = 12240 // 8.5 inches
	}
	if setup.PageHeight == 0 {
		setup.PageHeight = 15840 // 11 inches
	}
	if setup.MarginTop == 0 {
		setup.MarginTop = 1440 // 1 inch
	}
	if setup.MarginBottom == 0 {
		setup.MarginBottom = 1440
	}
	if setup.MarginLeft == 0 {
		setup.MarginLeft = 1440
	}
	if setup.MarginRight == 0 {
		setup.MarginRight = 1440
	}

	return fmt.Sprintf(`<w:sectPr>
	<w:pgSz w:w="%d" w:h="%d"/>
	<w:pgMar w:top="%d" w:right="%d" w:bottom="%d" w:left="%d" w:header="720" w:footer="720" w:gutter="0"/>
	<w:cols w:space="720"/>
</w:sectPr>`, setup.PageWidth, setup.PageHeight, setup.MarginTop, setup.MarginRight, setup.MarginBottom, setup.MarginLeft)
}

// SetReadOnly sets document protection to read-only
// This is a simplified implementation
func (t *Template) SetReadOnly(readOnly bool) error {
	if !readOnly {
		return nil // Nothing to do
	}

	// Would require modifying word/settings.xml to add documentProtection
	// For now, return informational error
	return fmt.Errorf("read-only protection requires settings.xml modification - not yet implemented")
}
