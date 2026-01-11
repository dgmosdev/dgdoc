package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dgmosdev/dgdoc/pptx"
)

// TestPPTXIntegration tests end-to-end PPTX functionality
func TestPPTXIntegration(t *testing.T) {
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
			name:     "Slide text replacement",
			template: "presentation_template.pptx",
			data: map[string]any{
				"title":     "Q1 Report",
				"presenter": "John Doe",
				"date":      "2026-01-11",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			outputPath := filepath.Join(os.TempDir(), "test_pptx_output.pptx")
			defer os.Remove(outputPath)

			templatePath := filepath.Join(fixturesDir, tc.template)
			if _, err := os.Stat(templatePath); os.IsNotExist(err) {
				t.Skipf("Template file %s not found - skipping", tc.template)
			}

			template, err := pptx.Open(templatePath)
			if err != nil {
				t.Fatalf("Failed to open template: %v", err)
			}
			defer template.Close()

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
