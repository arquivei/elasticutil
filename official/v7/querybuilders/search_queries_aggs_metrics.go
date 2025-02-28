package querybuilders

// MetricAggregation is a single-value metrics aggregation that operates on
// numeric values that are extracted from the aggregated documents.
// These values are extracted from specific numeric fields in the documents
// See: https://www.elastic.co/guide/en/elasticsearch/reference/7.0/search-aggregations-metrics.html
type MetricAggregation struct {
	name      string
	field     string
	operation string
}

// NewMinAggregation creates a new min metric aggregation
func NewMinAggregation(name, field string) MetricAggregation {
	return MetricAggregation{name, field, "min"}
}

// NewSumAggregation creates a new sum metric aggregation
func NewSumAggregation(name, field string) MetricAggregation {
	return MetricAggregation{name, field, "sum"}
}

// NewMaxAggregation creates a new max metric aggregation
func NewMaxAggregation(name, field string) MetricAggregation {
	return MetricAggregation{name, field, "max"}
}

// NewCountAggregation creates a new count metric aggregation
func NewCountAggregation(name, field string) MetricAggregation {
	return MetricAggregation{name, field, "value_count"}
}

// Source is a helper function used to build elastic query
// it returns a map that can be marshalled into json string
func (a *MetricAggregation) Source() (interface{}, error) {
	// This is the output of the source function:
	// It returns the json string that can be used in elastic query
	// Example:
	//	{
	//    "aggs" : {
	//      "intraday_return" : { "operation" : { "field" : "change" } }
	//    }
	//	}
	// This method returns only the { "operation" : { "field" : "change" } } part.

	source := make(map[string]interface{})
	opts := make(map[string]interface{})
	source[a.operation] = opts

	if a.field != "" {
		opts["field"] = a.field
	}

	return source, nil
}
