package http_client_test

import (
	"fmt"
	"testing"

	"github.com/sakamotoryou/api-agg-two/internal/common/http_client"
	"github.com/stretchr/testify/assert"
)

func TestRequestDataOperation(t *testing.T) {
	t.Run("Encode parameter", func(t *testing.T) {
		type TestParam struct {
			number  int
			decimal *float32
		}

		var dd float32 = 12.33
		newReq := http_client.Param(
			TestParam{number: 123, decimal: &dd},
		)(http_client.Request{})
		result := newReq.EncodeParam()
		assert.Equal(t, "number=123&decimal=12.33", result)
	})

	t.Run("Encode body", func(t *testing.T) {
		type TestBody struct {
			Number  int      `json:"number"`
			Decimal *float32 `json:"decimal"`
		}
		var dd float32 = 12.33
		newReq := http_client.Json(
			TestBody{Number: 123, Decimal: &dd},
		)(http_client.Request{})
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
		newReq := http_client.Body(
			TestBody{Number: 123, Decimal: &dd},
		)(http_client.Request{})
		_, err := newReq.EncodeBody()
		assert.ErrorIs(t, err, http_client.ErrUnknownContentEncode)
	})
}

func TestRequestOptions(t *testing.T) {
	t.Run("Domain Option", func(t *testing.T) {
		newReq := http_client.Domain("example.com")
		assert.Equal(t, "example.com", newReq(http_client.Request{}).GetDomain())
	})

	t.Run("Path Option", func(t *testing.T) {
		newReq := http_client.Path("/api/v1")
		assert.Equal(t, "/api/v1", newReq(http_client.Request{}).GetPath())
	})

	t.Run("Get Method", func(t *testing.T) {
		newReq := http_client.Get()
		assert.Equal(t, "GET", newReq(http_client.Request{}).GetMethod())
	})

	t.Run("Post Method", func(t *testing.T) {
		newReq := http_client.Post()
		assert.Equal(t, "POST", newReq(http_client.Request{}).GetMethod())
	})

	t.Run("Put Method", func(t *testing.T) {
		newReq := http_client.Put()
		assert.Equal(t, "PUT", newReq(http_client.Request{}).GetMethod())
	})

	t.Run("Delete Method", func(t *testing.T) {
		newReq := http_client.Delete()
		assert.Equal(t, "DELETE", newReq(http_client.Request{}).GetMethod())
	})

	t.Run("Patch Method", func(t *testing.T) {
		newReq := http_client.Patch()
		assert.Equal(t, "PATCH", newReq(http_client.Request{}).GetMethod())
	})

	// t.Run("Head Method", func(t *testing.T) {
	// 	newReq := http_client.Head()
	// 	assert.Equal(t, "HEAD", newReq(http_client.Request{}).GetMethod())
	// })
	//
	// t.Run("Options Method", func(t *testing.T) {
	// 	newReq := http_client.Options()
	// 	assert.Equal(t, "OPTIONS", newReq(http_client.Request{}).GetMethod())
	// })

	t.Run("Http Header", func(t *testing.T) {
		newReq := http_client.Header(
			http_client.SetRequestHeader("Content-Type", "application/json"),
			http_client.SetRequestHeader("x-api-key", "my-custom-key"),
		)
		assert.Equal(t, "application/json", newReq(http_client.Request{}).GetHeader("Content-Type"))
		assert.Equal(t, "my-custom-key", newReq(http_client.Request{}).GetHeader("x-api-key"))
		assert.Equal(t, "", newReq(http_client.Request{}).GetHeader("non-existent"))
	})

	t.Run("Set request parameter", func(t *testing.T) {
		type RequestParam struct {
			id  string
			num int
		}

		newReq := http_client.Param(
			RequestParam{"uuid-123", 123},
		)
		assert.EqualValues(t, RequestParam{"uuid-123", 123}, newReq(http_client.Request{}).GetRawRequestData().GetParam())
	})

	t.Run("Set body parameter", func(t *testing.T) {
		type RequestBody struct {
			id  string
			num int
		}

		newReq := http_client.Body(
			RequestBody{"uuid-123", 123},
		)
		assert.EqualValues(t, RequestBody{"uuid-123", 123}, newReq(http_client.Request{}).GetRawRequestData().GetBody())
	})

	t.Run("Combination of Domain, Path, and Get", func(t *testing.T) {
		newReq, err := http_client.Build(
			http_client.Domain("example.com"),
			http_client.Get(),
			http_client.Path("/api/v1"),
			http_client.Header(
				http_client.SetRequestHeader("Content-Type", "application/json"),
				http_client.SetRequestHeader("x-api-key", "my-custom-key"),
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

		newReq, err := http_client.Build(
			http_client.Domain("example.com"),
			http_client.Get(),
			http_client.Path("/api/v1"),
			http_client.Header(
				http_client.SetRequestHeader("x-api-key", "my-custom-key"),
			),
			http_client.Param(RequestParam{"param", 99.99}),
			http_client.Json(RequestBody{"body", 9.99}),
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
		_, err := http_client.Build(
			http_client.Get(),
			http_client.Path("/api/v1"),
		)

		assert.Error(t, err)
		assert.ErrorIs(t, err, http_client.ErrDomainNameEmpty)
	})

	t.Run("Missing method option", func(t *testing.T) {
		_, err := http_client.Build(
			http_client.Domain("example.com"),
			http_client.Path("/api/v1"),
		)

		assert.Error(t, err)
		assert.ErrorIs(t, err, http_client.ErrRequestMethodEmpty)
	})

	t.Run("Missing path option", func(t *testing.T) {
		_, err := http_client.Build(
			http_client.Domain("example.com"),
			http_client.Get(),
		)

		assert.Error(t, err)
		assert.ErrorIs(t, err, http_client.ErrPathEmpty)
	})

	t.Run("Request Param is not a struct", func(t *testing.T) {
		_, err := http_client.Build(
			http_client.Domain("example.com"),
			http_client.Get(),
			http_client.Path("/api/v1"),
			http_client.Param("some string"),
		)

		assert.Error(t, err)
		assert.ErrorIs(t, err, http_client.ErrRequestParamInvalidType)
	})

	t.Run("Empty request param", func(t *testing.T) {
		_, err := http_client.Build(
			http_client.Domain("example.com"),
			http_client.Path("/api/v1"),
			http_client.Get(),
			http_client.Json(
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
		_, err := http_client.Build(
			http_client.Domain("example.com"),
			http_client.Get(),
			http_client.Path("/api/v1"),
			http_client.Body("some string"),
		)

		assert.Error(t, err)
		assert.ErrorIs(t, err, http_client.ErrRequestBodyInvalidType)
	})

	t.Run("Empty request body", func(t *testing.T) {
		_, err := http_client.Build(
			http_client.Domain("example.com"),
			http_client.Get(),
			http_client.Path("/api/v1"),
		)

		assert.NoError(t, err)
	})
}

func TestBulkRequestGenerate(t *testing.T) {
	t.Run("Generate three different requests in bulk with all success", func(t *testing.T) {
		case_one := []http_client.RequestOptions{
			http_client.Post(),
			http_client.Json(
				struct {
					SomeData string `json:"some_data"`
				}{
					SomeData: "Hello World",
				},
			),
			http_client.Domain("https://some.domain.local"),
			http_client.Path("api/v1/save-message"),
		}
		case_two := []http_client.RequestOptions{
			http_client.Get(),
			http_client.Domain("https://some.domain.local"),
			http_client.Path("api/v1/get-message-details"),
		}
		case_three := []http_client.RequestOptions{
			http_client.Get(),
			http_client.Domain("https://some.domain-2.local"),
			http_client.Path("api/v2/fetch-with"),
			http_client.Param(
				struct {
					ID string
				}{
					ID: "uuid-1234",
				},
			),
		}

		cases := [][]http_client.RequestOptions{
			case_one, case_two, case_three,
		}

		bulk_requests := http_client.BulkBuild(
			http_client.Prepare(cases[0]...),
			http_client.Prepare(cases[1]...),
			http_client.Prepare(cases[2]...),
		)

		assert.False(t, bulk_requests.HasError())

		for i, c := range cases {
			r, _ := http_client.Build(c...)
			assert.Equal(t, fmt.Sprintf("%+v", r), fmt.Sprintf("%+v", bulk_requests.Requests()[i]))
		}
	})

	t.Run("Generate three different requests in bulk with one bad builder", func(t *testing.T) {
		cases := [][]http_client.RequestOptions{
			{
				http_client.Post(),
				http_client.Json(
					struct {
						SomeData string `json:"some_data"`
					}{
						SomeData: "Hello World",
					},
				),
			},
			{
				http_client.Get(),
				http_client.Domain("https://some.domain.local"),
				http_client.Path("api/v1/get-message-details"),
			},
			{
				http_client.Get(),
				http_client.Domain("https://some.domain-2.local"),
				http_client.Path("api/v2/fetch-with"),
				http_client.Param(
					struct {
						ID string
					}{
						ID: "uuid-1234",
					},
				),
			},
		}

		bulk_requests := http_client.BulkBuild(
			http_client.Prepare(cases[0]...),
			http_client.Prepare(cases[1]...),
			http_client.Prepare(cases[2]...),
		)

		assert.True(t, bulk_requests.HasError())
		for i, c := range cases {
			_, err := http_client.Build(c...)
      assert.ErrorIs(t, bulk_requests.Errors()[i] ,err)
		}
	})
}
