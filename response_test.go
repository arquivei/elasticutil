package elasticutil

import (
	"encoding/json"
	"testing"

	"github.com/arquivei/foundationkit/errors"
	"github.com/olivere/elastic/v7"
	"github.com/stretchr/testify/assert"
)

var (
	unknownErrString = "unknown elastic error"
	randomErr        = errors.New("err")
)

func TestGetErrorFromElasticResponse(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected error
	}{
		{
			name:     "Nil error",
			err:      nil,
			expected: nil,
		},
		{
			name:     "Non-Elastic Error",
			err:      randomErr,
			expected: randomErr,
		},
		{
			name: "Elastic Error with nil details",
			err: &elastic.Error{
				Details: nil,
			},
			expected: errors.New(unknownErrString),
		},
		{
			name: "Elastic Error with details",
			err: &elastic.Error{
				Details: &elastic.ErrorDetails{
					Type:   "type",
					Reason: "reason",
					RootCause: []*elastic.ErrorDetails{
						{
							Type:   "root",
							Reason: "root_reason",
						},
					},
				},
			},
			expected: errors.New("type[reason]: root[root_reason]"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := GetErrorFromElasticResponse(tt.err)
			assert.Equal(t, tt.expected, err)
		})
	}
}

func Test_getRootCauseFromElasticError(t *testing.T) {
	rootCauses := []*elastic.ErrorDetails{
		{
			Type:   "root1_type",
			Reason: "root1_reason",
		},
		{
			Type:   "root2_type",
			Reason: "root2_reason",
		},
	}

	tests := []struct {
		name       string
		errDetails *elastic.ErrorDetails
		want       string
	}{
		{
			name:       "nil details",
			errDetails: nil,
			want:       unknownErrString,
		},
		{
			name: "Error details without root causes",
			errDetails: &elastic.ErrorDetails{
				Type:   "type",
				Reason: "reason",
			},
			want: "type[reason]",
		},
		{
			name: "Error details without root causes",
			errDetails: &elastic.ErrorDetails{
				Type:   "type",
				Reason: "reason",
			},
			want: "type[reason]",
		},
		{
			name: "Error details with root causes",
			errDetails: &elastic.ErrorDetails{
				Type:   "type",
				Reason: "reason",
				RootCause: []*elastic.ErrorDetails{
					rootCauses[0],
				},
			},
			want: "type[reason]: root1_type[root1_reason]",
		},
		{
			name: "Error details with multiple root causes",
			errDetails: &elastic.ErrorDetails{
				Type:      "type",
				Reason:    "reason",
				RootCause: rootCauses,
			},
			want: "type[reason]: root1_type[root1_reason]: root2_type[root2_reason]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getRootCauseFromElasticError(tt.errDetails)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAllShardsMustReplyOnElasticSearch(t *testing.T) {
	type args struct {
		searchResult *elastic.SearchResult
		err          error
	}
	tests := []struct {
		name        string
		args        args
		expectedErr error
	}{
		{
			name:        "No search result, no error",
			expectedErr: nil,
		},
		{
			name: "Random error occurs",
			args: args{
				searchResult: &elastic.SearchResult{},
				err:          randomErr,
			},
			expectedErr: randomErr,
		},
		{
			name: "Search Result with no error",
			args: args{
				searchResult: &elastic.SearchResult{},
			},
			expectedErr: nil,
		},
		{
			name: "Search Result with no error: failed shards",
			args: args{
				searchResult: &elastic.SearchResult{
					Shards: &elastic.ShardsInfo{
						Total:      8,
						Successful: 4,
						Failed:     3,
						Skipped:    1,
					},
				},
			},
			expectedErr: errors.E(
				ErrNotAllShardsReplied,
				errors.KV("replied", 4),
				errors.KV("failed", 3),
				errors.KV("total", 8),
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := AllShardsMustReplyOnElasticSearch(tt.args.searchResult, tt.args.err)
			assert.Equal(t, tt.args.searchResult, got) // Shouldn't have changed
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}

func TestGetElasticPaginatorFromHits(t *testing.T) {
	op := errors.Op("elasticutil.GetElasticPaginatorFromHits")
	_, err := json.Marshal(make(chan int))
	jsonErr := errors.E(op, err)

	searchHits := []*elastic.SearchHit{
		{
			Sort: []interface{}{
				"123",
			},
		},
		{
			Sort: []interface{}{
				"456",
			},
		},
	}

	tests := []struct {
		name        string
		hits        []*elastic.SearchHit
		want        string
		expectedErr error
	}{
		{
			name:        "nil args",
			want:        "",
			expectedErr: nil,
		},
		{
			name:        "empty slice hits",
			hits:        []*elastic.SearchHit{},
			want:        "",
			expectedErr: nil,
		},
		{
			name: "single hit",
			hits: []*elastic.SearchHit{
				searchHits[0],
			},
			want:        `["123"]`,
			expectedErr: nil,
		},
		{
			name:        "multiple hits",
			hits:        searchHits,
			want:        `["456"]`,
			expectedErr: nil,
		},
		{
			name: "last hit got no sort",
			hits: []*elastic.SearchHit{
				searchHits[0],
				{},
			},
			want:        "",
			expectedErr: nil,
		},
		{
			name: "json marshal error",
			hits: []*elastic.SearchHit{
				{
					Sort: []interface{}{
						make(chan int),
					},
				},
			},
			expectedErr: jsonErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetElasticPaginatorFromHits(tt.hits)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}
