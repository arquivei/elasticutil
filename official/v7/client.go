package v7

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net/http"

	"github.com/arquivei/elasticutil/official/v7/internal/retrier"
	es "github.com/elastic/go-elasticsearch/v7"
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

// NewClientWithAuth returns a new Client using the @urls and some auth parameters.
// If @certPem is nil, it will try to create a client without authentication.
func NewClientWithAuth(urls []string, certPem []byte, username, password string) (Client, error) {
	client, err := es.NewClient(getElasticAuthConfig(urls, certPem, username, password))
	if err != nil {
		return nil, err
	}

	return &esClient{
		client: client,
	}, nil
}

func getElasticAuthConfig(urls []string, certPem []byte, username, password string) es.Config {
	esConfig := es.Config{
		Addresses:    urls,
		Username:     username,
		Password:     password,
		RetryBackoff: retrier.NewSimpleBackoff(10, 100),
	}

	if certPem == nil {
		esConfig.Transport = &http.Transport{
			DisableCompression: false,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}

		return esConfig
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(certPem)

	esConfig.Transport = &http.Transport{
		DisableCompression: false,
		TLSClientConfig: &tls.Config{
			RootCAs: caCertPool,
		},
	}

	return esConfig
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

// MustNewClientWithAuth returns a new Client using the @urls and some auth parameters.
// It panics instead of returning an error.
func MustNewClientWithAuth(urls []string, certPem []byte, username, password string) Client {
	client, err := NewClientWithAuth(urls, certPem, username, password)
	if err != nil {
		panic(err)
	}

	return client
}
