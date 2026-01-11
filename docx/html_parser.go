package docx

import (
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

func (t *Template) HTMLToOOXML(htmlContent string) (string, error) {
	// Parse the HTML
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Convert to OOXML
	var builder strings.Builder
	t.convertNode(doc, &builder, &textState{})

	return builder.String(), nil
}

// textState tracks the current text formatting state
type textState struct {
	bold      bool
	italic    bool
	underline bool
	strike    bool
	fontSize  string
	fontColor string
	bgColor   string // background/highlight color
	inList    bool
	listType  string // "ul" or "ol"
	listLevel int
	listNum   int
}

// copyState creates a copy of the current text state
func (s *textState) copy() *textState {
	return &textState{
		bold:      s.bold,
		italic:    s.italic,
		underline: s.underline,
		strike:    s.strike,
		fontSize:  s.fontSize,
		fontColor: s.fontColor,
		bgColor:   s.bgColor,
		inList:    s.inList,
		listType:  s.listType,
		listLevel: s.listLevel,
		listNum:   s.listNum,
	}
}

func (t *Template) convertNode(n *html.Node, builder *strings.Builder, state *textState) {
	switch n.Type {
	case html.ElementNode:
		t.handleElement(n, builder, state)
	case html.TextNode:
		parent := ""
		if n.Parent != nil {
			parent = strings.ToLower(n.Parent.Data)
		}

		var text string
		// List of tags where whitespace matters (inline contexts)
		// We preserve single spaces here to prevent "WordCombining" issues
		if parent == "p" || parent == "span" || parent == "a" || parent == "li" ||
			parent == "h1" || parent == "h2" || parent == "h3" || parent == "h4" ||
			parent == "h5" || parent == "h6" || parent == "b" || parent == "strong" ||
			parent == "i" || parent == "em" || parent == "u" || parent == "s" ||
			parent == "strike" || parent == "del" || parent == "td" || parent == "th" {

			// Collapse whitespace but preserve single spaces
			// standard HTML behavior: newlines/tabs -> space, multiple spaces -> single space
			re := regexp.MustCompile(`[\s\r\n]+`)
			text = re.ReplaceAllString(n.Data, " ")
		} else {
			// Block context - trim aggressively to avoid stray runs between blocks
			text = strings.TrimSpace(n.Data)
		}

		if text != "" {
			writeText(builder, text, state)
		}
	case html.DocumentNode:
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			t.convertNode(c, builder, state)
		}
	}
}

func (t *Template) handleElement(n *html.Node, builder *strings.Builder, state *textState) {
	tag := strings.ToLower(n.Data)
	newState := state.copy()

	// Apply any inline styles (color, background-color, etc.)
	applyStyleToState(n, newState)

	switch tag {
	case "html", "body", "head", "div", "article", "section", "main", "header", "footer", "nav", "aside":
		// Container elements - just process children
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			t.convertNode(c, builder, newState)
		}

	case "span":
		// Span with potential styling - process children with applied styles
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			t.convertNode(c, builder, newState)
		}

	case "p":
		t.writeParagraph(n, builder, newState)

	case "br":
		builder.WriteString(`<w:r><w:br/></w:r>`)

	case "strong", "b":
		newState.bold = true
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			t.convertNode(c, builder, newState)
		}

	case "em", "i":
		newState.italic = true
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			t.convertNode(c, builder, newState)
		}

	case "u":
		newState.underline = true
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			t.convertNode(c, builder, newState)
		}

	case "s", "strike", "del":
		newState.strike = true
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			t.convertNode(c, builder, newState)
		}

	case "ul":
		newState.inList = true
		newState.listType = "ul"
		newState.listLevel++
		newState.listNum = 0
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			t.convertNode(c, builder, newState)
		}

	case "ol":
		newState.inList = true
		newState.listType = "ol"
		newState.listLevel++
		newState.listNum = 0
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			t.convertNode(c, builder, newState)
		}

	case "li":
		t.writeListItem(n, builder, newState)

	case "table":
		t.writeTable(n, builder, newState)

	case "h1", "h2", "h3", "h4", "h5", "h6":
		t.writeHeading(n, builder, newState, tag)

	case "a":
		t.writeAnchor(n, builder, newState)

	case "img":
		t.writeImage(n, builder, newState)

	default:
		// Unknown element - process children
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			t.convertNode(c, builder, newState)
		}
	}
}

