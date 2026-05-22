package indexer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"

	"github.com/elastic/go-elasticsearch/v9"
)

type Indexer struct {
	client *elasticsearch.Client
}

func New(client *elasticsearch.Client) *Indexer {
	return &Indexer{client: client}
}

// EnsureAliasedIndex; alias yoksa index'i oluşturur ve alias'ı bağlar. Idempotent'tir.
func (i *Indexer) EnsureAliasedIndex(ctx context.Context, alias, index string, indexFile io.Reader) error {
	res, err := i.client.Indices.ExistsAlias(
		[]string{alias},
		i.client.Indices.ExistsAlias.WithContext(ctx),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode == 200 {
		return nil
	}

	if err := i.CreateIndex(ctx, index, indexFile); err != nil {
		return err
	}
	return i.AddAlias(ctx, index, alias)
}

// CreateIndex; verilen mapping ile yeni bir index oluşturur.
func (i *Indexer) CreateIndex(ctx context.Context, name string, body io.Reader) error {
	res, err := i.client.Indices.Create(
		name,
		i.client.Indices.Create.WithContext(ctx),
		i.client.Indices.Create.WithBody(body),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("create index %s: %s", name, res.String())
	}
	return nil
}

// DeleteIndex; bir index'i siler. Migrate akışında eski index temizlenmek için kullanılır.
func (i *Indexer) DeleteIndex(ctx context.Context, name string) error {
	res, err := i.client.Indices.Delete(
		[]string{name},
		i.client.Indices.Delete.WithContext(ctx),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("delete index %s: %s", name, res.String())
	}
	return nil
}

// AddAlias; bir index'e alias ekler.
func (i *Indexer) AddAlias(ctx context.Context, index, alias string) error {
	res, err := i.client.Indices.PutAlias(
		[]string{index}, alias,
		i.client.Indices.PutAlias.WithContext(ctx),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("put alias: %s", res.String())
	}
	return nil
}

// Reindex; kaynak index'ten hedef index'e tüm dökümanları kopyalar.
func (i *Indexer) Reindex(ctx context.Context, from, to string) error {
	buf, _ := json.Marshal(map[string]any{
		"source": map[string]any{"index": from, "size": 1000},
		"dest":   map[string]any{"index": to},
	})

	res, err := i.client.Reindex(
		bytes.NewReader(buf),
		i.client.Reindex.WithContext(ctx),
		i.client.Reindex.WithWaitForCompletion(true),
		i.client.Reindex.WithRefresh(true),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("reindex %s→%s: %s", from, to, res.String())
	}

	var out struct {
		Took     int   `json:"took"`
		Created  int   `json:"created"`
		Failures []any `json:"failures"`
	}
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return err
	}
	if len(out.Failures) > 0 {
		return fmt.Errorf("reindex failures: %v", out.Failures)
	}
	log.Printf("reindex: %d docs, %dms", out.Created, out.Took)
	return nil
}

// SwapAlias; alias'ı oldIndex'ten newIndex'e atomik olarak taşır (zero-downtime).
// Tek bir _aliases isteğiyle remove + add işlemi aynı anda gerçekleşir.
func (i *Indexer) SwapAlias(ctx context.Context, alias, oldIndex, newIndex string) error {
	buf, _ := json.Marshal(map[string]any{
		"actions": []map[string]any{
			{"remove": map[string]any{"index": oldIndex, "alias": alias}},
			{"add": map[string]any{"index": newIndex, "alias": alias}},
		},
	})

	res, err := i.client.Indices.UpdateAliases(
		bytes.NewReader(buf),
		i.client.Indices.UpdateAliases.WithContext(ctx),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("swap alias: %s", res.String())
	}
	return nil
}
