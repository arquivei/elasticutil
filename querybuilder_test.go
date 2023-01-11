//nolint
package elasticutil

import (
	"context"
	"testing"
	"time"

	"github.com/arquivei/foundationkit/errors"
	"github.com/arquivei/foundationkit/ref"
	"github.com/olivere/elastic/v7"
	"github.com/stretchr/testify/assert"
)

func Test_BuildElasticBoolQuery(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		filter        Filter
		expectedQuery string
		expectedError string
	}{
		{
			name:          "Match All",
			filter:        Filter{},
			expectedQuery: `{"match_all":{}}`,
		},
		{
			name: "[Must] One Field (slice string)",
			filter: Filter{
				Must: MockFilterMust{
					Strings: []string{"1", "2"},
				},
			},
			expectedQuery: `{"terms":{"String":["1","2"]}}`,
		},
		{
			name: "[Must] One Field (slice int)",
			filter: Filter{
				Must: MockFilterMust{
					Ints: []uint64{1, 2},
				},
			},
			expectedQuery: `{"terms":{"Int":[1,2]}}`,
		},
		{
			name: "[Must] One Field (slice int empty)",
			filter: Filter{
				Must: MockFilterMust{
					Ints: []uint64{},
				},
			},
			expectedQuery: `{"match_all":{}}`,
		},
		{
			name: "[Must] One Field (bool)",
			filter: Filter{
				Must: MockFilterMust{
					Bool: ref.Bool(true),
				},
			},
			expectedQuery: `{"term":{"Bool":true}}`,
		},
		{
			name: "[Must] One Field (single nested)",
			filter: Filter{
				Must: MockFilterMust{
					MockSingleNested: NewNested(MockSingleNested{
						Slice: []string{"1", "2"},
					}),
				},
			},
			expectedQuery: `{"nested":{"path":"SingleNestedField","query":{"terms":{"SingleNested.Slice":["1","2"]}}}}`,
		},
		{
			name: "[Must] One Field (multi nested)",
			filter: Filter{
				Must: MockFilterMust{
					MockMultiNested: NewNested(MockMultiNested{
						Bool:  ref.Bool(true),
						Slice: []string{"1", "2"},
						Range: &TimeRange{
							From: time.Date(1995, time.March, 1, 11, 35, 19, 29, time.UTC),
							To:   time.Date(2019, time.November, 28, 15, 27, 39, 49, time.UTC),
						},
					}),
				},
			},
			expectedQuery: `{"nested":{"path":"MultiNestedField","query":{"bool":{"must":[{"term":{"MultiNested.Bool":true}},{"terms":{"MultiNested.Slice":["1","2"]}},{"range":{"MultiNested.Range":{"from":"1995-03-01T11:35:19.000000029Z","include_lower":true,"include_upper":true,"to":"2019-11-28T15:27:39.000000049Z"}}}]}}}}`,
		},
		{
			name: "[Must] One Field (time range)",
			filter: Filter{
				Must: MockFilterMust{
					Times: &TimeRange{
						From: time.Date(1995, time.March, 1, 11, 35, 19, 29, time.UTC),
						To:   time.Date(2019, time.November, 28, 15, 27, 39, 49, time.UTC),
					},
				},
			},
			expectedQuery: `{"range":{"Time":{"from":"1995-03-01T11:35:19.000000029Z","include_lower":true,"include_upper":true,"to":"2019-11-28T15:27:39.000000049Z"}}}`,
		},
		{
			name: "[Must] One Field (int range)",
			filter: Filter{
				Must: MockFilterMust{
					Numbers: &IntRange{
						From: 1,
						To:   100,
					},
				},
			},
			expectedQuery: `{"range":{"Number":{"from":1,"include_lower":true,"include_upper":true,"to":100}}}`,
		},
		{
			name: "[Must] One Field (float range)",
			filter: Filter{
				Must: MockFilterMust{
					Values: &FloatRange{
						From: 1.0,
						To:   100.0,
					},
				},
			},
			expectedQuery: `{"range":{"Value":{"from":1,"include_lower":true,"include_upper":true,"to":100}}}`,
		},
		{
			name: "[Must] One Field (full text search should multi)",
			filter: Filter{
				Must: MockFilterMust{
					MultiShould: NewFullTextSearchShould([]string{"1", "2"}),
				},
			},
			expectedQuery: `{"bool":{"should":[{"multi_match":{"fields":["MultiShould1","MultiShould2","MultiShould3"],"max_expansions":1024,"query":"1","type":"phrase_prefix"}},{"multi_match":{"fields":["MultiShould1","MultiShould2","MultiShould3"],"max_expansions":1024,"query":"2","type":"phrase_prefix"}}]}}`,
		},
		{
			name: "[Must] One Field (full text search should single)",
			filter: Filter{
				Must: MockFilterMust{
					SingleShould: NewFullTextSearchShould([]string{"1", "2"}),
				},
			},
			expectedQuery: `{"bool":{"should":[{"multi_match":{"fields":["SingleShould1"],"max_expansions":1024,"query":"1","type":"phrase_prefix"}},{"multi_match":{"fields":["SingleShould1"],"max_expansions":1024,"query":"2","type":"phrase_prefix"}}]}}`,
		},
		{
			name: "[Must] One Field (full text search must multi)",
			filter: Filter{
				Must: MockFilterMust{
					MultiMust: NewFullTextSearchMust([]string{"1", "2"}),
				},
			},
			expectedQuery: `{"bool":{"must":[{"multi_match":{"fields":["MultiMust1","MultiMust2","MultiMust3"],"max_expansions":1024,"query":"1","type":"phrase_prefix"}},{"multi_match":{"fields":["MultiMust1","MultiMust2","MultiMust3"],"max_expansions":1024,"query":"2","type":"phrase_prefix"}}]}}`,
		},
		{
			name: "[Must] One Field (full text search must single)",
			filter: Filter{
				Must: MockFilterMust{
					SingleMust: NewFullTextSearchMust([]string{"1", "2"}),
				},
			},
			expectedQuery: `{"bool":{"must":[{"multi_match":{"fields":["SingleMust1"],"max_expansions":1024,"query":"1","type":"phrase_prefix"}},{"multi_match":{"fields":["SingleMust1"],"max_expansions":1024,"query":"2","type":"phrase_prefix"}}]}}`,
		},
		{
			name: "[Must] One field of each type",
			filter: Filter{
				Must: MockFilterMust{
					Strings: []string{"1", "2"},
					Ints:    []uint64{1, 2},
					Bool:    ref.Bool(true),
					MockSingleNested: NewNested(MockSingleNested{
						Slice: []string{"1", "2"},
					}),
					MockMultiNested: NewNested(MockMultiNested{
						Bool:  ref.Bool(true),
						Slice: []string{"1", "2"},
						Range: &TimeRange{
							From: time.Date(1995, time.March, 1, 11, 35, 19, 29, time.UTC),
							To:   time.Date(2019, time.November, 28, 15, 27, 39, 49, time.UTC),
						},
					}),
					Times: &TimeRange{
						From: time.Date(1995, time.March, 1, 11, 35, 19, 29, time.UTC),
						To:   time.Date(2019, time.November, 28, 15, 27, 39, 49, time.UTC),
					},
					Numbers: &IntRange{
						From: 1,
						To:   100,
					},
					Values: &FloatRange{
						From: 1.0,
						To:   100.0,
					},
					MultiShould:  NewFullTextSearchShould([]string{"1", "2"}),
					SingleShould: NewFullTextSearchShould([]string{"1", "2"}),
					MultiMust:    NewFullTextSearchMust([]string{"1", "2"}),
					SingleMust:   NewFullTextSearchMust([]string{"1", "2"}),
				},
			},
			expectedQuery: `{"bool":{"must":[{"terms":{"String":["1","2"]}},{"terms":{"Int":[1,2]}},{"term":{"Bool":true}},{"nested":{"path":"MultiNestedField","query":{"bool":{"must":[{"term":{"MultiNested.Bool":true}},{"terms":{"MultiNested.Slice":["1","2"]}},{"range":{"MultiNested.Range":{"from":"1995-03-01T11:35:19.000000029Z","include_lower":true,"include_upper":true,"to":"2019-11-28T15:27:39.000000049Z"}}}]}}}},{"nested":{"path":"SingleNestedField","query":{"terms":{"SingleNested.Slice":["1","2"]}}}},{"range":{"Time":{"from":"1995-03-01T11:35:19.000000029Z","include_lower":true,"include_upper":true,"to":"2019-11-28T15:27:39.000000049Z"}}},{"range":{"Number":{"from":1,"include_lower":true,"include_upper":true,"to":100}}},{"range":{"Value":{"from":1,"include_lower":true,"include_upper":true,"to":100}}},{"bool":{"should":[{"multi_match":{"fields":["MultiShould1","MultiShould2","MultiShould3"],"max_expansions":1024,"query":"1","type":"phrase_prefix"}},{"multi_match":{"fields":["MultiShould1","MultiShould2","MultiShould3"],"max_expansions":1024,"query":"2","type":"phrase_prefix"}}]}},{"bool":{"should":[{"multi_match":{"fields":["SingleShould1"],"max_expansions":1024,"query":"1","type":"phrase_prefix"}},{"multi_match":{"fields":["SingleShould1"],"max_expansions":1024,"query":"2","type":"phrase_prefix"}}]}},{"bool":{"must":[{"multi_match":{"fields":["MultiMust1","MultiMust2","MultiMust3"],"max_expansions":1024,"query":"1","type":"phrase_prefix"}},{"multi_match":{"fields":["MultiMust1","MultiMust2","MultiMust3"],"max_expansions":1024,"query":"2","type":"phrase_prefix"}}]}},{"bool":{"must":[{"multi_match":{"fields":["SingleMust1"],"max_expansions":1024,"query":"1","type":"phrase_prefix"}},{"multi_match":{"fields":["SingleMust1"],"max_expansions":1024,"query":"2","type":"phrase_prefix"}}]}}]}}`,
		},
		{
			name: "[Must] Custom query",
			filter: Filter{
				Must: MockFilterMust{
					Custom: NewCustomSearch(func() (*elastic.BoolQuery, error) {
						return elastic.NewBoolQuery().Must(elastic.NewMatchQuery("Strings", "1")), nil
					}),
				},
			},
			expectedQuery: `{"bool":{"must":{"match":{"Strings":{"query":"1"}}}}}`,
		},
		{
			name: "[Must Not] One Field (slice string)",
			filter: Filter{
				MustNot: MockFilterMustNot{
					Strings: []string{"1", "2"},
				},
			},
			expectedQuery: `{"bool":{"must_not":{"terms":{"String":["1","2"]}}}}`,
		},
		{
			name: "[Must Not] One Field (slice int)",
			filter: Filter{
				MustNot: MockFilterMustNot{
					Ints: []uint64{1, 2},
				},
			},
			expectedQuery: `{"bool":{"must_not":{"terms":{"Int":[1,2]}}}}`,
		},
		{
			name: "[Must Not] One Field (bool)",
			filter: Filter{
				MustNot: MockFilterMustNot{
					Bool: ref.Bool(true),
				},
			},
			expectedQuery: `{"bool":{"must_not":{"term":{"Bool":true}}}}`,
		},
		{
			name: "[Must Not] One Field (single nested)",
			filter: Filter{
				MustNot: MockFilterMustNot{
					SingleNested: NewNested(MockSingleNested{
						Slice: []string{"1", "2"},
					}),
				},
			},
			expectedQuery: `{"bool":{"must_not":{"nested":{"path":"SingleNestedField","query":{"terms":{"SingleNested.Slice":["1","2"]}}}}}}`,
		},
		{
			name: "[Must Not] One Field (multi nested)",
			filter: Filter{
				MustNot: MockFilterMustNot{
					MultiNested: NewNested(MockMultiNested{
						Bool:  ref.Bool(true),
						Slice: []string{"1", "2"},
						Range: &TimeRange{
							From: time.Date(1995, time.March, 1, 11, 35, 19, 29, time.UTC),
							To:   time.Date(2019, time.November, 28, 15, 27, 39, 49, time.UTC),
						},
					}),
				},
			},
			expectedQuery: `{"bool":{"must_not":{"nested":{"path":"MultiNestedField","query":{"bool":{"must":[{"term":{"MultiNested.Bool":true}},{"terms":{"MultiNested.Slice":["1","2"]}},{"range":{"MultiNested.Range":{"from":"1995-03-01T11:35:19.000000029Z","include_lower":true,"include_upper":true,"to":"2019-11-28T15:27:39.000000049Z"}}}]}}}}}}`,
		},
		{
			name: "[Must Not] One Field (multi nested empty)",
			filter: Filter{
				MustNot: MockFilterMustNot{
					MultiNested: NewNested(MockMultiNested{}),
				},
			},
			expectedQuery: `{"match_all":{}}`,
		},
		{
			name: "[Must Not] One Field (time range)",
			filter: Filter{
				MustNot: MockFilterMustNot{
					Times: &TimeRange{
						From: time.Date(1995, time.March, 1, 11, 35, 19, 29, time.UTC),
						To:   time.Date(2019, time.November, 28, 15, 27, 39, 49, time.UTC),
					},
				},
			},
			expectedQuery: `{"bool":{"must_not":{"range":{"Time":{"from":"1995-03-01T11:35:19.000000029Z","include_lower":true,"include_upper":true,"to":"2019-11-28T15:27:39.000000049Z"}}}}}`,
		},
		{
			name: "[Must Not] One Field (time range, 1 zero)",
			filter: Filter{
				MustNot: MockFilterMustNot{
					Times: &TimeRange{
						From: time.Time{},
						To:   time.Date(2019, time.November, 28, 15, 27, 39, 49, time.UTC),
					},
				},
			},
			expectedQuery: `{"bool":{"must_not":{"range":{"Time":{"from":null,"include_lower":true,"include_upper":true,"to":"2019-11-28T15:27:39.000000049Z"}}}}}`,
		},
		{
			name: "[Must Not] One Field (time range, 2 zeros)",
			filter: Filter{
				MustNot: MockFilterMustNot{
					Times: &TimeRange{
						From: time.Time{},
						To:   time.Time{},
					},
				},
			},
			expectedQuery: `{"bool":{"must_not":{"range":{"Time":{"from":null,"include_lower":true,"include_upper":true,"to":null}}}}}`,
		},
		{
			name: "[Must Not] One Field (int range)",
			filter: Filter{
				MustNot: MockFilterMustNot{
					Numbers: &IntRange{
						From: 1,
						To:   100,
					},
				},
			},
			expectedQuery: `{"bool":{"must_not":{"range":{"Number":{"from":1,"include_lower":true,"include_upper":true,"to":100}}}}}`,
		},
		{
			name: "[Must Not] One Field (int range, 1 zero)",
			filter: Filter{
				MustNot: MockFilterMustNot{
					Numbers: &IntRange{
						From: 0,
						To:   100,
					},
				},
			},
			expectedQuery: `{"bool":{"must_not":{"range":{"Number":{"from":null,"include_lower":true,"include_upper":true,"to":100}}}}}`,
		},
		{
			name: "[Must Not] One Field (int range, 2 zeros)",
			filter: Filter{
				MustNot: MockFilterMustNot{
					Numbers: &IntRange{
						From: 0,
						To:   0,
					},
				},
			},
			expectedQuery: `{"bool":{"must_not":{"range":{"Number":{"from":null,"include_lower":true,"include_upper":true,"to":null}}}}}`,
		},
		{
			name: "[Must Not] One Field (float range)",
			filter: Filter{
				MustNot: MockFilterMustNot{
					Values: &FloatRange{
						From: 1.0,
						To:   100.0,
					},
				},
			},
			expectedQuery: `{"bool":{"must_not":{"range":{"Value":{"from":1,"include_lower":true,"include_upper":true,"to":100}}}}}`,
		},
		{
			name: "[Must Not] One Field (float range, 1 zero)",
			filter: Filter{
				MustNot: MockFilterMustNot{
					Values: &FloatRange{
						From: 0,
						To:   100.0,
					},
				},
			},
			expectedQuery: `{"bool":{"must_not":{"range":{"Value":{"from":null,"include_lower":true,"include_upper":true,"to":100}}}}}`,
		},
		{
			name: "[Must Not] One Field (float range, 2 zeros)",
			filter: Filter{
				MustNot: MockFilterMustNot{
					Values: &FloatRange{
						From: 0.0,
						To:   0.0,
					},
				},
			},
			expectedQuery: `{"bool":{"must_not":{"range":{"Value":{"from":null,"include_lower":true,"include_upper":true,"to":null}}}}}`,
		},
		{
			name: "[Must Not] One Field (full text search should multi)",
			filter: Filter{
				MustNot: MockFilterMustNot{
					MultiShould: NewFullTextSearchShould([]string{"1", "2"}),
				},
			},
			expectedQuery: `{"bool":{"must_not":{"bool":{"should":[{"multi_match":{"fields":["MultiShould1","MultiShould2","MultiShould3"],"max_expansions":1024,"query":"1","type":"phrase_prefix"}},{"multi_match":{"fields":["MultiShould1","MultiShould2","MultiShould3"],"max_expansions":1024,"query":"2","type":"phrase_prefix"}}]}}}}`,
		},
		{
			name: "[Must Not] One Field (full text search should single)",
			filter: Filter{
				MustNot: MockFilterMustNot{
					SingleShould: NewFullTextSearchShould([]string{"1", "2"}),
				},
			},
			expectedQuery: `{"bool":{"must_not":{"bool":{"should":[{"multi_match":{"fields":["SingleShould1"],"max_expansions":1024,"query":"1","type":"phrase_prefix"}},{"multi_match":{"fields":["SingleShould1"],"max_expansions":1024,"query":"2","type":"phrase_prefix"}}]}}}}`,
		},
		{
			name: "[Must Not] One Field (full text search must multi)",
			filter: Filter{
				MustNot: MockFilterMustNot{
					MultiMust: NewFullTextSearchMust([]string{"1", "2"}),
				},
			},
			expectedQuery: `{"bool":{"must_not":{"bool":{"must":[{"multi_match":{"fields":["MultiMust1","MultiMust2","MultiMust3"],"max_expansions":1024,"query":"1","type":"phrase_prefix"}},{"multi_match":{"fields":["MultiMust1","MultiMust2","MultiMust3"],"max_expansions":1024,"query":"2","type":"phrase_prefix"}}]}}}}`,
		},
		{
			name: "[Must Not] One Field (full text search must single)",
			filter: Filter{
				MustNot: MockFilterMustNot{
					SingleMust: NewFullTextSearchMust([]string{"1", "2"}),
				},
			},
			expectedQuery: `{"bool":{"must_not":{"bool":{"must":[{"multi_match":{"fields":["SingleMust1"],"max_expansions":1024,"query":"1","type":"phrase_prefix"}},{"multi_match":{"fields":["SingleMust1"],"max_expansions":1024,"query":"2","type":"phrase_prefix"}}]}}}}`,
		},
		{
			name: "[Must Not] One field of each type",
			filter: Filter{
				MustNot: MockFilterMustNot{
					Strings: []string{"1", "2"},
					Ints:    []uint64{1, 2},
					Bool:    ref.Bool(true),
					SingleNested: NewNested(MockSingleNested{
						Slice: []string{"1", "2"},
					}),
					MultiNested: NewNested(MockMultiNested{
						Bool:  ref.Bool(true),
						Slice: []string{"1", "2"},
						Range: &TimeRange{
							From: time.Date(1995, time.March, 1, 11, 35, 19, 29, time.UTC),
							To:   time.Date(2019, time.November, 28, 15, 27, 39, 49, time.UTC),
						},
					}),
					Times: &TimeRange{
						From: time.Date(1995, time.March, 1, 11, 35, 19, 29, time.UTC),
						To:   time.Date(2019, time.November, 28, 15, 27, 39, 49, time.UTC),
					},
					Numbers: &IntRange{
						From: 1,
						To:   100,
					},
					Values: &FloatRange{
						From: 1.0,
						To:   100.0,
					},
					MultiShould:  NewFullTextSearchShould([]string{"1", "2"}),
					SingleShould: NewFullTextSearchShould([]string{"1", "2"}),
					MultiMust:    NewFullTextSearchMust([]string{"1", "2"}),
					SingleMust:   NewFullTextSearchMust([]string{"1", "2"}),
				},
			},
			expectedQuery: `{"bool":{"must_not":[{"terms":{"String":["1","2"]}},{"terms":{"Int":[1,2]}},{"term":{"Bool":true}},{"nested":{"path":"MultiNestedField","query":{"bool":{"must":[{"term":{"MultiNested.Bool":true}},{"terms":{"MultiNested.Slice":["1","2"]}},{"range":{"MultiNested.Range":{"from":"1995-03-01T11:35:19.000000029Z","include_lower":true,"include_upper":true,"to":"2019-11-28T15:27:39.000000049Z"}}}]}}}},{"nested":{"path":"SingleNestedField","query":{"terms":{"SingleNested.Slice":["1","2"]}}}},{"range":{"Time":{"from":"1995-03-01T11:35:19.000000029Z","include_lower":true,"include_upper":true,"to":"2019-11-28T15:27:39.000000049Z"}}},{"range":{"Number":{"from":1,"include_lower":true,"include_upper":true,"to":100}}},{"range":{"Value":{"from":1,"include_lower":true,"include_upper":true,"to":100}}},{"bool":{"should":[{"multi_match":{"fields":["MultiShould1","MultiShould2","MultiShould3"],"max_expansions":1024,"query":"1","type":"phrase_prefix"}},{"multi_match":{"fields":["MultiShould1","MultiShould2","MultiShould3"],"max_expansions":1024,"query":"2","type":"phrase_prefix"}}]}},{"bool":{"should":[{"multi_match":{"fields":["SingleShould1"],"max_expansions":1024,"query":"1","type":"phrase_prefix"}},{"multi_match":{"fields":["SingleShould1"],"max_expansions":1024,"query":"2","type":"phrase_prefix"}}]}},{"bool":{"must":[{"multi_match":{"fields":["MultiMust1","MultiMust2","MultiMust3"],"max_expansions":1024,"query":"1","type":"phrase_prefix"}},{"multi_match":{"fields":["MultiMust1","MultiMust2","MultiMust3"],"max_expansions":1024,"query":"2","type":"phrase_prefix"}}]}},{"bool":{"must":[{"multi_match":{"fields":["SingleMust1"],"max_expansions":1024,"query":"1","type":"phrase_prefix"}},{"multi_match":{"fields":["SingleMust1"],"max_expansions":1024,"query":"2","type":"phrase_prefix"}}]}}]}}`,
		},
		{
			name: "[Exists] One field (bool true)",
			filter: Filter{
				Exists: MockFilterMustExists{
					Bool1: ref.Bool(true),
				},
			},
			expectedQuery: `{"bool":{"must":{"exists":{"field":"Bool1"}}}}`,
		},
		{
			name: "[Exists] Two fields (bool true true)",
			filter: Filter{
				Exists: MockFilterMustExists{
					Bool1: ref.Bool(true),
					Bool2: ref.Bool(true),
				},
			},
			expectedQuery: `{"bool":{"must":[{"exists":{"field":"Bool1"}},{"exists":{"field":"Bool2"}}]}}`,
		},
		{
			name: "[Exists] Two fields (bool false false)",
			filter: Filter{
				Exists: MockFilterMustExists{
					Bool1: ref.Bool(false),
					Bool2: ref.Bool(false),
				},
			},
			expectedQuery: `{"bool":{"must_not":[{"exists":{"field":"Bool1"}},{"exists":{"field":"Bool2"}}]}}`,
		},
		{
			name: "[Exists] Two fields (bool true false)",
			filter: Filter{
				Exists: MockFilterMustExists{
					Bool1: ref.Bool(true),
					Bool2: ref.Bool(false),
				},
			},
			expectedQuery: `{"bool":{"must":{"exists":{"field":"Bool1"}},"must_not":{"exists":{"field":"Bool2"}}}}`,
		},
		{
			name: "[Exists] One field (bool false)",
			filter: Filter{
				Exists: MockFilterMustExists{
					Bool1: ref.Bool(false),
				},
			},
			expectedQuery: `{"bool":{"must_not":{"exists":{"field":"Bool1"}}}}`,
		},
		{
			name: "[Exists] One Field (multi nested true true)",
			filter: Filter{
				Exists: MockFilterMustExists{
					MultiNestedExists: NewNested(MockMultiNestedExists{
						Bool1: ref.Bool(true),
						Bool2: ref.Bool(true),
					}),
				},
			},
			expectedQuery: `{"bool":{"must":{"nested":{"path":"MultiNested","query":{"bool":{"must":[{"exists":{"field":"MultiNested.Bool1"}},{"exists":{"field":"MultiNested.Bool2"}}]}}}}}}`,
		},
		{
			name: "[Exists] One Field (multi nested false false)",
			filter: Filter{
				Exists: MockFilterMustExists{
					MultiNestedExists: NewNested(MockMultiNestedExists{
						Bool1: ref.Bool(false),
						Bool2: ref.Bool(false),
					}),
				},
			},
			expectedQuery: `{"bool":{"must_not":{"nested":{"path":"MultiNested","query":{"bool":{"must":[{"exists":{"field":"MultiNested.Bool1"}},{"exists":{"field":"MultiNested.Bool2"}}]}}}}}}`,
		},
		{
			name: "[Exists] One Field (multi nested true false)",
			filter: Filter{
				Exists: MockFilterMustExists{
					MultiNestedExists: NewNested(MockMultiNestedExists{
						Bool1: ref.Bool(true),
						Bool2: ref.Bool(false),
					}),
				},
			},
			expectedQuery: `{"bool":{"must":{"nested":{"path":"MultiNested","query":{"exists":{"field":"MultiNested.Bool1"}}}},"must_not":{"nested":{"path":"MultiNested","query":{"exists":{"field":"MultiNested.Bool2"}}}}}}`,
		},
		{
			name: "[Exists] One Field (single nested true)",
			filter: Filter{
				Exists: MockFilterMustExists{
					SingleNestedExists: NewNested(MockSingleNestedExists{
						Bool: ref.Bool(true),
					}),
				},
			},
			expectedQuery: `{"bool":{"must":{"nested":{"path":"SingleNested","query":{"exists":{"field":"SingleNested.Bool"}}}}}}`,
		},
		{
			name: "[Error][Must] Nested with invalid value",
			filter: Filter{
				Must: MockFilterMust{
					MockMultiNested: Nested{"a"},
				},
			},
			expectedError: `elasticutil.BuildElasticBoolQuery: getMustQuery: getMustNestedQuery: getMustQuery: [string] filter must be a struct`,
		},
		{
			name: "[Error][Must Not] FullTextSearch with invalid value",
			filter: Filter{
				Must: MockFilterMust{
					MultiMust: FullTextSearchMust{"a"},
				},
			},
			expectedError: `elasticutil.BuildElasticBoolQuery: getMustQuery: getFullTextSearchMustQuery: [MultiMust] full text search value is not supported`,
		},
		{
			name: "[Error][Must Not] FullTextSearchMust with invalid value",
			filter: Filter{
				MustNot: MockFilterMustNot{
					MultiMust: FullTextSearchMust{"a"},
				},
			},
			expectedError: `elasticutil.BuildElasticBoolQuery: getMustQuery: getFullTextSearchMustQuery: [MultiMust] full text search value is not supported`,
		},
		{
			name: "[Error][Must] FullTextSearchShould with invalid value",
			filter: Filter{
				Must: MockFilterMust{
					MultiShould: FullTextSearchShould{"a"},
				},
			},
			expectedError: `elasticutil.BuildElasticBoolQuery: getMustQuery: getFullTextSearchShouldQuery: [MultiShould] full text search value is not supported`,
		},
		{
			name: "[Error][Exists] Nested with invalid value",
			filter: Filter{
				Exists: MockFilterMustExists{
					MultiNestedExists: Nested{"a"},
				},
			},
			expectedError: `elasticutil.BuildElasticBoolQuery: getExistsQuery: getExistsNestedQuery: getExistsQuery: [string] filter must be a struct`,
		},
		{
			name: "[Error][Must] Struct not supported",
			filter: Filter{
				Must: MockInvalidFilter{
					NotSupportedStruct: struct{ A string }{A: "a"},
				},
			},
			expectedError: `elasticutil.BuildElasticBoolQuery: getMustQuery: [NotSupportedStruct] struct is not supported`,
		},
		{
			name: "[Error][Must] Type not supported",
			filter: Filter{
				Must: MockInvalidFilter{
					NotSupportedType: 1,
				},
			},
			expectedError: `elasticutil.BuildElasticBoolQuery: getMustQuery: [NotSupportedType] is of unknown type: int`,
		},
		{
			name: "[Error][Exists] Struct not supported",
			filter: Filter{
				Exists: MockInvalidFilter{
					NotSupportedStruct: struct{ A string }{A: "a"},
				},
			},
			expectedError: `elasticutil.BuildElasticBoolQuery: getExistsQuery: [NotSupportedStruct] struct is not supported`,
		},
		{
			name: "[Error][Exists] Type not supported",
			filter: Filter{
				Exists: MockInvalidFilter{
					NotSupportedType: 1,
				},
			},
			expectedError: `elasticutil.BuildElasticBoolQuery: getExistsQuery: [NotSupportedType] is of unknown type: int`,
		},
		{
			name: "[Error] Custom query",
			filter: Filter{
				Must: MockFilterMust{
					Custom: NewCustomSearch(func() (*elastic.BoolQuery, error) {
						return nil, errors.New("custom error")
					}),
				},
			},
			expectedError: `elasticutil.BuildElasticBoolQuery: getMustQuery: custom error`,
		},
		{
			name: "All filters",
			filter: Filter{
				Must: MockFilterMust{
					Strings: []string{"1", "2"},
					Ints:    []uint64{1, 2},
					Bool:    ref.Bool(true),
					MockSingleNested: NewNested(MockSingleNested{
						Slice: []string{"1", "2"},
					}),
					MockMultiNested: NewNested(MockMultiNested{
						Bool:  ref.Bool(true),
						Slice: []string{"1", "2"},
						Range: &TimeRange{
							From: time.Date(1995, time.March, 1, 11, 35, 19, 29, time.UTC),
							To:   time.Date(2019, time.November, 28, 15, 27, 39, 49, time.UTC),
						},
					}),
					Times: &TimeRange{
						From: time.Date(1995, time.March, 1, 11, 35, 19, 29, time.UTC),
						To:   time.Date(2019, time.November, 28, 15, 27, 39, 49, time.UTC),
					},
					Numbers: &IntRange{
						From: 1,
						To:   100,
					},
					Values: &FloatRange{
						From: 1.0,
						To:   100.0,
					},
					MultiShould:  NewFullTextSearchShould([]string{"1", "2"}),
					SingleShould: NewFullTextSearchShould([]string{"1", "2"}),
					MultiMust:    NewFullTextSearchMust([]string{"1", "2"}),
					SingleMust:   NewFullTextSearchMust([]string{"1", "2"}),
					Custom: NewCustomSearch(func() (*elastic.BoolQuery, error) {
						return elastic.NewBoolQuery().Must(elastic.NewMatchQuery("Strings", "1")), nil
					}),
				},
				MustNot: MockFilterMustNot{
					Strings: []string{"1", "2"},
					Ints:    []uint64{1, 2},
					Bool:    ref.Bool(true),
					SingleNested: NewNested(MockSingleNested{
						Slice: []string{"1", "2"},
					}),
					MultiNested: NewNested(MockMultiNested{
						Bool:  ref.Bool(true),
						Slice: []string{"1", "2"},
						Range: &TimeRange{
							From: time.Date(1995, time.March, 1, 11, 35, 19, 29, time.UTC),
							To:   time.Date(2019, time.November, 28, 15, 27, 39, 49, time.UTC),
						},
					}),
					Times: &TimeRange{
						From: time.Date(1995, time.March, 1, 11, 35, 19, 29, time.UTC),
						To:   time.Date(2019, time.November, 28, 15, 27, 39, 49, time.UTC),
					},
					Numbers: &IntRange{
						From: 1,
						To:   100,
					},
					Values: &FloatRange{
						From: 1.0,
						To:   100.0,
					},
					MultiShould:  NewFullTextSearchShould([]string{"1", "2"}),
					SingleShould: NewFullTextSearchShould([]string{"1", "2"}),
					MultiMust:    NewFullTextSearchMust([]string{"1", "2"}),
					SingleMust:   NewFullTextSearchMust([]string{"1", "2"}),
				},
				Exists: MockFilterMustExists{
					Bool1: ref.Bool(true),
					Bool2: ref.Bool(false),
					MultiNestedExists: NewNested(MockMultiNestedExists{
						Bool1: ref.Bool(false),
						Bool2: ref.Bool(true),
					}),
					SingleNestedExists: NewNested(MockSingleNestedExists{
						Bool: ref.Bool(true),
					}),
				},
			},
			expectedQuery: `{"bool":{"must":[{"terms":{"String":["1","2"]}},{"terms":{"Int":[1,2]}},{"term":{"Bool":true}},{"nested":{"path":"MultiNestedField","query":{"bool":{"must":[{"term":{"MultiNested.Bool":true}},{"terms":{"MultiNested.Slice":["1","2"]}},{"range":{"MultiNested.Range":{"from":"1995-03-01T11:35:19.000000029Z","include_lower":true,"include_upper":true,"to":"2019-11-28T15:27:39.000000049Z"}}}]}}}},{"nested":{"path":"SingleNestedField","query":{"terms":{"SingleNested.Slice":["1","2"]}}}},{"range":{"Time":{"from":"1995-03-01T11:35:19.000000029Z","include_lower":true,"include_upper":true,"to":"2019-11-28T15:27:39.000000049Z"}}},{"range":{"Number":{"from":1,"include_lower":true,"include_upper":true,"to":100}}},{"range":{"Value":{"from":1,"include_lower":true,"include_upper":true,"to":100}}},{"bool":{"should":[{"multi_match":{"fields":["MultiShould1","MultiShould2","MultiShould3"],"max_expansions":1024,"query":"1","type":"phrase_prefix"}},{"multi_match":{"fields":["MultiShould1","MultiShould2","MultiShould3"],"max_expansions":1024,"query":"2","type":"phrase_prefix"}}]}},{"bool":{"should":[{"multi_match":{"fields":["SingleShould1"],"max_expansions":1024,"query":"1","type":"phrase_prefix"}},{"multi_match":{"fields":["SingleShould1"],"max_expansions":1024,"query":"2","type":"phrase_prefix"}}]}},{"bool":{"must":[{"multi_match":{"fields":["MultiMust1","MultiMust2","MultiMust3"],"max_expansions":1024,"query":"1","type":"phrase_prefix"}},{"multi_match":{"fields":["MultiMust1","MultiMust2","MultiMust3"],"max_expansions":1024,"query":"2","type":"phrase_prefix"}}]}},{"bool":{"must":[{"multi_match":{"fields":["SingleMust1"],"max_expansions":1024,"query":"1","type":"phrase_prefix"}},{"multi_match":{"fields":["SingleMust1"],"max_expansions":1024,"query":"2","type":"phrase_prefix"}}]}},{"bool":{"must":{"match":{"Strings":{"query":"1"}}}}},{"nested":{"path":"MultiNested","query":{"exists":{"field":"MultiNested.Bool2"}}}},{"nested":{"path":"SingleNested","query":{"exists":{"field":"SingleNested.Bool"}}}},{"exists":{"field":"Bool1"}}],"must_not":[{"terms":{"String":["1","2"]}},{"terms":{"Int":[1,2]}},{"term":{"Bool":true}},{"nested":{"path":"MultiNestedField","query":{"bool":{"must":[{"term":{"MultiNested.Bool":true}},{"terms":{"MultiNested.Slice":["1","2"]}},{"range":{"MultiNested.Range":{"from":"1995-03-01T11:35:19.000000029Z","include_lower":true,"include_upper":true,"to":"2019-11-28T15:27:39.000000049Z"}}}]}}}},{"nested":{"path":"SingleNestedField","query":{"terms":{"SingleNested.Slice":["1","2"]}}}},{"range":{"Time":{"from":"1995-03-01T11:35:19.000000029Z","include_lower":true,"include_upper":true,"to":"2019-11-28T15:27:39.000000049Z"}}},{"range":{"Number":{"from":1,"include_lower":true,"include_upper":true,"to":100}}},{"range":{"Value":{"from":1,"include_lower":true,"include_upper":true,"to":100}}},{"bool":{"should":[{"multi_match":{"fields":["MultiShould1","MultiShould2","MultiShould3"],"max_expansions":1024,"query":"1","type":"phrase_prefix"}},{"multi_match":{"fields":["MultiShould1","MultiShould2","MultiShould3"],"max_expansions":1024,"query":"2","type":"phrase_prefix"}}]}},{"bool":{"should":[{"multi_match":{"fields":["SingleShould1"],"max_expansions":1024,"query":"1","type":"phrase_prefix"}},{"multi_match":{"fields":["SingleShould1"],"max_expansions":1024,"query":"2","type":"phrase_prefix"}}]}},{"bool":{"must":[{"multi_match":{"fields":["MultiMust1","MultiMust2","MultiMust3"],"max_expansions":1024,"query":"1","type":"phrase_prefix"}},{"multi_match":{"fields":["MultiMust1","MultiMust2","MultiMust3"],"max_expansions":1024,"query":"2","type":"phrase_prefix"}}]}},{"bool":{"must":[{"multi_match":{"fields":["SingleMust1"],"max_expansions":1024,"query":"1","type":"phrase_prefix"}},{"multi_match":{"fields":["SingleMust1"],"max_expansions":1024,"query":"2","type":"phrase_prefix"}}]}},{"nested":{"path":"MultiNested","query":{"exists":{"field":"MultiNested.Bool1"}}}},{"exists":{"field":"Bool2"}}]}}`,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			assert.NotPanics(t, func() {
				query, err := BuildElasticBoolQuery(context.Background(), test.filter)
				if test.expectedError == "" {
					assert.NoError(t, err)
					assert.Equal(t, test.expectedQuery, MarshalQuery(query))
				} else {
					assert.EqualError(t, err, test.expectedError)
				}
			})
		})
	}
}

