# Integration Test Fixtures

Bu klasör integration testler için gerçek doküman template'lerini içerir.

## Fixture Dosyaları (Opsiyonel)

Integration testler gerçek DOCX, XLSX, PPTX, ve ODT dosyalarıyla çalışmak üzere tasarlanmıştır. Test dosyaları aşağıdaki gibi olmalıdır:

### DOCX Templates
- `simple_template.docx` - Basit placeholder'lar: `{name}`, `{company}`
- `conditional_template.docx` - Conditional: `{#if premium}Premium{#else}Free{/if}`
- `loop_template.docx` - Loop: `{#items}{name}: {price}{/items}`
- `html_template.docx` - HTML içerik: `{content}`

### XLSX Templates
- `excel_template.xlsx` - Cell placeholders: `{company}`, `{revenue}`, `{year}`
- `excel_loop_template.xlsx` - Row loops: `{#items}{product}{price}{/items}`

### PPTX Templates  
- `presentation_template.pptx` - Slide text: `{title}`, `{presenter}`, `{date}`

### ODT Templates
- `odt_template.odt` - Text placeholders: `{title}`, `{author}`

## Test Çalıştırma

```bash
# Fixture'lar yoksa testler skip edilir
go test ./tests/integration -v

# Fixture'larla test etmek için önce template dosyalarını oluşturun
# Manuel olarak Word/Excel/PowerPoint ile template'ler hazırlayın
# ve fixtures/ klasörüne koyun
```

## Not

Integration testler **opsiyonel**dir. Fixture dosyaları yoksa testler otomatik olarak skip edilir. Bu sayede CI/CD'de veya fixture olmayan ortamlarda sorun çıkmaz.
