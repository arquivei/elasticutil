package v7

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/arquivei/foundationkit/errors"
	"github.com/elastic/go-elasticsearch/v9/esapi"
)

// SearchResponse represents the response for Search method.
type SearchResponse struct {
	IDs          []string
	Paginator    string
	Total        int
	Took         int
	Aggregations ResponseAggregation
}

type envelopeResponse struct {
	Took int
	Hits struct {
		Total struct {
			Value int
		}
		Hits []*envelopeHits `json:"Hits"`
	}
	Shards       *shardsInfo            `json:"_shards,omitempty"`
	Aggregations map[string]interface{} `json:"aggregations"`
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

	searchResponse.Aggregations, err = parseAggregations(r.Aggregations)
	if err != nil {
		return SearchResponse{}, errors.E(op, err)
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
		return errors.E("failed to decode error from response", errors.KV("error", responseBody["error"]))
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

func parseAggregations(
	aggs map[string]interface{},
) (ResponseAggregation, error) {
	const op = errors.Op("parseAggregations")

	var responseAggregation ResponseAggregation

	if aggs == nil {
		return responseAggregation, nil
	}

	for aggName, aggData := range aggs {
		parsedAggData, err := parseAggregationData(aggName, aggData)
		if err != nil {
			return responseAggregation, errors.E(op, err)
		}

		responseAggregation = responseAggregation.AppendAggregation(parsedAggData)
	}

	return responseAggregation, nil
}

func parseAggregationData(aggName string, aggData interface{}) (interface{}, error) {
	const op = errors.Op("parseAggregationData")

	aggMap, ok := aggData.(map[string]interface{})
	if !ok {
		return nil, errors.E(
			op,
			"unexpected type for aggregation data, expected: map[string]interface{}",
			errors.KV("given", fmt.Sprintf("%T", aggData)),
		)
	}

	if value, ok := aggMap["value"]; ok {
		return ResponseMetricAggregation{
			Name:  aggName,
			Value: value,
		}, nil
	}

	if buckets, ok := aggMap["buckets"].([]interface{}); ok {
		bucketAggregation := ResponseBucketAggregation{
			Name: aggName,
		}

		for _, bucket := range buckets {
			bucketMap, ok := bucket.(map[string]interface{})
			if !ok {
				return nil, errors.E(
					op,
					"unexpected type for bucket aggregation data, expected: map[string]interface{}",
					errors.KV("given", fmt.Sprintf("%T", bucket)),
				)
			}

			responseBucket := ResponseBucket{}

			for name, data := range bucketMap {
				switch name {
				case "doc_count":
					if docCount, ok := data.(float64); ok {
						responseBucket.DocCount = int(docCount)
					} else {
						return nil, errors.E(
							op,
							"unexpected type for doc_count field, expected: float64",
							errors.KV("given", fmt.Sprintf("%T", data)),
						)
					}
				case "key_as_string":
					if key, ok := data.(string); ok {
						responseBucket.Key = key
					} else {
						return nil, errors.E(
							op,
							"unexpected type for key_as_string field, expected: string",
							errors.KV("given", fmt.Sprintf("%T", data)),
						)
					}
				case "key":
					if responseBucket.Key == "" {
						if key, ok := data.(string); ok {
							responseBucket.Key = key
						} else if key, ok := data.(float64); ok {
							responseBucket.Key = fmt.Sprintf("%f", key)
						} else {
							return nil, errors.E(
								op,
								"unexpected type for key field, expected: string or float64",
								errors.KV("given", fmt.Sprintf("%T", data)),
							)
						}
					}
				default:
					// default case handles sub aggregations inside buckets
					if _, ok := data.(map[string]interface{}); ok {
						bucketSubAgg, err := parseAggregationData(name, data)
						if err != nil {
							return nil, errors.E(op, err)
						}

						if bucketMetricSubAgg, ok := bucketSubAgg.(ResponseMetricAggregation); ok {
							responseBucket.MetricsAggregations = append(
								responseBucket.MetricsAggregations,
								bucketMetricSubAgg,
							)
						}
					}
				}
			}

			bucketAggregation.Buckets = append(bucketAggregation.Buckets, responseBucket)
		}

		return bucketAggregation, nil
	}

	return nil, errors.E(op, "aggregation type is not supported", errors.KV("aggName", aggName))
}
