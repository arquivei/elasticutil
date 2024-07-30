package v7

import (
	"context"
	"strings"

	"github.com/arquivei/foundationkit/errors"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// SearchConfig hold all information to Search method.
type SearchConfig struct {
	Indexes           []string
	Size              int
	Filter            Filter
	IgnoreUnavailable bool
	AllowNoIndices    bool
	TrackTotalHits    bool
	Sort              Sorters
	SearchAfter       string
}

func (c *esClient) Search(ctx context.Context, config SearchConfig) (SearchResponse, error) {
	const op = errors.Op("v7.Client.Search")

	enrichLogWithIndexes(ctx, config.Indexes)

	response, err := c.doSearch(ctx, config)
	if err != nil {
		return SearchResponse{}, errors.E(op, err)
	}
	defer response.Body.Close()

	parsedResponse, err := parseResponse(ctx, response)
	if err != nil {
		if errors.GetCode(err) == errors.CodeEmpty {
			err = errors.E(err, ErrCodeUnexpectedResponse)
		}
		return SearchResponse{}, errors.E(op, err)
	}

	return parsedResponse, nil
}

func (c *esClient) doSearch(ctx context.Context, config SearchConfig) (*esapi.Response, error) {
	const op = errors.Op("doSearch")

	queryString, err := getQuery(ctx, config)
	if err != nil {
		return nil, errors.E(op, err, ErrCodeBadRequest)
	}

	response, err := c.client.Search(
		c.client.Search.WithIndex(config.Indexes...),
		c.client.Search.WithSize(config.Size),
		c.client.Search.WithBody(strings.NewReader(queryString)),
		c.client.Search.WithIgnoreUnavailable(config.IgnoreUnavailable),
		c.client.Search.WithAllowNoIndices(config.AllowNoIndices),
		c.client.Search.WithTrackTotalHits(config.TrackTotalHits),
		c.client.Search.WithSort(config.Sort.Strings()...),
	)
	if err != nil {
		return nil, errors.E(op, err, ErrCodeBadGateway)
	}

	err = checkErrorFromResponse(response)
	if err != nil {
		if errors.GetCode(err) == errors.CodeEmpty {
			err = errors.E(err, ErrCodeBadGateway)
		}
		return nil, errors.E(op, err)
	}

	return response, nil
}

func getQuery(ctx context.Context, config SearchConfig) (string, error) {
	query, err := buildElasticBoolQuery(config.Filter)
	if err != nil {
		return "", err
	}

	queryString := queryWithSearchAfter(query, config.SearchAfter)
	enrichLogWithQuery(ctx, queryString)
	return queryString, nil
}
