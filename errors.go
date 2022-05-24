package elasticutil

import "github.com/arquivei/foundationkit/errors"

var (
	// ErrNotAllShardsReplied ErrNotAllShardsReplied
	ErrNotAllShardsReplied = errors.New("not all shards replied")
)

func filterMustBeAStructError(kind string) error {
	return errors.New("[" + kind + "] filter must be a struct")
}

func structNotSupportedError(name string) error {
	return errors.New(name + " struct is not supported")
}

func typeNotSupportedError(name, t string) error {
	return errors.New(name + " is of unkown type: " + t)
}

func fullTextSearchTypeNotSupported(name string) error {
	return errors.New(name + " full text search is not supported")
}
