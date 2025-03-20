package request_test

import (
	"testing"

	"github.com/sakamotoryou/api-agg-two/internal/common/http_client/Entity/request"
	"github.com/stretchr/testify/assert"
)

func TestRequestDataOperation(t *testing.T) {
	t.Run("Encode parameter", func(t *testing.T) {
		type TestParam struct {
			number  int
			decimal *float32
		}

		var dd float32 = 12.33
		newReq := request.Param(
			TestParam{number: 123, decimal: &dd},
		)(request.Request{})
		result := newReq.EncodeParam()
		assert.Equal(t, "number=123&decimal=12.33", result)
	})

	t.Run("Encode body", func(t *testing.T) {
		type TestBody struct {
			Number  int      `json:"number"`
			Decimal *float32 `json:"decimal"`
		}
		var dd float32 = 12.33
		newReq := request.Json(
			TestBody{Number: 123, Decimal: &dd},
		)(request.Request{})
		result, err := newReq.EncodeBody()
		assert.NoError(t, err)
		assert.Equal(t, `{"number":123,"decimal":12.33}`, result)
	})

	t.Run("Encode body with empty 'Content-Type'", func(t *testing.T) {
		type TestBody struct {
			Number  int      `json:"number"`
			Decimal *float32 `json:"decimal"`
		}
		var dd float32 = 12.33
		newReq := request.Body(
			TestBody{Number: 123, Decimal: &dd},
		)(request.Request{})
		_, err := newReq.EncodeBody()
		assert.ErrorIs(t, err, request.ErrUnknownContentEncode)
	})
}

