# dgdoc

**HTML to Word Document Generator** - A Go package that replaces placeholders in Word DOCX templates with rich HTML content.

[![Go Reference](https://pkg.go.dev/badge/github.com/dgmos/dgdoc.svg)](https://pkg.go.dev/github.com/dgmos/dgdoc)

## Features

- 🔄 Replace `{{placeholder}}` values with HTML content
- 📊 Native Word tables from HTML `<table>`
- 📝 Full text formatting (bold, italic, underline, strikethrough)
- 🎨 Text and background colors via CSS styles
- 📋 Bullet and numbered lists
- 📑 Multiple placeholders support
- 💻 Use as CLI tool or Go library

## Installation

### As CLI Tool

```bash
go install github.com/dgmos/dgdoc/cmd/dgdoc@latest
```

### As Library

```bash
go get github.com/dgmos/dgdoc
```

## CLI Usage

```bash
# Using JSON file
dgdoc --template template.docx --output report.docx --data data.json

# Using inline JSON
dgdoc --template template.docx --json '{"content": "<h1>Hello</h1><p>World</p>"}'

# Multiple placeholders
dgdoc --template template.docx --json '{"title": "<h1>Report</h1>", "body": "<p>Content</p>"}'
```

### JSON Data Format

```json
{
  "content": "<h1>Main Title</h1><p>Paragraph text</p>",
  "table_data": "<table><tr><td>Cell 1</td><td>Cell 2</td></tr></table>",
  "author": "<span style=\"color: blue;\">John Doe</span>"
}
```

## Library Usage

```go
package main

import (
    "log"
    "github.com/dgmos/dgdoc/docx"
)

func main() {
    // Open template
    template, err := docx.Open("template.docx")
    if err != nil {
        log.Fatal(err)
    }
    defer template.Close()

    // Replace placeholders with HTML
    html := `
        <h1>Report Title</h1>
        <p>This is <strong>bold</strong> and <em>italic</em> text.</p>
        <ul>
            <li>Item 1</li>
            <li>Item 2</li>
        </ul>
        <table>
            <tr><th>Name</th><th>Value</th></tr>
            <tr><td>A</td><td>100</td></tr>
        </table>
    `
    template.SetContent("content", html)

    // Multiple placeholders
    template.SetContent("author", "<span style='color: blue;'>John Doe</span>")
    template.SetContent("date", "<strong>2024-01-15</strong>")

    // Save output
    template.Save("output.docx")
}
```

## Supported HTML

| HTML Element | Word Output |
|--------------|-------------|
| `<h1>` - `<h6>` | Headings with sizes |
| `<p>` | Paragraph |
| `<strong>`, `<b>` | **Bold** |
| `<em>`, `<i>` | *Italic* |
| `<u>` | Underline |
| `<s>`, `<strike>` | ~~Strikethrough~~ |
| `<ul>`, `<ol>`, `<li>` | Lists |
| `<table>`, `<tr>`, `<td>`, `<th>` | Native Word tables |
| `<br>` | Line break |
| `<span style="...">` | Inline styling |

## CSS Styles

```html
<!-- Text color -->
<span style="color: red;">Red text</span>
<span style="color: #FF5733;">Hex color</span>

<!-- Background color -->
<span style="background-color: yellow;">Highlighted</span>

<!-- Combined -->
<span style="color: white; background-color: navy;">Navy background</span>
```

### Supported Colors

Named colors: `red`, `blue`, `green`, `yellow`, `orange`, `pink`, `purple`, `navy`, `teal`, `lightblue`, `lightgreen`, `lightyellow`, and 30+ more.

## Template Format

Create a Word document with placeholders like `{{placeholder_name}}`:

```
Dear {{recipient}},

{{content}}

Best regards,
{{signature}}
```

## API Reference

### `docx.Open(path string) (*Template, error)`
Opens a DOCX template file.

### `template.SetContent(placeholder, html string) error`
Replaces `{{placeholder}}` with converted HTML content.

### `template.ReplaceText(placeholder, text string) error`
Replaces `{{placeholder}}` with plain text.

### `template.Save(path string) error`
Saves the modified document.

### `template.Close() error`
Closes the template and releases resources.

## License

MIT License
