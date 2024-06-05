package v7

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/arquivei/foundationkit/errors"
	"github.com/elastic/go-elasticsearch/v7/esapi"
)

// SearchResponse represents the response for Search method.
type SearchResponse struct {
	IDs       []string
	Paginator string
	Total     int
	Took      int
}

type envelopeResponse struct {
	Took int
	Hits struct {
		Total struct {
			Value int
		}
		Hits []*envelopeHits `json:"Hits"`
	}
	Shards *shardsInfo `json:"_shards,omitempty"`
}

type envelopeHits struct {
	ID   string        `json:"_id"`
	Sort []interface{} `json:"sort"`
}

type shardsInfo struct {
	Total      int `json:"total"`
	Successful int `json:"successful"`
	Failed     int `json:"failed"`
}

func parseResponse(ctx context.Context, response *esapi.Response) (SearchResponse, error) {
	const op = errors.Op("parseResponse")

	var searchResponse SearchResponse

	var r envelopeResponse
	err := json.NewDecoder(response.Body).Decode(&r)
	if err != nil {
		return SearchResponse{}, errors.E(op, err)
	}

	searchResponse.Total = r.Hits.Total.Value
	searchResponse.Took = r.Took

	enrichLogWithTook(ctx, r.Took)
	enrichLogWithShards(ctx, getTotalShards(r))

	err = checkShards(r)
	if err != nil {
		return searchResponse, errors.E(op, err, ErrCodeBadGateway)
	}

	if len(r.Hits.Hits) < 1 {
		return searchResponse, nil
	}

	for _, hit := range r.Hits.Hits {
		searchResponse.IDs = append(searchResponse.IDs, hit.ID)
	}

	paginator, err := getPaginatorFromHits(r.Hits.Hits)
	if err != nil {
		return SearchResponse{}, errors.E(op, err)
	}

	searchResponse.Paginator = paginator
	return searchResponse, nil
}

func getPaginatorFromHits(hits []*envelopeHits) (string, error) {
	const op errors.Op = "getPaginatorFromHits"

	if len(hits) == 0 {
		return "", nil
	}

	lastHit := hits[len(hits)-1]
	if lastHit == nil || lastHit.Sort == nil {
		return "", nil
	}
	paginator, err := json.Marshal(lastHit.Sort)
	if err != nil {
		return "", errors.E(op, err)
	}
	return string(paginator), nil
}

func getTotalShards(r envelopeResponse) int {
	if r.Shards != nil {
		return r.Shards.Total
	}
	return 0
}

func checkShards(r envelopeResponse) error {
	if r.Shards != nil && r.Shards.Failed > 0 {
		return errors.E(
			ErrNotAllShardsReplied,
			errors.KV("replied", r.Shards.Successful),
			errors.KV("failed", r.Shards.Failed),
			errors.KV("total", r.Shards.Total),
		)
	}
	return nil
}

func checkErrorFromResponse(response *esapi.Response) error {
	if response == nil {
		return ErrNilResponse
	}

	if !response.IsError() {
		return nil
	}

	var responseBody map[string]interface{}
	decodeError := json.NewDecoder(response.Body).Decode(&responseBody)
	if decodeError != nil {
		return decodeError
	}

	errorBody, ok := responseBody["error"].(map[string]interface{})
	if !ok {
		return errors.New("failed to decode error from response")
	}

	rootCause, ok := errorBody["root_cause"].([]interface{})
	if !ok {
		return errors.New("failed to decode root cause from error response")
	}

	var reasonCause interface{}
	if len(rootCause) > 0 {
		rootCauses, ok := rootCause[0].(map[string]interface{})
		if !ok {
			return errors.New("failed to decode root cause map from error response")
		}

		reasonCause = rootCauses["reason"]
	}

	err := fmt.Errorf("[%s] %s: %s: %s", response.Status(), errorBody["type"], errorBody["reason"], reasonCause)

	if response.StatusCode == http.StatusBadRequest {
		err = errors.E(err, ErrCodeBadRequest)
	}

	return err
}
