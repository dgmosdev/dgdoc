package docx

import (
	"fmt"
	"regexp"
	"strings"
)

// PreprocessTemplate processes template directives (conditionals and loops) in the document content
// This should be called before regular placeholder replacement
func PreprocessTemplate(content string, data map[string]any) (string, error) {
	var err error

	// Process loops first (they can contain conditionals)
	content, err = processLoops(content, data)
	if err != nil {
		return "", fmt.Errorf("loop processing failed: %w", err)
	}

	// Process conditionals
	content, err = processConditionals(content, data)
	if err != nil {
		return "", fmt.Errorf("conditional processing failed: %w", err)
	}

	return content, nil
}

// processConditionals processes {#if}...{/if} and {#if}...{#else}...{/if} blocks
func processConditionals(content string, data map[string]any) (string, error) {
	// Pattern to match conditional blocks (including nested ones)
	// We'll process from innermost to outermost
	maxIterations := 100 // Prevent infinite loops
	iteration := 0

	for {
		if iteration >= maxIterations {
			return "", fmt.Errorf("too many nested conditionals or infinite loop detected")
		}
		iteration++

		// Find the innermost {#if} block (one that doesn't contain another {#if})
		changed := false
		content, changed = processOneConditional(content, data)

		if !changed {
			break // No more conditionals to process
		}
	}

	return content, nil
}

// processOneConditional finds and processes one conditional block
func processOneConditional(content string, data map[string]any) (string, bool) {
	// Extract plain text to find conditional blocks
	textContent := extractTextContent(content)

	// Find {#if condition}...{/if} pattern
	// Look for the last {#if that doesn't have another {#if before its {/if}
	ifPattern := regexp.MustCompile(`\{#if\s+([^}]+)\}`)
	elsePattern := regexp.MustCompile(`\{#else\}`)

	ifMatches := ifPattern.FindAllStringIndex(textContent, -1)
	if len(ifMatches) == 0 {
		return content, false // No conditionals found
	}

	// Find the last {#if}
	lastIfIdx := ifMatches[len(ifMatches)-1][0]
	lastIfEnd := ifMatches[len(ifMatches)-1][1]

	// Extract condition
	condMatch := ifPattern.FindStringSubmatch(textContent[lastIfIdx:])
	if len(condMatch) < 2 {
		return content, false
	}
	condition := strings.TrimSpace(condMatch[1])

	// Find corresponding {/if}
	endifIdx := strings.Index(textContent[lastIfEnd:], "{/if}")
	if endifIdx == -1 {
		// Malformed: no closing {/if}
		return content, false
	}
	endifIdx += lastIfEnd

	// Check for {#else}
	elseIdx := -1
	elseMatch := elsePattern.FindStringIndex(textContent[lastIfEnd:endifIdx])
	if elseMatch != nil {
		elseIdx = lastIfEnd + elseMatch[0]
	}

	// Evaluate condition
	result, err := evaluateCondition(condition, data)
	if err != nil {
		// If evaluation fails, treat as false
		result = false
	}

	// Determine which content to keep
	var keepStart, keepEnd int
	if result {
		// Keep content between {#if} and {#else} or {/if}
		keepStart = lastIfEnd
		if elseIdx != -1 {
			keepEnd = elseIdx
		} else {
			keepEnd = endifIdx
		}
	} else {
		// Keep content between {#else} and {/if}, or nothing
		if elseIdx != -1 {
			keepStart = elseIdx + len("{#else}")
			keepEnd = endifIdx
		} else {
			keepStart = endifIdx
			keepEnd = endifIdx
		}
	}

	// Extract the content to keep from the text representation
	textToKeep := textContent[keepStart:keepEnd]

	// Now we need to replace the entire conditional block in the actual XML content
	// We'll use the placeholder replacement logic to find and replace

	// Replace the conditional block in the XML content
	content = replaceConditionalBlock(content, textContent, lastIfIdx, endifIdx+len("{/if}"), textToKeep)

	return content, true
}

