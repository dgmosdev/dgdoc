package benchmarks

// Helper functions shared across benchmarks

// processMapData processes a map and returns count
func processMapData(data map[string]any) int {
	count := 0
	for range data {
		count++
	}
	return count
}

// processArrayData processes an array and returns count
func processArrayData(items []any) int {
	count := 0
	for range items {
		count++
	}
	return count
}
