package main

import (
	"context"
	"fmt"
	"time"

	"github.com/arquivei/elasticutil"
	"github.com/olivere/elastic/v7"
)

type ExampleFilterMust struct {
	Names            []string `es:"Name"`
	SocialNames      []string `es:"SocialName"`
	Ages             []uint64 `es:"Age"`
	HasCovid         *bool
	CreatedAt        *elasticutil.TimeRange
	AgeRange         *elasticutil.IntRange            `es:"Age"`
	CovidInfo        elasticutil.Nested               `es:"Covid"`
	NameOrSocialName elasticutil.FullTextSearchShould `es:"Name,SocialName"`
	MyCustomSearch   elasticutil.CustomSearch
}

type ExampleFilterExists struct {
	HasCovidInfo elasticutil.Nested `es:"Covid"`
	HasAge       *bool              `es:"Age"`
}

type ExampleCovidInfo struct {
	HasCovidInfo     *bool                  `es:"Covid"`
	Symptoms         []string               `es:"Covid.Symptom"`
	FirstSymptomDate *elasticutil.TimeRange `es:"Covid.Date"`
}

func main() {
	requestFilter := elasticutil.Filter{
		Must: ExampleFilterMust{
			Names:    []string{"John", "Mary"},
			Ages:     []uint64{16, 17, 18, 25, 26},
			HasCovid: refBool(true),
			CovidInfo: elasticutil.NewNested(
				ExampleCovidInfo{
					Symptoms: []string{"cough"},
					FirstSymptomDate: &elasticutil.TimeRange{
						From: time.Date(2019, time.November, 28, 15, 27, 39, 49, time.UTC),
						To:   time.Date(2020, time.November, 28, 15, 27, 39, 49, time.UTC),
					},
				},
			),
			CreatedAt: &elasticutil.TimeRange{
				From: time.Date(2020, time.November, 28, 15, 27, 39, 49, time.UTC),
				To:   time.Date(2021, time.November, 28, 15, 27, 39, 49, time.UTC),
			},
			AgeRange: &elasticutil.IntRange{
				From: 15,
				To:   30,
			},
			NameOrSocialName: elasticutil.NewFullTextSearchShould([]string{"John", "Mary", "Rebecca"}),
			MyCustomSearch: elasticutil.NewCustomSearch(func() (*elastic.BoolQuery, error) {
				return elastic.NewBoolQuery().Must(elastic.NewMatchQuery("Name", "John")), nil
			}),
		},
		MustNot: ExampleFilterMust{
			Names: []string{"Lary"},
			AgeRange: &elasticutil.IntRange{
				From: 29,
				To:   30,
			},
		},
		Exists: ExampleFilterExists{
			HasCovidInfo: elasticutil.NewNested(
				ExampleCovidInfo{
					HasCovidInfo: refBool(true),
				},
			),
			HasAge: refBool(true),
		},
	}

	elasticQuery, err := elasticutil.BuildElasticBoolQuery(context.Background(), requestFilter)
	if err != nil {
		panic(err)
	}

	verboseElasticQuery := elasticutil.MarshalQuery(elasticQuery)

	fmt.Println(verboseElasticQuery)
}

func refBool(b bool) *bool {
	return &b
}

/*
Expected output:

{
  "bool": {
    "must": [
      {
        "terms": {
          "Name": [
            "John",
            "Mary"
          ]
        }
      },
      {
        "terms": {
          "Age": [
            16,
            17,
            18,
            25,
            26
          ]
        }
      },
      {
        "term": {
          "HasCovid": true
        }
      },
      {
        "range": {
          "CreatedAt": {
            "from": "2020-11-28T15:27:39.000000049Z",
            "include_lower": true,
            "include_upper": true,
            "to": "2021-11-28T15:27:39.000000049Z"
          }
        }
      },
      {
        "range": {
          "Age": {
            "from": 15,
            "include_lower": true,
            "include_upper": true,
            "to": 30
          }
        }
      },
      {
        "nested": {
          "path": "Covid",
          "query": {
            "bool": {
              "must": [
                {
                  "terms": {
                    "Covid.Symptom": [
                      "cough"
                    ]
                  }
                },
                {
                  "range": {
                    "Covid.Date": {
                      "from": "2019-11-28T15:27:39.000000049Z",
                      "include_lower": true,
                      "include_upper": true,
                      "to": "2020-11-28T15:27:39.000000049Z"
                    }
                  }
                }
              ]
            }
          }
        }
      },
      {
        "bool": {
          "should": [
            {
              "multi_match": {
                "fields": [
                  "Name",
                  "SocialName"
                ],
                "max_expansions": 1024,
                "query": "John",
                "type": "phrase_prefix"
              }
            },
            {
              "multi_match": {
                "fields": [
                  "Name",
                  "SocialName"
                ],
                "max_expansions": 1024,
                "query": "Mary",
                "type": "phrase_prefix"
              }
            },
            {
              "multi_match": {
                "fields": [
                  "Name",
                  "SocialName"
                ],
                "max_expansions": 1024,
                "query": "Rebecca",
                "type": "phrase_prefix"
              }
            }
          ]
        }
      },
      {
        "bool": {
          "must": {
            "match": {
              "Name": {
                "query": "John"
              }
            }
          }
        }
      },
      {
        "nested": {
          "path": "Covid",
          "query": {
            "exists": {
              "field": "Covid"
            }
          }
        }
      },
      {
        "exists": {
          "field": "Age"
        }
      }
    ],
    "must_not": [
      {
        "terms": {
          "Name": [
            "Lary"
          ]
        }
      },
      {
        "range": {
          "Age": {
            "from": 29,
            "include_lower": true,
            "include_upper": true,
            "to": 30
          }
        }
      }
    ]
  }
}

*/
