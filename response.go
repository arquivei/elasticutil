package elasticutil

import (
	"encoding/json"
	"strings"

	"github.com/arquivei/foundationkit/errors"
	"github.com/olivere/elastic/v7"
)

// AllShardsMustReplyOnElasticSearch checks if any shard failed to respond.
func AllShardsMustReplyOnElasticSearch(
	results *elastic.SearchResult,
	err error,
) (*elastic.SearchResult, error) {
	if err == nil &&
		results != nil &&
		results.Shards != nil &&
		results.Shards.Failed > 0 {
		err = errors.E(
			ErrNotAllShardsReplied,
			errors.KV("replied", results.Shards.Successful),
			errors.KV("failed", results.Shards.Failed),
			errors.KV("total", results.Shards.Total),
		)
	}
	return results, err
}

// GetErrorFromElasticResponse checks if err is an *elastic.Error and returns an error with a formatted message.
//nolint: gocritic, errorlint
func GetErrorFromElasticResponse(err error) error {
	switch e := err.(type) {
	case *elastic.Error:
		err = errors.New(getRootCauseFromElasticError(e.Details))
	}
	return err
}

func getRootCauseFromElasticError(errorDetails *elastic.ErrorDetails) string {
	if errorDetails == nil {
		return "unknown elastic error"
	}
	errors := []string{errorDetails.Type + "[" + errorDetails.Reason + "]"}

	for _, rootCause := range errorDetails.RootCause {
		errors = append(errors, rootCause.Type+"["+rootCause.Reason+"]")
	}

	return strings.Join(errors, ": ")
}

// GetElasticPaginatorFromHits gets the elastic sort in the last hit as a json string.
func GetElasticPaginatorFromHits(hits []*elastic.SearchHit) (string, error) {
	const op errors.Op = "elasticutil.GetElasticPaginatorFromHits"

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
