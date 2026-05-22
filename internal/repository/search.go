package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/AkifhanIlgaz/articlehub/internal/model"
)

func (r *ArticleRepository) SearchByTitle(ctx context.Context, title string) ([]model.Article, error) {
	body, err := json.Marshal(map[string]any{
		"query": map[string]any{
			"match": map[string]any{"title": map[string]any{"query": title}},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("search by title: %w", err)
	}
	res, err := r.es.Search().Index(IndexName).Raw(bytes.NewReader(body)).Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("search by title: %w", err)
	}
	return hitsToArticles(res.Hits.Hits)
}

// FullTextSearch; title, content ve tags alanlarında multi_match query ile arama yapar.
// Türkçe analyzer sayesinde kök eşleşmesi ve eş anlamlı kelime desteği sağlar.
func (r *ArticleRepository) FullTextSearch(ctx context.Context, query string) ([]model.Article, error) {
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
	res, err := r.es.Search().Index(IndexName).Raw(bytes.NewReader(body)).Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("full text search: %w", err)
	}
	return hitsToArticles(res.Hits.Hits)
}

// GetByTags; verilen tag listesindeki en az biriyle eşleşen makaleleri döndürür (terms query).
func (r *ArticleRepository) GetByTags(ctx context.Context, tags []string) ([]model.Article, error) {
	body, err := json.Marshal(map[string]any{
		"query": map[string]any{"terms": map[string]any{"tags": tags}},
	})
	if err != nil {
		return nil, fmt.Errorf("get by tags: %w", err)
	}
	res, err := r.es.Search().Index(IndexName).Raw(bytes.NewReader(body)).Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("get by tags: %w", err)
	}
	return hitsToArticles(res.Hits.Hits)
}

// GetByAuthor; belirli bir yazara ait makaleleri fuzzy match ile döndürür.
func (r *ArticleRepository) GetByAuthor(ctx context.Context, author string) ([]model.Article, error) {
	body, err := json.Marshal(map[string]any{
		"query": map[string]any{
			"match": map[string]any{"author": map[string]any{"query": author, "fuzziness": "AUTO"}},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("get by author: %w", err)
	}
	res, err := r.es.Search().Index(IndexName).Raw(bytes.NewReader(body)).Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("get by author: %w", err)
	}
	return hitsToArticles(res.Hits.Hits)
}

// GetPopular; view_count ve like_count'a göre function_score ile en popüler makaleleri döndürür.
func (r *ArticleRepository) GetPopular(ctx context.Context, limit int) ([]model.Article, error) {
	body, err := json.Marshal(map[string]any{
		"query": map[string]any{
			"function_score": map[string]any{
				"query": map[string]any{"match_all": map[string]any{}},
				"functions": []map[string]any{
					{"field_value_factor": map[string]any{"field": "view_count", "factor": 1.0, "modifier": "log1p"}},
					{"field_value_factor": map[string]any{"field": "like_count", "factor": 1.0, "modifier": "log1p"}},
				},
				"score_mode": "sum",
				"boost_mode": "replace",
			},
		},
		"size": limit,
	})
	if err != nil {
		return nil, fmt.Errorf("get popular: %w", err)
	}
	res, err := r.es.Search().Index(IndexName).Raw(bytes.NewReader(body)).Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("get popular: %w", err)
	}
	return hitsToArticles(res.Hits.Hits)
}

// GetRecent; published_at alanına göre azalan sırada en yeni makaleleri döndürür.
func (r *ArticleRepository) GetRecent(ctx context.Context, limit int) ([]model.Article, error) {
	body, err := json.Marshal(map[string]any{
		"query": map[string]any{"match_all": map[string]any{}},
		"sort":  []map[string]any{{"published_at": map[string]any{"order": "desc"}}},
		"size":  limit,
	})
	if err != nil {
		return nil, fmt.Errorf("get recent: %w", err)
	}
	res, err := r.es.Search().Index(IndexName).Raw(bytes.NewReader(body)).Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("get recent: %w", err)
	}
	return hitsToArticles(res.Hits.Hits)
}

// GetByDateRange; published_at alanı from ile to arasında olan makaleleri döndürür.
func (r *ArticleRepository) GetByDateRange(ctx context.Context, from, to string) ([]model.Article, error) {
	body, err := json.Marshal(map[string]any{
		"query": map[string]any{
			"bool": map[string]any{
				"filter": map[string]any{
					"range": map[string]any{"published_at": map[string]any{"gte": from, "lte": to}},
				},
			},
		},
		"sort": []map[string]any{{"published_at": map[string]any{"order": "desc"}}},
	})
	if err != nil {
		return nil, fmt.Errorf("get by date range: %w", err)
	}
	res, err := r.es.Search().Index(IndexName).Raw(bytes.NewReader(body)).Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("get by date range: %w", err)
	}
	return hitsToArticles(res.Hits.Hits)
}