func (t *Template) writeParagraph(n *html.Node, builder *strings.Builder, state *textState) {
	builder.WriteString(`<w:p><w:pPr></w:pPr>`)

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		t.convertNode(c, builder, state)
	}

	builder.WriteString(`</w:p>`)
}

func writeText(builder *strings.Builder, text string, state *textState) {
	builder.WriteString(`<w:r>`)

	// Write run properties if any formatting is applied
	if state.bold || state.italic || state.underline || state.strike || state.fontSize != "" || state.fontColor != "" || state.bgColor != "" {
		builder.WriteString(`<w:rPr>`)
		if state.bold {
			builder.WriteString(`<w:b/><w:bCs/>`)
		}
		if state.italic {
			builder.WriteString(`<w:i/><w:iCs/>`)
		}
		if state.underline {
			builder.WriteString(`<w:u w:val="single"/>`)
		}
		if state.strike {
			builder.WriteString(`<w:strike/>`)
		}
		if state.fontSize != "" {
			fmt.Fprintf(builder, `<w:sz w:val="%s"/><w:szCs w:val="%s"/>`, state.fontSize, state.fontSize)
		}
		if state.fontColor != "" {
			fmt.Fprintf(builder, `<w:color w:val="%s"/>`, state.fontColor)
		}
		if state.bgColor != "" {
			// w:shd is used for background/highlight color in Word
			fmt.Fprintf(builder, `<w:shd w:val="clear" w:color="auto" w:fill="%s"/>`, state.bgColor)
		}
		builder.WriteString(`</w:rPr>`)
	}

	// Escape and write text
	text = escapeXML(text)

	// Preserve spaces
	if strings.HasPrefix(text, " ") || strings.HasSuffix(text, " ") {
		fmt.Fprintf(builder, `<w:t xml:space="preserve">%s</w:t>`, text)
	} else {
		fmt.Fprintf(builder, `<w:t>%s</w:t>`, text)
	}

	builder.WriteString(`</w:r>`)
}

func (t *Template) writeListItem(n *html.Node, builder *strings.Builder, state *textState) {
	state.listNum++

	builder.WriteString(`<w:p>`)
	builder.WriteString(`<w:pPr>`)

	// Add list formatting
	builder.WriteString(`<w:pStyle w:val="ListParagraph"/>`)
	fmt.Fprintf(builder, `<w:numPr><w:ilvl w:val="%d"/>`, state.listLevel-1)

	// Use different numId for ordered vs unordered
	if state.listType == "ol" {
		builder.WriteString(`<w:numId w:val="1"/>`)
	} else {
		builder.WriteString(`<w:numId w:val="2"/>`)
	}
	builder.WriteString(`</w:numPr>`)

	// Add indentation
	indent := state.listLevel * 720 // 720 twips = 0.5 inch
	fmt.Fprintf(builder, `<w:ind w:left="%d" w:hanging="360"/>`, indent)

	builder.WriteString(`</w:pPr>`)

	// Write list marker for bullet
	if state.listType == "ul" {
		builder.WriteString(`<w:r><w:rPr></w:rPr><w:t>• </w:t></w:r>`)
	} else {
		fmt.Fprintf(builder, `<w:r><w:rPr></w:rPr><w:t>%d. </w:t></w:r>`, state.listNum)
	}

	// Process children
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		t.convertNode(c, builder, state)
	}
	builder.WriteString(`</w:p>`)
}

