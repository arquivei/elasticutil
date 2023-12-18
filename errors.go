package elasticutil

import "github.com/arquivei/foundationkit/errors"

// ErrNotAllShardsReplied is returned when no all elasticsearch's shards
// successfully reply.
var ErrNotAllShardsReplied = errors.New("not all shards replied")

func filterMustBeAStructError(kind string) error {
	return errors.New("[" + kind + "] filter must be a struct")
}

func structNotSupportedError(name string) error {
	return errors.New("[" + name + "] struct is not supported")
}

func typeNotSupportedError(name, t string) error {
	return errors.New("[" + name + "] is of unknown type: " + t)
}

func fullTextSearchTypeNotSupported(name string) error {
	return errors.New("[" + name + "] full text search value is not supported")
}

func multiMatchSearchTypeNotSupported(name string) error {
	return errors.New("[" + name + "] multi match search value is not supported")
}
