package repository

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/AkifhanIlgaz/articlehub/internal/model"
	"github.com/elastic/go-elasticsearch/v9"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types"
)

//go:embed articles_index.json
var articlesIndexConfig []byte

const IndexName = "articles"

type ArticleRepository struct {
	es *elasticsearch.TypedClient
}

func New(es *elasticsearch.TypedClient) (*ArticleRepository, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := es.Indices.Create(IndexName).
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

func hitsToArticles(hits []types.Hit) ([]model.Article, error) {
	articles := make([]model.Article, len(hits))
	for i, hit := range hits {
		if err := json.Unmarshal(hit.Source_, &articles[i]); err != nil {
			return nil, err
		}
	}
	return articles, nil
}
