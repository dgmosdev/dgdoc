# dgdoc

**HTML to Word Document Generator** - A Go package that replaces placeholders in Word DOCX templates with rich HTML content.

[![Go Reference](https://pkg.go.dev/badge/github.com/dgmos/dgdoc.svg)](https://pkg.go.dev/github.com/dgmos/dgdoc)

## Features

- 🔄 Replace `{{placeholder}}` values with HTML **or plain text** (normal fields)
- 📊 Native Word tables from HTML `<table>`
- 📝 Full text formatting (bold, italic, underline, strikethrough)
- 🎨 Text and background colors via CSS styles
- 📋 Bullet and numbered lists
- 📑 Multiple placeholders support (mix HTML and plain text)
- 💻 Use as CLI tool or Go library

## Installation

### As CLI Tool

```bash
go install github.com/dgmos/dgdoc/cmd/dgdoc@latest
```

### Direct Download (Windows, macOS, Linux)

If you don't have Go installed, you can download the latest pre-built binaries from the [GitHub Releases](https://github.com/dgmos/dgdoc/releases) page.

1. Download the archive for your operating system.
2. Extract the `dgdoc` (or `dgdoc.exe`) binary.
3. Move it to a folder in your system PATH (e.g., `/usr/local/bin` on macOS/Linux or a custom folder added to PATH on Windows).

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

# Mixed HTML and normal fields
dgdoc --template template.docx --json '{"customer": "Ahmet Yılmaz", "address": "İstanbul, Türkiye", "content": "<h1>Title</h1><p>Body</p>"}'
```

### JSON Data Format

```json
{
  "customer": "Ahmet Yılmaz",
  "address": "Atatürk Mah. No:1, İstanbul",
  "content": "<h1>Main Title</h1><p>Paragraph text</p>",
  "table_data": "<table><tr><td>Cell 1</td><td>Cell 2</td></tr></table>",
  "status": "<span style=\"color: green;\">✓ Aktif</span>"
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

    // Prepare data (supports HTML or plain text)
    data := map[string]any{
        "content": `
            <h1>Report</h1>
            <p>This is <strong>bold</strong> text.</p>
        `,
        "author": "<span style='color: blue;'>John Doe</span>",
        "date":   "2024-01-15", // Plain text
    }

    // Apply all placeholders at once
    if err := template.Apply(data); err != nil {
        log.Fatal(err)
    }

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

### `template.Apply(data map[string]any) error`
Replaces multiple placeholders at once. Useful for batch updates.

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
