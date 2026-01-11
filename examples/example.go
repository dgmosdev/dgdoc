package main

import (
	"fmt"
	"log"

	"github.com/dgmos/dgdoc/docx"
)

func main() {
	// Open template
	template, err := docx.Open("examples/input.docx")
	if err != nil {
		log.Fatalf("Failed to open template: %v", err)
	}
	defer template.Close()

	// Main content with full HTML support
	mainContent := `
		<h1>Proje Raporu</h1>
		<p>Bu rapor <strong>otomatik olarak</strong> oluşturulmuştur.</p>
		
		<h2>Özet</h2>
		<ul>
			<li>Proje tamamlandı</li>
			<li>Bütçe dahilinde kalındı</li>
			<li><span style="color: green;">✓ Başarılı</span></li>
		</ul>
		
		<h2>Detaylar</h2>
		<table>
			<tr><th>Görev</th><th>Durum</th><th>Tarih</th></tr>
			<tr><td>Analiz</td><td><span style="background-color: lightgreen;">Tamamlandı</span></td><td>01/01/2026</td></tr>
			<tr><td>Geliştirme</td><td><span style="background-color: lightgreen;">Tamamlandı</span></td><td>15/01/2026</td></tr>
			<tr><td>Test</td><td><span style="background-color: lightyellow;">Devam Ediyor</span></td><td>20/01/2026</td></tr>
		</table>
	`

	// Replace multiple placeholders at once using Apply
	data := map[string]any{
		"customer_name": "Ahmet Can Bilgay",
		"address":       "Atatürk Mah. Marmara Sok. No:5, Ümraniye, İstanbul",
		"date":          "<strong>11 Ocak 2026</strong>",
		"status":        `<span style="color: white; background-color: green;"> AKTİF </span>`,
		"content":       mainContent,
	}

	if err := template.Apply(data); err != nil {
		log.Fatalf("Failed to apply data: %v", err)
	}

	// Save
	if err := template.Save("output.docx"); err != nil {
		log.Fatalf("Failed to save: %v", err)
	}

	fmt.Println("✓ output.docx oluşturuldu!")
}
