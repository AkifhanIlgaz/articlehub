package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/AkifhanIlgaz/articlehub/internal/model"
)

// SuggestTitle; title.suggest alanı üzerinde edge n-gram tabanlı autocomplete döndürür.
// Kullanıcı yazmaya devam ederken anlık öneri sunmak için kullanılır.
func (r *ArticleRepository) SuggestTitle(ctx context.Context, prefix string, size int) ([]string, error) {
	body, err := json.Marshal(map[string]any{
		"query":   map[string]any{"match": map[string]any{"title.suggest": prefix}},
		"_source": []string{"title"},
		"size":    size,
	})
	if err != nil {
		return nil, fmt.Errorf("suggest title: %w", err)
	}

	res, err := r.es.Search().Index(IndexName).Raw(bytes.NewReader(body)).Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("suggest title: %w", err)
	}

	titles := make([]string, 0, len(res.Hits.Hits))
	for _, hit := range res.Hits.Hits {
		var a model.Article
		if err := json.Unmarshal(hit.Source_, &a); err != nil {
			return nil, fmt.Errorf("suggest title: %w", err)
		}
		titles = append(titles, a.Title)
	}
	return titles, nil
}

// SuggestCompletion; completion suggester field üzerinde prefix tabanlı öneri döndürür.
// Edge n-gram'dan farkı: FST yapısı kullanır, daha hızlıdır ancak mapping'e özel completion type gerektirir.
func (r *ArticleRepository) SuggestCompletion(ctx context.Context, prefix string, size int) ([]string, error) {
	panic("not implemented")
}
