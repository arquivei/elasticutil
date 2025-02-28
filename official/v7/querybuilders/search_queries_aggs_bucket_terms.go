package querybuilders

var showTermDocCountError bool = true

// BucketTermsAggregation is a multi-bucket value source based aggregation
// where buckets are dynamically built - one per unique value.
//
// See: http://www.elastic.co/guide/en/elasticsearch/reference/7.0/search-aggregations-bucket-terms-aggregation.html
type BucketTermsAggregation struct {
	name                   string
	field                  string
	subMetricsAggregations []MetricAggregation
	size                   *int  //optional
	minDocCount            *int  //optional
	showTermDocCountError  *bool //always true
}

// NewBucketTermsAggregation returns a pointer to a new BucketTermsAggregation
// with showTermDocCountError always true and an empty subMetricsAggregations list
func NewBucketTermsAggregation(name, field string) *BucketTermsAggregation {
	return &BucketTermsAggregation{
		name:                   name,
		field:                  field,
		subMetricsAggregations: make([]MetricAggregation, 0),
		showTermDocCountError:  &showTermDocCountError,
	}
}

// SubMetrics appends a MetricAggregation list to the subMetricsAggregations field
func (a *BucketTermsAggregation) SubMetrics(
	aggs ...MetricAggregation,
) *BucketTermsAggregation {
	a.subMetricsAggregations = append(a.subMetricsAggregations, aggs...)
	return a
}

// Size sets the size field of the BucketTermsAggregation
func (a *BucketTermsAggregation) Size(size int) *BucketTermsAggregation {
	a.size = &size
	return a
}

// MinDocCount sets the minimum document count per bucket.
// Buckets with less documents than this min value will not be returned.
func (a *BucketTermsAggregation) MinDocCount(minDocCount int) *BucketTermsAggregation {
	a.minDocCount = &minDocCount
	return a
}

// Source is a helper function used to build elastic query
// it returns a map that can be marshalled into json string
func (a *BucketTermsAggregation) Source() (interface{}, error) {
	// This is the output of the source function:
	// It returns the json string that can be used in elastic query
	// Example:
	//	{
	//    "aggs" : {
	//      "genders" : {
	//        "terms" : { "field" : "gender" }
	//      }
	//    }
	//	}
	// This method returns only the { "terms" : { "field" : "gender" } } part.

	source := make(map[string]interface{})
	opts := make(map[string]interface{})
	source["terms"] = opts

	if a.field != "" {
		opts["field"] = a.field
	}
	if a.size != nil && *a.size >= 0 {
		opts["size"] = *a.size
	}
	if a.minDocCount != nil && *a.minDocCount >= 0 {
		opts["min_doc_count"] = *a.minDocCount
	}
	if a.showTermDocCountError != nil {
		opts["show_term_doc_count_error"] = *a.showTermDocCountError
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
