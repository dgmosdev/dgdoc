#!/bin/bash

# dgdoc CLI Usage Examples
# This script demonstrates various ways to use the dgdoc CLI tool

echo "=== dgdoc CLI Examples ==="
echo ""

# Example 1: Basic usage with JSON file
echo "1. Basic Invoice Generation"
echo "   Command:"
echo "   dgdoc --template templates/invoice_template.docx \\"
echo "         --output output/invoice.docx \\"
echo "         --data data/invoice_data.json"
echo ""

# Example 2: Using inline JSON
echo "2. Quick Report with Inline JSON"
echo "   Command:"
echo "   dgdoc --template templates/simple_report.docx \\"
echo "         --output output/report.docx \\"
echo "         --json '{\"title\": \"Monthly Report\", \"date\": \"2026-01-11\", \"content\": \"<h1>Summary</h1><p>All systems operational.</p>\"}'"
echo ""

# Example 3: Conditionals and loops
echo "3. Complex Document with Conditionals"
echo "   Command:"
echo "   dgdoc --template templates/customer_report.docx \\"
echo "         --output output/customer_report.docx \\"
echo "         --json '{\"customer\": \"Acme Corp\", \"premium\": true, \"orders\": [{\"id\": 1, \"total\": 500}, {\"id\": 2, \"total\": 750}]}'"
echo ""

# Example 4: Excel template
echo "4. Excel Report Generation"
echo "   Command:"
echo "   dgdoc --template templates/sales_template.xlsx \\"
echo "         --output output/sales_report.xlsx \\"
echo "         --data data/sales_data.json"
echo ""

# Example 5: PowerPoint presentation
echo "5. PowerPoint Presentation"
echo "   Command:"
echo "   dgdoc --template templates/presentation.pptx \\"
echo "         --output output/Q1_presentation.pptx \\"
echo "         --data data/presentation_data.json"
echo ""

# Example 6: Batch processing
echo "6. Batch Processing Multiple Documents"
echo "   Command:"
echo "   for customer in customer1 customer2 customer3; do"
echo "     dgdoc --template templates/report.docx \\"
echo "           --output \"output/\${customer}_report.docx\" \\"
echo "           --data \"data/\${customer}_data.json\""
echo "   done"
echo ""

# Example 7: Using environment variables
echo "7. Dynamic Content from Environment Variables"
echo "   Command:"
echo "   AUTHOR=\$(whoami)"
echo "   DATE=\$(date +%Y-%m-%d)"
echo "   dgdoc --template templates/document.docx \\"
echo "         --output output/document.docx \\"
echo "         --json \"{\\\"author\\\": \\\"\$AUTHOR\\\", \\\"date\\\": \\\"\$DATE\\\"}\""
echo ""

echo "=== Advanced Features ==="
echo ""

# Example 8: HTML content
echo "8. Rich HTML Content"
echo "   Template placeholder: {content}"
echo "   JSON:"
echo "   {\"content\": \"<h1>Title</h1><p>Text with <b>bold</b> and <span style='color: red;'>colored</span> text.</p>\"}"
echo ""

# Example 9: Images
echo "9. Image Embedding"
echo "   Template placeholder: {%logo}"
echo "   JSON:"
echo "   {\"%logo\": \"https://example.com/logo.png\"}"
echo "   OR"
echo "   {\"%logo\": \"./assets/logo.png\"}"
echo ""

# Example 10: Links
echo "10. Hyperlinks"
echo "    Template placeholder: {%website}"
echo "    JSON:"
echo "    {\"%website\": \"Visit Us|https://example.com\"}"
echo ""

echo "=== Installation Reminder ==="
echo ""
echo "Install dgdoc CLI:"
echo "  go install github.com/dgmosdev/dgdoc/cmd/dgdoc@latest"
echo ""
echo "Or download from releases:"
echo "  https://github.com/dgmosdev/dgdoc/releases"
