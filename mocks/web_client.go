package mocks

import "net/http"

// HttpClientDo is the name of the mock function key for http.Client.Do
var HTTPClientDo = "http_client_do"

type WebClient interface {
	Do(*http.Request) (*http.Response, error)
}

type MockClient struct{}

func (c *MockClient) Do(_ *http.Request) (*http.Response, error) {
	res, err := GetResult(HTTPClientDo)
	if err != nil {
		return nil, err
	} else if _, ok := res.(error); ok {
		return nil, res.(error)
	}
	return res.(*http.Response), nil
}
