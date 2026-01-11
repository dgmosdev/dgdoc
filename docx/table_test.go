package docx

import (
	"strings"
	"testing"
)

func TestCellMerge_Colspan(t *testing.T) {
	template := &Template{
		files: map[string][]byte{},
	}

	html := `<table>
		<tr>
			<td colspan="2">Merged</td>
			<td>Normal</td>
		</tr>
		<tr>
			<td>Cell 1</td>
			<td>Cell 2</td>
			<td>Cell 3</td>
		</tr>
	</table>`

	result, err := template.HTMLToOOXML(html)
	if err != nil {
		t.Fatalf("HTMLToOOXML failed: %v", err)
	}

	// Should contain gridSpan for colspan
	if !strings.Contains(result, "gridSpan") {
		t.Errorf("Expected result to contain gridSpan for colspan, got: %s", result)
	}

	// Should contain the merged cell content
	if !strings.Contains(result, "Merged") {
		t.Errorf("Expected result to contain 'Merged', got: %s", result)
	}
}

func TestCellMerge_Rowspan(t *testing.T) {
	template := &Template{
		files: map[string][]byte{},
	}

	html := `<table>
		<tr>
			<td rowspan="2">Tall</td>
			<td>Cell 1</td>
		</tr>
		<tr>
			<td>Cell 2</td>
		</tr>
	</table>`

	result, err := template.HTMLToOOXML(html)
	if err != nil {
		t.Fatalf("HTMLToOOXML failed: %v", err)
	}

	// Should contain vMerge for rowspan
	if !strings.Contains(result, "vMerge") {
		t.Errorf("Expected result to contain vMerge for rowspan, got: %s", result)
	}

	// Should contain restart and continue markers
	if !strings.Contains(result, "restart") {
		t.Errorf("Expected result to contain vMerge restart, got: %s", result)
	}
	if !strings.Contains(result, "continue") {
		t.Errorf("Expected result to contain vMerge continue, got: %s", result)
	}
}

func TestCellMerge_Combined(t *testing.T) {
	template := &Template{
		files: map[string][]byte{},
	}

	html := `<table>
		<tr>
			<td colspan="2" rowspan="2">Big Cell</td>
			<td>A</td>
		</tr>
		<tr>
			<td>B</td>
		</tr>
		<tr>
			<td>C</td>
			<td>D</td>
			<td>E</td>
		</tr>
	</table>`

	result, err := template.HTMLToOOXML(html)
	if err != nil {
		t.Fatalf("HTMLToOOXML failed: %v", err)
	}

	// Should contain both gridSpan and vMerge
	if !strings.Contains(result, "gridSpan") {
		t.Errorf("Expected result to contain gridSpan")
	}
	if !strings.Contains(result, "vMerge") {
		t.Errorf("Expected result to contain vMerge")
	}
	if !strings.Contains(result, "Big Cell") {
		t.Errorf("Expected result to contain cell content")
	}
}

func TestTableWithNoMerge(t *testing.T) {
	template := &Template{
		files: map[string][]byte{},
	}

	html := `<table>
		<tr>
			<td>A</td>
			<td>B</td>
		</tr>
		<tr>
			<td>C</td>
			<td>D</td>
		</tr>
	</table>`

	result, err := template.HTMLToOOXML(html)
	if err != nil {
		t.Fatalf("HTMLToOOXML failed: %v", err)
	}

	// Should not contain merge markers
	if strings.Contains(result, "gridSpan") {
		t.Errorf("Did not expect gridSpan in simple table")
	}
	if strings.Contains(result, "vMerge") {
		t.Errorf("Did not expect vMerge in simple table")
	}

	// Should contain all cells
	for _, text := range []string{"A", "B", "C", "D"} {
		if !strings.Contains(result, text) {
			t.Errorf("Expected result to contain '%s'", text)
		}
	}
}
