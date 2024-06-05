package v7

import "github.com/arquivei/foundationkit/errors"

// ---------- Errors

// ErrNotAllShardsReplied is returned when no all elasticsearch's shards
// successfully reply.
var ErrNotAllShardsReplied = errors.New("not all shards replied")

// ErrNilResponse is returned when elasticsearch returns no error, but
// returns a nil response.
var ErrNilResponse = errors.New("nil response")

// ---------- Codes

// ErrCodeBadRequest is returned when elasticsearch returns a
// status 400.
var ErrCodeBadRequest = errors.Code("bad request")

// ErrCodeBadGateway is returned when elasticsearch client returns an error.
var ErrCodeBadGateway = errors.Code("bad gateway")

// ErrCodeUnexpectedResponse is returned when the elasticsearch returned
// unexpected data
var ErrCodeUnexpectedResponse = errors.Code("unexpected response")

// ----------

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
