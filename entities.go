package elasticutil

import "time"

type Ranges interface {
	time.Time | uint64 | float64
}

// TimeRange represents a time range with a beginning and an end.
type TimeRange struct {
	From time.Time
	To   time.Time
}

// IntRange represents an int range with a beginning and an end.
type IntRange struct {
	From uint64
	To   uint64
}

// FloatRange represents a float range with a beginning and an end.
type FloatRange struct {
	From float64
	To   float64
}

// Nested represents a nested query.
type Nested struct {
	payload interface{}
}

// NewNested creates a Nested struct with the given payload.
func NewNested(payload interface{}) Nested {
	return Nested{payload}
}

// FullTextSearchMust Represents a Must's Full Text Search.
type FullTextSearchMust struct {
	payload interface{}
}

// NewFullTextSearchMust creates a FullTextSearchMust struct with the given payload.
func NewFullTextSearchMust(payload interface{}) FullTextSearchMust {
	return FullTextSearchMust{payload}
}

// FullTextSearchMust Represents a Should's Full Text Search.
type FullTextSearchShould struct {
	payload interface{}
}

// NewFullTextSearchShould creates a FullTextSearchShould struct with the given payload.
func NewFullTextSearchShould(payload interface{}) FullTextSearchShould {
	return FullTextSearchShould{payload}
}