func (t *Template) writeTable(n *html.Node, builder *strings.Builder, state *textState) {
	// First, count columns by examining the first row
	colCount := countTableColumns(n)

	builder.WriteString(`<w:tbl>`)

	// Table properties
	builder.WriteString(`<w:tblPr>`)
	builder.WriteString(`<w:tblStyle w:val="TableGrid"/>`)
	builder.WriteString(`<w:tblW w:w="5000" w:type="pct"/>`) // 100% width
	builder.WriteString(`<w:tblBorders>`)
	builder.WriteString(`<w:top w:val="single" w:sz="4" w:space="0" w:color="000000"/>`)
	builder.WriteString(`<w:left w:val="single" w:sz="4" w:space="0" w:color="000000"/>`)
	builder.WriteString(`<w:bottom w:val="single" w:sz="4" w:space="0" w:color="000000"/>`)
	builder.WriteString(`<w:right w:val="single" w:sz="4" w:space="0" w:color="000000"/>`)
	builder.WriteString(`<w:insideH w:val="single" w:sz="4" w:space="0" w:color="000000"/>`)
	builder.WriteString(`<w:insideV w:val="single" w:sz="4" w:space="0" w:color="000000"/>`)
	builder.WriteString(`</w:tblBorders>`)
	builder.WriteString(`<w:tblLook w:val="04A0" w:firstRow="1" w:lastRow="0" w:firstColumn="1" w:lastColumn="0" w:noHBand="0" w:noVBand="1"/>`)
	builder.WriteString(`</w:tblPr>`)

	// Table grid - define columns
	builder.WriteString(`<w:tblGrid>`)
	colWidth := 9000 / colCount // Distribute width evenly (9000 twips ≈ full page)
	for i := 0; i < colCount; i++ {
		fmt.Fprintf(builder, `<w:gridCol w:w="%d"/>`, colWidth)
	}
	builder.WriteString(`</w:tblGrid>`)

	// Track rowspan state: map[colIndex]remainingSpans
	rowspanTracker := make(map[int]int)

	// Process rows
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode {
			tag := strings.ToLower(c.Data)
			switch tag {
			case "tbody", "thead", "tfoot":
				for tr := c.FirstChild; tr != nil; tr = tr.NextSibling {
					if tr.Type == html.ElementNode && strings.ToLower(tr.Data) == "tr" {
						t.writeTableRowWithMerge(tr, builder, state, colWidth, rowspanTracker)
					}
				}
			case "tr":
				t.writeTableRowWithMerge(c, builder, state, colWidth, rowspanTracker)
			}
		}
	}

	builder.WriteString(`</w:tbl>`)
}

// countTableColumns counts the number of columns in the table
func countTableColumns(n *html.Node) int {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode {
			tag := strings.ToLower(c.Data)
			switch tag {
			case "tbody", "thead", "tfoot":
				for tr := c.FirstChild; tr != nil; tr = tr.NextSibling {
					if tr.Type == html.ElementNode && strings.ToLower(tr.Data) == "tr" {
						return countRowCells(tr)
					}
				}
			case "tr":
				return countRowCells(c)
			}
		}
	}
	return 1
}

func countRowCells(tr *html.Node) int {
	count := 0
	for c := tr.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode {
			tag := strings.ToLower(c.Data)
			if tag == "td" || tag == "th" {
				count++
			}
		}
	}
	if count == 0 {
		return 1
	}
	return count
}

// writeTableRow is deprecated - use writeTableRowWithMerge for all cases
// Kept for backwards compatibility but not used

