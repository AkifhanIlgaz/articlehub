package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/AkifhanIlgaz/articlehub/internal/model"
)

func (r *ArticleRepository) Index(ctx context.Context, article model.Article) error {
	_, err := r.es.Index(IndexName).Document(article).Do(ctx)
	if err != nil {
		return fmt.Errorf("index article: %w", err)
	}
	return nil
}

func (r *ArticleRepository) GetArticle(ctx context.Context, id string) (*model.Article, error) {
	res, err := r.es.Get(IndexName, id).Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("get article: %w", err)
	}

	var article model.Article
	if err := json.Unmarshal(res.Source_, &article); err != nil {
		return nil, fmt.Errorf("get article: %w", err)
	}
	return &article, nil
}

// ReplaceArticle; dökümanın tamamını Index API ile yeniden yazar.
// Gönderilmeyen alanlar silinir. PUT /articles/:id semantiği.
func (r *ArticleRepository) ReplaceArticle(ctx context.Context, id string, article model.Article) error {
	_, err := r.es.Index(IndexName).Id(id).Document(article).Do(ctx)
	if err != nil {
		return fmt.Errorf("replace article: %w", err)
	}
	return nil
}

// PatchArticle; Update API'nin doc özelliğiyle sadece gönderilen alanları günceller.
// Gönderilmeyen alanlar mevcut değerlerini korur. PATCH /articles/:id semantiği.
func (r *ArticleRepository) PatchArticle(ctx context.Context, id string, fields map[string]any) error {
	_, err := r.es.Update(IndexName, id).Doc(fields).Do(ctx)
	if err != nil {
		return fmt.Errorf("patch article: %w", err)
	}
	return nil
}

func (r *ArticleRepository) DeleteArticle(ctx context.Context, id string) error {
	_, err := r.es.Delete(IndexName, id).Do(ctx)
	if err != nil {
		return fmt.Errorf("delete article: %w", err)
	}
	return nil
}

// IncrementViewCount; view_count alanını painless script ile atomik olarak 1 artırır.
func (r *ArticleRepository) IncrementViewCount(ctx context.Context, id string) error {
	return r.scriptUpdate(ctx, id, "ctx._source.view_count++")
}

// IncrementLikeCount; like_count alanını painless script ile atomik olarak 1 artırır.
func (r *ArticleRepository) IncrementLikeCount(ctx context.Context, id string) error {
	return r.scriptUpdate(ctx, id, "ctx._source.like_count++")
}

func (r *ArticleRepository) scriptUpdate(ctx context.Context, id, source string) error {
	body, err := json.Marshal(map[string]any{
		"script": map[string]any{"source": source, "lang": "painless"},
	})
	if err != nil {
		return fmt.Errorf("script update: %w", err)
	}
	_, err = r.es.Update(IndexName, id).Raw(bytes.NewReader(body)).Do(ctx)
	if err != nil {
		return fmt.Errorf("script update: %w", err)
	}
	return nil
}
