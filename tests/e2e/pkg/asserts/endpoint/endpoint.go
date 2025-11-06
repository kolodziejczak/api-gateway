package endpoint

import (
	"fmt"
	"io"
	"net/http"
	"testing"

	httphelper "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/http"
	"github.com/stretchr/testify/assert"
)

func AssertEndpoint(t *testing.T, method, url string, requestHeaders map[string]string, expectedHttpCode int, expectedResponseHeaders map[string]string) error {
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

	if expectedResponseHeaders != nil {
		for headerName, headerValue := range expectedResponseHeaders {
			assert.Equal(t, headerValue, response.Header.Get(headerName))
		}
	}

	return nil
}
