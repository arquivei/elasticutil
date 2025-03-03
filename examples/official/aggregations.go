package main

import elasticutil "github.com/arquivei/elasticutil/official/v7"

func createAggregations() elasticutil.RequestAggregation {
	return elasticutil.RequestAggregation{
		Metrics: []elasticutil.RequestMetricAggregation{
			{
				Name:  "sum_aggregation_name",
				Type:  "sum",
				Field: "sum_field",
			},
			{
				Name:  "max_aggregation_name",
				Type:  "max",
				Field: "max_field",
			},
			{
				Name:  "max_aggregation_name_2",
				Type:  "max",
				Field: "max_field_2",
			},
			{
				Name:  "min_aggregation_name",
				Type:  "min",
				Field: "min_field",
			},
			{
				Name:  "count_aggregation",
				Type:  "count",
				Field: "count_field",
			},
		},
		Buckets: []elasticutil.RequestBucketAggregation{
			{
				Name:  "some_term_agg",
				Field: "CompanyRole",
				Type:  "term",
				MetricsSubAgg: []elasticutil.RequestMetricAggregation{
					{
						Name:  "min_agg_in_term_name",
						Field: "min_in_term_field",
						Type:  "min",
					},
				},
			},
			{
				Name:  "some_histogram_agg",
				Field: "EmissionDate",
				Type:  "monthlyhistogram",
				MetricsSubAgg: []elasticutil.RequestMetricAggregation{
					{
						Name:  "max_agg_in_histogram_name",
						Field: "max_in_histogram_field",
						Type:  "max",
					},
					{
						Name:  "min_agg_in_histogram_name",
						Field: "min_in_histogram_field",
						Type:  "min",
					},
				},
			},
		},
	}
}
