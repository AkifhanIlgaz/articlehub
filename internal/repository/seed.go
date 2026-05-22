package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/AkifhanIlgaz/articlehub/internal/model"
	"github.com/elastic/go-elasticsearch/v9/esutil"
)

func (r *ArticleRepository) Seed(ctx context.Context, count int, articleFn func() model.Article) error {
	var indexed, failed atomic.Uint64

	bi, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Index:         IndexName,
		Client:        r.es.Transport,
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
			article := articleFn()
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
				OnSuccess:  func(_ context.Context, _ esutil.BulkIndexerItem, _ esutil.BulkIndexerResponseItem) { indexed.Add(1) },
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
