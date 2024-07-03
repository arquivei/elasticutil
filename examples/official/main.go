package main

import (
	"context"
	"fmt"

	elasticutil "github.com/arquivei/elasticutil/official/v7"
)

func main() {
	client := elasticutil.MustNewClient("")

	response, err := client.Search(
		context.Background(),
		elasticutil.SearchConfig{
			Filter:            createFilter(),
			Indexes:           []string{""},
			Size:              5,
			IgnoreUnavailable: true,
			AllowNoIndices:    true,
			TrackTotalHits:    true,
			Sort: elasticutil.Sorters{
				Sorters: []elasticutil.Sorter{
					{
						Field:     "ID",
						Ascending: true,
					},
				},
			},
			SearchAfter: "",
		},
	)

	fmt.Println(response, err)

}
