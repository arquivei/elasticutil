package querybuilders

// Query is an interface for every type of Elasticsearch's query.
type Query interface {
	Source() (interface{}, error)
}
