package docx

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

// Template represents a DOCX document template that can be manipulated by
// replacing placeholders with HTML or plain text content.
type Template struct {
	path    string
	files   map[string][]byte
	zipFile *zip.ReadCloser
}

// Open reads a DOCX file from the given path and prepares a Template for manipulation.
// It loads all document parts into memory for processing.
func Open(path string) (*Template, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open docx: %w", err)
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
	}

	return t, nil
}

// Close releases the underlying ZIP file reader resources.
// It should be called when the template is no longer needed.
func (t *Template) Close() error {
	if t.zipFile != nil {
		return t.zipFile.Close()
	}
	return nil
}

// SetContent replaces a placeholder in the document with HTML-formatted content.
// The placeholder should be provided without curly braces (e.g., "content" for {{content}}).
// Blocks like tables and lists will be inserted as native Word elements.
func (t *Template) SetContent(placeholder, htmlContent string) error {
	// Convert HTML to OOXML
	ooxml, err := HTMLToOOXML(htmlContent)
	if err != nil {
		return fmt.Errorf("failed to convert HTML: %w", err)
	}

	// Process document.xml
	docContent, ok := t.files["word/document.xml"]
	if !ok {
		return fmt.Errorf("document.xml not found in docx")
	}

	// Replace placeholder in document
	newContent := replacePlaceholder(string(docContent), placeholder, ooxml)
	t.files["word/document.xml"] = []byte(newContent)

	return nil
}

// replacePlaceholder handles the replacement of placeholders.
// It uses mergeSplitPlaceholder which effectively handles both single-run and multi-run placeholders.
func replacePlaceholder(content, placeholder, replacement string) string {
	return mergeSplitPlaceholder(content, placeholder, replacement)
}

