package main

import (
	"context"
	"fmt"
	"io"

	"github.com/elastic/go-elasticsearch/v9"
)

type ArticleRepository struct {
	client *elasticsearch.Client
}

func NewArticleRepository(client *elasticsearch.Client) *ArticleRepository {
	return &ArticleRepository{client: client}
}

func (r *ArticleRepository) EnsureIndex(ctx context.Context, name string, body io.Reader) error {
	exists, err := r.client.Indices.Exists([]string{name})
	if err != nil {
		return err
	}
	if exists.StatusCode == 200 {
		del, err := r.client.Indices.Delete([]string{name})
		if err != nil {
			return err
		}
		defer del.Body.Close()
		if del.IsError() {
			return fmt.Errorf("%s", del.String())
		}
	}

	res, err := r.client.Indices.Create(
		name,
		r.client.Indices.Create.WithContext(ctx),
		r.client.Indices.Create.WithBody(body),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("%s", res.String())
	}

	return nil
}

func (r *ArticleRepository) IndexArticle(ctx context.Context, article Article) error {
	panic("not implemented")
}

func (r *ArticleRepository) GetArticle(ctx context.Context, id string) (*Article, error) {
	panic("not implemented")
}

func (r *ArticleRepository) UpdateArticle(ctx context.Context, id string, article Article) error {
	panic("not implemented")
}

func (r *ArticleRepository) DeleteArticle(ctx context.Context, id string) error {
	panic("not implemented")
}
