package v7

import (
	"context"
	"net/http"

	"github.com/arquivei/elasticutil/official/v7/internal/retrier"
	es "github.com/elastic/go-elasticsearch/v8"
)

// Client represents the Elasticsearch's client of the official lib.
type Client interface {
	Search(context.Context, SearchConfig) (SearchResponse, error)
}

type esClient struct {
	client *es.Client
}

// NewClient returns a new Client using the @urls.
func NewClient(urls ...string) (Client, error) {
	client, err := es.NewClient(es.Config{
		Addresses:    urls,
		RetryBackoff: retrier.NewSimpleBackoff(10, 100),
		Transport: &http.Transport{
			DisableCompression: false,
		},
	})
	if err != nil {
		return nil, err
	}
	return &esClient{
		client: client,
	}, nil
}

// MustNewClient returns a new Client using the @urls. It panics
// instead of returning an error.
func MustNewClient(urls ...string) Client {
	client, err := NewClient(urls...)
	if err != nil {
		panic(err)
	}
	return client
}
