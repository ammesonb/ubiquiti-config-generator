package mocks

import "net/http"

// HttpClientDo is the name of the mock function key for http.Client.Do
var HTTPClientDo = "http_client_do"

type WebClient interface {
	Do(*http.Request) (*http.Response, error)
}

type MockClient struct{}

func (c *MockClient) Do(_ *http.Request) (*http.Response, error) {
	return nil, nil
}
