package elasticutil

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"time"

	"github.com/arquivei/foundationkit/errors"
	"github.com/olivere/elastic/v7"
)

const maxExpansions = 1024

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

// BuildElasticBoolQuery builds a olivere/elastic's query based on Filter.
func BuildElasticBoolQuery(
	ctx context.Context,
	filter Filter,
) (elastic.Query, error) {
	const op = errors.Op("elasticutil.BuildElasticBoolQuery")

	var mustQueries, mustNotQueries, existsQueries, notExistsQueries []elastic.Query

	if filter.Must != nil {
		var err error
		mustQueries, err = getMustQuery(ctx, filter.Must)
		if err != nil {
			return nil, errors.E(op, err)
		}
	}

	if filter.MustNot != nil {
		var err error
		mustNotQueries, err = getMustQuery(ctx, filter.MustNot)
		if err != nil {
			return nil, errors.E(op, err)
		}
	}

	if filter.Exists != nil {
		var err error
		existsQueries, notExistsQueries, err = getExistsQuery(filter.Exists)
		if err != nil {
			return nil, errors.E(op, err)
		}
	}

	if shouldReturnMatchAllQuery(
		mustQueries,
		mustNotQueries,
		existsQueries,
		notExistsQueries,
	) {
		return elastic.NewMatchAllQuery(), nil
	}

	if shouldReturnOnlyMustQuery(
		mustQueries,
		mustNotQueries,
		existsQueries,
		notExistsQueries,
	) {
		return mustQueries[0], nil
	}

	return getBoolQuery(
		mustQueries,
		mustNotQueries,
		existsQueries,
		notExistsQueries,
	), nil
}

// MarshalQuery transforms a olivere/elastic's query in a string for log and test
// purpose.
func MarshalQuery(query elastic.Query) string {
	if query != nil {
		q, e := query.Source()
		if e != nil {
			return e.Error()
		}
		qs, e := json.Marshal(q)
		if e != nil {
			return e.Error()
		}
		return string(qs)
	}
	return ""
}

// nolint: funlen, cyclop, gocognit
func getMustQuery(
	ctx context.Context,
	filter interface{},
) ([]elastic.Query, error) {
	const op = errors.Op("getMustQuery")

	var queries []elastic.Query

	rv := reflect.ValueOf(filter)
	if rv.Kind() != reflect.Struct {
		return nil, errors.E(op, filterMustBeAStructError(rv.Kind().String()))
	}

	nFields := rv.Type().NumField()
	for i := 0; i < nFields; i++ {
		// Get value, type and tags
		fvalue := rv.Field(i)
		ftype := rv.Type().Field(i)
		fnames := parseFieldNames(ftype.Tag.Get("es"))

		// Skip zero values
		if fvalue.IsZero() {
			continue
		}

		// Rename field if specified a new name inside the tag
		structFieldName := ftype.Name
		names := []string{structFieldName}
		if len(fnames) > 0 {
			names = fnames
		}

		// Fix kind and value if field is a pointer
		fkind := ftype.Type.Kind()
		if fkind == reflect.Ptr {
			fvalue = fvalue.Elem()
			fkind = fvalue.Kind()
		}

		switch fkind {
		case reflect.Slice:
			if rv.Field(i).Len() == 0 {
				continue
			}
			switch ftype.Type.Elem().String() {
			case "string":
				queries = append(queries, elastic.NewTermsQuery(names[0],
					stringsToInterfaces(fvalue.Interface())...))
			case "uint64":
				queries = append(queries, elastic.NewTermsQuery(names[0],
					utin64ToInterfaces(fvalue.Interface())...))
			}
		case reflect.Bool:
			queries = append(queries, elastic.NewTermQuery(names[0],
				fvalue.Bool()))
		case reflect.Struct:
			switch v := fvalue.Interface().(type) {
			case TimeRange:
				queries = append(queries, getTimeRangeQuery(v.From, v.To, names[0]))
			case FloatRange:
				queries = append(queries, getFloatRangeQuery(v.From, v.To, names[0]))
			case IntRange:
				queries = append(queries, getIntRangeQuery(v.From, v.To, names[0]))
			case Nested:
				var err error
				queries, err = getMustNestedQuery(ctx, v.payload, names[0], queries)
				if err != nil {
					return nil, errors.E(op, err)
				}
			case FullTextSearchShould:
				boolQuery, err := getFullTextSearchShouldQuery(
					v.payload,
					structFieldName,
					names,
				)
				if err != nil {
					return nil, errors.E(op, err)
				}
				queries = append(queries, boolQuery)
			case FullTextSearchMust:
				boolQuery, err := getFullTextSearchMustQuery(
					v.payload,
					structFieldName,
					names,
				)
				if err != nil {
					return nil, errors.E(op, err)
				}
				queries = append(queries, boolQuery)
			default:
				return nil, errors.E(op, structNotSupportedError(names[0]))
			}
		default:
			return nil, errors.E(op, typeNotSupportedError(structFieldName, fkind.String()))
		}
	}

	return queries, nil
}

