package main

import (
	"context"
	"flag"
	"log"
	"os"
	"time"

	"github.com/AkifhanIlgaz/articlehub/internal/indexer"
	"github.com/elastic/go-elasticsearch/v9"
)

// Zero-downtime migration akışı:
//  1. articles_v2 oluştur (yeni mapping ile)
//  2. _reindex: v1 → v2
//  3. Alias'ı atomik olarak v1'den çıkar, v2'ye bağla
//  4. articles_v1'i sil

func main() {
	newMapping := flag.String("mapping", "articles_index_v2.json", "path to new index mapping JSON")
	flag.Parse()

	es, err := elasticsearch.New(elasticsearch.WithAddresses("http://localhost:9200"))
	if err != nil {
		log.Fatalf("elasticsearch: %v", err)
	}

	idx := indexer.New(es)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	if err := migrate(ctx, idx, *newMapping); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	log.Println("migration complete — alias now points to articles_v2")
}

func migrate(ctx context.Context, idx *indexer.Indexer, mappingFile string) error {
	const (
		alias  = "articles"
		oldIdx = "articles_v1"
		newIdx = "articles_v2"
	)

	f, err := os.Open(mappingFile)
	if err != nil {
		return err
	}
	defer f.Close()

	log.Printf("step 1: creating %s...", newIdx)
	if err := idx.CreateIndex(ctx, newIdx, f); err != nil {
		return err
	}

	log.Printf("step 2: reindexing %s → %s...", oldIdx, newIdx)
	if err := idx.Reindex(ctx, oldIdx, newIdx); err != nil {
		return err
	}

	log.Printf("step 3: swapping alias %q: %s → %s...", alias, oldIdx, newIdx)
	if err := idx.SwapAlias(ctx, alias, oldIdx, newIdx); err != nil {
		return err
	}

	log.Printf("step 4: deleting %s...", oldIdx)
	if err := idx.DeleteIndex(ctx, oldIdx); err != nil {
		return err
	}

	return nil
}
