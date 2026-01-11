package docx

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Template represents a DOCX document template that can be manipulated by
// replacing placeholders with HTML or plain text content.
type Template struct {
	path    string
	files   map[string][]byte
	zipFile *zip.ReadCloser
	nextRID int
	mediaID int
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

	t.initTracking()

	return t, nil
}

func (t *Template) initTracking() {
	// Find highest rId in word/_rels/document.xml.rels
	rels, ok := t.files["word/_rels/document.xml.rels"]
	if ok {
		re := regexp.MustCompile(`Id="rId(\d+)"`)
		matches := re.FindAllStringSubmatch(string(rels), -1)
		max := 0
		for _, m := range matches {
			var id int
			fmt.Sscanf(m[1], "%d", &id)
			if id > max {
				max = id
			}
		}
		t.nextRID = max + 1
	} else {
		t.nextRID = 1
	}

	// Find highest media index
	maxMedia := 0
	for name := range t.files {
		if strings.HasPrefix(name, "word/media/image") {
			var id int
			fmt.Sscanf(strings.TrimPrefix(name, "word/media/image"), "%d", &id)
			if id > maxMedia {
				maxMedia = id
			}
		}
	}
	t.mediaID = maxMedia + 1
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
	var ooxml string
	var err error

	// Check if this is a special placeholder (e.g., %image or %link)
	if strings.HasPrefix(placeholder, "%") {
		ooxml, err = t.handleSpecialPlaceholder(placeholder, htmlContent)
	} else {
		// Convert HTML to OOXML
		ooxml, err = t.HTMLToOOXML(htmlContent)
	}

	if err != nil {
		return fmt.Errorf("failed to process content for %s: %w", placeholder, err)
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
	fullPlaceholder := "{" + placeholder + "}"

	for {
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
			content = content[:replaceStart] + replacement + content[replaceEnd:]
		} else {
			// Inline replacement
			startSeg := segments[startSegmentIdx]
			endSeg := segments[endSegmentIdx]

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

			var result strings.Builder
			result.WriteString(content[:startSeg.textStart+offsetInStart])

			cleanReplacement := replacement
			if strings.HasPrefix(replacement, "<w:t") && strings.HasSuffix(replacement, "</w:t>") {
				tMatch := regexp.MustCompile(`<w:t[^>]*>(.*)</w:t>`).FindStringSubmatch(replacement)
				if len(tMatch) > 1 {
					cleanReplacement = tMatch[1]
				}
			}

			result.WriteString(cleanReplacement)
			result.WriteString(content[endSeg.textStart+offsetInEnd:])
			content = result.String()
		}
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

// Apply replaces multiple placeholders in the document using a map.
// Keys are placeholder names (without braces), and values can be any type
// that can be stringified (string, int, float, bool, etc.).
// If a value is a string, it's processed as HTML (SetContent).
// For other types, it's processed as plain text.
func (t *Template) Apply(data map[string]any) error {
	for placeholder, val := range data {
		// Clean placeholder name
		placeholder = strings.TrimPrefix(placeholder, "{")
		placeholder = strings.TrimSuffix(placeholder, "}")

		var htmlContent string
		if s, ok := val.(string); ok {
			htmlContent = s
		} else {
			htmlContent = fmt.Sprintf("%v", val)
		}

		if err := t.SetContent(placeholder, htmlContent); err != nil {
			return fmt.Errorf("failed to apply %s: %w", placeholder, err)
		}
	}

	// Cleanup remaining placeholders
	return t.Cleanup()
}

// Cleanup removes any remaining placeholders from the document
func (t *Template) Cleanup() error {
	docContent, ok := t.files["word/document.xml"]
	if !ok {
		return nil
	}
	content := string(docContent)

	// we need to find all unique placeholders effectively
	placeholders := extractPlaceholders(content)

	for _, p := range placeholders {
		// Replace with empty string
		content = replacePlaceholder(content, p, "")
	}

	t.files["word/document.xml"] = []byte(content)
	return nil
}

func extractPlaceholders(content string) []string {
	// Extract plain text to handle split placeholders
	textPattern := regexp.MustCompile(`<w:t[^>]*>([^<]*)</w:t>`)
	allMatches := textPattern.FindAllStringSubmatch(content, -1)

	var textBuilder strings.Builder
	for _, m := range allMatches {
		textBuilder.WriteString(m[1])
	}
	fullText := textBuilder.String()

	// Find all patterns looking like placeholders: {name}
	// We support alphanumeric, underscores, hyphens, dots
	// We do NOT include % because special placeholders like %image are usually handled or if not maybe should be kept?
	// User said {example} so standard text placeholders.
	// But if user has {%image} and didn't provide it, they probably want it gone too.
	// So I will include % in the regex.
	re := regexp.MustCompile(`\{([a-zA-Z0-9_%\-\.]+)\}`)
	matches := re.FindAllStringSubmatch(fullText, -1)

	seen := make(map[string]bool)
	var result []string

	for _, m := range matches {
		name := m[1]
		if !seen[name] {
			seen[name] = true
			result = append(result, name)
		}
	}

	return result
}

func (t *Template) handleSpecialPlaceholder(placeholder, content string) (string, error) {
	// Detect if it is an image
	isImage := false
	lower := strings.ToLower(content)
	if strings.HasPrefix(lower, "data:image/") ||
		strings.HasSuffix(lower, ".png") ||
		strings.HasSuffix(lower, ".jpg") ||
		strings.HasSuffix(lower, ".jpeg") ||
		strings.HasSuffix(lower, ".gif") {
		isImage = true
	}

	if isImage {
		var imgData []byte
		var ext string
		var err error

		if strings.HasPrefix(content, "data:image/") {
			// Handle Base64
			parts := strings.SplitN(content, ",", 2)
			if len(parts) != 2 {
				return "", fmt.Errorf("invalid base64 image")
			}
			imgData, err = base64.StdEncoding.DecodeString(parts[1])
			ext = ".png" // Default, could be more specific
		} else if strings.HasPrefix(content, "http") {
			// Handle URL
			resp, err := http.Get(content)
			if err != nil {
				return "", fmt.Errorf("failed to download image: %w", err)
			}
			defer resp.Body.Close()
			imgData, err = io.ReadAll(resp.Body)
			ext = filepath.Ext(content)
		} else {
			// Handle local file
			imgData, err = os.ReadFile(content)
			ext = filepath.Ext(content)
		}

		if err != nil {
			return "", err
		}

		rId, err := t.addImage(imgData, ext)
		if err != nil {
			return "", err
		}

		// Get dimensions
		img, _, err := image.Decode(bytes.NewReader(imgData))
		width, height := 200, 100 // Default
		if err == nil {
			bounds := img.Bounds()
			width, height = bounds.Dx(), bounds.Dy()
		}

		return generateImageXML(rId, width, height), nil
	}

	// Handle as Link
	url := content
	text := content
	if strings.Contains(content, "|") {
		parts := strings.SplitN(content, "|", 2)
		text = parts[0]
		url = parts[1]
	}

	return t.generateLinkXML(url, text), nil
}

func (t *Template) addRelationship(target, relType string, external bool) string {
	rId := fmt.Sprintf("rId%d", t.nextRID)
	t.nextRID++

	relPath := "word/_rels/document.xml.rels"
	rels, ok := t.files[relPath]
	if !ok {
		// Create minimal rels if missing
		rels = []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"></Relationships>`)
	}

	targetMode := ""
	if external {
		targetMode = ` TargetMode="External"`
	}

	newRel := fmt.Sprintf(`<Relationship Id="%s" Type="%s" Target="%s"%s/>`, rId, relType, target, targetMode)
	content := string(rels)
	if strings.Contains(content, "</Relationships>") {
		content = strings.Replace(content, "</Relationships>", newRel+"</Relationships>", 1)
	} else {
		content += newRel
	}
	t.files[relPath] = []byte(content)

	return rId
}

func (t *Template) addImage(data []byte, extension string) (string, error) {
	if extension == "" {
		extension = ".png"
	}
	imageName := fmt.Sprintf("image%d%s", t.mediaID, extension)
	t.mediaID++

	imagePath := "word/media/" + imageName
	t.files[imagePath] = data

	// Add relationship
	rId := t.addRelationship("media/"+imageName, "http://schemas.openxmlformats.org/officeDocument/2006/relationships/image", false)
	return rId, nil
}

func generateImageXML(rId string, widthPx, heightPx int) string {
	// Convert px to EMUs (1 px ~= 9525 EMUs)
	width := widthPx * 9525
	height := heightPx * 9525

	// Limit size if too large (e.g., max width 6 inches ~= 5486400 EMUs)
	maxWidth := 5486400
	if width > maxWidth {
		ratio := float64(maxWidth) / float64(width)
		width = maxWidth
		height = int(float64(height) * ratio)
	}

	return fmt.Sprintf(`<w:r><w:drawing><wp:inline distT="0" distB="0" distL="0" distR="0"><wp:extent cx="%d" cy="%d"/><wp:docPr id="1" name="Image"/><wp:cNvGraphicFramePr><a:graphicFrameLocks xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" noChangeAspect="1"/></wp:cNvGraphicFramePr><a:graphic xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main"><a:graphicData uri="http://schemas.openxmlformats.org/drawingml/2006/picture"><pic:pic xmlns:pic="http://schemas.openxmlformats.org/drawingml/2006/picture"><pic:nvPicPr><pic:cNvPr id="0" name="Picture"/><pic:cNvPicPr/></pic:nvPicPr><pic:blipFill><a:blip r:embed="%s"/><a:stretch><a:fillRect/></a:stretch></pic:blipFill><pic:spPr><a:xfrm><a:off x="0" y="0"/><a:ext cx="%d" cy="%d"/></a:xfrm><a:prstGeom prst="rect"><a:avLst/></a:prstGeom></pic:spPr></pic:pic></a:graphicData></a:graphic></wp:inline></w:drawing></w:r>`,
		width, height, rId, width, height)
}

func (t *Template) generateLinkXML(url, text string) string {
	rId := t.addRelationship(url, "http://schemas.openxmlformats.org/officeDocument/2006/relationships/hyperlink", true)

	// If text already contains OOXML (from HTML parser), we need to inject the style into each run.
	// For simplicity, let's wrap it in a <w:hyperlink> element which handles the clickability.
	// Word expects runs inside <w:hyperlink>.

	var content string
	if strings.Contains(text, "<w:r") {
		// Already has runs, just use them.
		content = text
	} else {
		// Wrap text in a run with Hyperlink style
		content = fmt.Sprintf(`<w:r><w:rPr><w:rStyle w:val="Hyperlink"/><w:u w:val="single"/><w:color w:val="0563C1"/></w:rPr><w:t xml:space="preserve">%s</w:t></w:r>`, escapeXML(text))
	}

	return fmt.Sprintf(`<w:hyperlink r:id="%s" w:history="1">%s</w:hyperlink>`, rId, content)
}

func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	return s
}