// nolint: funlen, cyclop
func getExistsQuery(
	filter interface{},
) (existsQueries, notExistsQueries []elastic.Query, err error) {
	const op = errors.Op("getExistsQuery")

	rv := reflect.ValueOf(filter)
	if rv.Kind() != reflect.Struct {
		return nil, nil, errors.E(op, filterMustBeAStructError(rv.Kind().String()))
	}

	nFields := rv.Type().NumField()
	for i := 0; i < nFields; i++ {
		// Get value, type and tags
		fvalue := rv.Field(i)
		ftype := rv.Type().Field(i)
		fnames := parseFieldNames(ftype.Tag.Get("es"))

		// Skip zero values
		if fvalue.IsZero() {
			continue
		}

		// Rename field if specified a new name inside the tag
		structFieldName := ftype.Name
		names := []string{structFieldName}
		if len(fnames) > 0 {
			names = fnames
		}

		// Fix kind and value if field is a pointer
		fkind := ftype.Type.Kind()
		if fkind == reflect.Ptr {
			fvalue = fvalue.Elem()
			fkind = fvalue.Kind()
		}

		switch fkind {
		case reflect.Bool:
			if fvalue.Bool() {
				existsQueries = append(existsQueries, elastic.NewExistsQuery(names[0]))
			} else {
				notExistsQueries = append(notExistsQueries, elastic.NewExistsQuery(names[0]))
			}
		case reflect.Struct:
			switch v := fvalue.Interface().(type) {
			case Nested:
				var err error
				existsQueries, notExistsQueries, err = getExistsNestedQuery(
					v.payload,
					names[0],
					existsQueries,
					notExistsQueries,
				)
				if err != nil {
					return nil, nil, errors.E(op, err)
				}
			default:
				return nil, nil, errors.E(op, structNotSupportedError(names[0]))
			}
		default:
			return nil, nil, errors.E(op,
				typeNotSupportedError(structFieldName, fkind.String()))
		}
	}

	return existsQueries, notExistsQueries, nil
}

func shouldReturnMatchAllQuery(
	mustQueries,
	mustNotQueries,
	existsQueries,
	notExistsQueries []elastic.Query,
) bool {
	return len(mustNotQueries) == 0 &&
		len(existsQueries) == 0 &&
		len(notExistsQueries) == 0 &&
		len(mustQueries) == 0
}

func shouldReturnOnlyMustQuery(
	mustQueries,
	mustNotQueries,
	existsQueries,
	notExistsQueries []elastic.Query,
) bool {
	return len(mustNotQueries) == 0 &&
		len(existsQueries) == 0 &&
		len(notExistsQueries) == 0 &&
		len(mustQueries) == 1
}

func getBoolQuery(
	mustQueries,
	mustNotQueries,
	existsQueries,
	notExistsQueries []elastic.Query,
) *elastic.BoolQuery {
	boolQuery := elastic.NewBoolQuery()
	if len(mustQueries) > 0 {
		boolQuery.Must(mustQueries...)
	}
	if len(mustNotQueries) > 0 {
		boolQuery.MustNot(mustNotQueries...)
	}
	if len(existsQueries) > 0 {
		boolQuery.Must(existsQueries...)
	}
	if len(notExistsQueries) > 0 {
		boolQuery.MustNot(notExistsQueries...)
	}

	return boolQuery
}

func getFullTextSearchShouldQuery(
	payload interface{},
	structName string,
	names []string,
) (*elastic.BoolQuery, error) {
	const op = errors.Op("getFullTextSearchShouldQuery")
	contents, ok := payload.([]string)
	if !ok {
		return nil, errors.E(op,
			fullTextSearchTypeNotSupported(structName))
	}

	boolQuery := elastic.NewBoolQuery()
	for _, content := range contents {
		boolQuery.Should(
			elastic.NewMultiMatchQuery(content, names...).
				Type("phrase_prefix").
				MaxExpansions(maxExpansions),
		)
	}
	return boolQuery, nil
}

