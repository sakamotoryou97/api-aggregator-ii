package http_client_test

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/sakamotoryou/api-agg-two/internal/common/http_client"
	"github.com/stretchr/testify/assert"
)

type Response_reader_stub struct {
	s        string
	i        int64
	prevRune int64
}

func newResponseReaderStub(s string) *Response_reader_stub {
	return &Response_reader_stub{s, 0, -1}
}

func (r *Response_reader_stub) Read(p []byte) (n int, err error) {
	if r.i >= int64(len(r.s)) {
		return 0, io.EOF
	}
	r.prevRune = -1
	n = copy(p, r.s[r.i:])
	r.i += int64(n)
	return
}

func (r *Response_reader_stub) Close() error {
	return nil
}

func newBlankHttpResponse() *http.Response {
	return &http.Response{}
}

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
		respFunc := http_client.ResponseHeader(header)
		newResp := respFunc(http_client.Response{}).GetHeader()

    for i := range tests {
      assert.Equal(t, tests[i].value, newResp.Get(tests[i].key))
    }
	})

	t.Run("Response body option", func(t *testing.T) {
		respFunc := http_client.ResponseBody(
			strings.NewReader("stub"),
		)
		newResp := respFunc(http_client.Response{}).GetBody()

		data := make([]byte, 4)
		newResp.Read(data)
		assert.Equal(t, "stub", string(data))
	})

	t.Run("Response content type option", func(t *testing.T) {
		respFunc := http_client.ResponseContentType("application/json")
		assert.Equal(t, "application/json", respFunc(http_client.Response{}).GetContentType())
	})

	t.Run("Response status code option", func(t *testing.T) {
		respFunc := http_client.ResponseStatusCode(200)
		assert.Equal(t, 200, respFunc(http_client.Response{}).GetStatusCode())
	})
}

// func TestResponseBinder(t *testing.T) {
// 	t.Run("Response body bind decoder", func(t *testing.T) {
// 		t.Run("Decode text format", func(t *testing.T) {
// 			new_response_with_plain_text := func() http_client.Response {
//         r, _ := http_client.NewResponse(
// 					http_client.ResponseContentType("text/plain"),
// 					http_client.ResponseBody(strings.NewReader("hello to text binder")),
// 				)
//
//         return r
// 			}
//
// 			t.Run("Valid text bind with string", func(t *testing.T) {
// 				newResp := new_response_with_plain_text()
// 				var data string
// 				err := newResp.Resolve(
// 					http_client.BindData(&data),
// 				)
// 				assert.NoError(t, err)
// 				assert.Equal(t, "hello to text binder", data)
// 			})
//
// 			t.Run("Text bind with not string pointer", func(t *testing.T) {
// 				newResp := new_response_with_plain_text()
//
// 				var data string
// 				err := newResp.Resolve(
// 					http_client.BindData(data),
// 				)
// 				assert.ErrorIs(t, err, http_client.ErrDataBindInvalid)
// 			})
//
// 			t.Run("Text bind with not string", func(t *testing.T) {
// 				newResp := new_response_with_plain_text()
//
// 				var data int
// 				err := newResp.Resolve(
// 					http_client.BindData(&data),
// 				)
// 				assert.ErrorIs(t, err, http_client.ErrDefaultBindNotString)
// 			})
// 		})
//
// 		t.Run("Decode json format", func(t *testing.T) {
// 			new_response_with_json := func() http_client.Response {
//         r, _ := http_client.NewResponse(
// 					http_client.ResponseContentType("application/json"),
// 					http_client.ResponseBody(strings.NewReader(`{"itemId": 1234, "name": "onetwothreefour"}`)),
// 				)
//
//         return r
// 			}
// 			t.Run("Valid json bind to struct", func(t *testing.T) {
// 				newResp := new_response_with_json()
//
// 				type expect struct {
// 					ItemId int    `json:"itemId"`
// 					Name   string `json:"name"`
// 				}
//
// 				var data expect
// 				err := newResp.Resolve(
// 					http_client.BindData(&data),
// 				)
// 				assert.NoError(t, err)
// 				assert.Equal(t, expect{1234, "onetwothreefour"}, data)
// 			})
// 		})
// 	})
//
// 	t.Run("Response body bind default status code", func(t *testing.T) {
// 		tests := []struct {
// 			statusCode int
// 			expected   error
// 		}{
//       {http.StatusOK, nil},
//       {http.StatusCreated, nil},
//       {http.StatusNoContent, nil},
// 			{http.StatusBadRequest, http_client.ErrBadRequest},
// 			{http.StatusUnauthorized, http_client.ErrUnauthorized},
// 			{http.StatusForbidden, http_client.ErrForbidden},
// 			{http.StatusNotFound, http_client.ErrNotFound},
// 			{http.StatusMethodNotAllowed, http_client.ErrMethodNotAllowed},
// 			{http.StatusConflict, http_client.ErrConflict},
// 			{http.StatusInternalServerError, http_client.ErrInternalServer},
// 			{http.StatusNotImplemented, http_client.ErrNotImplemented},
// 			{http.StatusServiceUnavailable, http_client.ErrServiceUnavailable},
// 			{999, http_client.ErrInternalServer},
// 		}
//
// 		for _, test := range tests {
// 			respFunc := http_client.ResponseStatusCode(test.statusCode)
// 			newResp := respFunc(http_client.Response{})
// 			err := newResp.Resolve(
// 				http_client.BindError(),
// 			)
// 			assert.ErrorIs(t, err, test.expected)
// 		}
// 	})
// }
