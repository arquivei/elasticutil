package querybuilders

// AggsQuery is used to build an aggregation elastic query
// For more details, see:
// https://www.elastic.co/guide/en/elasticsearch/reference/7.0/search-aggregations.html
type AggsQuery struct {
	metricAggs              []MetricAggregation
	bucketTermsAggs         []BucketTermsAggregation
	bucketDateHistogramAggs []BucketDateHistogramAggregation
}

// Creates a new aggs query with empty aggregations lists
func NewAggsQuery() *AggsQuery {
	return &AggsQuery{
		metricAggs:              make([]MetricAggregation, 0),
		bucketTermsAggs:         make([]BucketTermsAggregation, 0),
		bucketDateHistogramAggs: make([]BucketDateHistogramAggregation, 0),
	}
}

// BucketTerms appends a MetricAggregation list to the metricAggs field
func (q *AggsQuery) Metric(aggs ...MetricAggregation) *AggsQuery {
	q.metricAggs = append(q.metricAggs, aggs...)
	return q
}

// BucketTerms appends a BucketTermsAggregation list to the bucketTermsAggs field
func (q *AggsQuery) BucketTerms(aggs ...BucketTermsAggregation) *AggsQuery {
	q.bucketTermsAggs = append(q.bucketTermsAggs, aggs...)
	return q
}

// BucketDateHistogram appends a BucketDateHistogramAggregation list to the bucketDateHistogramAggs field
func (q *AggsQuery) BucketDateHistogram(aggs ...BucketDateHistogramAggregation) *AggsQuery {
	q.bucketDateHistogramAggs = append(q.bucketDateHistogramAggs, aggs...)
	return q
}

// Source is a helper function used to build elastic query
// it returns a map that can be marshalled into json string
func (q *AggsQuery) Source() (interface{}, error) {
	// This is the output of the source function:
	// It returns the json string that can be used in elastic query
	// {
	//     "sum_agg_name": {
	//         "sum": {
	//             "field": "field_name"
	//         }
	//     },
	//     "min_agg_name": {
	//         "min": {
	//             "field": "field_name"
	//         }
	//     },
	//     "max_agg_name": {
	//         "max": {
	//             "field": "field_name"
	//         }
	//     },
	//     "count_agg_name": {
	//         "value_count": {
	//             "field": "field_name"
	//         }
	//     },
	//     "bucket_histogram_name": {
	//         "date_histogram": {
	//             "field": "EmissionDate",
	//             "interval": "month",
	//             "min_doc_count": 300
	//         },
	//         "aggs": {
	//             "max_agg_inside_bucket_histogram_name": {
	//                 "max": {
	//                     "field": "field_name"
	//                 }
	//             },
	//             "min_agg_inside_bucket_histogram_name": {
	//                 "min": {
	//                     "field": "field_name"
	//                 }
	//             }
	//         }
	//     },
	//     "bucket_term_name": {
	//         "terms": {
	//             "field": "field_name",
	//             "size": 10,
	//             "show_term_doc_count_error": true
	//         },
	//         "aggs": {
	//             "sum_agg_inside_bucket_term_name": {
	//                 "sum": {
	//                     "field": "field_name"
	//                 }
	//             }
	//         }
	//     }
	// }

	aggs := make(map[string]interface{})

	for _, metricAgg := range q.metricAggs {
		src, err := metricAgg.Source()
		if err != nil {
			return nil, err
		}
		aggs[metricAgg.name] = src
	}

	for _, bucketTermsAgg := range q.bucketTermsAggs {
		src, err := bucketTermsAgg.Source()
		if err != nil {
			return nil, err
		}
		aggs[bucketTermsAgg.name] = src
	}

	for _, bucketHistogramAggs := range q.bucketDateHistogramAggs {
		src, err := bucketHistogramAggs.Source()
		if err != nil {
			return nil, err
		}
		aggs[bucketHistogramAggs.name] = src
	}

	return aggs, nil
}
