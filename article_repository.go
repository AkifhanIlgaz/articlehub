package main

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/elastic/go-elasticsearch/v9"
	"github.com/elastic/go-elasticsearch/v9/esutil"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types"
)

//go:embed articles_index.json
var articlesIndexConfig []byte

const articlesIndexName = "articles"

type ArticleRepository struct {
	es *elasticsearch.TypedClient
}

func NewArticleRepository(es *elasticsearch.TypedClient) (*ArticleRepository, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := es.Indices.Create(articlesIndexName).
		Raw(bytes.NewReader(articlesIndexConfig)).
		Header("Content-Type", "application/json").
		Header("Accept", "application/json").
		Do(ctx)
	if err != nil {
		var esErr *types.ElasticsearchError
		if errors.As(err, &esErr) && esErr.ErrorCause.Type == "resource_already_exists_exception" {
			return &ArticleRepository{es: es}, nil
		}

		return nil, fmt.Errorf("new article repository: %w", err)
	}

	return &ArticleRepository{es: es}, nil
}

func (r *ArticleRepository) Index(ctx context.Context, article Article) error {
	_, err := r.es.Index(articlesIndexName).
		Document(article).
		Do(ctx)
	if err != nil {
		return fmt.Errorf("index article: %w", err)
	}

	return nil
}

func (r *ArticleRepository) GetArticle(ctx context.Context, id string) (*Article, error) {
	res, err := r.es.Get(articlesIndexName, id).
		Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("get article: %w", err)
	}

	var article Article
	if err := json.Unmarshal(res.Source_, &article); err != nil {
		return nil, fmt.Errorf("get article: %w", err)
	}

	return &article, nil
}

func (r *ArticleRepository) UpdateArticle(ctx context.Context, id string, article Article) error {
	_, err := r.es.Update(articlesIndexName, id).
		Doc(article).
		Do(ctx)
	if err != nil {
		return fmt.Errorf("update article: %w", err)
	}

	return nil
}

func (r *ArticleRepository) DeleteArticle(ctx context.Context, id string) error {
	_, err := r.es.Delete(articlesIndexName, id).
		Do(ctx)
	if err != nil {
		return fmt.Errorf("delete article: %w", err)
	}

	return nil
}

