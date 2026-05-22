package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/AkifhanIlgaz/articlehub/internal/model"
	"github.com/AkifhanIlgaz/articlehub/internal/repository"
	"github.com/brianvoe/gofakeit"
	"github.com/elastic/go-elasticsearch/v9"
)

func main() {
	count := flag.Int("count", 500, "number of articles to seed")
	flag.Parse()

	es, err := elasticsearch.NewTyped(elasticsearch.WithAddresses("http://localhost:9200"))
	if err != nil {
		log.Fatalf("elasticsearch: %v", err)
	}

	repo, err := repository.New(es)
	if err != nil {
		log.Fatalf("repository: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	start := time.Now()
	log.Printf("seeding %d articles...", *count)

	if err := repo.Seed(ctx, *count, fakeArticle); err != nil {
		log.Fatalf("seed: %v", err)
	}

	log.Printf("done in %s", time.Since(start))
}

func fakeArticle() model.Article {
	return model.Article{
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
