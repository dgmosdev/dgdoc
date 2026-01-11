package xlsx

import (
	"regexp"
)

// ConvertHTMLToRichText converts basic HTML to Excel rich text format
// Supports: <b>, <i>, <u>, <span style="color:...">
func ConvertHTMLToRichText(html string) string {
	// For now, we'll do a simplified conversion - strip HTML tags
	// A full implementation would parse HTML and create multiple <r> (run) elements
	// with different formatting (<rPr> for properties)

	// This is a basic implementation that strips tags but preserves content
	result := html

	// Remove HTML tags but keep the text
	result = stripHTMLTags(result)

	return result
}

// stripHTMLTags removes HTML tags from text
func stripHTMLTags(s string) string {
	// Remove all HTML tags
	tagPattern := regexp.MustCompile(`<[^>]*>`)
	return tagPattern.ReplaceAllString(s, "")
}

// NOTE: A full implementation of HTML to Excel rich text would involve:
// 1. Parsing HTML into a tree structure
// 2. For each formatting tag (<b>, <i>, etc.), create an Excel <r> (run) element
// 3. Add <rPr> (run properties) with appropriate formatting
// 4. Example Excel rich text:
//    <is>
//      <r><t>Normal text </t></r>
//      <r><rPr><b/></rPr><t>bold text</t></r>
//      <r><t> more normal</t></r>
//    </is>
