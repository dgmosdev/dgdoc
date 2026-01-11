# dgdoc

**Professional HTML to Word Generator**

A high-performance Go library and CLI tool designed to dynamically populate DOCX templates with rich HTML content. **dgdoc** seamlessly converts HTML tables, lists, images, and styling into native Word elements, bridging the gap between web content and professional document reporting.

[![Go Reference](https://pkg.go.dev/badge/github.com/dgmosdev/dgdoc.svg)](https://pkg.go.dev/github.com/dgmosdev/dgdoc)

## Key Features

### 📝 DOCX (Word) Support
- **Conditional Rendering**: Show/hide content based on data conditions using `{#if}`.
- **Loop Support**: Repeat content for arrays with `{#items}` and access loop metadata.
- **Dynamic Content Injection**: Seamlessly replace placeholders with rich HTML or plain text.
- **Native Table Generation**: Automatically convert HTML tables into native Word table structures.
- **Rich Text Formatting**: Full support for bold, italic, underline, lists, and headings.
- **Visual Styling**: Apply custom text and background colors using standard CSS.
- **Media & Assets**: Embed images (URL, Base64, local paths) and signatures effortlessly.
- **Smart Hyperlinks**: Generate clickable, styled links from standard HTML `<a>` tags.
- **Subtemplates**: Merge external DOCX files with `{@include:filename.docx}`.
- **Chart Detection**: Detect and count charts in documents.
- **Document Metadata**: Set properties (title, author, keywords) and page setup (margins, size).

### 📗 XLSX (Excel) Support
- **Cell Placeholders**: Replace placeholders in Excel cells.
- **Formula Preservation**: Automatically skip cells containing formulas to prevent breaking calculations.
- **Loop Support**: Duplicate rows for array data with `{#items}...{/items}`.
- **HTML Conversion**: Convert HTML content to plain text for cells.

### 📘 PPTX (PowerPoint) Support
- **Slide Placeholders**: Replace text placeholders in PowerPoint slides.
- **Content Duplication**: Duplicate slide content sections for array data.
- **HTML Conversion**: Insert HTML content (tags stripped) into slides.

### 📙 ODT (OpenDocument) Support
- **Text Replacement**: Replace placeholders in ODT documents.
- **Full Loop Support**: Iterate over arrays in OpenDocument format.
- **HTML Conversion**: Convert HTML to plain text for ODT files.

### 🔧 Developer Tools
- **Versatile Integration**: Available as a high-performance Go library and a standalone CLI tool.
- **Consistent API**: Same interface across all document formats (DOCX, XLSX, PPTX, ODT).
- **Comprehensive Testing**: 22 unit tests covering all modules.

## Installation

### As CLI Tool

```bash
go install github.com/dgmosdev/dgdoc/cmd/dgdoc@latest
```

### Direct Download (Pre-built Binaries)

If you don't have Go installed, you can download the latest pre-built binaries from the [GitHub Releases](https://github.com/dgmosdev/dgdoc/releases) page.

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
go get github.com/dgmosdev/dgdoc
```

## CLI Usage

```bash
# Using JSON file
dgdoc --template template.docx --output report.docx --data data.json

# Using inline JSON
dgdoc --template template.docx --json '{"content": "<h1>Hello</h1><p>World</p>"}'

# Mixed HTML and normal fields
dgdoc --template template.docx --json '{"customer": "Ahmet Yılmaz", "address": "İstanbul, Türkiye", "content": "<h1>Title</h1><p>Body</p>"}'

# With conditionals and loops
dgdoc --template template.docx --json '{"premium": true, "items": [{"name": "Item1"}, {"name": "Item2"}]}'
```

### JSON Data Format

```json
{
  "customer": "Ahmet Yılmaz",
  "address": "Atatürk Mah. No:1, İstanbul",
  "content": "<h1>Main Title</h1><p>Paragraph text</p>",
  "table_data": "<table><tr><td>Cell 1</td><td>Cell 2</td></tr></table>",
  "status": "<span style=\"color: green;\">✓ Aktif</span>",
  "premium": true,
  "items": [
    {"name": "Product A", "price": 100},
    {"name": "Product B", "price": 200}
  ]
}
```

## Library Usage

```go
package main

import (
    "log"
    "github.com/dgmosdev/dgdoc/docx"
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

## Conditional Rendering & Loops

dgdoc supports powerful template directives for dynamic content generation.

### Conditional Rendering

Use `{#if condition}...{/if}` to conditionally include content:

**Template**:
```
Welcome, {#if premium}Premium{#else}Free{/if} User!
```

**Data**:
```json
{"premium": true}
```

**Output**: "Welcome, Premium User!"

#### Supported Conditions

- **Truthiness**: `{#if variable}` - checks if variable exists and is truthy
- **Comparisons**: `age > 18`, `status == 'active'`, `price <= 100`
- **Logical operators**: `age > 18 && active`, `premium || trial`
- **Negation**: `!disabled`

### Loops

Use `{#arrayName}...{/arrayName}` to iterate over arrays:

**Template**:
```
{#items}
- {name}: {price}TL
{/items}
```

**Data**:
```json
{
  "items": [
    {"name": "Product A", "price": 100},
    {"name": "Product B", "price": 200}
  ]
}
```

**Output**:
```
- Product A: 100TL
- Product B: 200TL
```

#### Loop Metadata

Access special variables within loops:

- `{@index}` - Current index (0-based)
- `{@first}` - True for first iteration
- `{@last}` - True for last iteration
- `{@length}` - Total number of items

**Example**:
```
{#items}
{@index}. {name}{#if @last} (Last item){/if}
{/items}
```

### Combined Example

```go
data := map[string]any{
    "user": map[string]any{
        "name": "Ahmet",
        "premium": true,
    },
    "orders": []any{
        map[string]any{"id": 1, "total": 150},
        map[string]any{"id": 2, "total": 200},
    },
}

template.Apply(data)
```

**Template**:
```
Hello {user.name}!

{#if user.premium}
Your orders:
{#orders}
- Order #{id}: {total}TL
{/orders}
{#else}
Upgrade to premium to see your orders.
{/if}
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

---

## Multi-Format Support

dgdoc now supports multiple document formats beyond DOCX!

### Excel (XLSX) Templates

```go
import "github.com/dgmosdev/dgdoc/xlsx"

// Open Excel template
template, _ := xlsx.Open("template.xlsx")
defer template.Close()

// Replace cell values
template.SetCellValue("company_name", "Acme Corp")
template.SetCellValue("revenue", "1,000,000")

// Apply batch data
data := map[string]any{
    "name": "Q1 Report",
    "date": "2026-01-11",
}
template.Apply(data)

// Loop support - duplicate rows
loopData := map[string]any{
    "items": []any{
        map[string]any{"product": "Widget A", "sales": "500"},
        map[string]any{"product": "Widget B", "sales": "750"},
    },
}
template.ProcessLoops(loopData)

template.Save("output.xlsx")
```

**Features:**
- ✅ Formula preservation (formulas are not replaced)
- ✅ Supports both text and numeric cells
- ✅ HTML content (basic tag stripping)
- ✅ Loop support for row duplication

---

### PowerPoint (PPTX) Templates

```go
import "github.com/dgmosdev/dgdoc/pptx"

// Open PowerPoint template
template, _ := pptx.Open("template.pptx")
defer template.Close()

// Replace text in slides
template.SetTextValue("title", "Q1 Presentation")
template.SetTextValue("subtitle", "2026 Report")

// Apply batch data
data := map[string]any{
    "presenter": "John Doe",
    "date": "January 11, 2026",
}
template.Apply(data)

// Loop support - duplicate slide content
loopData := map[string]any{
    "slides": []any{
        map[string]any{"topic": "Revenue", "amount": "$1M"},
        map[string]any{"topic": "Growth", "amount": "25%"},
    },
}
template.ProcessLoops(loopData)

template.Save("output.pptx")
```

**Features:**
- ✅ Text placeholder replacement in slides
- ✅ HTML content support
- ✅ Loop support for content duplication

---

### OpenDocument (ODT) Templates

```go
import "github.com/dgmosdev/dgdoc/odt"

// Open ODT template
template, _ := odt.Open("template.odt")
defer template.Close()

// Replace text
template.SetTextValue("title", "Document Title")

// Apply batch data
data := map[string]any{
    "author": "Jane Smith",
    "content": "<b>Bold text</b> and normal text",
}
template.Apply(data)

// Loop support
loopData := map[string]any{
    "sections": []any{
        map[string]any{"heading": "Section 1"},
        map[string]any{"heading": "Section 2"},
    },
}
template.ProcessLoops(loopData)

template.Save("output.odt")
```

**Features:**
- ✅ Supports paragraphs, spans, and headings
- ✅ HTML content support
- ✅ Loop support

---

## Advanced DOCX Features

### Document Merging (Subtemplates)

Merge external DOCX files into your main document:

**Template** (`main.docx`):
```
Company Report

{@include:header.docx}

Main content here...

{@include:footer.docx}
```

**Code**:
```go
template, _ := docx.Open("main.docx")

// Specify paths to external files
includes := map[string]string{
    "header.docx": "./templates/header.docx",
    "footer.docx": "./templates/footer.docx",
}

template.MergeDocuments(includes)
template.Save("output.docx")
```

### Chart Detection

Detect and count charts in your documents:

```go
template, _ := docx.Open("report.docx")

// Detect all charts
charts, _ := template.DetectCharts()
fmt.Printf("Found %d charts\n", len(charts))

// Check if document has charts
if template.HasCharts() {
    count := template.GetChartCount()
    fmt.Printf("Document contains %d charts\n", count)
}
```

### Document Metadata

Set document properties and page setup:

```go
template, _ := docx.Open("document.docx")

// Set metadata
meta := docx.Metadata{
    Title:       "Q1 Financial Report",
    Author:      "Finance Department",
    Subject:     "Quarterly Results",
    Keywords:    "finance, Q1, 2026, report",
    Description: "Detailed financial analysis for Q1 2026",
    Category:    "Financial Reports",
}
template.SetMetadata(meta)

// Configure page setup (values in twips: 1 inch = 1440 twips)
pageSetup := docx.PageSetup{
    PageWidth:    11906, // A4 width
    PageHeight:   16838, // A4 height
    MarginTop:    1440,  // 1 inch
    MarginBottom: 1440,
    MarginLeft:   1440,
    MarginRight:  1440,
}
template.SetPageSetup(pageSetup)

template.Save("output.docx")
```

---

## Supported Document Formats

| Format | Extension | Features |
|--------|-----------|----------|
| **Microsoft Word** | `.docx` | ✅ Full HTML support, conditionals, loops, tables, images, links, subtemplates, metadata |
| **Microsoft Excel** | `.xlsx` | ✅ Cell placeholders, formula preservation, loops, HTML conversion |
| **Microsoft PowerPoint** | `.pptx` | ✅ Slide placeholders, content duplication, HTML conversion |
| **OpenDocument Text** | `.odt` | ✅ Text replacement, loops, HTML conversion |

---

## License

MIT License
