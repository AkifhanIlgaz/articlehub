package main

import (
	"context"
	"time"

	"github.com/brianvoe/gofakeit"
	"github.com/elastic/go-elasticsearch/v9"
)

func main() {
	es, err := connectToES()
	if err != nil {
		panic(err)
	}

	repo, err := NewArticleRepository(es)
	if err != nil {
		panic(err)
	}

	if err := repo.Seed(context.Background(), es, 50); err != nil {
		panic(err)
	}
}

func fakeArticle() Article {
	return Article{
		ID:          gofakeit.UUID(),
		Title:       gofakeit.Sentence(6),
		Slug:        gofakeit.Word() + "-" + gofakeit.Word() + "-" + gofakeit.Word(),
		Content:     gofakeit.Paragraph(3, 5, 10, " "),
		Author:      gofakeit.Name(),
		Tags:        []string{gofakeit.Word(), gofakeit.Word(), gofakeit.Word()},
		ReadingTime: gofakeit.Number(1, 15),
		ViewCount:   gofakeit.Number(0, 10000),
		LikeCount:   gofakeit.Number(0, 1000),
		PublishedAt: gofakeit.DateRange(
			time.Now().AddDate(-1, 0, 0),
			time.Now(),
		).Format(time.RFC3339),
	}
}

func connectToES() (*elasticsearch.TypedClient, error) {
	es, err := elasticsearch.NewTyped(elasticsearch.WithAddresses("http://localhost:9200"))
	if err != nil {
		return nil, err
	}

	return es, nil
}
