package main

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/elastic/go-elasticsearch/v9"
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

func (r *ArticleRepository) GetArticles(ctx context.Context) ([]Article, error) {
	res, err := r.es.Search().
		Index(articlesIndexName).
		Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("get articles: %w", err)
	}

	articles := make([]Article, 0, len(res.Hits.Hits))
	for _, hit := range res.Hits.Hits {
		var article Article
		if err := json.Unmarshal(hit.Source_, &article); err != nil {
			return nil, fmt.Errorf("get articles: %w", err)
		}
		articles = append(articles, article)
	}

	return articles, nil
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
