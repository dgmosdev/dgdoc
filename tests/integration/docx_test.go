package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dgmosdev/dgdoc/docx"
)

// TestDOCXIntegration tests end-to-end DOCX functionality
func TestDOCXIntegration(t *testing.T) {
	// Skip if fixtures are not available
	fixturesDir := filepath.Join("..", "fixtures")
	if _, err := os.Stat(fixturesDir); os.IsNotExist(err) {
		t.Skip("Fixtures directory not found - skipping integration tests")
	}

	testCases := []struct {
		name     string
		template string
		data     map[string]any
		expected []string // strings that should be in the output
	}{
		{
			name:     "Basic placeholder replacement",
			template: "simple_template.docx",
			data: map[string]any{
				"name":    "John Doe",
				"company": "Acme Corp",
			},
			expected: []string{"John Doe", "Acme Corp"},
		},
		{
			name:     "Conditionals",
			template: "conditional_template.docx",
			data: map[string]any{
				"premium": true,
				"name":    "Jane Smith",
			},
			expected: []string{"Premium", "Jane Smith"},
		},
		{
			name:     "Loops",
			template: "loop_template.docx",
			data: map[string]any{
				"items": []any{
					map[string]any{"name": "Item 1", "price": "100"},
					map[string]any{"name": "Item 2", "price": "200"},
				},
			},
			expected: []string{"Item 1", "Item 2", "100", "200"},
		},
		{
			name:     "HTML content",
			template: "html_template.docx",
			data: map[string]any{
				"content": "<h1>Title</h1><p>Paragraph with <b>bold</b> text</p>",
			},
			expected: []string{"Title", "bold"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create temp output file
			outputPath := filepath.Join(os.TempDir(), "test_output.docx")
			defer func() { _ = os.Remove(outputPath) }()

			// Open template (if it exists)
			templatePath := filepath.Join(fixturesDir, tc.template)
			if _, err := os.Stat(templatePath); os.IsNotExist(err) {
				t.Skipf("Template file %s not found - skipping", tc.template)
			}

			template, err := docx.Open(templatePath)
			if err != nil {
				t.Fatalf("Failed to open template: %v", err)
			}
			defer func() { _ = template.Close() }()

			// Apply data
			if err := template.Apply(tc.data); err != nil {
				t.Fatalf("Failed to apply data: %v", err)
			}

			// Save output
			if err := template.Save(outputPath); err != nil {
				t.Fatalf("Failed to save output: %v", err)
			}

			// Verify output file exists
			if _, err := os.Stat(outputPath); os.IsNotExist(err) {
				t.Fatal("Output file was not created")
			}

			// TODO: Open and verify content
			// For now, just check file was created successfully
			t.Logf("Successfully processed %s", tc.name)
		})
	}
}

// TestDOCXMetadata tests metadata functionality
func TestDOCXMetadata(t *testing.T) {
	// Create a temporary test file
	outputPath := filepath.Join(os.TempDir(), "metadata_test.docx")
	defer func() { _ = os.Remove(outputPath) }()

	// Since we need a real DOCX to start with, we'll skip if no fixtures
	fixturesDir := filepath.Join("..", "fixtures")
	templatePath := filepath.Join(fixturesDir, "simple_template.docx")

	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		t.Skip("Fixtures not available - skipping metadata test")
	}

	template, err := docx.Open(templatePath)
	if err != nil {
		t.Fatalf("Failed to open template: %v", err)
	}
	defer func() { _ = template.Close() }()

	// Set metadata
	meta := docx.Metadata{
		Title:    "Test Document",
		Author:   "Test Author",
		Keywords: "test, integration, dgdoc",
	}

	if err := template.SetMetadata(meta); err != nil {
		t.Fatalf("Failed to set metadata: %v", err)
	}

	// Set page setup
	pageSetup := docx.PageSetup{
		PageWidth:  11906,
		PageHeight: 16838,
		MarginTop:  1440,
	}

	if err := template.SetPageSetup(pageSetup); err != nil {
		t.Fatalf("Failed to set page setup: %v", err)
	}

	// Save
	if err := template.Save(outputPath); err != nil {
		t.Fatalf("Failed to save: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatal("Output file was not created")
	}

	t.Log("Metadata test completed successfully")
}