// writeTableRowWithMerge handles rows with cell merging (rowspan tracking)
func (t *Template) writeTableRowWithMerge(n *html.Node, builder *strings.Builder, state *textState, colWidth int, rowspanTracker map[int]int) {
	builder.WriteString(`<w:tr>`)

	colIdx := 0
	cellIdx := 0

	// Get all cells in this row first
	cells := []*html.Node{}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode {
			tag := strings.ToLower(c.Data)
			if tag == "td" || tag == "th" {
				cells = append(cells, c)
			}
		}
	}

	// Process each column position
	for cellIdx < len(cells) || rowspanTracker[colIdx] > 0 {
		// If this column is part of a rowspan from above, write continue cell
		if rowspanTracker[colIdx] > 0 {
			t.writeMergedCell(builder, colWidth, "continue")
			colIdx++
			continue
		}

		// No more cells to process
		if cellIdx >= len(cells) {
			break
		}

		// Process the next cell from HTML
		c := cells[cellIdx]
		cellIdx++

		tag := strings.ToLower(c.Data)

		// Get colspan and rowspan attributes
		colspan := 1
		rowspan := 1
		for _, attr := range c.Attr {
			switch attr.Key {
			case "colspan":
				_, _ = fmt.Sscanf(attr.Val, "%d", &colspan)
			case "rowspan":
				_, _ = fmt.Sscanf(attr.Val, "%d", &rowspan)
			}
		}

		// Write the cell with merge properties
		vMergeType := ""
		if rowspan > 1 {
			vMergeType = "restart"
			// Track this rowspan for future rows
			for i := 0; i < colspan; i++ {
				rowspanTracker[colIdx+i] = rowspan - 1
			}
		}

		t.writeTableCellWithMerge(c, builder, state, tag == "th", colWidth, colspan, vMergeType)
		colIdx += colspan
	}

	// Decrement all rowspan counters for this row
	for col := range rowspanTracker {
		if rowspanTracker[col] > 0 {
			rowspanTracker[col]--
			if rowspanTracker[col] == 0 {
				delete(rowspanTracker, col)
			}
		}
	}

	builder.WriteString(`</w:tr>`)
}

// writeTableCell is deprecated - use writeTableCellWithMerge for all cases
// Kept for backwards compatibility but not used

// writeTableCellWithMerge writes a table cell with merge properties
func (t *Template) writeTableCellWithMerge(n *html.Node, builder *strings.Builder, state *textState, isHeader bool, colWidth int, colspan int, vMergeType string) {
	builder.WriteString(`<w:tc>`)
	builder.WriteString(`<w:tcPr>`)

	// Set cell width (multiply by colspan)
	totalWidth := colWidth * colspan
	fmt.Fprintf(builder, `<w:tcW w:w="%d" w:type="dxa"/>`, totalWidth)

	// Horizontal merge (colspan)
	if colspan > 1 {
		fmt.Fprintf(builder, `<w:gridSpan w:val="%d"/>`, colspan)
	}

	// Vertical merge (rowspan)
	if vMergeType != "" {
		fmt.Fprintf(builder, `<w:vMerge w:val="%s"/>`, vMergeType)
	}

	// Apply cell styling from style attribute
	t.applyTableCellStyling(n, builder)

	builder.WriteString(`</w:tcPr>`)

	// Start paragraph in cell
	builder.WriteString(`<w:p><w:pPr></w:pPr>`)

	// If header, make text bold
	cellState := state.copy()
	if isHeader {
		cellState.bold = true
	}

	// Process cell content
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		t.convertNode(c, builder, cellState)
	}

	builder.WriteString(`</w:p>`)
	builder.WriteString(`</w:tc>`)
}

