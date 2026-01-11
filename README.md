# dgdoc

**Professional HTML to Word Generator**

A high-performance Go library and CLI tool designed to dynamically populate DOCX templates with rich HTML content. **dgdoc** seamlessly converts HTML tables, lists, images, and styling into native Word elements, bridging the gap between web content and professional document reporting.

[![Go Reference](https://pkg.go.dev/badge/github.com/dgmos/dgdoc.svg)](https://pkg.go.dev/github.com/dgmos/dgdoc)

## Key Features

- **Dynamic Content Injection**: Seamlessly replace placeholders with rich HTML or plain text.
- **Native Table Generation**: Automatically convert HTML tables into native Word table structures.
- **Rich Text Formatting**: Full support for bold, italic, underline, lists, and headings.
- **Visual Styling**: Apply custom text and background colors using standard CSS.
- **Media & Assets**: Embed images (URL, Base64, local paths) and signatures effortlessly.
- **Smart Hyperlinks**: Generate clickable, styled links from standard HTML `<a>` tags.
- **Versatile Integration**: Available as a high-performance Go library and a standalone CLI tool.

## Installation

### As CLI Tool

```bash
go install github.com/dgmos/dgdoc/cmd/dgdoc@latest
```

### Direct Download (Pre-built Binaries)

If you don't have Go installed, you can download the latest pre-built binaries from the [GitHub Releases](https://github.com/dgmos/dgdoc/releases) page.

#### Windows
1. Download `dgdoc_windows_amd64.zip`.
2. Extract the `dgdoc.exe` file to a folder (e.g., `C:\dgdoc`).
3. Add this folder to your system **PATH**:
   - Search for "Environment Variables" in the Start menu.
   - Edit the "Path" variable and add `C:\dgdoc`.
4. Open PowerShell or CMD and type: `dgdoc --version`

#### macOS (Intel & Apple Silicon)
1. Download `dgdoc_darwin_arm64.tar.gz` (Apple Silicon) or `dgdoc_darwin_amd64.tar.gz` (Intel).
2. Extract the archive.
3. Open Terminal and run:
   ```bash
   sudo cp dgdoc /usr/local/bin/
   sudo chmod +x /usr/local/bin/dgdoc
   ```
4. Verify by typing: `dgdoc --version`

#### Linux
1. Download `dgdoc_linux_amd64.tar.gz`.
2. Extract the archive.
3. Run:
   ```bash
   sudo cp dgdoc /usr/local/bin/
   sudo chmod +x /usr/local/bin/dgdoc
   ```
4. Verify by typing: `dgdoc --version`

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

    // Prepare data
    data := map[string]any{
        "customer_name": "Antigravity Tech Corp",
        "description": `
            <h1>Project Report</h1>
            <p>This document includes <span style="color: blue">rich content</span>.</p>
            <ul>
                <li>HTML Support</li>
                <li>Image Embedding</li>
            </ul>
        `,
        // Special Syntax for Images
        "%signature": "https://example.com/signature.png",
        
        // Special Syntax for Links
        "%website": "Visit Website|https://example.com",
    }

    // Apply all placeholders at once
    // Note: Replaces ALL occurrences of a placeholder in the document
    if err := template.Apply(data); err != nil {
        log.Fatal(err)
    }

    // Save output
    if err := template.Save("output.docx"); err != nil {
        log.Fatal(err)
    }
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

The library supports the following placeholder formats:

- `{placeholder}`: Replaced with HTML or plain text.
- `{%placeholder}`: Special syntax for **Images** and **Hyperlinks**.

### Special Syntax (`{%placeholder}`)

When using the `{%}` prefix, `dgdoc` automatically detects the type of content:

| Content Type | Example Value | Result |
|--------------|---------------|--------|
| **Image (URL)** | `https://example.com/sig.png` | Embedded image in document |
| **Image (Base64)** | `data:image/png;base64,...` | Embedded image from base64 data |
| **Image (Local)** | `./assets/signature.jpg` | Embedded image from file path |
| **Hyperlink** | `https://google.com` | Clickable link (Text is the URL) |
| **Link with Text**| `Google|https://google.com` | Clickable link with custom text |

### Supported HTML Features

| Feature | HTML Tags/Attributes |
|---------|-----------------------|
| **Text Styles** | `<b>`, `<strong>`, `<i>`, `<em>`, `<u>`, `<s>`, `<strike>`, `<span>` |
| **Colors** | `<span style="color: #ff0000; background-color: #ffff00">` |
| **Headings** | `<h1>` through `<h6>` |
| **Lists** | `<ul>` (unordered), `<ol>` (ordered), `<li>` |
| **Tables** | `<table>`, `<tr>`, `<td>`, `<th>`, `<thead>`, `<tbody>` |
| **Links** | `<a href="https://...">Link Text</a>` |
| **Images** | `<img src="https://..." />` (URL, Local, or Base64) |
| **Break** | `<br/>` |

## API Reference

### `docx.Open(path string) (*Template, error)`
Opens a DOCX template file.

### `template.Apply(data map[string]any) error`
Replaces multiple placeholders at once. Useful for batch updates.

### `template.SetContent(placeholder, html string) error`
Replaces `{placeholder}` with converted HTML content.

### `template.ReplaceText(placeholder, text string) error`
Replaces `{placeholder}` with plain text.

### `template.Save(path string) error`
Saves the modified document.

### `template.Close() error`
Closes the template and releases resources.

## License

MIT License
