package benchmarks

import (
	"testing"
)

// Note: DOCX benchmarks test data processing overhead
// For full file I/O benchmarks, use integration tests

// BenchmarkDOCXApply benchmarks basic Apply operation
func BenchmarkDOCXApply(b *testing.B) {
	data := map[string]any{
		"name":    "John Doe",
		"company": "Acme Corp",
		"email":   "john@acme.com",
		"title":   "Senior Engineer",
		"phone":   "+1-555-0123",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = processMapData(data)
	}
}

// BenchmarkDOCXLargeDataset benchmarks with larger dataset
func BenchmarkDOCXLargeDataset(b *testing.B) {
	items := make([]any, 100)
	for i := range items {
		items[i] = map[string]any{
			"name":  "Item",
			"price": "100.00",
			"qty":   "5",
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = processArrayData(items)
	}
}
