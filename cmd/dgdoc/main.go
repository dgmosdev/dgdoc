package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/dgmos/dgdoc/docx"
)

var (
	version = "1.0.0"
)

func main() {
	// Define flags
	templatePath := flag.String("template", "", "Path to the DOCX template file (required)")
	outputPath := flag.String("output", "output.docx", "Path for the output DOCX file")
	dataPath := flag.String("data", "", "Path to JSON file with placeholder data")
	dataJSON := flag.String("json", "", "Inline JSON string with placeholder data")
	showVersion := flag.Bool("version", false, "Show version information")
	showHelp := flag.Bool("help", false, "Show help message")

	// Custom usage message
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `dgdoc - HTML to Word Document Generator

Usage:
  dgdoc --template <file.docx> --output <output.docx> --data <data.json>
  dgdoc --template <file.docx> --json '{"placeholder": "<p>HTML content</p>"}' 

Options:
`)
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, `
Examples:
  # Using JSON file
  dgdoc --template template.docx --output report.docx --data data.json

  # Using inline JSON
  dgdoc --template template.docx --json '{"content": "<h1>Hello</h1><p>World</p>"}'

  # Multiple placeholders
  dgdoc --template template.docx --json '{"title": "<h1>Report</h1>", "body": "<p>Content here</p>"}'

JSON Format:
  {
    "placeholder_name": "<html>content</html>",
    "another_placeholder": "<p>more content</p>"
  }

Supported HTML:
  - Headings: <h1> to <h6>
  - Paragraphs: <p>
  - Formatting: <strong>, <em>, <u>, <s>
  - Lists: <ul>, <ol>, <li>
  - Tables: <table>, <tr>, <td>, <th>
  - Colors: style="color: red; background-color: yellow"

`)
	}

	flag.Parse()

	// Handle version
	if *showVersion {
		fmt.Printf("dgdoc version %s\n", version)
		os.Exit(0)
	}

	// Handle help
	if *showHelp {
		flag.Usage()
		os.Exit(0)
	}

	// Validate required flags
	if *templatePath == "" {
		fmt.Fprintln(os.Stderr, "Error: --template is required")
		flag.Usage()
		os.Exit(1)
	}

	// Check template exists
	if _, err := os.Stat(*templatePath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: template file not found: %s\n", *templatePath)
		os.Exit(1)
	}

	// Parse data
	var data map[string]string
	if *dataPath != "" {
		// Read from JSON file
		content, err := os.ReadFile(*dataPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading data file: %v\n", err)
			os.Exit(1)
		}
		if err := json.Unmarshal(content, &data); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing JSON file: %v\n", err)
			os.Exit(1)
		}
	} else if *dataJSON != "" {
		// Parse inline JSON
		if err := json.Unmarshal([]byte(*dataJSON), &data); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Fprintln(os.Stderr, "Error: either --data or --json is required")
		flag.Usage()
		os.Exit(1)
	}

	// Open template
	template, err := docx.Open(*templatePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening template: %v\n", err)
		os.Exit(1)
	}
	defer template.Close()

	// Apply all placeholders
	for placeholder, htmlContent := range data {
		placeholder = strings.TrimPrefix(placeholder, "{{")
		placeholder = strings.TrimSuffix(placeholder, "}}")

		if err := template.SetContent(placeholder, htmlContent); err != nil {
			fmt.Fprintf(os.Stderr, "Error setting content for '%s': %v\n", placeholder, err)
			os.Exit(1)
		}
		fmt.Printf("✓ Replaced {{%s}}\n", placeholder)
	}

	// Save output
	if err := template.Save(*outputPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving output: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Output saved to: %s\n", *outputPath)
}
