package endpoint

import (
	"fmt"
	"io"
	"net/http"
	"testing"

	httphelper "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/http"
	"github.com/stretchr/testify/assert"
)

func AssertEndpoint(t *testing.T, method, url string, expectedHttpCode int) error {
	t.Helper()
	httpClient := httphelper.NewHTTPClient(t, httphelper.WithPrefix("ext-auth-client"))
	request, err := http.NewRequest(method, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	response, err := httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("failed to perform request: %w", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(response.Body)
	assert.Equal(t, expectedHttpCode, response.StatusCode, "unexpected status code")

	//if expectedResponseHeaders != nil {
	//	for headerName, headerValue := range expectedResponseHeaders {
	//		assert.Equal(t, headerValue, response.Header.Get(headerName))
	//	}
	//}

	return nil
}

func AssertEndpointWithoutResponseHeaders(t *testing.T, method, url string, requestHeaders map[string]string, expectedHttpCode int, expectedMissingHeaders []string) error {
	t.Helper()
	httpClient := httphelper.NewHTTPClient(t, httphelper.WithPrefix("ext-auth-client"))
	request, err := http.NewRequest(method, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	for headerName, headerValue := range requestHeaders {
		request.Header.Set(headerName, headerValue)
	}

	response, err := httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("failed to perform request: %w", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(response.Body)
	assert.Equal(t, expectedHttpCode, response.StatusCode, "unexpected status code")

	if expectedMissingHeaders != nil && len(expectedMissingHeaders) != 0 {
		for _, header := range expectedMissingHeaders {
			assert.Empty(t, response.Header.Get(header))
		}
	}

	return nil
}

func AssertEndpointWithResponseHeaders(t *testing.T, method, url string, requestHeaders map[string]string, expectedHttpCode int, expectedResponseHeaders map[string]string) error {
	t.Helper()
	httpClient := httphelper.NewHTTPClient(t, httphelper.WithPrefix("ext-auth-client"))
	request, err := http.NewRequest(method, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	for headerName, headerValue := range requestHeaders {
		request.Header.Set(headerName, headerValue)
	}

	response, err := httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("failed to perform request: %w", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(response.Body)
	assert.Equal(t, expectedHttpCode, response.StatusCode, "unexpected status code")
	for k, v := range response.Header {
		println("HIHIHI: ", k, v)
	}
	if expectedResponseHeaders != nil {
		for headerName, headerValue := range expectedResponseHeaders {
			responseHeaderValue := response.Header.Get(headerName)
			println("XDDDDDD DEBUG SECTION")
			println("EXPECTED: ", headerName, headerValue)
			println("GOT:", responseHeaderValue)
			println("XDDDDDD END OF DEBUG SECTION")
			//assert.Equal(t, headerValue, responseHeaderValue)
			if headerValue != responseHeaderValue {
				t.Fatalf("Didn't get the expected response header: %s: %s, got %s", headerName, headerValue, responseHeaderValue)
			}
		}
	}

	return nil
}
