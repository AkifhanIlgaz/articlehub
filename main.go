package main

import (
	"context"
	"encoding/json"
	"os"
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

	articles, err := repo.GetPopular(context.Background(), 10)
	if err != nil {
		panic(err)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(articles); err != nil {
		panic(err)
	}
}

func fakeArticle() Article {
	return Article{
		ID:          gofakeit.UUID(),
		Title:       gofakeit.HipsterSentence(6),
		Slug:        gofakeit.HipsterWord() + "-" + gofakeit.HipsterWord() + "-" + gofakeit.HipsterWord(),
		Content:     gofakeit.HipsterParagraph(3, 5, 10, " "),
		Author:      gofakeit.Name(),
		Tags:        []string{gofakeit.HipsterWord(), gofakeit.HipsterWord(), gofakeit.HipsterWord()},
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
