package pptx

import (
	"regexp"
)

// ConvertHTMLToText converts basic HTML to plain text for PowerPoint
// Strips HTML tags and preserves content
func ConvertHTMLToText(html string) string {
	// Remove HTML tags but keep the text
	result := html

	// Remove all HTML tags
	tagPattern := regexp.MustCompile(`<[^>]*>`)
	result = tagPattern.ReplaceAllString(result, "")

	return result
}

// NOTE: A full implementation of HTML to PowerPoint rich text would involve:
// 1. Parsing HTML into a tree structure
// 2. For each formatting tag (<b>, <i>, etc.), create PowerPoint <a:r> (run) elements with <a:rPr> formatting
// 3. Example PowerPoint rich text:
//    <a:p>
//      <a:r><a:t>Normal text </a:t></a:r>
//      <a:r><a:rPr><a:b/></a:rPr><a:t>bold text</a:t></a:r>
//    </a:p>
