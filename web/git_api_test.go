package web

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/ammesonb/ubiquiti-config-generator/internal/errors"
	"github.com/ammesonb/ubiquiti-config-generator/mocks"
	"github.com/stretchr/testify/assert"
)

func closeBody(response *http.Response) {
	if response != nil && response.Body != nil {
		if err := response.Body.Close(); err != nil {
			fmt.Printf("Failed to close response body: %v\n", err)
		}
	}
}

func TestMakeGitRequest(t *testing.T) {
	client := &mocks.MockClient{}

	body := map[string]any{
		// client is not JSON-serializable
		"foo": http.Client{},
	}
	response, err := makeGitRequest(client, "test-obj", "abc", "def", "/test-url", "GET", body)
	assert.Nil(t, response)
	assert.ErrorIs(t, err, errors.ErrWithCtx(errMarshalBody, "test-obj"))
	defer closeBody(response)

	// invalid method and URL
	response, err = makeGitRequest(client, "test-obj", "abc", "def", "http:/test-url-foo", "[}(", nil)
	assert.Nil(t, response)
	assert.ErrorIs(t, err, errors.ErrWithVarCtx(errCreateRequest, "[}(", "test-obj"))
	defer closeBody(response)
}
