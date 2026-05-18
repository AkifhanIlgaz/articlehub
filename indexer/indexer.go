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

const articlesAlias = "articles"
const articlesInitialIndex = "articles_v1"

type Indexer struct {
	client *elasticsearch.Client
}

func NewIndexer(client *elasticsearch.Client) *Indexer {
	return &Indexer{client: client}
}

func (i *Indexer) EnsureAliasedIndex(ctx context.Context, alias string, index string, indexFile io.Reader) error {
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

	createRes, err := i.client.Indices.Create(
		index,
		i.client.Indices.Create.WithContext(ctx),
		i.client.Indices.Create.WithBody(indexFile),
	)
	if err != nil {
		return err
	}

	defer createRes.Body.Close()

	if createRes.IsError() {
		return fmt.Errorf("create %s: %s", index, createRes.String())
	}

	return i.addAlias(ctx, index, alias)
}

func (i *Indexer) addAlias(ctx context.Context, index, alias string) error {
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

func (i *Indexer) createIndex(ctx context.Context, name string, body io.Reader) error {
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
		return fmt.Errorf("%s", res.String())
	}

	return nil
}

func (i *Indexer) reindex(ctx context.Context, from, to string) (string, error) {
	body := map[string]any{
		"source": map[string]any{
			"index": from,
			"size":  1000, // batch boyutu
		},
		"dest": map[string]any{
			"index": to,
		},
	}
	buf, _ := json.Marshal(body)

	res, err := i.client.Reindex(
		bytes.NewReader(buf),
		i.client.Reindex.WithContext(ctx),
		i.client.Reindex.WithWaitForCompletion(true), // küçük indeksler için OK
		i.client.Reindex.WithRefresh(true),
	)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.IsError() {
		return "", fmt.Errorf("%s", res.String())
	}

	var out struct {
		Task     string `json:"task"`
		Took     int    `json:"took"`
		Total    int    `json:"total"`
		Created  int    `json:"created"`
		Failures []any  `json:"failures"`
	}
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return "", err
	}
	if len(out.Failures) > 0 {
		return "", fmt.Errorf("reindex failures: %v", out.Failures)
	}
	log.Printf("reindex: %d doc, %dms", out.Created, out.Took)
	return out.Task, nil
}

func (i *Indexer) swapAlias(ctx context.Context, alias, oldIndex, newIndex string) error {
	body := map[string]any{
		"actions": []map[string]any{
			{"remove": map[string]any{"index": oldIndex, "alias": alias}},
			{"add": map[string]any{"index": newIndex, "alias": alias}},
		},
	}
	buf, _ := json.Marshal(body)

	res, err := i.client.Indices.UpdateAliases(
		bytes.NewReader(buf),
		i.client.Indices.UpdateAliases.WithContext(ctx),
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