// replaceConditionalBlock replaces a conditional block in the XML content
// This is complex because we need to handle split placeholders across XML runs
func replaceConditionalBlock(xmlContent, textContent string, textStart, textEnd int, replacement string) string {
	// Extract all <w:t> segments
	textPattern := regexp.MustCompile(`<w:t[^>]*>([^<]*)</w:t>`)
	matches := textPattern.FindAllStringSubmatchIndex(xmlContent, -1)

	// Build character position map
	charPos := 0
	var startXMLPos, endXMLPos int = -1, -1

	for _, match := range matches {
		text := xmlContent[match[2]:match[3]]
		textLen := len(text)

		// Check if this segment contains the start of our block
		if startXMLPos == -1 && charPos+textLen > textStart {
			// Start is in this segment
			startXMLPos = match[0] // Start of <w:t>
		}

		// Check if this segment contains the end of our block
		if charPos+textLen >= textEnd {
			endXMLPos = match[1] // End of </w:t>
			break
		}

		charPos += textLen
	}

	if startXMLPos == -1 || endXMLPos == -1 {
		// Couldn't find positions, return unchanged
		return xmlContent
	}

	// Find the containing paragraph(s)
	// Look backwards for <w:p>
	pStart := strings.LastIndex(xmlContent[:startXMLPos], "<w:p>")
	if pStart == -1 {
		pStart = strings.LastIndex(xmlContent[:startXMLPos], "<w:p ")
		if pStart == -1 {
			return xmlContent
		}
	}

	// Look forwards for </w:p>
	pEnd := strings.Index(xmlContent[endXMLPos:], "</w:p>")
	if pEnd == -1 {
		return xmlContent
	}
	pEnd = endXMLPos + pEnd + len("</w:p>")

	// Generate XML for the replacement text
	var replacementXML string
	if replacement != "" {
		// Simple implementation: wrap in <w:r><w:t>
		escapedText := escapeXML(replacement)
		replacementXML = fmt.Sprintf(`<w:p><w:r><w:t xml:space="preserve">%s</w:t></w:r></w:p>`, escapedText)
	} else {
		replacementXML = "" // Empty replacement
	}

	// Replace the block
	result := xmlContent[:pStart] + replacementXML + xmlContent[pEnd:]
	return result
}

// processLoops processes {#arrayName}...{/arrayName} blocks
func processLoops(content string, data map[string]any) (string, error) {
	maxIterations := 100
	iteration := 0

	for {
		if iteration >= maxIterations {
			return "", fmt.Errorf("too many nested loops or infinite loop detected")
		}
		iteration++

		changed := false
		content, changed = processOneLoop(content, data)

		if !changed {
			break
		}
	}

	return content, nil
}

// processOneLoop finds and processes one loop block
func processOneLoop(content string, data map[string]any) (string, bool) {
	textContent := extractTextContent(content)

	// Find {#varName}...{/varName} pattern
	loopPattern := regexp.MustCompile(`\{#([a-zA-Z0-9_\.]+)\}`)
	matches := loopPattern.FindAllStringSubmatchIndex(textContent, -1)

	if len(matches) == 0 {
		return content, false
	}

	// Process the last (innermost) loop
	lastMatch := matches[len(matches)-1]
	startIdx := lastMatch[0]
	endIdx := lastMatch[1]
	varName := textContent[lastMatch[2]:lastMatch[3]]

	// Find closing tag
	closeTag := fmt.Sprintf("{/%s}", varName)
	closeIdx := strings.Index(textContent[endIdx:], closeTag)
	if closeIdx == -1 {
		return content, false // Malformed
	}
	closeIdx += endIdx

	// Get the loop variable value
	loopData, err := getNestedValue(varName, data)
	if err != nil {
		// Variable doesn't exist or error, skip this loop (render nothing)
		content = replaceConditionalBlock(content, textContent, startIdx, closeIdx+len(closeTag), "")
		return content, true
	}

	// Check if it's an array
	arr, ok := loopData.([]any)
	if !ok {
		// Not an array, skip
		content = replaceConditionalBlock(content, textContent, startIdx, closeIdx+len(closeTag), "")
		return content, true
	}

	// Extract loop body
	bodyText := textContent[endIdx:closeIdx]

	// Repeat the body for each item
	var result strings.Builder
	for i, item := range arr {
		// Create context for this iteration
		itemData := make(map[string]any)
		for k, v := range data {
			itemData[k] = v
		}

		// Add loop metadata
		itemData["@index"] = i
		itemData["@first"] = i == 0
		itemData["@last"] = i == len(arr)-1
		itemData["@length"] = len(arr)

		// Add item to context
		// If item is a map, merge its properties
		if itemMap, ok := item.(map[string]any); ok {
			for k, v := range itemMap {
				itemData[k] = v
			}
		} else {
			// Otherwise, make it available as a special variable
			itemData["@item"] = item
		}

		// Process the body with this context
		// For now, we'll just replace placeholders in the bodyText
		// Note: This is a simplified version; a full implementation would need to
		// reconstruct the XML properly
		processedBody := bodyText

		// Replace any {varName} placeholders in the body
		for key, val := range itemData {
			placeholder := fmt.Sprintf("{%s}", key)
			if strings.Contains(processedBody, placeholder) {
				processedBody = strings.ReplaceAll(processedBody, placeholder, fmt.Sprintf("%v", val))
			}
		}

		result.WriteString(processedBody)
		if i < len(arr)-1 {
			result.WriteString("\n") // Add separator
		}
	}

	// Replace the loop block
	content = replaceConditionalBlock(content, textContent, startIdx, closeIdx+len(closeTag), result.String())
	return content, true
}

// extractTextContent extracts plain text from XML content (from <w:t> tags)
func extractTextContent(xmlContent string) string {
	textPattern := regexp.MustCompile(`<w:t[^>]*>([^<]*)</w:t>`)
	matches := textPattern.FindAllStringSubmatch(xmlContent, -1)

	var builder strings.Builder
	for _, match := range matches {
		builder.WriteString(match[1])
	}
	return builder.String()
}