// applyTableCellStyling applies styling from HTML style attribute to cell
func (t *Template) applyTableCellStyling(n *html.Node, builder *strings.Builder) {
	for _, attr := range n.Attr {
		if attr.Key == "style" {
			styles := parseStyle(attr.Val)

			// Background color
			if bgColor, ok := styles["background-color"]; ok {
				hexColor := colorToHex(bgColor)
				fmt.Fprintf(builder, `<w:shd w:val="clear" w:color="auto" w:fill="%s"/>`, hexColor)
			}

			// Custom borders
			hasBorder := false
			if _, ok := styles["border"]; ok {
				hasBorder = true
			}
			if _, ok := styles["border-top"]; ok {
				hasBorder = true
			}
			if _, ok := styles["border-left"]; ok {
				hasBorder = true
			}
			if _, ok := styles["border-bottom"]; ok {
				hasBorder = true
			}
			if _, ok := styles["border-right"]; ok {
				hasBorder = true
			}

			if hasBorder {
				builder.WriteString(`<w:tcBorders>`)

				// Parse border style (simplified - use single or double)
				borderStyle := "single"
				borderSize := "4"       // default thin
				borderColor := "000000" // default black

				if border, ok := styles["border"]; ok {
					// Simple parsing: "2px solid red" -> extract width and color
					parts := strings.Fields(border)
					for _, part := range parts {
						if strings.Contains(part, "px") {
							width := strings.TrimSuffix(part, "px")
							if w := parseInt(width); w > 2 {
								borderSize = "8"
							}
						} else if part == "double" {
							borderStyle = "double"
						} else {
							// Try as color
							if hex := colorToHex(part); hex != "000000" || part == "black" {
								borderColor = hex
							}
						}
					}
				}

				borderXML := fmt.Sprintf(`<w:top w:val="%s" w:sz="%s" w:color="%s"/>`, borderStyle, borderSize, borderColor)
				builder.WriteString(borderXML)
				builder.WriteString(strings.Replace(borderXML, "top", "left", 1))
				builder.WriteString(strings.Replace(borderXML, "top", "bottom", 1))
				builder.WriteString(strings.Replace(borderXML, "top", "right", 1))

				builder.WriteString(`</w:tcBorders>`)
			}
		}
	}
}

// writeMergedCell writes an empty cell that continues a vertical merge
func (t *Template) writeMergedCell(builder *strings.Builder, colWidth int, vMergeType string) {
	builder.WriteString(`<w:tc>`)
	builder.WriteString(`<w:tcPr>`)
	fmt.Fprintf(builder, `<w:tcW w:w="%d" w:type="dxa"/>`, colWidth)
	fmt.Fprintf(builder, `<w:vMerge w:val="%s"/>`, vMergeType)
	builder.WriteString(`</w:tcPr>`)
	builder.WriteString(`<w:p><w:pPr></w:pPr></w:p>`)
	builder.WriteString(`</w:tc>`)
}

func (t *Template) writeHeading(n *html.Node, builder *strings.Builder, state *textState, tag string) {
	// Map heading level to font size (in half-points)
	sizes := map[string]string{
		"h1": "48", // 24pt
		"h2": "40", // 20pt
		"h3": "32", // 16pt
		"h4": "28", // 14pt
		"h5": "24", // 12pt
		"h6": "22", // 11pt
	}

	builder.WriteString(`<w:p><w:pPr>`)
	fmt.Fprintf(builder, `<w:pStyle w:val="Heading%s"/>`, tag[1:])
	builder.WriteString(`</w:pPr>`)

	headingState := state.copy()
	headingState.bold = true
	headingState.fontSize = sizes[tag]

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		t.convertNode(c, builder, headingState)
	}

	builder.WriteString(`</w:p>`)
}

func (t *Template) writeAnchor(n *html.Node, builder *strings.Builder, state *textState) {
	href := ""
	for _, attr := range n.Attr {
		if attr.Key == "href" {
			href = attr.Val
			break
		}
	}

	linkState := state.copy()
	linkState.fontColor = "0563C1" // Standard Word Hyperlink Blue
	linkState.underline = true

	var content strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		t.convertNode(c, &content, linkState)
	}

	if href != "" {
		builder.WriteString(t.generateLinkXML(href, content.String()))
	} else {
		builder.WriteString(content.String())
	}
}

