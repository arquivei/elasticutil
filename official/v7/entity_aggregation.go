package v7

// RequestAggregation contains all the aggregations that will be requested to elastic search
// This struct will be transformed in an elastic query
type RequestAggregation struct {
	Metrics []RequestMetricAggregation
	Buckets []RequestBucketAggregation
}

// RequestMetricAggregation is an aggregation that computes metrics over a set of documents
type RequestMetricAggregation struct {
	Name  string
	Type  string
	Field string
}

// RequestBucketAggregation is a aggregation that groups documents using either a histogram or a terms search.
// It can contains a metrics aggregation that will compute metrics over all documents inside a bucket
type RequestBucketAggregation struct {
	Name          string
	Type          string
	Field         string
	MetricsSubAgg []RequestMetricAggregation
}

// ResponseAggregation contains all the aggregations returned from elastic search
type ResponseAggregation struct {
	MetricAggregations []ResponseMetricAggregation
	BucketAggregations []ResponseBucketAggregation
}

// AppendAggregation adds a parsed aggregation to the appropriate slice
// in the ResponseAggregation. It handles ResponseMetricAggregation and
// ResponseBucketAggregation types. If the input 'aggs' is of any other type,
// it is ignored.
func (a ResponseAggregation) AppendAggregation(
	aggs interface{},
) ResponseAggregation {
	if metricAggs, ok := aggs.(ResponseMetricAggregation); ok {
		a.MetricAggregations = append(a.MetricAggregations, metricAggs)
	} else if bucketAggs, ok := aggs.(ResponseBucketAggregation); ok {
		a.BucketAggregations = append(a.BucketAggregations, bucketAggs)
	}
	return a
}

type ResponseMetricAggregation struct {
	Name  string
	Value any
}

type ResponseBucketAggregation struct {
	Name    string
	Buckets []ResponseBucket
}

type ResponseBucket struct {
	Key                 string
	DocCount            int
	MetricsAggregations []ResponseMetricAggregation
}