func (r *ArticleRepository) SearchByTitle(ctx context.Context, title string) ([]Article, error) {
	body, err := json.Marshal(map[string]any{
		"query": map[string]any{
			"match": map[string]any{
				"title": map[string]any{
					"query":     title,
					"fuzziness": "AUTO",
				},
			},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("search by title: %w", err)
	}

	res, err := r.es.Search().
		Index(articlesIndexName).
		Raw(bytes.NewReader(body)).
		Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("search by title: %w", err)
	}

	articles, err := hitsToArticles(res.Hits.Hits)
	if err != nil {
		return nil, fmt.Errorf("search by title: %w", err)
	}

	return articles, nil
}

// FullTextSearch; title, content ve tags alanlarında multi_match query ile arama yapar.
// Türkçe analyzer sayesinde kök eşleşmesi ve eş anlamlı kelime desteği sağlar.
// Sonuçları relevance skoruna göre sıralı döndürür.
func (r *ArticleRepository) FullTextSearch(ctx context.Context, query string) ([]Article, error) {
	body, err := json.Marshal(map[string]any{
		"query": map[string]any{
			"multi_match": map[string]any{
				"query":  query,
				"fields": []string{"title^3", "content", "tags^2"},
				"type":   "best_fields",
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("full text search: %w", err)
	}

	res, err := r.es.Search().
		Index(articlesIndexName).
		Raw(bytes.NewReader(body)).
		Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("full text search: %w", err)
	}

	articles, err := hitsToArticles(res.Hits.Hits)
	if err != nil {
		return nil, fmt.Errorf("search by title: %w", err)
	}

	return articles, nil
}

// GetByTags; verilen tag listesindeki en az biriyle eşleşen makaleleri döndürür (terms query).
// Tüm tag'lerle eşleşmesi isteniyorsa her tag için ayrı term query bool/must ile zincirlenir.
func (r *ArticleRepository) GetByTags(ctx context.Context, tags []string) ([]Article, error) {
	body, err := json.Marshal(map[string]any{
		"query": map[string]any{
			"terms": map[string]any{
				"tags": tags,
			},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("get by tags: %w", err)
	}

	res, err := r.es.Search().
		Index(articlesIndexName).
		Raw(bytes.NewReader(body)).
		Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("get by tags: %w", err)
	}

	articles, err := hitsToArticles(res.Hits.Hits)
	if err != nil {
		return nil, fmt.Errorf("get by tags: %w", err)
	}

	return articles, nil
}

// GetByAuthor; belirli bir yazara ait makaleleri döndürür.
// author.keyword alanı üzerinde term query kullanır; büyük/küçük harf duyarlıdır.
func (r *ArticleRepository) GetByAuthor(ctx context.Context, author string) ([]Article, error) {
	body, err := json.Marshal(map[string]any{
		"query": map[string]any{
			"match": map[string]any{
				"author": map[string]any{
					"query":     author,
					"fuzziness": "AUTO",
				},
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("get by author: %w", err)
	}

	res, err := r.es.Search().
		Index(articlesIndexName).
		Raw(bytes.NewReader(body)).
		Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("get by author: %w", err)
	}

	articles := make([]Article, 0, len(res.Hits.Hits))
	for _, hit := range res.Hits.Hits {
		var article Article
		if err := json.Unmarshal(hit.Source_, &article); err != nil {
			return nil, fmt.Errorf("get by author: %w", err)
		}
		articles = append(articles, article)
	}

	return articles, nil
}

// GetPopular; view_count ve like_count'a göre function_score veya basit sort ile en popüler
// makaleleri döndürür. limit parametresi kaç makale getirileceğini belirler.
func (r *ArticleRepository) GetPopular(ctx context.Context, limit int) ([]Article, error) {
	body, err := json.Marshal(map[string]any{
		"query": map[string]any{
			"function_score": map[string]any{
				"query": map[string]any{
					"match_all": map[string]any{},
				},
				"functions": []map[string]any{
					{
						"field_value_factor": map[string]any{
							"field":    "view_count",
							"factor":   1.0,
							"modifier": "log1p",
						},
					},
					{
						"field_value_factor": map[string]any{
							"field":    "like_count",
							"factor":   1.0,
							"modifier": "log1p",
						},
					},
				},
				"score_mode": "sum",
				"boost_mode": "replace",
			},
		},
		"size": limit,
	})
	if err != nil {
		return nil, fmt.Errorf("get by author: %w", err)
	}

	res, err := r.es.Search().
		Index(articlesIndexName).
		Raw(bytes.NewReader(body)).
		Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("get by author: %w", err)
	}

	articles := make([]Article, 0, len(res.Hits.Hits))
	for _, hit := range res.Hits.Hits {
		var article Article
		if err := json.Unmarshal(hit.Source_, &article); err != nil {
			return nil, fmt.Errorf("get by author: %w", err)
		}
		articles = append(articles, article)
	}

	return articles, nil
}

// GetRecent; published_at alanına göre azalan sırada en yeni makaleleri döndürür.
// limit parametresi kaç makale getirileceğini belirler.
func (r *ArticleRepository) GetRecent(ctx context.Context, limit int) ([]Article, error) {
	body, err := json.Marshal(map[string]any{
		"query": map[string]any{
			"match_all": map[string]any{},
		},
		"sort": []map[string]any{
			{"published_at": map[string]any{"order": "desc"}},
		},
		"size": limit,
	})
	if err != nil {
		return nil, fmt.Errorf("get recent: %w", err)
	}

	res, err := r.es.Search().
		Index(articlesIndexName).
		Raw(bytes.NewReader(body)).
		Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("get recent: %w", err)
	}

	articles, err := hitsToArticles(res.Hits.Hits)
	if err != nil {
		return nil, fmt.Errorf("get recent: %w", err)
	}

	return articles, nil
}

// GetByDateRange; published_at alanı from ile to arasında olan makaleleri döndürür.
// Tarihler RFC3339/ISO8601 formatında olmalıdır; range query kullanılır.
func (r *ArticleRepository) GetByDateRange(ctx context.Context, from, to string) ([]Article, error) {
	body, err := json.Marshal(map[string]any{
		"query": map[string]any{
			"bool": map[string]any{
				"filter": map[string]any{
					"range": map[string]any{
						"published_at": map[string]any{
							"gte": from,
							"lte": to,
						},
					},
				},
			},
		},
		"sort": []map[string]any{
			{"published_at": map[string]any{"order": "desc"}},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("get by date range: %w", err)
	}

	res, err := r.es.Search().
		Index(articlesIndexName).
		Raw(bytes.NewReader(body)).
		Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("get by date range: %w", err)
	}

	articles, err := hitsToArticles(res.Hits.Hits)
	if err != nil {
		return nil, fmt.Errorf("get by date range: %w", err)
	}

	return articles, nil
}

// GetByReadingTime; okuma süresi minMin ile maxMin dakika arasındaki makaleleri döndürür.
// Kısa içerik (<5 dk) veya uzun okuma (>15 dk) filtreleri için kullanışlıdır.
//
// Query: bool/filter içinde range — reading_time integer field, skor hesaplamaya gerek yok.
func (r *ArticleRepository) GetByReadingTime(ctx context.Context, minMin, maxMin int) ([]Article, error) {
	body, err := json.Marshal(map[string]any{
		"query": map[string]any{
			"bool": map[string]any{
				"filter": map[string]any{
					"range": map[string]any{
						"reading_time": map[string]any{
							"gte": minMin,
							"lte": maxMin,
						},
					},
				},
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("get by reading time: %w", err)
	}

	res, err := r.es.Search().
		Index(articlesIndexName).
		Raw(bytes.NewReader(body)).
		Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("get by reading time: %w", err)
	}

	articles, err := hitsToArticles(res.Hits.Hits)
	if err != nil {
		return nil, fmt.Errorf("get by reading time: %w", err)
	}

	return articles, nil
}

// SuggestTitle; title.suggest alanı üzerinde edge n-gram tabanlı autocomplete döndürür.
// Kullanıcı yazmaya devam ederken arama kutusuna anlık öneri sunmak için kullanılır.
//
// Query: title.suggest field'ına match — index anında "ela","elas","elast"... token'ları üretilmiş,
// search anında sadece lowercase+ascii yapılır. "ela" yazınca "Elasticsearch" bulunur.
func (r *ArticleRepository) SuggestTitle(ctx context.Context, prefix string, size int) ([]string, error) {
	body, err := json.Marshal(map[string]any{
		"query": map[string]any{
			"match": map[string]any{
				"title.suggest": prefix,
			},
		},
		"_source": []string{"title"},
		"size":    size,
	})
	if err != nil {
		return nil, fmt.Errorf("suggest title: %w", err)
	}

	res, err := r.es.Search().
		Index(articlesIndexName).
		Raw(bytes.NewReader(body)).
		Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("suggest title: %w", err)
	}

	titles := make([]string, 0, len(res.Hits.Hits))
	for _, hit := range res.Hits.Hits {
		var a Article
		if err := json.Unmarshal(hit.Source_, &a); err != nil {
			return nil, fmt.Errorf("suggest title: %w", err)
		}
		titles = append(titles, a.Title)
	}

	return titles, nil
}

// GetRelated; verilen makaleyle aynı tag'leri paylaşan diğer makaleleri döndürür.
// Mevcut makale sonuçlardan hariç tutulur; terms query veya more_like_this kullanılabilir.
//
// Query: bool/filter içinde iki koşul:
//   - terms: verilen tag'lerden en az biri eşleşmeli
//   - must_not/term: articleID'ye sahip makale hariç tutulur
func (r *ArticleRepository) GetRelated(ctx context.Context, articleID string, tags []string, limit int) ([]Article, error) {
	body, err := json.Marshal(map[string]any{
		"query": map[string]any{
			"bool": map[string]any{
				"filter": map[string]any{
					"terms": map[string]any{"tags": tags},
				},
				"must_not": map[string]any{
					"term": map[string]any{"_id": articleID},
				},
			},
		},
		"size": limit,
	})
	if err != nil {
		return nil, fmt.Errorf("get related: %w", err)
	}

	res, err := r.es.Search().
		Index(articlesIndexName).
		Raw(bytes.NewReader(body)).
		Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("get related: %w", err)
	}

	articles, err := hitsToArticles(res.Hits.Hits)
	if err != nil {
		return nil, fmt.Errorf("get related: %w", err)
	}

	return articles, nil
}

// IncrementViewCount; belirli bir makalenin view_count alanını script update ile atomik olarak 1 artırır.
// Tüm belgeyi yeniden index'lemek yerine update API'sinin painless script özelliği kullanılır.
//
// Query: Update API + inline painless script — ctx.source ile dokümanın mevcut değeri okunur,
// +1 eklenir. Tüm dokümanı GET edip tekrar PUT etmeye gerek kalmaz.
func (r *ArticleRepository) IncrementViewCount(ctx context.Context, id string) error {
	body, err := json.Marshal(map[string]any{
		"script": map[string]any{
			"source": "ctx._source.view_count++",
			"lang":   "painless",
		},
	})
	if err != nil {
		return fmt.Errorf("increment view count: %w", err)
	}

	_, err = r.es.Update(articlesIndexName, id).
		Raw(bytes.NewReader(body)).
		Do(ctx)
	if err != nil {
		return fmt.Errorf("increment view count: %w", err)
	}

	return nil
}

// IncrementLikeCount; belirli bir makalenin like_count alanını script update ile atomik olarak 1 artırır.
// IncrementViewCount ile aynı yaklaşımı kullanır; gerekirse azaltma için delta parametresi eklenebilir.
func (r *ArticleRepository) IncrementLikeCount(ctx context.Context, id string) error {
	body, err := json.Marshal(map[string]any{
		"script": map[string]any{
			"source": "ctx._source.like_count++",
			"lang":   "painless",
		},
	})
	if err != nil {
		return fmt.Errorf("increment like count: %w", err)
	}

	_, err = r.es.Update(articlesIndexName, id).
		Raw(bytes.NewReader(body)).
		Do(ctx)
	if err != nil {
		return fmt.Errorf("increment like count: %w", err)
	}

	return nil
}

// ListPaginated; makaleleri from/size parametreleriyle sayfalı olarak döndürür.
// sortField ("published_at", "view_count", "like_count") ve sortOrder ("asc"/"desc") ile sıralama yapılır.
//
// Query: match_all + sort + from/size. Toplam hit sayısı (int) da döner — UI'da "X sonuç bulundu"
// göstermek için kullanılır. from+size > 10.000 ise ES hata verir; derin sayfalama için search_after tercih edilir.
func (r *ArticleRepository) ListPaginated(ctx context.Context, from, size int, sortField, sortOrder string) ([]Article, int, error) {
	body, err := json.Marshal(map[string]any{
		"query": map[string]any{
			"match_all": map[string]any{},
		},
		"sort": []map[string]any{
			{sortField: map[string]any{"order": sortOrder}},
		},
		"from": from,
		"size": size,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list paginated: %w", err)
	}

	res, err := r.es.Search().
		Index(articlesIndexName).
		Raw(bytes.NewReader(body)).
		Do(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("list paginated: %w", err)
	}

	articles, err := hitsToArticles(res.Hits.Hits)
	if err != nil {
		return nil, 0, fmt.Errorf("list paginated: %w", err)
	}

	total := int(res.Hits.Total.Value)

	return articles, total, nil
}

func (r *ArticleRepository) Seed(ctx context.Context, es *elasticsearch.TypedClient, count int) error {
	var (
		indexed atomic.Uint64
		failed  atomic.Uint64
	)

	bi, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Index:         articlesIndexName,
		Client:        es.Transport,
		NumWorkers:    runtime.NumCPU(),
		FlushBytes:    5e6,
		FlushInterval: 5 * time.Second,
	})
	if err != nil {
		return fmt.Errorf("seed: %w", err)
	}

	var wg sync.WaitGroup
	for range count {
		wg.Add(1)
		go func() {
			defer wg.Done()

			article := fakeArticle()
			data, err := json.Marshal(article)
			if err != nil {
				log.Printf("seed: marshal error: %v", err)
				failed.Add(1)
				return
			}

			_ = bi.Add(ctx, esutil.BulkIndexerItem{
				Action:     "index",
				DocumentID: article.ID,
				Body:       bytes.NewReader(data),
				OnSuccess: func(_ context.Context, _ esutil.BulkIndexerItem, _ esutil.BulkIndexerResponseItem) {
					indexed.Add(1)
				},
				OnFailure: func(_ context.Context, _ esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem, err error) {
					failed.Add(1)
					if err != nil {
						log.Printf("seed: %v", err)
					} else {
						log.Printf("seed: %s %s", res.Error.Type, res.Error.Reason)
					}
				},
			})
		}()
	}

	wg.Wait()

	if err := bi.Close(ctx); err != nil {
		return fmt.Errorf("seed: flush: %w", err)
	}

	fmt.Printf("seed: indexed=%d failed=%d\n", indexed.Load(), failed.Load())
	return nil
}

func hitsToArticles(hits []types.Hit) ([]Article, error) {
	articles := make([]Article, len(hits))
	for i, hit := range hits {
		if err := json.Unmarshal(hit.Source_, &articles[i]); err != nil {
			return nil, err
		}
	}
	return articles, nil
}