func (t *Template) writeImage(n *html.Node, builder *strings.Builder, state *textState) {
	src := ""
	for _, attr := range n.Attr {
		if attr.Key == "src" {
			src = attr.Val
			break
		}
	}

	if src != "" {
		ooxml, err := t.handleSpecialPlaceholder("%img", src)
		if err == nil {
			builder.WriteString(ooxml)
		}
	}
}

// ParseStyle parses inline CSS style attribute
func parseStyle(style string) map[string]string {
	result := make(map[string]string)
	parts := strings.Split(style, ";")
	for _, part := range parts {
		kv := strings.SplitN(strings.TrimSpace(part), ":", 2)
		if len(kv) == 2 {
			result[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}
	return result
}

// colorToHex converts a CSS color to hex format
func colorToHex(color string) string {
	color = strings.TrimSpace(strings.ToLower(color))

	// Already hex
	if strings.HasPrefix(color, "#") {
		hex := strings.TrimPrefix(color, "#")
		if len(hex) == 3 {
			return string(hex[0]) + string(hex[0]) + string(hex[1]) + string(hex[1]) + string(hex[2]) + string(hex[2])
		}
		return hex
	}

	// RGB format
	if strings.HasPrefix(color, "rgb") {
		re := regexp.MustCompile(`\d+`)
		matches := re.FindAllString(color, 3)
		if len(matches) >= 3 {
			r := parseInt(matches[0])
			g := parseInt(matches[1])
			b := parseInt(matches[2])
			return fmt.Sprintf("%02x%02x%02x", r, g, b)
		}
	}

	// Named colors (extended list)
	colors := map[string]string{
		"black":       "000000",
		"white":       "FFFFFF",
		"red":         "FF0000",
		"green":       "008000",
		"blue":        "0000FF",
		"yellow":      "FFFF00",
		"cyan":        "00FFFF",
		"magenta":     "FF00FF",
		"gray":        "808080",
		"grey":        "808080",
		"orange":      "FFA500",
		"pink":        "FFC0CB",
		"purple":      "800080",
		"brown":       "A52A2A",
		"lime":        "00FF00",
		"navy":        "000080",
		"teal":        "008080",
		"olive":       "808000",
		"maroon":      "800000",
		"aqua":        "00FFFF",
		"silver":      "C0C0C0",
		"fuchsia":     "FF00FF",
		"lightgray":   "D3D3D3",
		"lightgrey":   "D3D3D3",
		"darkgray":    "A9A9A9",
		"darkgrey":    "A9A9A9",
		"lightblue":   "ADD8E6",
		"lightgreen":  "90EE90",
		"lightyellow": "FFFFE0",
		"lightpink":   "FFB6C1",
		"darkblue":    "00008B",
		"darkgreen":   "006400",
		"darkred":     "8B0000",
		"gold":        "FFD700",
		"coral":       "FF7F50",
		"salmon":      "FA8072",
		"khaki":       "F0E68C",
		"violet":      "EE82EE",
		"indigo":      "4B0082",
		"crimson":     "DC143C",
		"beige":       "F5F5DC",
		"ivory":       "FFFFF0",
		"lavender":    "E6E6FA",
		"turquoise":   "40E0D0",
	}

	if hex, ok := colors[color]; ok {
		return hex
	}

	return "000000" // Default to black
}

func parseInt(s string) int {
	var n int
	_, _ = fmt.Sscanf(s, "%d", &n)
	return n
}

// applyStyleToState parses the style attribute of an HTML node and applies colors to the state
func applyStyleToState(n *html.Node, state *textState) {
	for _, attr := range n.Attr {
		if attr.Key == "style" {
			styles := parseStyle(attr.Val)

			// Apply text color
			if color, ok := styles["color"]; ok {
				state.fontColor = colorToHex(color)
			}

			// Apply background color
			if bgColor, ok := styles["background-color"]; ok {
				state.bgColor = colorToHex(bgColor)
			}
			if bgColor, ok := styles["background"]; ok {
				// Simple case: just a color
				state.bgColor = colorToHex(bgColor)
			}
		}
	}
}