// GetByReadingTime; okuma süresi minMin ile maxMin dakika arasındaki makaleleri döndürür.
func (r *ArticleRepository) GetByReadingTime(ctx context.Context, minMin, maxMin int) ([]model.Article, error) {
	body, err := json.Marshal(map[string]any{
		"query": map[string]any{
			"bool": map[string]any{
				"filter": map[string]any{
					"range": map[string]any{"reading_time": map[string]any{"gte": minMin, "lte": maxMin}},
				},
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("get by reading time: %w", err)
	}
	res, err := r.es.Search().Index(IndexName).Raw(bytes.NewReader(body)).Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("get by reading time: %w", err)
	}
	return hitsToArticles(res.Hits.Hits)
}

// GetRelated; verilen makaleyle aynı tag'leri paylaşan diğer makaleleri döndürür.
func (r *ArticleRepository) GetRelated(ctx context.Context, articleID string, tags []string, limit int) ([]model.Article, error) {
	body, err := json.Marshal(map[string]any{
		"query": map[string]any{
			"bool": map[string]any{
				"filter":   map[string]any{"terms": map[string]any{"tags": tags}},
				"must_not": map[string]any{"term": map[string]any{"_id": articleID}},
			},
		},
		"size": limit,
	})
	if err != nil {
		return nil, fmt.Errorf("get related: %w", err)
	}
	res, err := r.es.Search().Index(IndexName).Raw(bytes.NewReader(body)).Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("get related: %w", err)
	}
	return hitsToArticles(res.Hits.Hits)
}

// ListPaginated; makaleleri from/size parametreleriyle sayfalı olarak döndürür.
// from+size > 10.000 ise ES hata verir; derin sayfalama için SearchAfter tercih edilir.
func (r *ArticleRepository) ListPaginated(ctx context.Context, from, size int, sortField, sortOrder string) ([]model.Article, int, error) {
	body, err := json.Marshal(map[string]any{
		"query": map[string]any{"match_all": map[string]any{}},
		"sort":  []map[string]any{{sortField: map[string]any{"order": sortOrder}}},
		"from":  from,
		"size":  size,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list paginated: %w", err)
	}
	res, err := r.es.Search().Index(IndexName).Raw(bytes.NewReader(body)).Do(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("list paginated: %w", err)
	}
	articles, err := hitsToArticles(res.Hits.Hits)
	if err != nil {
		return nil, 0, err
	}
	return articles, int(res.Hits.Total.Value), nil
}

// SearchByCategoryAndDateRange; belirtilen kategorideki ve tarih aralığındaki makaleleri döndürür.
// bool/filter içinde term (category) + range (published_at) kombinasyonu kullanılır.
func (r *ArticleRepository) SearchByCategoryAndDateRange(ctx context.Context, category string, from, to time.Time) ([]model.Article, error) {
	panic("not implemented")
}

// FuzzySearch; yazım hatalarına toleranslı arama yapar.
// fuzziness "AUTO" ile "elasitc" yazınca "elastic" bulunur.
func (r *ArticleRepository) FuzzySearch(ctx context.Context, q string) ([]model.Article, error) {
	panic("not implemented")
}

// SearchAfter; derin sayfalama için search_after token'ını kullanır.
// from+size > 10.000 limitini aşmaz; bir önceki sayfanın son hit sort değerleri cursor olarak verilir.
// Dönen []any bir sonraki sayfanın searchAfter parametresi olarak kullanılır.
func (r *ArticleRepository) SearchAfter(ctx context.Context, size int, sortField, sortOrder string, searchAfter []any) ([]model.Article, []any, error) {
	panic("not implemented")
}

// SearchWithHighlight; q ile eşleşen term'leri <em>...</em> ile sararak döndürür.
func (r *ArticleRepository) SearchWithHighlight(ctx context.Context, q string) ([]model.HighlightedArticle, error) {
	panic("not implemented")
}

// SearchWithFacets; arama sonuçlarıyla birlikte kategori/tag/tarih/istatistik aggregation'larını döndürür.
func (r *ArticleRepository) SearchWithFacets(ctx context.Context, q string) (*model.FacetResult, error) {
	panic("not implemented")
}
