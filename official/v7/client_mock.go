package v7

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type mockClient struct {
	mock.Mock
}

func (m *mockClient) Search(_ context.Context, sc SearchConfig) (SearchResponse, error) {
	args := m.Called(sc)
	return args.Get(0).(SearchResponse), args.Error(1)
}

// MustNewClientMock returns a mocked Client that uses Search method and returns @expectedResponse
// and @expectedError for the giving @input.
func MustNewClientMockSearch(input SearchConfig, expectedResponse SearchResponse, expectedError error) Client {
	m := mockClient{}
	m.On("Search", input).Return(expectedResponse, expectedError)
	return &m
}
