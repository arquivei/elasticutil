package main

import (
	"time"

	elasticutil "github.com/arquivei/elasticutil/official/v7"
	"github.com/arquivei/elasticutil/official/v7/querybuilders"
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

func createFilter() elasticutil.Filter {
	return elasticutil.Filter{
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
			MyCustomSearch: elasticutil.NewCustomSearch(func() (querybuilders.Query, error) {
				return querybuilders.NewBoolQuery().Must(querybuilders.NewTermQuery("Name", "John")), nil
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
}

func refBool(b bool) *bool {
	return &b
}