// mergeSplitPlaceholder finds placeholders split across Word XML runs and replaces them
// This handles cases where Word inserts bookmarks, formatting changes, or other elements
// between parts of the placeholder text
func mergeSplitPlaceholder(content, placeholder, replacement string) string {
	fullPlaceholder := "{{" + placeholder + "}}"

	// Extract all <w:t> text content with positions
	textPattern := regexp.MustCompile(`<w:t[^>]*>([^<]*)</w:t>`)
	allMatches := textPattern.FindAllStringSubmatchIndex(content, -1)

	if len(allMatches) == 0 {
		return content
	}

	// Build a map of text content to find the placeholder
	var textBuilder strings.Builder
	type textSegment struct {
		start, end         int // Position in original content
		textStart, textEnd int // Position of text within the segment
	}
	segments := make([]textSegment, 0, len(allMatches))

	for _, match := range allMatches {
		text := content[match[2]:match[3]]
		textBuilder.WriteString(text)
		segments = append(segments, textSegment{
			start:     match[0],
			end:       match[1],
			textStart: match[2],
			textEnd:   match[3],
		})
	}

	allText := textBuilder.String()

	// Find the placeholder in the concatenated text
	placeholderIdx := strings.Index(allText, fullPlaceholder)
	if placeholderIdx == -1 {
		return content
	}

	placeholderEndIdx := placeholderIdx + len(fullPlaceholder)

	// Find which segments contain the placeholder
	var startSegmentIdx, endSegmentIdx int = -1, -1
	var charPos int

	for i, seg := range segments {
		segmentText := content[seg.textStart:seg.textEnd]
		segmentLen := len(segmentText)

		if startSegmentIdx == -1 && charPos+segmentLen > placeholderIdx {
			startSegmentIdx = i
		}
		if charPos+segmentLen >= placeholderEndIdx {
			endSegmentIdx = i
			break
		}
		charPos += segmentLen
	}

	if startSegmentIdx == -1 || endSegmentIdx == -1 {
		return content
	}

	// Check if the replacement contains block-level elements
	isBlockReplacement := strings.Contains(replacement, "<w:p") || strings.Contains(replacement, "<w:tbl")

	if isBlockReplacement {
		// Existing behavior: find and replace the whole container (paragraph or table)
		startPos := segments[startSegmentIdx].start
		endPos := segments[endSegmentIdx].end

		contentBefore := content[:startPos]
		contentAfter := content[endPos:]

		// Check if we're inside a table cell
		lastTcStart := strings.LastIndex(contentBefore, "<w:tc>")
		lastTcEnd := strings.LastIndex(contentBefore, "</w:tc>")
		insideTableCell := lastTcStart > lastTcEnd && lastTcStart != -1

		var replaceStart, replaceEnd int

		if insideTableCell {
			lastTblStart := strings.LastIndex(contentBefore, "<w:tbl>")
			if lastTblStart == -1 {
				return content
			}
			tblEndIdx := strings.Index(contentAfter, "</w:tbl>")
			if tblEndIdx == -1 {
				return content
			}
			replaceStart = lastTblStart
			replaceEnd = endPos + tblEndIdx + len("</w:tbl>")
		} else {
			pStartPattern := regexp.MustCompile(`<w:p>|<w:p\s[^>]*>`)
			pStarts := pStartPattern.FindAllStringIndex(contentBefore, -1)
			if len(pStarts) == 0 {
				return content
			}
			replaceStart = pStarts[len(pStarts)-1][0]
			pEndIdx := strings.Index(contentAfter, "</w:p>")
			if pEndIdx == -1 {
				return content
			}
			replaceEnd = endPos + pEndIdx + len("</w:p>")
		}
		return content[:replaceStart] + replacement + content[replaceEnd:]
	} else {
		// New behavior: Inline replacement
		// Only replace the text inside the <w:t> elements
		// We need to handle the fact that the placeholder might start/end mid-segment

		startSeg := segments[startSegmentIdx]
		endSeg := segments[endSegmentIdx]

		// Calculate relative positions in the first and last segments
		var currentPos int
		for i := 0; i < startSegmentIdx; i++ {
			currentPos += segments[i].textEnd - segments[i].textStart
		}
		offsetInStart := placeholderIdx - currentPos

		currentPos = 0
		for i := 0; i < endSegmentIdx; i++ {
			currentPos += segments[i].textEnd - segments[i].textStart
		}
		offsetInEnd := placeholderEndIdx - currentPos

		// Rebuild the XML but replacing only the placeholder part
		var result strings.Builder
		result.WriteString(content[:startSeg.textStart+offsetInStart])

		// If the replacement is just <w:t>text</w:t>, we only want the 'text' part
		// because we are already inside a <w:t> or at least at a position where
		// we expect text content.
		// Actually, if we are doing inline replacement, we might want to replace
		// the WHOLE sequence of <w:t> elements with the new OOXML snippet if it's multiple runs.
		// But if it's just raw text wrapped in <w:t>, let's strip the <w:t> if we are injecting
		// into an existing <w:t>.

		// Let's be smart: if replacement starts with <w:r>, it's a full run.
		// If it's just <w:t>, we can inject the text content.

		cleanReplacement := replacement
		if strings.HasPrefix(replacement, "<w:t") && strings.HasSuffix(replacement, "</w:t>") {
			// Extract content from <w:t>
			tMatch := regexp.MustCompile(`<w:t[^>]*>(.*)</w:t>`).FindStringSubmatch(replacement)
			if len(tMatch) > 1 {
				cleanReplacement = tMatch[1]
			}
		}

		result.WriteString(cleanReplacement)
		result.WriteString(content[endSeg.textStart+offsetInEnd:])

		return result.String()
	}
}

// Save writes the modified document to the specified output path.
// It compresses all document parts back into a DOCX (ZIP) file.
func (t *Template) Save(path string) error {
	// Create the output file
	outFile, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	// Create a new zip writer
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

// ReplaceText is a simpler function that just replaces text without HTML parsing
func (t *Template) ReplaceText(placeholder, text string) error {
	docContent, ok := t.files["word/document.xml"]
	if !ok {
		return fmt.Errorf("document.xml not found in docx")
	}

	// Escape XML special characters
	text = escapeXML(text)

	// Wrap in Word text element
	ooxml := fmt.Sprintf(`<w:t>%s</w:t>`, text)

	newContent := replacePlaceholder(string(docContent), placeholder, ooxml)
	t.files["word/document.xml"] = []byte(newContent)

	return nil
}

func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	return s
}
