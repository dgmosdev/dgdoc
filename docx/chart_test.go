package docx

import (
	"testing"
)

func TestDetectCharts(t *testing.T) {
	// Create a document with chart reference
	template := &Template{
		files: make(map[string][]byte),
	}
	template.files["word/document.xml"] = []byte(`<?xml version="1.0"?>
<w:document>
	<w:body>
		<w:p>
			<w:r>
				<w:drawing>
					<wp:inline>
						<a:graphic>
							<a:graphicData uri="http://schemas.openxmlformats.org/drawingml/2006/chart">
								<c:chart r:id="rId5"/>
							</a:graphicData>
						</a:graphic>
					</wp:inline>
				</w:drawing>
			</w:r>
		</w:p>
	</w:body>
</w:document>`)

	charts, err := template.DetectCharts()
	if err != nil {
		t.Fatalf("DetectCharts failed: %v", err)
	}

	if len(charts) != 1 {
		t.Errorf("Expected 1 chart, found %d", len(charts))
	}

	if len(charts) > 0 && charts[0] != "rId5" {
		t.Errorf("Expected chart rId5, got %s", charts[0])
	}
}

func TestHasCharts(t *testing.T) {
	// Document with chart
	withChart := &Template{
		files: make(map[string][]byte),
	}
	withChart.files["word/document.xml"] = []byte(`<w:document><c:chart r:id="rId5"/></w:document>`)

	if !withChart.HasCharts() {
		t.Error("Expected document to have charts")
	}

	// Document without chart
	withoutChart := &Template{
		files: make(map[string][]byte),
	}
	withoutChart.files["word/document.xml"] = []byte(`<w:document><w:p><w:r><w:t>No charts here</w:t></w:r></w:p></w:document>`)

	if withoutChart.HasCharts() {
		t.Error("Expected document to have no charts")
	}
}

func TestGetChartCount(t *testing.T) {
	template := &Template{
		files: make(map[string][]byte),
	}
	template.files["word/document.xml"] = []byte(`<?xml version="1.0"?>
<w:document>
	<c:chart r:id="rId5"/>
	<c:chart r:id="rId6"/>
	<c:chart r:id="rId7"/>
</w:document>`)

	count := template.GetChartCount()
	if count != 3 {
		t.Errorf("Expected 3 charts, found %d", count)
	}
}
