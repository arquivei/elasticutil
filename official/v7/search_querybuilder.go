package v7

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/arquivei/foundationkit/errors"
	"github.com/rs/zerolog/log"

	"github.com/arquivei/elasticutil/official/v7/querybuilders"
)

const maxExpansions = 1024

func buildElasticBoolQuery(filter Filter) (querybuilders.Query, error) {
	const op = errors.Op("buildElasticBoolQuery")

	var mustQueries, mustNotQueries, existsQueries, notExistsQueries []querybuilders.Query

	if filter.Must != nil {
		var err error
		mustQueries, err = getMustQuery(filter.Must)
		if err != nil {
			return nil, errors.E(op, err)
		}
	}

	if filter.MustNot != nil {
		var err error
		mustNotQueries, err = getMustQuery(filter.MustNot)
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
		return querybuilders.NewMatchAllQuery(), nil
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

func buildElasticAggsQuery(aggregation RequestAggregation) (querybuilders.Query, error) {
	const op = errors.Op("buildElasticAggsQuery")

	aggsQuery := querybuilders.NewAggsQuery()

	metricAggsQueries, err := buildElasticMetricAggsQueries(aggregation.Metrics)
	if err != nil {
		return nil, errors.E(op, err)
	}
	aggsQuery.Metric(metricAggsQueries...)

	for _, bucketAgg := range aggregation.Buckets {
		switch bucketAgg.Type {
		case "term":
			termsAggQuery := querybuilders.NewBucketTermsAggregation(bucketAgg.Name, bucketAgg.Field)

			subMetricAggsQueries, err := buildElasticMetricAggsQueries(bucketAgg.MetricsSubAgg)
			if err != nil {
				return nil, errors.E(op, err)
			}

			termsAggQuery = termsAggQuery.SubMetrics(subMetricAggsQueries...)
			aggsQuery.BucketTerms(*termsAggQuery)
		case "monthlyhistogram":
			histogramAggQuery := querybuilders.NewBucketHistogramAggregation(
				bucketAgg.Name,
				bucketAgg.Field,
				"month",
			)

			subMetricAggsQueries, err := buildElasticMetricAggsQueries(bucketAgg.MetricsSubAgg)
			if err != nil {
				return nil, errors.E(op, err)
			}

			histogramAggQuery = histogramAggQuery.SubMetrics(subMetricAggsQueries...)
			aggsQuery.BucketDateHistogram(*histogramAggQuery)
		default:
			return nil, errors.E(op, aggregationTypeNotSupported(bucketAgg.Type))
		}
	}

	return aggsQuery, nil
}

func buildElasticMetricAggsQueries(
	metricsAggs []RequestMetricAggregation,
) ([]querybuilders.MetricAggregation, error) {
	const op = errors.Op("buildElasticMetricAggsQuery")

	metricAggsQueries := make([]querybuilders.MetricAggregation, 0)

	for _, metricAgg := range metricsAggs {
		switch metricAgg.Type {
		case "min":
			metricAggsQueries = append(metricAggsQueries, querybuilders.NewMinAggregation(metricAgg.Name, metricAgg.Field))
		case "max":
			metricAggsQueries = append(metricAggsQueries, querybuilders.NewMaxAggregation(metricAgg.Name, metricAgg.Field))
		case "sum":
			metricAggsQueries = append(metricAggsQueries, querybuilders.NewSumAggregation(metricAgg.Name, metricAgg.Field))
		case "count":
			metricAggsQueries = append(metricAggsQueries, querybuilders.NewCountAggregation(metricAgg.Name, metricAgg.Field))
		default:
			return nil, errors.E(op, aggregationTypeNotSupported(metricAgg.Type))
		}
	}

	return metricAggsQueries, nil
}

func marshalQuery(query querybuilders.Query) string {
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
func getMustQuery(filter interface{}) ([]querybuilders.Query, error) {
	const op = errors.Op("getMustQuery")

	var queries []querybuilders.Query

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
				queries = append(queries, querybuilders.NewTermsQuery(names[0],
					extractSliceFromInterface[string](fvalue.Interface())...))
			case "uint64":
				queries = append(queries, querybuilders.NewTermsQuery(names[0],
					extractSliceFromInterface[uint64](fvalue.Interface())...))
			}
		case reflect.Bool:
			queries = append(queries, querybuilders.NewTermQuery(names[0],
				fvalue.Bool()))
		case reflect.Struct:
			switch v := fvalue.Interface().(type) {
			case TimeRange:
				queries = append(queries, getRangeQuery(v.From, v.To, names[0]))
			case FloatRange:
				queries = append(queries, getRangeQuery(v.From, v.To, names[0]))
			case IntRange:
				queries = append(queries, getRangeQuery(v.From, v.To, names[0]))
			case Nested:
				var err error
				queries, err = getMustNestedQuery(v.Payload, names[0], queries)
				if err != nil {
					return nil, errors.E(op, err)
				}
			case FullTextSearchShould:
				boolQuery, err := getFullTextSearchShouldQuery(
					v.Payload,
					structFieldName,
					names,
				)
				if err != nil {
					return nil, errors.E(op, err)
				}
				queries = append(queries, boolQuery)
			case FullTextSearchMust:
				boolQuery, err := getFullTextSearchMustQuery(
					v.Payload,
					structFieldName,
					names,
				)
				if err != nil {
					return nil, errors.E(op, err)
				}
				queries = append(queries, boolQuery)
			case MultiMatchSearchShould:
				boolQuery, err := getMultiMatchSearchShouldQuery(
					v.Payload,
					structFieldName,
					names,
				)
				if err != nil {
					return nil, errors.E(op, err)
				}
				queries = append(queries, boolQuery)
			case CustomSearch:
				boolQuery, err := v.GetQuery()
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
) (existsQueries, notExistsQueries []querybuilders.Query, err error) {
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
				existsQueries = append(existsQueries, querybuilders.NewExistsQuery(names[0]))
			} else {
				notExistsQueries = append(notExistsQueries, querybuilders.NewExistsQuery(names[0]))
			}
		case reflect.Struct:
			switch v := fvalue.Interface().(type) {
			case Nested:
				var err error
				existsQueries, notExistsQueries, err = getExistsNestedQuery(
					v.Payload,
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
	notExistsQueries []querybuilders.Query,
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
	notExistsQueries []querybuilders.Query,
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
	notExistsQueries []querybuilders.Query,
) *querybuilders.BoolQuery {
	boolQuery := querybuilders.NewBoolQuery()
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
) (*querybuilders.BoolQuery, error) {
	const op = errors.Op("getFullTextSearchShouldQuery")
	contents, ok := payload.([]string)
	if !ok {
		return nil, errors.E(op,
			fullTextSearchTypeNotSupported(structName))
	}

	boolQuery := querybuilders.NewBoolQuery()
	for _, content := range contents {
		boolQuery.Should(
			querybuilders.NewMultiMatchQuery(content, names...).
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
) (*querybuilders.BoolQuery, error) {
	const op = errors.Op("getFullTextSearchMustQuery")
	contents, ok := payload.([]string)
	if !ok {
		return nil, errors.E(op,
			fullTextSearchTypeNotSupported(structName))
	}

	boolQuery := querybuilders.NewBoolQuery()
	for _, content := range contents {
		boolQuery.Must(
			querybuilders.NewMultiMatchQuery(content, names...).
				Type("phrase_prefix").
				MaxExpansions(maxExpansions),
		)
	}
	return boolQuery, nil
}

func getMultiMatchSearchShouldQuery(
	payload interface{},
	structName string,
	names []string,
) (*querybuilders.BoolQuery, error) {
	const op = errors.Op("getMultiMatchSearchShouldQuery")
	contents, ok := payload.([]string)
	if !ok {
		return nil, errors.E(op,
			multiMatchSearchTypeNotSupported(structName))
	}

	boolQuery := querybuilders.NewBoolQuery()
	for _, content := range contents {
		boolQuery.Should(
			querybuilders.NewMultiMatchQuery(content, names...).
				Type("best_fields"). // default, but we explicitly specify this choice
				MaxExpansions(maxExpansions),
		)
	}
	return boolQuery, nil
}

func getRangeQuery[T Ranges](from T, to T, name string) *querybuilders.RangeQuery {
	var zero T
	query := querybuilders.NewRangeQuery(name)
	if from != zero {
		query = query.From(from)
	}
	if to != zero {
		query = query.To(to)
	}
	return query
}

func getMustNestedQuery(
	payload interface{},
	name string,
	queries []querybuilders.Query,
) ([]querybuilders.Query, error) {
	const op = errors.Op("getMustNestedQuery")
	nestedQuery, err := getMustQuery(payload)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return appendNestedQuery(
		name,
		queries,
		nestedQuery,
	), nil
}

func getExistsNestedQuery(
	payload interface{},
	name string,
	existsQueries []querybuilders.Query,
	notExistsQueries []querybuilders.Query,
) ([]querybuilders.Query, []querybuilders.Query, error) {
	const op = errors.Op("getExistsNestedQuery")
	nestedQueryExists, nestedQueryNotExists, err := getExistsQuery(payload)
	if err != nil {
		return nil, nil, errors.E(op, err)
	}

	existsQueries = appendNestedQuery(
		name,
		existsQueries,
		nestedQueryExists,
	)

	notExistsQueries = appendNestedQuery(
		name,
		notExistsQueries,
		nestedQueryNotExists,
	)

	return existsQueries, notExistsQueries, nil
}

func appendNestedQuery(
	queryName string,
	queries []querybuilders.Query,
	nestedQuery []querybuilders.Query,
) []querybuilders.Query {
	if len(nestedQuery) == 0 {
		return queries
	}

	if len(nestedQuery) == 1 {
		return append(queries,
			querybuilders.NewNestedQuery(
				queryName,
				nestedQuery[0],
			),
		)
	}

	return append(queries,
		querybuilders.NewNestedQuery(
			queryName,
			querybuilders.NewBoolQuery().Must(nestedQuery...),
		))
}

func parseFieldNames(tag string) []string {
	if tag == "" {
		return nil
	}
	return strings.Split(tag, ",")
}

func extractSliceFromInterface[T any](input interface{}) []interface{} {
	s, _ := input.([]T)
	is := make([]interface{}, len(s))
	for i, v := range s {
		is[i] = v
	}
	return is
}

func queryWithSearchAfter(q querybuilders.Query, aggs string, searchAfter string) string {
	var b strings.Builder

	b.WriteString(`{"query":`)

	b.WriteString(marshalQuery(q))

	if len(aggs) > 0 {
		b.WriteString(", ")
		b.WriteString(fmt.Sprintf(`	"aggs": %s`, aggs))
	}

	if len(searchAfter) > 0 {
		b.WriteString(", ")
		b.WriteString(fmt.Sprintf(`	"search_after": %s`, searchAfter))
	}

	b.WriteString("}")

	log.Debug().Str("elastic_query", b.String()).Msg("Elastic query")

	return b.String()
}

func hasAggregations(aggregation RequestAggregation) bool {
	return len(aggregation.Buckets) > 0 || len(aggregation.Metrics) > 0
}
