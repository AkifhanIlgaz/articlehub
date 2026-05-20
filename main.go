package main

import (
	"context"
	"fmt"
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

	for range 50 {
		article := Article{
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

		if err := repo.Index(context.Background(), article); err != nil {
			panic(err)
		}
	}

	fmt.Println("50 article indexed successfully")

	articles, err := repo.GetArticle(context.Background(), "fHHRRZ4BeHoowvibMA4y")
	if err != nil {
		panic(err)
	}
	fmt.Println("articles:", articles)
}

func connectToES() (*elasticsearch.TypedClient, error) {
	es, err := elasticsearch.NewTyped(elasticsearch.WithAddresses("http://localhost:9200"))
	if err != nil {
		return nil, err
	}

	return es, nil
}
