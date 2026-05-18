package main

import (
	"context"
	"os"

	"github.com/elastic/go-elasticsearch/v9"
)

func main() {
	es, err := connectToES()
	if err != nil {
		panic(err)
	}

	file, err := os.Open("articles_index.json")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	repo := NewArticleRepository(es)
	err = repo.EnsureIndex(context.Background(), "articles", file)
	if err != nil {
		panic(err)
	}

}

func connectToES() (*elasticsearch.Client, error) {
	es, err := elasticsearch.New(elasticsearch.WithAddresses("http://localhost:9200"))
	if err != nil {
		return nil, err
	}

	res, err := es.Info()
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	return es, nil
}
