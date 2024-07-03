package v7

import (
	"time"

	"github.com/arquivei/elasticutil/official/v7/querybuilders"
)

// Filter is a struct that eill be transformed in a olivere/elastic's query.
//
// "Must" and "MustNot" is for the terms, range and multi match query.
// "Exists" is for the exists query.
// For nested queries, uses the Nested type.
type Filter struct {
	Must    interface{}
	MustNot interface{}
	Exists  interface{}
}

// Ranges is an interface that represents one of the following range type:
// time.time, uint64 and float64.
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

// CustomSearch is the struct that contains the CustomQuery function.
type CustomSearch struct {
	GetQuery CustomQuery
}

// CustomQuery is the type function that will return the custom query.
type CustomQuery func() (querybuilders.Query, error)

// NewCustomSearch creates a CustomSearch struct with the given CustomQuery function.
func NewCustomSearch(query CustomQuery) CustomSearch {
	return CustomSearch{query}
}

// MultiMatchSearchShould Represents a Should's Multi Match Search.
type MultiMatchSearchShould struct {
	payload interface{}
}

// NewMultiMatchSearchShould creates a MultiMatchSearchShould struct with the given payload.
func NewMultiMatchSearchShould(payload interface{}) MultiMatchSearchShould {
	return MultiMatchSearchShould{payload}
}
