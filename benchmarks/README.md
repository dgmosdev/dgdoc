# Performance Benchmarks

Bu klasör dgdoc için performance benchmark'larını içerir.

## Benchmark Çalıştırma

Tüm benchmark'ları çalıştırmak için:

```bash
go test -bench=. -benchmem ./benchmarks
```

Belirli bir modül için:

```bash
# Sadece DOCX
go test -bench=BenchmarkDOCX -benchmem ./benchmarks

# Sadece XLSX
go test -bench=BenchmarkXLSX -benchmem ./benchmarks

# Sadece PPTX
go test -bench=BenchmarkPPTX -benchmem ./benchmarks
```

## Benchmark Kategorileri

### Data Processing Benchmarks

Bu benchmark'lar template engine'in veri işleme performansını ölçer:

**DOCX Benchmarks:**
- `BenchmarkDOCXApply` - Temel veri uygulama (5 field)
- `BenchmarkDOCXLargeDataset` - Büyük veri seti (100 item loop)

**XLSX Benchmarks:**
- `BenchmarkXLSXDataProcessing` - Hücre veri işleme
- `BenchmarkXLSXLargeSpreadsheet` - Büyük spreadsheet (1000 satır)

**PPTX Benchmarks:**
- `BenchmarkPPTXDataProcessing` - Slayt veri işleme
- `BenchmarkPPTXMultipleSlides` - Çoklu slayt (50 slayt)

> **Not:** Bu benchmark'lar veri işleme overhead'ini ölçer. Gerçek dosya I/O performansı 
> için integration testlerle gerçek DOCX/XLSX/PPTX dosyaları kullanılmalıdır.

## Örnek Çıktı

```
goos: darwin
goarch: arm64
pkg: github.com/dgmosdev/dgdoc/benchmarks

BenchmarkDOCXApply-8                 5000000    250 ns/op      64 B/op     1 allocs/op
BenchmarkDOCXLargeDataset-8           500000   3500 ns/op    1024 B/op    15 allocs/op

BenchmarkXLSXDataProcessing-8        8000000    180 ns/op      48 B/op     1 allocs/op
BenchmarkXLSXLargeSpreadsheet-8        50000  35000 ns/op    8192 B/op   100 allocs/op

BenchmarkPPTXDataProcessing-8        7000000    200 ns/op      56 B/op     1 allocs/op
BenchmarkPPTXMultipleSlides-8         100000  15000 ns/op    4096 B/op    50 allocs/op
```

## Performans Metrikler

- **ns/op**: İşlem başına nanosaniye
- **B/op**: İşlem başına ayrılan byte
- **allocs/op**: İşlem başına memory allocation sayısı

## CPU Profiling

CPU profiling için:

```bash
go test -bench=. -cpuprofile=cpu.prof ./benchmarks
go tool pprof cpu.prof
```

## Memory Profiling

Memory profiling için:

```bash
go test -bench=. -memprofile=mem.prof ./benchmarks
go tool pprof mem.prof
```

## Optimizasyon İçin İpuçları

1. **Batch Processing**: Tek tek değişim yerine `Apply()` kullanın
2. **Memory Reuse**: Büyük döngülerde template'i yeniden kullanın
3. **Lazy Loading**: Sadece ihtiyaç olan dosyaları açın
4. **Parallel Processing**: Çok sayıda doküman için goroutine'ler kullanın

## Karşılaştırma

Farklı versiyonları karşılaştırmak için:

```bash
# v1 benchmark
go test -bench=. ./benchmarks > old.txt

# Kod değişikliği yap

# v2 benchmark
go test -bench=. ./benchmarks > new.txt

# Karşılaştır
benchcmp old.txt new.txt
```

## CI/CD Entegrasyonu

GitHub Actions veya benzeri CI'da:

```yaml
- name: Run Benchmarks
  run: go test -bench=. -benchmem ./benchmarks
  
- name: Performance Regression Check
  run: |
    go test -bench=. ./benchmarks > new.txt
    # Compare with baseline
```
