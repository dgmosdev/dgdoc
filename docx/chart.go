package docx

import (
	"fmt"
	"regexp"
)

// ChartData represents data for a chart
type ChartData struct {
	Categories []string
	Series     []ChartSeries
}

// ChartSeries represents a data series in a chart
type ChartSeries struct {
	Name   string
	Values []float64
}

// DetectCharts finds all charts in the document
func (t *Template) DetectCharts() ([]string, error) {
	var charts []string

	// Check document.xml for chart relationships
	content := string(t.files["word/document.xml"])

	// Pattern to find chart references: <c:chart r:id="rId..."/>
	chartPattern := regexp.MustCompile(`<c:chart[^>]*r:id="([^"]+)"[^>]*/>`)
	matches := chartPattern.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) >= 2 {
			charts = append(charts, match[1])
		}
	}

	return charts, nil
}

// ReplaceChartData replaces data in a chart
// This is a simplified implementation that demonstrates the concept
func (t *Template) ReplaceChartData(chartName string, data ChartData) error {
	// In a full implementation, this would:
	// 1. Find the chart file (word/charts/chart*.xml)
	// 2. Find the embedded Excel data (xl/worksheets/sheet*.xml in chart relationships)
	// 3. Replace the data values
	// 4. Update the chart cache

	// For now, we'll provide a basic structure
	// Real implementation would require parsing the chart XML structure

	return fmt.Errorf("chart data replacement not yet fully implemented - requires chart XML parsing")
}

// HasCharts checks if the document contains any charts
func (t *Template) HasCharts() bool {
	charts, _ := t.DetectCharts()
	return len(charts) > 0
}

// GetChartCount returns the number of charts in the document
func (t *Template) GetChartCount() int {
	charts, _ := t.DetectCharts()
	return len(charts)
}

// Note: Full chart implementation would include:
// - Parsing chart*.xml files
// - Finding embedded Excel data files
// - Updating both the data and the chart cache
// - Supporting different chart types (bar, line, pie, etc.)
// - Preserving chart styling and options
