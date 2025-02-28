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
