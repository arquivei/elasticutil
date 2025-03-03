package v7

import (
	"encoding/json"
	"testing"

	"github.com/arquivei/foundationkit/errors"
	"github.com/r3labs/diff/v3"
	"github.com/stretchr/testify/assert"
)

func Test_parseAggregations(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                        string
		elasticAggregationsResponse string
		expectedResponse            ResponseAggregation
		expectedError               error
	}{
		{
			name:                        "empty aggregations",
			elasticAggregationsResponse: "{}",
			expectedResponse:            ResponseAggregation{},
			expectedError:               nil,
		},
		{
			name:                        "one metric aggregation",
			elasticAggregationsResponse: `{"max_aggregation_name":{"value":1025000.0}}`,
			expectedResponse: ResponseAggregation{
				MetricAggregations: []ResponseMetricAggregation{
					{
						Name:  "max_aggregation_name",
						Value: 1025000.0,
					},
				},
			},
			expectedError: nil,
		},
		{
			name: "multiple metrics aggregations",
			elasticAggregationsResponse: `{
											"sum_aggregation_name" : {
											"value" : 3.063405636E7
											},
											"min_aggregation_name" : {
											"value" : 0.0
											},
											"max_aggregation_name_2" : {
											"value" : 1025000.0
											},
											"max_aggregation_name" : {
											"value" : 1025000.0
											},
											"count_aggregation_name" : {
											"value" : 3396
											}
										}`,
			expectedResponse: ResponseAggregation{
				MetricAggregations: []ResponseMetricAggregation{
					{
						Name:  "sum_aggregation_name",
						Value: 3.063405636e7,
					},
					{
						Name:  "min_aggregation_name",
						Value: 0.0,
					},
					{
						Name:  "max_aggregation_name_2",
						Value: 1025000.0,
					},
					{
						Name:  "max_aggregation_name",
						Value: 1025000.0,
					},
					{
						Name:  "count_aggregation_name",
						Value: float64(3396),
					},
				},
			},
			expectedError: nil,
		},
		{
			name: "metrics, histogram and terms aggregations with sub aggregations",
			elasticAggregationsResponse: `{
											"sum_aggregation_name": {
												"value": 3.063405636E7
											},
											"min_aggregation_name": {
												"value": 0.0
											},
											"max_aggregation_name_2": {
												"value": 1025000.0
											},
											"max_aggregation_name": {
												"value": 1025000.0
											},
											"count_aggregation": {
												"value": 3396
											},
											"some_histogram_agg": {
												"buckets": [
													{
														"key_as_string": "2024-01-01T00:00:00.000Z",
														"key": 1704067200000,
														"doc_count": 377,
														"max_agg_in_histogram_name": {
															"value": 250000.0
														},
														"min_agg_in_histogram_name": {
															"value": 11.5
														}
													},
													{
														"key_as_string": "2024-02-01T00:00:00.000Z",
														"key": 1706745600000,
														"doc_count": 373,
														"max_agg_in_histogram_name": {
															"value": 240000.0
														},
														"min_agg_in_histogram_name": {
															"value": 5.0
														}
													}
												]
											},
											"some_term_agg": {
												"doc_count_error_upper_bound": 0,
												"sum_other_doc_count": 0,
												"buckets": [
													{
														"key": "received",
														"doc_count": 2121,
														"doc_count_error_upper_bound": 0,
														"min_agg_in_term_name": {
															"value": 0.0
														}
													},
													{
														"key": "transporter",
														"doc_count": 551,
														"doc_count_error_upper_bound": 0,
														"min_agg_in_term_name": {
															"value": 45.0
														}
													},
													{
														"key": "authorized",
														"doc_count": 507,
														"doc_count_error_upper_bound": 0,
														"min_agg_in_term_name": {
															"value": 45.0
														}
													},
													{
														"key": "emitted",
														"doc_count": 217,
														"doc_count_error_upper_bound": 0,
														"min_agg_in_term_name": {
															"value": 0.0
														}
													}
												]
											}
										}`,
			expectedResponse: ResponseAggregation{
				MetricAggregations: []ResponseMetricAggregation{
					{
						Name:  "sum_aggregation_name",
						Value: 3.063405636e7,
					},
					{
						Name:  "min_aggregation_name",
						Value: 0.0,
					},
					{
						Name:  "max_aggregation_name_2",
						Value: 1025000.0,
					},
					{
						Name:  "max_aggregation_name",
						Value: 1025000.0,
					},
					{
						Name:  "count_aggregation",
						Value: float64(3396),
					},
				},
				BucketAggregations: []ResponseBucketAggregation{
					{
						Name: "some_histogram_agg",
						Buckets: []ResponseBucket{
							{
								Key:      "2024-01-01T00:00:00.000Z",
								DocCount: 377,
								MetricsAggregations: []ResponseMetricAggregation{
									{
										Name:  "max_agg_in_histogram_name",
										Value: 250000.0,
									},
									{
										Name:  "min_agg_in_histogram_name",
										Value: 11.5,
									},
								},
							},
							{
								Key:      "2024-02-01T00:00:00.000Z",
								DocCount: 373,
								MetricsAggregations: []ResponseMetricAggregation{
									{
										Name:  "max_agg_in_histogram_name",
										Value: 240000.0,
									},
									{
										Name:  "min_agg_in_histogram_name",
										Value: 5.0,
									},
								},
							},
						},
					},
					{
						Name: "some_term_agg",
						Buckets: []ResponseBucket{
							{
								Key:      "received",
								DocCount: 2121,
								MetricsAggregations: []ResponseMetricAggregation{
									{
										Name:  "min_agg_in_term_name",
										Value: 0.0,
									},
								},
							},
							{
								Key:      "transporter",
								DocCount: 551,
								MetricsAggregations: []ResponseMetricAggregation{
									{
										Name:  "min_agg_in_term_name",
										Value: 45.0,
									},
								},
							},
							{
								Key:      "authorized",
								DocCount: 507,
								MetricsAggregations: []ResponseMetricAggregation{
									{
										Name:  "min_agg_in_term_name",
										Value: 45.0,
									},
								},
							},
							{
								Key:      "emitted",
								DocCount: 217,
								MetricsAggregations: []ResponseMetricAggregation{
									{
										Name:  "min_agg_in_term_name",
										Value: 0.0,
									},
								},
							},
						},
					},
				},
			},
			expectedError: nil,
		},
		{
			name:                        "error - invalid aggregation content",
			elasticAggregationsResponse: `{"max_aggregation_name_2":1025000.0}`,
			expectedResponse:            ResponseAggregation{},
			expectedError:               errors.New("parseAggregations: parseAggregationData: unexpected type for aggregation data, expected: map[string]interface{} [given=float64]"),
		},
		{
			name:                        "error - unknown aggregation type",
			elasticAggregationsResponse: `{"max_aggregation_name":{"test":1025000.0}}`,
			expectedResponse:            ResponseAggregation{},
			expectedError:               errors.New("parseAggregations: parseAggregationData: aggregation type is not supported [aggName=max_aggregation_name]"),
		},
		{
			name:                        "error - invalid buckets content",
			elasticAggregationsResponse: `{"some_histogram_agg":{"buckets":["test"]}}`,
			expectedResponse:            ResponseAggregation{},
			expectedError:               errors.New("parseAggregations: parseAggregationData: unexpected type for bucket aggregation data, expected: map[string]interface{} [given=string]"),
		},
		{
			name: "error - invalid doc_count type inside bucket agg",
			elasticAggregationsResponse: `{"some_term_agg": {
												"doc_count_error_upper_bound": 0,
												"sum_other_doc_count": 0,
												"buckets": [
													{
														"key": "received",
														"doc_count": "2121"
													}
												]
											}
											}`,
			expectedResponse: ResponseAggregation{},
			expectedError:    errors.New("parseAggregations: parseAggregationData: unexpected type for doc_count field, expected: float64 [given=string]"),
		},
		{
			name: "error - invalid key type inside bucket agg",
			elasticAggregationsResponse: `{"some_term_agg": {
												"doc_count_error_upper_bound": 0,
												"sum_other_doc_count": 0,
												"buckets": [
													{
														"key": true,
														"doc_count": 2121
													}
												]
											}
											}`,
			expectedResponse: ResponseAggregation{},
			expectedError:    errors.New("parseAggregations: parseAggregationData: unexpected type for key field, expected: string or float64 [given=bool]"),
		},
		{
			name: "error - invalid key_as_string type inside bucket agg",
			elasticAggregationsResponse: `{"some_term_agg": {
												"doc_count_error_upper_bound": 0,
												"sum_other_doc_count": 0,
												"buckets": [
													{
														"key_as_string": true,
														"doc_count": 2121
													}
												]
											}
											}`,
			expectedResponse: ResponseAggregation{},
			expectedError:    errors.New("parseAggregations: parseAggregationData: unexpected type for key_as_string field, expected: string [given=bool]"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			assert.NotPanics(t, func() {
				aggregationsMap := make(map[string]interface{})
				err := json.Unmarshal([]byte(test.elasticAggregationsResponse), &aggregationsMap)
				assert.NoError(t, err, "unmarshal function should not return error - [%s]", test.name)

				response, err := parseAggregations(aggregationsMap)
				if test.expectedError == nil {
					assert.NoError(t, err, "unexpected error - [%s]", test.name)
				} else {
					assert.EqualError(t, err, test.expectedError.Error(), "unexpected error message - [%s]", test.name)
				}
				assert.ElementsMatch(
					t,
					test.expectedResponse.MetricAggregations,
					response.MetricAggregations,
					"unexpected metric aggregations - [%s]",
					test.name,
				)

				changelog, err := diff.Diff(response.BucketAggregations, test.expectedResponse.BucketAggregations)
				assert.NoError(t, err, "diff function should not return error - [%s]", test.name)
				assert.Len(t, changelog, 0, "unexpected bucket aggregations - [%s]", test.name)
			})
		})
	}
}