func TestRequestOptions(t *testing.T) {
	t.Run("Domain Option", func(t *testing.T) {
		newReq := request.Domain("example.com")
		assert.Equal(t, "example.com", newReq(request.Request{}).GetDomain())
	})

	t.Run("Path Option", func(t *testing.T) {
		newReq := request.Path("/api/v1")
		assert.Equal(t, "/api/v1", newReq(request.Request{}).GetPath())
	})

	t.Run("Get Method", func(t *testing.T) {
		newReq := request.Get()
		assert.Equal(t, "GET", newReq(request.Request{}).GetMethod())
	})

	t.Run("Post Method", func(t *testing.T) {
		newReq := request.Post()
		assert.Equal(t, "POST", newReq(request.Request{}).GetMethod())
	})

	t.Run("Put Method", func(t *testing.T) {
		newReq := request.Put()
		assert.Equal(t, "PUT", newReq(request.Request{}).GetMethod())
	})

	t.Run("Delete Method", func(t *testing.T) {
		newReq := request.Delete()
		assert.Equal(t, "DELETE", newReq(request.Request{}).GetMethod())
	})

	t.Run("Patch Method", func(t *testing.T) {
		newReq := request.Patch()
		assert.Equal(t, "PATCH", newReq(request.Request{}).GetMethod())
	})

	// t.Run("Head Method", func(t *testing.T) {
	// 	newReq := request.Head()
	// 	assert.Equal(t, "HEAD", newReq(request.Request{}).GetMethod())
	// })
	//
	// t.Run("Options Method", func(t *testing.T) {
	// 	newReq := request.Options()
	// 	assert.Equal(t, "OPTIONS", newReq(request.Request{}).GetMethod())
	// })

	t.Run("Http Header", func(t *testing.T) {
		newReq := request.Header(
			request.SetRequestHeader("Content-Type", "application/json"),
			request.SetRequestHeader("x-api-key", "my-custom-key"),
		)
		assert.Equal(t, "application/json", newReq(request.Request{}).GetHeader("Content-Type"))
		assert.Equal(t, "my-custom-key", newReq(request.Request{}).GetHeader("x-api-key"))
		assert.Equal(t, "", newReq(request.Request{}).GetHeader("non-existent"))
	})

	t.Run("Set request parameter", func(t *testing.T) {
		type RequestParam struct {
			id  string
			num int
		}

		newReq := request.Param(
			RequestParam{"uuid-123", 123},
		)
		assert.EqualValues(t, RequestParam{"uuid-123", 123}, newReq(request.Request{}).GetRawRequestData().GetParam())
	})

	t.Run("Set body parameter", func(t *testing.T) {
		type RequestBody struct {
			id  string
			num int
		}

		newReq := request.Body(
			RequestBody{"uuid-123", 123},
		)
		assert.EqualValues(t, RequestBody{"uuid-123", 123}, newReq(request.Request{}).GetRawRequestData().GetBody())
	})

	t.Run("Combination of Domain, Path, and Get", func(t *testing.T) {
		newReq, err := request.Build(
			request.Domain("example.com"),
			request.Get(),
			request.Path("/api/v1"),
			request.Header(
				request.SetRequestHeader("Content-Type", "application/json"),
				request.SetRequestHeader("x-api-key", "my-custom-key"),
			),
		)
		assert.NoError(t, err)
		assert.Equal(t, "example.com", newReq.GetDomain())
		assert.Equal(t, "/api/v1", newReq.GetPath())
		assert.Equal(t, "GET", newReq.GetMethod())
		assert.Equal(t, "application/json", newReq.GetHeader("Content-Type"))
		assert.Equal(t, "my-custom-key", newReq.GetHeader("x-api-key"))
	})

	t.Run("Combination of Domain, Path, Post, Param and Body", func(t *testing.T) {
		type RequestParam struct {
			Name     string
			SomeData float64
		}

		type RequestBody struct {
			BodyName string  `json:"body_name"`
			BodyData float64 `json:"body_data"`
		}

		newReq, err := request.Build(
			request.Domain("example.com"),
			request.Get(),
			request.Path("/api/v1"),
			request.Header(
				request.SetRequestHeader("x-api-key", "my-custom-key"),
			),
			request.Param(RequestParam{"param", 99.99}),
			request.Json(RequestBody{"body", 9.99}),
		)
		assert.NoError(t, err)
		assert.Equal(t, "example.com", newReq.GetDomain())
		assert.Equal(t, "/api/v1", newReq.GetPath())
		assert.Equal(t, "GET", newReq.GetMethod())
		assert.Equal(t, "application/json", newReq.GetHeader("Content-Type"))
		assert.Equal(t, "my-custom-key", newReq.GetHeader("x-api-key"))
		assert.EqualValues(t, "name=param&somedata=99.99", newReq.GetParam())
		assert.EqualValues(t, `{"body_name":"body","body_data":9.99}`, newReq.GetBody())
	})

	t.Run("Missing domain option", func(t *testing.T) {
		_, err := request.Build(
			request.Get(),
			request.Path("/api/v1"),
		)

		assert.Error(t, err)
		assert.ErrorIs(t, err, request.ErrDomainNameEmpty)
	})

	t.Run("Missing method option", func(t *testing.T) {
		_, err := request.Build(
			request.Domain("example.com"),
			request.Path("/api/v1"),
		)

		assert.Error(t, err)
		assert.ErrorIs(t, err, request.ErrRequestMethodEmpty)
	})

	t.Run("Missing path option", func(t *testing.T) {
		_, err := request.Build(
			request.Domain("example.com"),
			request.Get(),
		)

		assert.Error(t, err)
		assert.ErrorIs(t, err, request.ErrPathEmpty)
	})

	t.Run("Request Param is not a struct", func(t *testing.T) {
		_, err := request.Build(
			request.Domain("example.com"),
			request.Get(),
			request.Path("/api/v1"),
			request.Param("some string"),
		)

		assert.Error(t, err)
		assert.ErrorIs(t, err, request.ErrRequestParamInvalidType)
	})

	t.Run("Empty request param", func(t *testing.T) {
		_, err := request.Build(
			request.Domain("example.com"),
			request.Path("/api/v1"),
			request.Get(),
			request.Json(
				struct {
					Hello string `json:"hello"`
				}{
					Hello: "Hello world!",
				},
			),
		)

		assert.NoError(t, err)
	})

	t.Run("Request Body is not a struct", func(t *testing.T) {
		_, err := request.Build(
			request.Domain("example.com"),
			request.Get(),
			request.Path("/api/v1"),
			request.Body("some string"),
		)

		assert.Error(t, err)
		assert.ErrorIs(t, err, request.ErrRequestBodyInvalidType)
	})

	t.Run("Empty request body", func(t *testing.T) {
		_, err := request.Build(
			request.Domain("example.com"),
			request.Get(),
			request.Path("/api/v1"),
		)

		assert.NoError(t, err)
	})
}
