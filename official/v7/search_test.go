package v7

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/arquivei/foundationkit/errors"
	"github.com/arquivei/foundationkit/ref"
	es "github.com/elastic/go-elasticsearch/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/arquivei/elasticutil/official/v7/querybuilders"
)

func Test_Search(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name              string
		config            SearchConfig
		expectedResponse  SearchResponse
		transport         *mockTransport
		expectedError     string
		expectedErrorCode errors.Code
	}{
		{
			name: "success",
			config: SearchConfig{
				Indexes:           []string{"index1", "index2"},
				Size:              19,
				Filter:            getMockFilter(),
				IgnoreUnavailable: true,
				AllowNoIndices:    true,
				TrackTotalHits:    true,
				Sort: Sorters{
					Sorters: []Sorter{
						{
							Field:     "Date",
							Ascending: true,
						},
						{
							Field:     "ID",
							Ascending: false,
						},
					},
				},
				SearchAfter: `{"paginator"}`,
			},
			transport: func() *mockTransport {
				server := new(mockTransport)
				server.On(
					"RoundTrip",
					"http://localhost:9200/index1,index2/_search?allow_no_indices=true&ignore_unavailable=true&size=19&sort=Date%3Aasc%2CID%3Adesc&track_total_hits=true",
					`{"query":{"bool":{"must":[{"terms":{"Name":["John","Mary"]}},{"terms":{"Age":[16,17,18,25,26]}},{"term":{"HasCovid":true}},{"range":{"CreatedAt":{"from":"2020-11-28T15:27:39.000000049Z","include_lower":true,"include_upper":true,"to":"2021-11-28T15:27:39.000000049Z"}}},{"range":{"Age":{"from":15,"include_lower":true,"include_upper":true,"to":30}}},{"range":{"Age":{"from":0.5,"include_lower":true,"include_upper":true,"to":1.9}}},{"nested":{"path":"Covid","query":{"bool":{"must":[{"terms":{"Covid.Symptom":["cough"]}},{"range":{"Covid.Date":{"from":"2019-11-28T15:27:39.000000049Z","include_lower":true,"include_upper":true,"to":"2020-11-28T15:27:39.000000049Z"}}}]}}}},{"bool":{"should":[{"multi_match":{"fields":["Name","SocialName"],"max_expansions":1024,"query":"John","type":"phrase_prefix"}},{"multi_match":{"fields":["Name","SocialName"],"max_expansions":1024,"query":"Mary","type":"phrase_prefix"}},{"multi_match":{"fields":["Name","SocialName"],"max_expansions":1024,"query":"Rebecca","type":"phrase_prefix"}}]}},{"bool":{"must":[{"multi_match":{"fields":["Name","SocialName"],"max_expansions":1024,"query":"Lennon","type":"phrase_prefix"}},{"multi_match":{"fields":["Name","SocialName"],"max_expansions":1024,"query":"McCartney","type":"phrase_prefix"}}]}},{"bool":{"should":[{"multi_match":{"fields":["Any"],"max_expansions":1024,"query":"Beatles","type":"best_fields"}},{"multi_match":{"fields":["Any"],"max_expansions":1024,"query":"Stones","type":"best_fields"}}]}},{"bool":{"must":{"term":{"Name":"John"}}}},{"nested":{"path":"Covid","query":{"exists":{"field":"Covid"}}}},{"exists":{"field":"Age"}}],"must_not":[{"terms":{"Name":["Lary"]}},{"range":{"Age":{"from":29,"include_lower":true,"include_upper":true,"to":30}}}]}}, 	"search_after": {"paginator"}}`,
				).Once().Return(
					`{"took":10,"_shards":{"total":1,"successful":1,"skipped":0,"failed":0},"hits":{"total":{"value":2},"max_score":null,"hits":[{"_index":"tiramisu_cte-2022101","_id":"elastic-id-1","_score":null,"sort":["pag2"]},{"_index":"tiramisu_cte-2019","_id":"elastic-id-2","_score":null,"sort":["pag3"]}]}}`,
					200,
					nil,
				)

				return server
			}(),
			expectedResponse: SearchResponse{
				IDs:       []string{"elastic-id-1", "elastic-id-2"},
				Paginator: `["pag3"]`,
				Total:     2,
				Took:      10,
			},
		},
		{
			name: "1 shard failed",
			config: SearchConfig{
				Indexes:           []string{"index1", "index2"},
				Size:              19,
				Filter:            getMockFilter(),
				IgnoreUnavailable: true,
				AllowNoIndices:    true,
				TrackTotalHits:    true,
				Sort: Sorters{
					Sorters: []Sorter{
						{
							Field:     "Date",
							Ascending: true,
						},
						{
							Field:     "ID",
							Ascending: false,
						},
					},
				},
				SearchAfter: `{"paginator"}`,
			},
			transport: func() *mockTransport {
				server := new(mockTransport)
				server.On(
					"RoundTrip",
					"http://localhost:9200/index1,index2/_search?allow_no_indices=true&ignore_unavailable=true&size=19&sort=Date%3Aasc%2CID%3Adesc&track_total_hits=true",
					`{"query":{"bool":{"must":[{"terms":{"Name":["John","Mary"]}},{"terms":{"Age":[16,17,18,25,26]}},{"term":{"HasCovid":true}},{"range":{"CreatedAt":{"from":"2020-11-28T15:27:39.000000049Z","include_lower":true,"include_upper":true,"to":"2021-11-28T15:27:39.000000049Z"}}},{"range":{"Age":{"from":15,"include_lower":true,"include_upper":true,"to":30}}},{"range":{"Age":{"from":0.5,"include_lower":true,"include_upper":true,"to":1.9}}},{"nested":{"path":"Covid","query":{"bool":{"must":[{"terms":{"Covid.Symptom":["cough"]}},{"range":{"Covid.Date":{"from":"2019-11-28T15:27:39.000000049Z","include_lower":true,"include_upper":true,"to":"2020-11-28T15:27:39.000000049Z"}}}]}}}},{"bool":{"should":[{"multi_match":{"fields":["Name","SocialName"],"max_expansions":1024,"query":"John","type":"phrase_prefix"}},{"multi_match":{"fields":["Name","SocialName"],"max_expansions":1024,"query":"Mary","type":"phrase_prefix"}},{"multi_match":{"fields":["Name","SocialName"],"max_expansions":1024,"query":"Rebecca","type":"phrase_prefix"}}]}},{"bool":{"must":[{"multi_match":{"fields":["Name","SocialName"],"max_expansions":1024,"query":"Lennon","type":"phrase_prefix"}},{"multi_match":{"fields":["Name","SocialName"],"max_expansions":1024,"query":"McCartney","type":"phrase_prefix"}}]}},{"bool":{"should":[{"multi_match":{"fields":["Any"],"max_expansions":1024,"query":"Beatles","type":"best_fields"}},{"multi_match":{"fields":["Any"],"max_expansions":1024,"query":"Stones","type":"best_fields"}}]}},{"bool":{"must":{"term":{"Name":"John"}}}},{"nested":{"path":"Covid","query":{"exists":{"field":"Covid"}}}},{"exists":{"field":"Age"}}],"must_not":[{"terms":{"Name":["Lary"]}},{"range":{"Age":{"from":29,"include_lower":true,"include_upper":true,"to":30}}}]}}, 	"search_after": {"paginator"}}`,
				).Once().Return(
					`{"took":10,"_shards":{"total":1,"successful":0,"skipped":0,"failed":1},"hits":{"total":{"value":2},"max_score":null,"hits":[{"_index":"tiramisu_cte-2022101","_id":"elastic-id-1","_score":null,"sort":["pag2"]},{"_index":"tiramisu_cte-2019","_id":"elastic-id-2","_score":null,"sort":["pag3"]}]}}`,
					200,
					nil,
				)

				return server
			}(),
			expectedError:     "v7.Client.Search: parseResponse: not all shards replied [replied=0,failed=1,total=1]",
			expectedErrorCode: ErrCodeBadGateway,
		},
		{
			name: "elastic return an error",
			config: SearchConfig{
				Indexes:           []string{"index1", "index2"},
				Size:              19,
				Filter:            getMockFilter(),
				IgnoreUnavailable: true,
				AllowNoIndices:    true,
				TrackTotalHits:    true,
				Sort: Sorters{
					Sorters: []Sorter{
						{
							Field:     "Date",
							Ascending: true,
						},
						{
							Field:     "ID",
							Ascending: false,
						},
					},
				},
				SearchAfter: `{"paginator"}`,
			},
			transport: func() *mockTransport {
				server := new(mockTransport)
				server.On(
					"RoundTrip",
					"http://localhost:9200/index1,index2/_search?allow_no_indices=true&ignore_unavailable=true&size=19&sort=Date%3Aasc%2CID%3Adesc&track_total_hits=true",
					`{"query":{"bool":{"must":[{"terms":{"Name":["John","Mary"]}},{"terms":{"Age":[16,17,18,25,26]}},{"term":{"HasCovid":true}},{"range":{"CreatedAt":{"from":"2020-11-28T15:27:39.000000049Z","include_lower":true,"include_upper":true,"to":"2021-11-28T15:27:39.000000049Z"}}},{"range":{"Age":{"from":15,"include_lower":true,"include_upper":true,"to":30}}},{"range":{"Age":{"from":0.5,"include_lower":true,"include_upper":true,"to":1.9}}},{"nested":{"path":"Covid","query":{"bool":{"must":[{"terms":{"Covid.Symptom":["cough"]}},{"range":{"Covid.Date":{"from":"2019-11-28T15:27:39.000000049Z","include_lower":true,"include_upper":true,"to":"2020-11-28T15:27:39.000000049Z"}}}]}}}},{"bool":{"should":[{"multi_match":{"fields":["Name","SocialName"],"max_expansions":1024,"query":"John","type":"phrase_prefix"}},{"multi_match":{"fields":["Name","SocialName"],"max_expansions":1024,"query":"Mary","type":"phrase_prefix"}},{"multi_match":{"fields":["Name","SocialName"],"max_expansions":1024,"query":"Rebecca","type":"phrase_prefix"}}]}},{"bool":{"must":[{"multi_match":{"fields":["Name","SocialName"],"max_expansions":1024,"query":"Lennon","type":"phrase_prefix"}},{"multi_match":{"fields":["Name","SocialName"],"max_expansions":1024,"query":"McCartney","type":"phrase_prefix"}}]}},{"bool":{"should":[{"multi_match":{"fields":["Any"],"max_expansions":1024,"query":"Beatles","type":"best_fields"}},{"multi_match":{"fields":["Any"],"max_expansions":1024,"query":"Stones","type":"best_fields"}}]}},{"bool":{"must":{"term":{"Name":"John"}}}},{"nested":{"path":"Covid","query":{"exists":{"field":"Covid"}}}},{"exists":{"field":"Age"}}],"must_not":[{"terms":{"Name":["Lary"]}},{"range":{"Age":{"from":29,"include_lower":true,"include_upper":true,"to":30}}}]}}, 	"search_after": {"paginator"}}`,
				).Once().Return(
					`{"took":10,"_shards":{"total":1,"successful":1,"skipped":0,"failed":0},"hits":{"total":{"value":2},"max_score":null,"hits":[{"_index":"tiramisu_cte-2022101","_id":"elastic-id-1","_score":null,"sort":["pag2"]},{"_index":"tiramisu_cte-2019","_id":"elastic-id-2","_score":null,"sort":["pag3"]}]}}`,
					200,
					errors.New("elastic error"),
				)

				return server
			}(),
			expectedError:     "v7.Client.Search: doSearch: elastic error",
			expectedErrorCode: ErrCodeBadGateway,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			assert.NotPanics(t, func() {
				client := mustNewClientTest(test.transport)
				response, err := client.Search(context.Background(), test.config)
				if test.expectedError == "" {
					assert.NoError(t, err)
				} else {
					assert.EqualError(t, err, test.expectedError)
					assert.Equal(t, test.expectedErrorCode, errors.GetCode(err))
				}
				assert.Equal(t, test.expectedResponse, response)
			})
		})
	}
}