func Test_MarshalQuery(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		query    elastic.Query
		expected string
	}{
		{
			name:     "nil query",
			query:    nil,
			expected: "",
		},
		{
			name:     "error query",
			query:    &mockQuery{value: "a", err: errors.New("error query")},
			expected: "error query",
		},
		{
			name:     "error unmarshal",
			query:    &mockQuery{value: make(chan int)},
			expected: "json: unsupported type: chan int",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			assert.NotPanics(t, func() {
				assert.Equal(t, test.expected, MarshalQuery(test.query))
			})
		})
	}
}

type mockQuery struct {
	value interface{}
	err   error
}

func (m *mockQuery) Source() (interface{}, error) {
	return m.value, m.err
}

type MockFilterMust struct {
	Strings []string `es:"String"`
	Ints    []uint64 `es:"Int"`
	Bool    *bool

	MockMultiNested  Nested `es:"MultiNestedField"`
	MockSingleNested Nested `es:"SingleNestedField"`

	Times   *TimeRange  `es:"Time"`
	Numbers *IntRange   `es:"Number"`
	Values  *FloatRange `es:"Value"`

	MultiShould  FullTextSearchShould `es:"MultiShould1,MultiShould2,MultiShould3"`
	SingleShould FullTextSearchShould `es:"SingleShould1"`
	MultiMust    FullTextSearchMust   `es:"MultiMust1,MultiMust2,MultiMust3"`
	SingleMust   FullTextSearchMust   `es:"SingleMust1"`
	Custom       CustomSearch         `es:"Custom"`
}

