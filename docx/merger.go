package docx

import (
	"archive/zip"
	"fmt"
	"io"
	"regexp"
	"strings"
)

// MergeDocuments merges content from another DOCX file at a placeholder location
// Syntax: {@include:filename.docx}
func (t *Template) MergeDocuments(placeholders map[string]string) error {
	content := string(t.files["word/document.xml"])

	// Find {@include:filename} patterns
	includePattern := regexp.MustCompile(`\{@include:([^}]+)\}`)
	matches := includePattern.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		placeholder := match[0]
		filename := strings.TrimSpace(match[1])

		// Get the file path from placeholders map
		filepath, ok := placeholders[filename]
		if !ok {
			// Skip if not provided
			continue
		}

		// Open the external document
		externalDoc, err := openExternalDocument(filepath)
		if err != nil {
			return fmt.Errorf("failed to open external document %s: %w", filepath, err)
		}

		// Extract the body content
		externalContent, err := extractBodyContent(externalDoc)
		if err != nil {
			return fmt.Errorf("failed to extract content from %s: %w", filepath, err)
		}

		// Replace the placeholder with the external content
		content = strings.ReplaceAll(content, placeholder, externalContent)
		externalDoc.Close()
	}

	t.files["word/document.xml"] = []byte(content)
	return nil
}

// openExternalDocument opens an external DOCX file
func openExternalDocument(path string) (*zip.ReadCloser, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open docx: %w", err)
	}
	return r, nil
}

// extractBodyContent extracts the content from inside <w:body> tags
func extractBodyContent(zipFile *zip.ReadCloser) (string, error) {
	// Find document.xml
	var documentXML []byte
	for _, f := range zipFile.File {
		if f.Name == "word/document.xml" {
			rc, err := f.Open()
			if err != nil {
				return "", err
			}

			content, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				return "", err
			}

			documentXML = content
			break
		}
	}

	if documentXML == nil {
		return "", fmt.Errorf("document.xml not found")
	}

	// Extract content between <w:body> and </w:body>
	content := string(documentXML)
	bodyStart := strings.Index(content, "<w:body>")
	bodyEnd := strings.Index(content, "</w:body>")

	if bodyStart == -1 || bodyEnd == -1 {
		return "", fmt.Errorf("body tags not found")
	}

	// Extract just the content inside the body (excluding the tags themselves)
	bodyContent := content[bodyStart+len("<w:body>") : bodyEnd]

	// Remove the last sectPr (section properties) as it will conflict with the main document
	sectPrPattern := regexp.MustCompile(`<w:sectPr>.*?</w:sectPr>`)
	bodyContent = sectPrPattern.ReplaceAllString(bodyContent, "")

	return bodyContent, nil
}

// IncludeHeader includes a header from an external DOCX file
// This is a simplified implementation
func (t *Template) IncludeHeader(headerPath string) error {
	// Open external document
	externalDoc, err := openExternalDocument(headerPath)
	if err != nil {
		return fmt.Errorf("failed to open header document: %w", err)
	}
	defer externalDoc.Close()

	// Find header file in external document
	for _, f := range externalDoc.File {
		if strings.HasPrefix(f.Name, "word/header") && strings.HasSuffix(f.Name, ".xml") {
			rc, err := f.Open()
			if err != nil {
				return err
			}

			content, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				return err
			}

			// Copy header to our document
			// Use the same filename
			t.files[f.Name] = content

			// Note: In a full implementation, we would need to:
			// 1. Update document.xml.rels to add relationship to the header
			// 2. Add header reference in document.xml sectPr
			// For now, this is a basic copy

			break
		}
	}

	return nil
}

// IncludeFooter includes a footer from an external DOCX file
// This is a simplified implementation
func (t *Template) IncludeFooter(footerPath string) error {
	// Open external document
	externalDoc, err := openExternalDocument(footerPath)
	if err != nil {
		return fmt.Errorf("failed to open footer document: %w", err)
	}
	defer externalDoc.Close()

	// Find footer file in external document
	for _, f := range externalDoc.File {
		if strings.HasPrefix(f.Name, "word/footer") && strings.HasSuffix(f.Name, ".xml") {
			rc, err := f.Open()
			if err != nil {
				return err
			}

			content, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				return err
			}

			// Copy footer to our document
			t.files[f.Name] = content

			break
		}
	}

	return nil
}
