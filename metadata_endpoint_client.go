package main

import (
	"net/http"

	"github.com/stretchr/testify/mock"
)

type MetadataEndpointClient interface {
	Get(url string) (resp *http.Response, err error)
}

type RealMetadataEndpointClient struct{}

func NewRealMetadataEndpointClient() *RealMetadataEndpointClient {
	return &RealMetadataEndpointClient{}
}

func (c *RealMetadataEndpointClient) Get(url string) (resp *http.Response, err error) {
	return http.Get(url)
}

type MockedMetadataEndpointClient struct {
	mock.Mock
}

func NewMockedMetadataEndpointClient() *MockedMetadataEndpointClient {
	return &MockedMetadataEndpointClient{}
}

func (m *MockedMetadataEndpointClient) Get(url string) (resp *http.Response, err error) {
	args := m.Called(url)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*http.Response), args.Error(1)
}