type MockCustomQuery struct{}

type MockFilterMustNot struct {
	Strings []string `es:"String"`
	Ints    []uint64 `es:"Int"`
	Bool    *bool

	MultiNested  Nested `es:"MultiNestedField"`
	SingleNested Nested `es:"SingleNestedField"`

	Times   *TimeRange  `es:"Time"`
	Numbers *IntRange   `es:"Number"`
	Values  *FloatRange `es:"Value"`

	MultiShould  FullTextSearchShould `es:"MultiShould1,MultiShould2,MultiShould3"`
	SingleShould FullTextSearchShould `es:"SingleShould1"`
	MultiMust    FullTextSearchMust   `es:"MultiMust1,MultiMust2,MultiMust3"`
	SingleMust   FullTextSearchMust   `es:"SingleMust1"`
}

type MockFilterMustExists struct {
	MultiNestedExists  Nested `es:"MultiNested"`
	SingleNestedExists Nested `es:"SingleNested"`
	Bool1              *bool
	Bool2              *bool
}

type MockSingleNested struct {
	Slice []string `es:"SingleNested.Slice"`
}

type MockMultiNested struct {
	Bool  *bool      `es:"MultiNested.Bool"`
	Slice []string   `es:"MultiNested.Slice"`
	Range *TimeRange `es:"MultiNested.Range"`
}

type MockMultiNestedExists struct {
	Bool1 *bool `es:"MultiNested.Bool1"`
	Bool2 *bool `es:"MultiNested.Bool2"`
}

type MockSingleNestedExists struct {
	Bool *bool `es:"SingleNested.Bool"`
}

type MockInvalidFilter struct {
	NotSupportedStruct struct {
		A string
	}
	NotSupportedType int
}