func getFullTextSearchMustQuery(
	payload interface{},
	structName string,
	names []string,
) (*elastic.BoolQuery, error) {
	const op = errors.Op("getFullTextSearchMustQuery")
	contents, ok := payload.([]string)
	if !ok {
		return nil, errors.E(op,
			fullTextSearchTypeNotSupported(structName))
	}

	boolQuery := elastic.NewBoolQuery()
	for _, content := range contents {
		boolQuery.Must(
			elastic.NewMultiMatchQuery(content, names...).
				Type("phrase_prefix").
				MaxExpansions(maxExpansions),
		)
	}
	return boolQuery, nil
}

func getIntRangeQuery(from, to uint64, name string) *elastic.RangeQuery {
	query := elastic.NewRangeQuery(name)
	if from != 0 {
		query = query.From(from)
	}
	if to != 0 {
		query = query.To(to)
	}
	return query
}

func getFloatRangeQuery(from, to float64, name string) *elastic.RangeQuery {
	query := elastic.NewRangeQuery(name)
	if from != 0 {
		query = query.From(from)
	}
	if to != 0 {
		query = query.To(to)
	}
	return query
}

func getTimeRangeQuery(from, to time.Time, name string) *elastic.RangeQuery {
	query := elastic.NewRangeQuery(name)
	if !from.IsZero() {
		query = query.From(from)
	}
	if !to.IsZero() {
		query = query.To(to)
	}
	return query
}

func getMustNestedQuery(
	ctx context.Context,
	payload interface{},
	name string,
	queries []elastic.Query,
) ([]elastic.Query, error) {
	const op = errors.Op("getMustNestedQuery")
	nestedQuery, err := getMustQuery(ctx, payload)
	if err != nil {
		return nil, errors.E(op, err)
	}

	switch len(nestedQuery) {
	case 0:
		// do nothing
	case 1:
		queries = append(queries,
			elastic.NewNestedQuery(
				name,
				nestedQuery[0],
			))
	default:
		queries = append(queries,
			elastic.NewNestedQuery(
				name,
				elastic.NewBoolQuery().Must(nestedQuery...),
			))
	}
	return queries, nil
}

func getExistsNestedQuery(
	payload interface{},
	name string,
	existsQueries []elastic.Query,
	notExistsQueries []elastic.Query,
) ([]elastic.Query, []elastic.Query, error) {
	const op = errors.Op("getExistsNestedQuery")
	nestedQueryExists, nestedQueryNotExists, err := getExistsQuery(payload)
	if err != nil {
		return nil, nil, errors.E(op, err)
	}

	if len(nestedQueryExists) > 0 {
		switch len(nestedQueryExists) {
		case 1:
			existsQueries = append(existsQueries,
				elastic.NewNestedQuery(
					name,
					nestedQueryExists[0],
				),
			)
		default:
			existsQueries = append(existsQueries,
				elastic.NewNestedQuery(
					name,
					elastic.NewBoolQuery().Must(nestedQueryExists...),
				))
		}
	}
	if len(nestedQueryNotExists) > 0 {
		switch len(nestedQueryNotExists) {
		case 1:
			notExistsQueries = append(notExistsQueries,
				elastic.NewNestedQuery(
					name,
					nestedQueryNotExists[0],
				),
			)
		default:
			notExistsQueries = append(notExistsQueries,
				elastic.NewNestedQuery(
					name,
					elastic.NewBoolQuery().Must(nestedQueryNotExists...),
				))
		}
	}
	return existsQueries, notExistsQueries, nil
}

func parseFieldNames(tag string) []string {
	if tag == "" {
		return nil
	}
	return strings.Split(tag, ",")
}

func stringsToInterfaces(input interface{}) []interface{} {
	s, _ := input.([]string)
	is := make([]interface{}, len(s))
	for i, v := range s {
		is[i] = v
	}
	return is
}

func utin64ToInterfaces(input interface{}) []interface{} {
	u, _ := input.([]uint64)
	is := make([]interface{}, len(u))
	for i, v := range u {
		is[i] = v
	}
	return is
}
