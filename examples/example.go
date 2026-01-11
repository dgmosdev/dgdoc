package main

import (
	"fmt"
	"log"

	"github.com/dgmos/dgdoc/docx"
)

func main() {
	// Open template
	template, err := docx.Open("template.docx")
	if err != nil {
		log.Fatalf("Failed to open template: %v", err)
	}
	defer template.Close()

	// Replace multiple placeholders with HTML content

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
	template.SetContent("content", mainContent)

	// Normal fields (plain text)
	template.ReplaceText("customer_name", "Ahmet Can Bilgay")
	template.ReplaceText("address", "Atatürk Mah. Marmara Sok. No:5, Ümraniye, İstanbul")

	// Date in bold using SetContent
	template.SetContent("date", `<strong>11 Ocak 2026</strong>`)

	// Status indicator using SetContent
	template.SetContent("status", `<span style="color: white; background-color: green;"> AKTİF </span>`)

	// Save
	if err := template.Save("output.docx"); err != nil {
		log.Fatalf("Failed to save: %v", err)
	}

	fmt.Println("✓ output.docx oluşturuldu!")
}
