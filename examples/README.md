# dgdoc Examples

Bu klasör `dgdoc` kullanımını gösteren örnek dosyalar içerir.

## Klasör Yapısı

```
examples/
├── templates/          # Örnek şablon dosyaları (DOCX, XLSX, PPTX, ODT)
├── data/              # Örnek JSON veri dosyaları
├── cli_usage/         # CLI kullanım örnekleri
└── README.md          # Bu dosya
```

## Örnek Veri Dosyaları

### `data/invoice_data.json`
Fatura oluşturma için örnek veri:
- Customer bilgileri
- Ürün listesi (loop örneği)
- Conditional rendering (premium müşteri)
- HTML içerik (notlar)

### `data/report_data.json`
Rapor oluşturma için örnek veri:
- Bölge bazlı satış verileri (loop)
- Executive summary (HTML formatında)
- Conditional sections

## CLI Kullanım Örnekleri

`cli_usage/examples.sh` dosyası çeşitli kullanım senaryolarını gösterir:

1. **Basit kullanım** - JSON dosyası ile
2. **Inline JSON** - Hızlı test için
3. **Conditionals ve Loops** - Gelişmiş özellikler
4. **Multi-format** - Excel, PowerPoint desteği
5. **Batch processing** - Toplu işlem
6. **Environment variables** - Dinamik içerik

## Şablon Dosyaları Oluşturma

### DOCX Şablonu

Word'de placeholder'lar ekleyin:
```
Sayın {customer_name},

{#if premium}
Premium müşterimiz olduğunuz için teşekkür ederiz!
{#else}
Müşterimiz olduğunuz için teşekkür ederiz!
{/if}

Siparişleriniz:
{#items}
- {description}: {quantity} x {unit_price} = {total}
{/items}

Toplam: {total}
```

### XLSX Şablonu

Excel'de hücrelere placeholder'lar yerleştirin:
```
A1: {company_name}
A2: {revenue}
A3: {year}

Döngüler için:
A5: {#items}
B5: {product}
C5: {price}
A6: {/items}
```

### PPTX Şablonu

PowerPoint'te text kutularına:
```
Slide 1: {title}
Slide 2: {presenter}
         {date}

İçerik döngüsü:
{#slides}
{topic}: {amount}
{/slides}
```

## Hızlı Başlangıç

1. **Şablon hazırlayın** (Word/Excel/PowerPoint)
   - Placeholder'lar ekleyin: `{variable_name}`
   - Conditionals: `{#if condition}...{/if}`
   - Loops: `{#arrayName}...{/arrayName}`

2. **JSON verisi oluşturun**
   ```json
   {
     "variable_name": "value",
     "condition": true,
     "arrayName": [
       {"item": "value1"},
       {"item": "value2"}
     ]
   }
   ```

3. **CLI ile çalıştırın**
   ```bash
   dgdoc --template template.docx \
         --output output.docx \
         --data data.json
   ```

## Go Library Kullanımı

```go
package main

import (
    "github.com/dgmosdev/dgdoc/docx"
)

func main() {
    template, _ := docx.Open("template.docx")
    defer template.Close()
    
    data := map[string]any{
        "customer_name": "Acme Corp",
        "premium": true,
        "items": []any{
            map[string]any{"description": "Product A", "price": "100"},
        },
    }
    
    template.Apply(data)
    template.Save("output.docx")
}
```

## Özellikler

- ✅ **HTML içerik**: `<h1>`, `<p>`, `<b>`, `<i>`, `<ul>`, `<table>` vb.
- ✅ **Conditionals**: `{#if}`, `{#else}`, `{/if}`
- ✅ **Loops**: `{#items}...{/items}` 
- ✅ **Images**: `{%logo}` → URL, Base64, veya dosya yolu
- ✅ **Links**: `{%website}` → "Text|https://url"
- ✅ **Multi-format**: DOCX, XLSX, PPTX, ODT

## Daha Fazla Bilgi

- [GitHub Repository](https://github.com/dgmosdev/dgdoc)
- [Go Package Documentation](https://pkg.go.dev/github.com/dgmosdev/dgdoc)
- [README.md](../README.md) - Detaylı dokümantasyon
