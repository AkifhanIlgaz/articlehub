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
