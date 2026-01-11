package benchmarks

import (
	"testing"
)

// BenchmarkXLSXDataProcessing benchmarks XLSX data processing
func BenchmarkXLSXDataProcessing(b *testing.B) {
	data := map[string]any{
		"company": "Acme Corp",
		"revenue": "1000000",
		"year":    "2026",
		"quarter": "Q1",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = processMapData(data)
	}
}

// BenchmarkXLSXLargeSpreadsheet benchmarks with many rows
func BenchmarkXLSXLargeSpreadsheet(b *testing.B) {
	items := make([]any, 1000)
	for i := range items {
		items[i] = map[string]any{
			"product": "Widget",
			"price":   "99.99",
			"stock":   "100",
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = processArrayData(items)
	}
}
