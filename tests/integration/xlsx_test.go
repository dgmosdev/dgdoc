package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dgmosdev/dgdoc/xlsx"
)

// TestXLSXIntegration tests end-to-end XLSX functionality
func TestXLSXIntegration(t *testing.T) {
	fixturesDir := filepath.Join("..", "fixtures")
	if _, err := os.Stat(fixturesDir); os.IsNotExist(err) {
		t.Skip("Fixtures directory not found - skipping integration tests")
	}

	testCases := []struct {
		name     string
		template string
		data     map[string]any
	}{
		{
			name:     "Cell value replacement",
			template: "excel_template.xlsx",
			data: map[string]any{
				"company": "Test Corp",
				"revenue": "1000000",
				"year":    "2026",
			},
		},
		{
			name:     "Loop support",
			template: "excel_loop_template.xlsx",
			data: map[string]any{
				"items": []any{
					map[string]any{"product": "Widget A", "price": "100"},
					map[string]any{"product": "Widget B", "price": "200"},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			outputPath := filepath.Join(os.TempDir(), "test_excel_output.xlsx")
			defer func() { _ = os.Remove(outputPath) }()

			templatePath := filepath.Join(fixturesDir, tc.template)
			if _, err := os.Stat(templatePath); os.IsNotExist(err) {
				t.Skipf("Template file %s not found - skipping", tc.template)
			}

			template, err := xlsx.Open(templatePath)
			if err != nil {
				t.Fatalf("Failed to open template: %v", err)
			}
			defer func() { _ = template.Close() }()

			if err := template.Apply(tc.data); err != nil {
				t.Fatalf("Failed to apply data: %v", err)
			}

			if err := template.Save(outputPath); err != nil {
				t.Fatalf("Failed to save output: %v", err)
			}

			if _, err := os.Stat(outputPath); os.IsNotExist(err) {
				t.Fatal("Output file was not created")
			}

			t.Logf("Successfully processed %s", tc.name)
		})
	}
}
