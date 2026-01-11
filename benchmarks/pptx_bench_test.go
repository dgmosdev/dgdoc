package benchmarks

import (
	"testing"
)

// BenchmarkPPTXDataProcessing benchmarks PPTX data processing
func BenchmarkPPTXDataProcessing(b *testing.B) {
	data := map[string]any{
		"title":     "Q1 Report",
		"presenter": "John Doe",
		"date":      "2026-01-11",
		"subtitle":  "Annual Review",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = processMapData(data)
	}
}

// BenchmarkPPTXMultipleSlides benchmarks with many slides
func BenchmarkPPTXMultipleSlides(b *testing.B) {
	slides := make([]any, 50)
	for i := range slides {
		slides[i] = map[string]any{
			"title":   "Slide Title",
			"content": "Slide content here",
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = processArrayData(slides)
	}
}
