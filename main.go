package main

import (
	"fmt"

	"github.com/elastic/go-elasticsearch/v9"
)

func main() {
	es, err := elasticsearch.New(elasticsearch.WithAddresses("http://localhost:9200"))
	if err != nil {
		panic(err)
	}

	res, err := es.Info()
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	fmt.Println(res)
}
