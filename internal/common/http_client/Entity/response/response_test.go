package response_test

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/sakamotoryou/api-agg-two/internal/common/http_client/Entity/response"
	"github.com/stretchr/testify/assert"
)

func TestResponseOptions(t *testing.T) {
	t.Run("Response header option", func(t *testing.T) {
		type test struct {
			key   string
			value string
		}

		var tests []test = []test{
			{"x-api-key", "12345"},
			{"some-other-key", "keykeykey"},
		}

		header := make(http.Header, 0)
		for _, h := range tests {
			header.Set(h.key, h.value)
		}
		respFunc := response.ResponseHeader(header)
		newResp := respFunc(response.Response{}).GetHeader()

		for i := range tests {
			assert.Equal(t, tests[i].value, newResp.Get(tests[i].key))
		}
	})

	t.Run("Response body option", func(t *testing.T) {
		respFunc := response.ResponseBody(
			strings.NewReader("stub"),
		)
		newResp := respFunc(response.Response{}).GetBody()

		data := make([]byte, 4)
		newResp.Read(data)
		assert.Equal(t, "stub", string(data))
	})

	t.Run("Response content type option", func(t *testing.T) {
		respFunc := response.ResponseContentType("application/json")
		assert.Equal(t, "application/json", respFunc(response.Response{}).GetContentType())
	})

	t.Run("Response status code option", func(t *testing.T) {
		respFunc := response.ResponseStatusCode(200)
		assert.Equal(t, 200, respFunc(response.Response{}).GetStatusCode())
	})

	t.Run("On success bind data", func(t *testing.T) {
		var data string
		respFunc := response.ResponseBody(
			strings.NewReader("stub"),
		)
		newResp := respFunc(response.Response{})
		newResp = response.ResponseStatusCode(200)(newResp)
		newResp = response.OnSuccess(
			response.Decode(&data),
		)(newResp)

		err := newResp.ResolveSuccess()
		assert.NoError(t, err)
		assert.Equal(t, "stub", string(data))
	})

	t.Run("On success run custom logic", func(t *testing.T) {
		var data string
		respFunc := response.ResponseBody(
			strings.NewReader("stub"),
		)
		newResp := respFunc(response.Response{})
		newResp = response.ResponseStatusCode(200)(newResp)
		newResp = response.OnSuccess(
			response.Decode(&data),
      func() response.SuccessResolver {
        return func(resp response.Response) response.ResponseErrorFunc {
          data += " with custom logic"
          return response.Skip()
        }
      }(),
		)(newResp)

		err := newResp.ResolveSuccess()
		assert.NoError(t, err)
		assert.Equal(t, "stub with custom logic", string(data))
	})

	t.Run("On failure receive error", func(t *testing.T) {
		expect := "Oops! Something went wrong on the server."
		respFunc := response.ResponseBody(
			strings.NewReader(`{"error":"internal server error"}`),
		)
		newResp := respFunc(response.Response{})
		newResp = response.ResponseStatusCode(500)(newResp)
		newResp = response.OnReject(
			response.OnInternalServerError(
				response.SetClientError(expect),
			),
		)(newResp)

		err := newResp.ResolveReject()
		assert.ErrorAs(t, err, &response.ResponseError{})
		assert.Equal(t, err.Error(), expect)
	})

	t.Run("On error with default error", func(t *testing.T) {
		respFunc := response.ResponseBody(
			strings.NewReader(`{"error":"bad bad request"}`),
		)
		newResp := respFunc(response.Response{})
		newResp = response.ResponseStatusCode(400)(newResp)
		newResp = response.OnReject()(newResp)

		err := newResp.ResolveReject()
		actual_err := err.(response.ResponseError)
		assert.Equal(t, fmt.Sprint(`
    Front Message: Bad request,
    Stack Message: 
    `), actual_err.Log())
	})
}