type mockTransport struct {
	mock.Mock
}

func (m *mockTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	buf := new(strings.Builder)
	io.Copy(buf, r.Body)
	request := buf.String()

	args := m.Called(r.URL.String(), request)
	bodyContent := args.String(0)
	httpCode := args.Int(1)

	response := &http.Response{
		Header:     make(http.Header),
		StatusCode: httpCode,
		Body:       io.NopCloser(strings.NewReader(bodyContent)),
	}

	response.Header.Add("X-Elastic-Product", "Elasticsearch")

	return response, args.Error(2)
}

func getMockFilter() Filter {
	type ExampleFilterMust struct {
		Names             []string `es:"Name"`
		Ages              []uint64 `es:"Age"`
		HasCovid          *bool
		CreatedAt         *TimeRange
		AgeRange          *IntRange              `es:"Age"`
		ValueRange        *FloatRange            `es:"Age"`
		CovidInfo         Nested                 `es:"Covid"`
		NameOrSocialName  FullTextSearchShould   `es:"Name,SocialName"`
		NameAndSocialName FullTextSearchMust     `es:"Name,SocialName"`
		Any               MultiMatchSearchShould `es:"Any"`
		MyCustomSearch    CustomSearch
	}

	type ExampleFilterExists struct {
		HasCovidInfo Nested `es:"Covid"`
		HasAge       *bool  `es:"Age"`
	}

	type ExampleCovidInfo struct {
		HasCovidInfo     *bool      `es:"Covid"`
		Symptoms         []string   `es:"Covid.Symptom"`
		FirstSymptomDate *TimeRange `es:"Covid.Date"`
	}

	return Filter{
		Must: ExampleFilterMust{
			Names:    []string{"John", "Mary"},
			Ages:     []uint64{16, 17, 18, 25, 26},
			HasCovid: ref.Bool(true),
			CovidInfo: NewNested(
				ExampleCovidInfo{
					Symptoms: []string{"cough"},
					FirstSymptomDate: &TimeRange{
						From: time.Date(2019, time.November, 28, 15, 27, 39, 49, time.UTC),
						To:   time.Date(2020, time.November, 28, 15, 27, 39, 49, time.UTC),
					},
				},
			),
			CreatedAt: &TimeRange{
				From: time.Date(2020, time.November, 28, 15, 27, 39, 49, time.UTC),
				To:   time.Date(2021, time.November, 28, 15, 27, 39, 49, time.UTC),
			},
			AgeRange: &IntRange{
				From: 15,
				To:   30,
			},
			NameOrSocialName: NewFullTextSearchShould([]string{"John", "Mary", "Rebecca"}),
			ValueRange: &FloatRange{
				From: 0.5,
				To:   1.9,
			},
			NameAndSocialName: NewFullTextSearchMust([]string{"Lennon", "McCartney"}),
			Any:               NewMultiMatchSearchShould([]string{"Beatles", "Stones"}),
			MyCustomSearch: NewCustomSearch(func() (querybuilders.Query, error) {
				return querybuilders.NewBoolQuery().Must(querybuilders.NewTermQuery("Name", "John")), nil
			}),
		},
		MustNot: ExampleFilterMust{
			Names: []string{"Lary"},
			AgeRange: &IntRange{
				From: 29,
				To:   30,
			},
		},
		Exists: ExampleFilterExists{
			HasCovidInfo: NewNested(
				ExampleCovidInfo{
					HasCovidInfo: ref.Bool(true),
				},
			),
			HasAge: ref.Bool(true),
		},
	}
}

func mustNewClientTest(roundTripper http.RoundTripper) Client {
	client, err := es.NewClient(
		es.Config{
			Transport:            roundTripper,
			UseResponseCheckOnly: true,
		},
	)
	if err != nil {
		panic(err)
	}
	return &esClient{
		client: client,
	}
}
