# ArticleHub — Elasticsearch Search API (Go)

Bir makale platformunun arama servisini ES + Go ile uçtan uca uygulayan örnek proje.

## Stack

- Go 1.22+
- Elasticsearch 8.13 (Docker)
- Kibana 8.13 (Docker)
- [`github.com/elastic/go-elasticsearch/v8`](https://github.com/elastic/go-elasticsearch)
- [`github.com/gin-gonic/gin`](https://github.com/gin-gonic/gin)
- [`github.com/joho/godotenv`](https://github.com/joho/godotenv)
- [`github.com/brianvoe/gofakeit/v7`](https://github.com/brianvoe/gofakeit)

## Features

- Full-text search (Turkish custom analyzer)
- Multi-field, fuzzy, faceted search
- Aggregations (categories, tags, time-series, stats)
- Autocomplete (completion suggester + edge n-gram)
- Zero-downtime reindex with aliases
- Bulk indexing (manual NDJSON + BulkIndexer)
- Pagination (from/size + search_after), sorting, highlighting

## Quick Start

```bash
docker compose up -d
go run ./cmd/seed --count 500
go run ./cmd/api
```

## Architecture

```
┌─────────────┐     HTTP      ┌──────────────┐     ┌───────────────────┐
│   Client    │ ────────────▶ │  Gin API     │ ──▶ │  Elasticsearch    │
│  (curl/UI)  │               │  :8080       │     │  :9200            │
└─────────────┘               └──────────────┘     └───────────────────┘
                                                             │
                                                    ┌───────────────────┐
                                                    │     Kibana        │
                                                    │     :5601         │
                                                    └───────────────────┘
```

## API Reference

| Method | Endpoint | Açıklama |
|--------|----------|----------|
| `POST` | `/articles` | Yeni makale ekle |
| `GET` | `/articles/:id` | ID ile getir |
| `PUT` | `/articles/:id` | Tam güncelle |
| `PATCH` | `/articles/:id` | Partial güncelle |
| `DELETE` | `/articles/:id` | Sil |
| `GET` | `/search?q=&category=&from=&size=&sort=` | Arama |
| `GET` | `/search/facets?q=` | Facet'li arama |
| `GET` | `/suggest?q=` | Autocomplete |
| `POST` | `/admin/seed?count=500` | Seed data (dev) |
| `GET` | `/healthz` | Cluster sağlığı |

### Örnek curl'ler

```bash
# Yeni makale ekle
curl -X POST http://localhost:8080/articles \
  -H "Content-Type: application/json" \
  -d '{"title":"Elasticsearch Nedir","content":"...","author":"Ahmet","tags":["search","go"],"category":"tech"}'

# Arama
curl "http://localhost:8080/search?q=elasticsearch&category=tech&from=0&size=10"

# Facet'li arama
curl "http://localhost:8080/search/facets?q=golang"

# Autocomplete
curl "http://localhost:8080/suggest?q=elas"

# Sağlık kontrolü
curl http://localhost:8080/healthz
```

## Index Design

```json
{
  "settings": {
    "number_of_shards": 1,
    "number_of_replicas": 0,
    "analysis": {
      "analyzer": {
        "turkish_custom": {
          "type": "custom",
          "tokenizer": "standard",
          "filter": ["lowercase", "turkish_stop", "turkish_stemmer"]
        }
      },
      "filter": {
        "turkish_stop": { "type": "stop", "stopwords": "_turkish_" },
        "turkish_stemmer": { "type": "stemmer", "language": "turkish" }
      }
    }
  },
  "mappings": {
    "properties": {
      "title":        { "type": "text", "analyzer": "turkish_custom",
                        "fields": { "keyword": { "type": "keyword" } } },
      "content":      { "type": "text", "analyzer": "turkish_custom" },
      "author":       { "type": "keyword" },
      "tags":         { "type": "keyword" },
      "category":     { "type": "keyword" },
      "published_at": { "type": "date" },
      "view_count":   { "type": "integer" }
    }
  }
}
```

---

## Task List

### Görev 1: Ortamı Kur
- [ ] `articlehub/` proje klasörünü oluştur
- [ ] `docker-compose.yml` yaz (ES single-node, security kapalı + Kibana)
- [ ] `go mod init github.com/<kullanıcıadın>/articlehub`
- [ ] `main.go` — ES client kur, cluster health endpoint'ine bas, response'u yazdır
- [ ] `docker compose up -d` ile her şeyi ayağa kaldır, `http://localhost:9200` ve `http://localhost:5601` erişimini doğrula

> **Öğrendiklerin:** Container'lardan ES ayağa kaldırma, Kibana Dev Tools'a erişim, Go client kurma.

---

### Görev 2: Index'i Mapping ile Oluştur
- [ ] `indexer` paketi oluştur
- [ ] `CreateArticlesIndex()` fonksiyonunu yaz
- [ ] Turkish custom analyzer tanımla (`turkish_stop` + `turkish_stemmer` filter)
- [ ] `title` field'ını hem `text` hem `keyword` (multi-field) olarak tanımla
- [ ] `--recreate` flag'iyle mevcut index'i silip yeniden oluştur
- [ ] Kibana Dev Tools'ta `GET /articles/_mapping` ile mapping'i doğrula

> **Öğrendiklerin:** Custom analyzer, `text` vs `keyword` ayrımı, multi-field, shard/replica ayarı, idempotent migration.

---

### Görev 3: CRUD İşlemleri
- [ ] `Article` struct'ını tanımla
- [ ] `IndexArticle(ctx, article)` — yeni döküman ekle
- [ ] `GetArticle(ctx, id)` — ID ile getir
- [ ] `UpdateArticle(ctx, id, partial)` — partial update (Update API, `doc` field)
- [ ] `DeleteArticle(ctx, id)` — sil
- [ ] Full reindex (Index API) ile partial update (Update API) farkını dene ve anla
- [ ] `?refresh=wait_for` parametresini test sırasında kullan
- [ ] README'ye `if_seq_no` / `if_primary_term` ile optimistic concurrency control hakkında not ekle

> **Öğrendiklerin:** Index/Get/Update/Delete API'leri, refresh semantiği, optimistic concurrency control.

---

### Görev 4: Bulk Insert ile Seed Data
- [ ] `github.com/brianvoe/gofakeit/v7` ekle
- [ ] Rastgele `Article` üretecek fonksiyon yaz (title, content, author, category, 1-5 tag, son 2 yıl içinde tarih, rastgele view_count)
- [ ] **Manuel bulk:** NDJSON formatında body oluştur, `_bulk` endpoint'ine POST et
- [ ] **BulkIndexer:** `esutil.BulkIndexer` ile goroutine'lerle paralel index'le
- [ ] 500 dökümanın index'lenme süresini yazdır, iki yöntemi karşılaştır
- [ ] Bazı dökümanlar başarısız olursa hata yönetimini handle et

> **Öğrendiklerin:** Bulk API'nin NDJSON formatı, throughput optimizasyonu, BulkIndexer ile concurrent indexing.

---

### Görev 5: Arama — Temel Query'ler
- [ ] `SearchByTitle(q string)` — `match` query, sadece `title`
- [ ] `SearchMultiField(q string)` — `multi_match`, `title^3` + `content`
- [ ] `SearchExactTag(tag string)` — `term` query, `tags` field (keyword)
- [ ] `SearchByCategoryAndDateRange(cat string, from, to time.Time)` — `bool` query: `filter` içinde `term` (category) + `range` (published_at)
- [ ] `FuzzySearch(q string)` — `match` query, `fuzziness: "AUTO"` ("elasitc" → "elastic")
- [ ] Her fonksiyon için en az 2 test case'i README'ye ekle
- [ ] `map[string]any` ile başla, sonra struct'lara geç — her ikisini de gör

> **Öğrendiklerin:** `match` vs `term`, `bool` query yapısı, `must` vs `filter` ayrımı, field boosting, fuzzy matching.

---

### Görev 6: Pagination, Sorting, Highlighting
- [ ] `from`/`size` ile basit sayfalama
- [ ] `search_after` ile derin sayfalama
- [ ] README'de farkı açıkla: `from + size > 10.000` neden hata verir, `search_after` neden tercih edilir
- [ ] `published_at` ile azalan/artan sıralama
- [ ] `view_count` ile azalan sıralama
- [ ] Score ile ikincil sıralama dene
- [ ] Highlighting: eşleşen kelimeleri `<em>` ile sar, `pre_tags`/`post_tags` ile özelleştir

> **Öğrendiklerin:** Pagination stratejileri ve sınırları, sort kombinasyonları, highlighting.

---

### Görev 7: Aggregations (Facet'ler)
- [ ] `SearchWithFacets` fonksiyonunu yaz — hem `hits` hem `aggregations` dönsün
- [ ] Terms aggregation: En çok yazı olan 10 kategori (`category`)
- [ ] Terms aggregation: En çok kullanılan 20 tag
- [ ] Date histogram: Aya göre makale sayısı (son 12 ay)
- [ ] Stats aggregation: `view_count` üzerinde min/max/avg/sum
- [ ] Sub-aggregation: Her kategori içinde en popüler 3 yazar (terms içinde terms)

> **Öğrendiklerin:** Aggregation'ların gücü, nested aggregation, analytics use case'leri.

---

### Görev 8: REST API'yi Bina Et (Gin)
- [ ] Gin router kur, tüm endpoint'leri bağla (bkz. API Reference tablosu)
- [ ] Request/response için ayrı DTO struct'ları yaz — ES iç modelini dışarı sızdırma
- [ ] Error middleware — ES hatalarını kullanıcıya direkt gösterme
- [ ] `context.Context` her fonksiyona geç, timeout kullan
- [ ] Config: ES URL, index adı, port — `.env` dosyasından `godotenv` ile oku
- [ ] Query parametrelerini ES query'sine dönüştüren katmanı yaz (`q`, `category`, `from`, `size`, `sort`)

> **Öğrendiklerin:** ES'i servis katmanının arkasında konumlandırma, query param → ES query dönüşümü.

---

### Görev 9: Autocomplete / Suggestions
- [ ] **Completion suggester:** Mapping'e `suggest` adında `completion` type field ekle, bulk insert sırasında doldur
- [ ] `GET /suggest?q=` endpoint'ini completion suggester ile implement et
- [ ] **Edge n-gram analyzer:** `el`, `ela`, `elas`, `elast`... prefix'lerini index'leyen custom analyzer yaz
- [ ] Edge n-gram ile arama yapan alternatif endpoint veya flag ekle
- [ ] README'de iki yaklaşımı karşılaştır: ne zaman hangisi tercih edilir?

> **Öğrendiklerin:** Suggester API, n-gram analyzer, UX odaklı ES konfigürasyonu.

---

### Görev 10: Alias ve Reindex (Zero-Downtime Migration)
- [ ] Index'i `articles_v1` adıyla oluştur, üstüne `articles` alias'ı bağla
- [ ] Tüm kodun alias'ı (`articles`) kullansın — direkt index adı olmasın
- [ ] Bir mapping değişikliği yap (yeni field ekle veya analyzer değiştir)
- [ ] `articles_v2` index'ini yeni mapping ile oluştur
- [ ] `_reindex` API'si ile v1'den v2'ye kopyala
- [ ] Alias'ı atomik olarak v1'den çıkar, v2'ye bağla (tek istekte, downtime sıfır)
- [ ] `articles_v1`'i sil
- [ ] Bu akışı `migrate.go` script'i olarak yaz
- [ ] README'de "Zero-Downtime Migration" başlığı altında açıkla

> **Öğrendiklerin:** Alias yönetimi, reindex, zero-downtime deployment — seni junior'lar arasından ayıran konu.

---

### Bonus Görevler
- [ ] **Integration testler:** `testcontainers-go` ile gerçek ES container'ı kaldırıp test koş
- [ ] **Prometheus metrikleri:** Her endpoint için latency/error rate ölç
- [ ] **Structured logging:** `zerolog` veya `slog` ile
- [ ] **Rate limiting:** Search endpoint'ine basit bir rate limit
- [ ] **Docker image:** API'yi containerize et, `docker compose up` ile her şey tek komutla kalksın

---

## What I Learned

> Bu bölümü her görevi tamamladıkça dolduracağım. Mülakat öncesi kendimi test etmek için kullanıyorum.

1. **ES neden hızlı?** — Inverted index, BM25 skoring, segment yapısı.
2. **`text` vs `keyword` farkı?** — `text` analyze edilir (full-text search için); `keyword` olduğu gibi index'lenir (exact match, aggregation, sort için).
3. **`must` vs `filter` farkı?** — `must` relevance score hesaplar; `filter` score hesaplamaz, cache'lenir, daha hızlıdır. Filtreleme için her zaman `filter` tercih et.
4. **`match` vs `term` farkı?** — `match` analiz yapar (tokenize, lowercase vb.); `term` analiz etmez, exact value arar.
5. **Shard nedir, replica nedir?** — Shard: index'in yatay parçası, paralel yazma/okuma sağlar. Replica: shard'ın kopyası, yüksek erişilebilirlik ve okuma performansı için.
6. **Mapping değişikliği nasıl yapılır?** — Mevcut field'lar değiştirilemez. Yeni index oluştur, reindex et, alias'ı taşı.
7. **Bulk API neden hızlı?** — Network round-trip sayısını düşürür; tek HTTP isteğinde binlerce döküman index'lenir.
8. **`from+size` limiti ve alternatifi?** — Varsayılan limit 10.000. Aşılırsa ES hata verir. Alternatif: `search_after` (stateless cursor) veya `scroll` (stateful, deprecated).
9. **Alias ne işe yarar?** — Index'in üstüne sanal bir isim; zero-downtime reindex, A/B index geçişi, birden fazla index'i tek isimle sorgulama.
10. **Custom analyzer içinde neler olur?** — Tokenizer (metni parçalar) + Token filter'lar (lowercase, stop words, stemmer vb.).
11. **Aggregation nedir?** — SQL'deki GROUP BY + COUNT/SUM/AVG karşılığı; facet, istatistik, time-series analizi için kullanılır.
12. **ES birincil DB olarak kullanılır mı?** — Hayır. Durability garantisi zayıf, transaction yok, join yok. Birincil DB (PostgreSQL, MongoDB vb.) + ES ikincil (arama/analitik) mimarisi doğrudur.

---

## Kaynaklar

- [Elasticsearch Official Docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/index.html)
- [go-elasticsearch GitHub](https://github.com/elastic/go-elasticsearch)
- [Kibana Dev Tools](http://localhost:5601/app/dev_tools)

