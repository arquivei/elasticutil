package v7

import (
	"encoding/json"
	"time"

	"github.com/arquivei/elasticutil/official/v7/querybuilders"
)

// Filter is a struct that will be transformed in a olivere/elastic's query.
//
// "Must" and "MustNot" is for the terms, range and multi match query.
// "Exists" is for the exists query.
// For nested queries, uses the Nested type.
type Filter struct {
	Must    any
	MustNot any
	Exists  any
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
	Payload any
}

// NewNested creates a Nested struct with the given payload.
func NewNested(payload any) Nested {
	return Nested{payload}
}

func (m Nested) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.Payload)
}

// FullTextSearchMust represents a Must's Full Text Search.
type FullTextSearchMust struct {
	Payload any
}

// NewFullTextSearchMust creates a FullTextSearchMust struct with the given payload.
func NewFullTextSearchMust(payload any) FullTextSearchMust {
	return FullTextSearchMust{payload}
}

func (m FullTextSearchMust) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.Payload)
}

// FullTextSearchShould represents a Should's Full Text Search.
type FullTextSearchShould struct {
	Payload any
}

// NewFullTextSearchShould creates a FullTextSearchShould struct with the given payload.
func NewFullTextSearchShould(payload any) FullTextSearchShould {
	return FullTextSearchShould{payload}
}

// CustomSearch is the struct that contains the CustomQuery function.
type CustomSearch struct {
	GetQuery CustomQuery
	Payload  any
}

// CustomQuery is the type function that will return the custom query.
type CustomQuery func() (querybuilders.Query, error)

// NewCustomSearch creates a new CustomSearch instance with the provided query and payload.
// The @payload is any serializable data that will be used for custom JSON marshaling.
func NewCustomSearch(query CustomQuery, payload any) CustomSearch {
	return CustomSearch{
		GetQuery: query,
		Payload:  payload,
	}
}

func (m CustomSearch) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.Payload)
}

// MultiMatchSearchShould Represents a Should's Multi Match Search.
type MultiMatchSearchShould struct {
	Payload any
}

// NewMultiMatchSearchShould creates a MultiMatchSearchShould struct with the given payload.
func NewMultiMatchSearchShould(payload any) MultiMatchSearchShould {
	return MultiMatchSearchShould{payload}
}

func (m MultiMatchSearchShould) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.Payload)
}
