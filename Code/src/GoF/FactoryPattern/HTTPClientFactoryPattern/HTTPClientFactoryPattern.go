package HTTPClientFactoryPattern

import (
	"net/http"
	"net/http/httptest"
)

// Doer We define a Doer interface, that has the method signature
// of the `http.Client` structs `Do` method
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

// NewHTTPClient This gives us a regular HTTP client from the `net/http` package
func NewHTTPClient() Doer {
	return &http.Client{}
}

type mockHTTPClient struct{}

func (*mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	// The `NewRecorder` method of the httptest package gives us
	// a new mock request generator
	res := httptest.NewRecorder()
	// calling the `Result` method gives us
	// the default empty *http.Response object
	return res.Result(), nil
}

// NewMockHTTPClient This gives us a mock HTTP client, which returns
// an empty response for any request sent to it
func NewMockHTTPClient() Doer {
	return &mockHTTPClient{}
}
