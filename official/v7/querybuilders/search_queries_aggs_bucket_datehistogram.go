package querybuilders

// BucketDateHistogramAggregation is a multi-bucket values source based aggregation
// that can be applied on date values extracted from the documents.
// It dynamically builds fixed size (a.k.a. interval) buckets over the
// values.
// See: https://www.elastic.co/guide/en/elasticsearch/reference/7.0/search-aggregations-bucket-datehistogram-aggregation.html
type BucketDateHistogramAggregation struct {
	name                   string
	field                  string
	subMetricsAggregations []MetricAggregation
	interval               string
	minDocCount            *int //optional
}

// NewBucketHistogramAggregation returns a pointer to a new BucketDateHistogramAggregation
// with an empty subMetricsAggregations list
func NewBucketHistogramAggregation(name, field, interval string) *BucketDateHistogramAggregation {
	return &BucketDateHistogramAggregation{
		name:                   name,
		field:                  field,
		interval:               interval,
		subMetricsAggregations: make([]MetricAggregation, 0),
	}
}

// SubMetrics appends a MetricAggregation list to the subMetricsAggregations field
func (a *BucketDateHistogramAggregation) SubMetrics(
	agg ...MetricAggregation,
) *BucketDateHistogramAggregation {
	a.subMetricsAggregations = append(a.subMetricsAggregations, agg...)
	return a
}

// MinDocCount sets the minimum document count per bucket.
// Buckets with less documents than this min value will not be returned.
func (a *BucketDateHistogramAggregation) MinDocCount(minDocCount int) *BucketDateHistogramAggregation {
	a.minDocCount = &minDocCount
	return a
}

// Source is a helper function used to build elastic query
// it returns a map that can be marshalled into json string
func (a *BucketDateHistogramAggregation) Source() (interface{}, error) {
	// This is the output of the source function:
	// It returns the json string that can be used in elastic query
	// Example:
	// {
	//     "aggs" : {
	//         "articles_over_time" : {
	//             "date_histogram" : {
	//                 "field" : "date",
	//                 "interval" : "month"
	//             }
	//         }
	//     }
	// }
	//
	// This method returns only the { "date_histogram" : { ... } } part.

	source := make(map[string]interface{})
	opts := make(map[string]interface{})

	// date_histogram is deprecated and future updates to the Elasticsearch client
	// will necessitate a migration to either fixed_interval or calendar_interval.
	source["date_histogram"] = opts

	if a.field != "" {
		opts["field"] = a.field
	}
	if s := a.interval; s != "" {
		opts["interval"] = s
	}
	if a.minDocCount != nil {
		opts["min_doc_count"] = *a.minDocCount
	}

	if len(a.subMetricsAggregations) > 0 {
		aggsMap := make(map[string]interface{})
		source["aggs"] = aggsMap
		for _, subAggs := range a.subMetricsAggregations {
			src, err := subAggs.Source()
			if err != nil {
				return nil, err
			}
			aggsMap[subAggs.name] = src
		}
	}

	return source, nil
}
