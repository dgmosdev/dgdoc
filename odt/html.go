package odt

import (
	"regexp"
)

// ConvertHTMLToText converts basic HTML to plain text for ODT
// Strips HTML tags and preserves content
func ConvertHTMLToText(html string) string {
	// Remove HTML tags but keep the text
	result := html

	// Remove all HTML tags
	tagPattern := regexp.MustCompile(`<[^>]*>`)
	result = tagPattern.ReplaceAllString(result, "")

	return result
}

// NOTE: A full implementation of HTML to ODT rich text would involve:
// 1. Parsing HTML into a tree structure
// 2. For each formatting tag (<b>, <i>, etc.), create ODT <text:span> elements with style properties
// 3. Example ODT rich text:
//    <text:p>
//      <text:span>Normal text </text:span>
//      <text:span text:style-name="Bold">bold text</text:span>
//    </text:p>
