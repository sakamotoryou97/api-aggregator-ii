package http_client_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sakamotoryou/api-agg-two/internal/common/http_client"
	"github.com/stretchr/testify/assert"
)

func TestAsyncDefaultClient(t *testing.T) {
	t.Run("Call two get request and one post request concurrently", func(t *testing.T) {
		type Data struct {
			SomeData string `json:"some_data"`
		}
		type RecommendedResponse struct {
			RecommendedResponse string `json:"recommended_response"`
		}
		type AddtionalMsg struct {
			AddtionalMsg string `json:"additional_response"`
		}

		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/recommended-response", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			b := RecommendedResponse{
				RecommendedResponse: "recommended msg",
			}

			return_b, _ := json.Marshal(b)
			w.Write(return_b)
		})
		mux.HandleFunc("GET /api/v1/additional-response", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			b := AddtionalMsg{
				AddtionalMsg: "additional msg",
			}

			return_b, _ := json.Marshal(b)
			w.Write(return_b)
		})
		mux.HandleFunc("POST /api/v1/save-message", func(w http.ResponseWriter, r *http.Request) {
			var data Data
			json.NewDecoder(r.Body).Decode(&data)

			data.SomeData += " Appended"

			b, _ := json.Marshal(data)

			w.Header().Set("Content-Type", "application/json")
			w.Write(b)
		})
		server := httptest.NewServer(mux)
		defer server.Close()

		done := make(chan struct{})
		defer close(done)

		var recommended_response_data RecommendedResponse
		var additional_response_data AddtionalMsg
		var returnData Data

		got := map[string]any{
			"/api/v1/save-message":         &returnData,
			"/api/v1/additional_response":  &additional_response_data,
			"/api/v1/recommended-response": &recommended_response_data,
		}
		async_client := http_client.AsyncDefault(
			done,
			http_client.PrepareClientGenerator(
				http_client.PrepareRequest(
					http_client.Get(),
					http_client.Domain(server.URL),
					http_client.Path("/api/v1/recommended-response"),
				),

				http_client.PrepareResponse(
					http_client.ResponseFuncBodyBind(&recommended_response_data),
				),
			),
			http_client.PrepareClientGenerator(
				http_client.PrepareRequest(
					http_client.Get(),
					http_client.Domain(server.URL),
					http_client.Path("/api/v1/additional-response"),
				),

				http_client.PrepareResponse(
					http_client.ResponseFuncBodyBind(&additional_response_data),
				),
			),
			http_client.PrepareClientGenerator(
				http_client.PrepareRequest(
					http_client.Post(),
					http_client.Json(
						Data{
							SomeData: "Hello World",
						},
					),
					http_client.Domain(server.URL),
					http_client.Path("/api/v1/save-message"),
				),

				http_client.PrepareResponse(
					http_client.ResponseFuncBodyBind(&returnData),
				),
			),
		)

		expect := map[string]any{
			"/api/v1/save-message":         &Data{"Hello World Appended"},
			"/api/v1/additional_response":  &AddtionalMsg{"additional msg"},
			"/api/v1/recommended-response": &RecommendedResponse{"recommended msg"},
		}

		for client := range async_client {
			assert.NoError(t, client.Err)

			path := client.Client.Request.GetPath()
			assert.EqualValues(t, expect[path], got[path])
		}
	})
}

// func TestHttpClientAsyncRequest(t *testing.T) {
// 	t.Run("Generate three Request struct into pipeline success", func(t *testing.T) {
// 		case_post_req_one := func() (http_client.Request, error) {
// 			return http_client.Build(
// 				http_client.Post(),
// 				http_client.Json(
// 					struct {
// 						SomeData string `json:"some_data"`
// 					}{
// 						SomeData: "Hello World",
// 					},
// 				),
// 				http_client.Domain("https://some.domain.local"),
// 				http_client.Path("api/v1/save-message"),
// 			)
// 		}
//
// 		case_get_req_two := func() (http_client.Request, error) {
// 			return http_client.Build(
// 				http_client.Get(),
// 				http_client.Domain("https://some.domain.local"),
// 				http_client.Path("api/v1/get-message-details"),
// 			)
// 		}
//
// 		case_get_req_three := func() (http_client.Request, error) {
// 			return http_client.Build(
// 				http_client.Get(),
// 				http_client.Domain("https://some.domain-2.local"),
// 				http_client.Path("api/v2/fetch-with"),
// 				http_client.Param(
// 					struct {
// 						ID string
// 					}{
// 						ID: "uuid-1234",
// 					},
// 				),
// 			)
// 		}
//
// 		cases := []TestCase{
// 			{
// 				expect: case_post_req_one,
// 			},
// 			{
// 				expect: case_get_req_two,
// 			},
// 			{
// 				expect: case_get_req_three,
// 			},
// 		}
//
// 		var request_async <-chan http_client.RequestAsync
// 		// done := make(chan struct{})
// 		// defer close(done)
// 		request_async = http_client.GenerateRequests(
// 			// done,
// 			case_post_req_one,
// 			case_get_req_two,
// 			case_get_req_three,
// 		)
//
// 		var get []http_client.RequestAsync
// 		for result := range request_async {
// 			get = append(get, result)
// 		}
//
// 		assert.Equal(t, 3, len(get))
// 		for i, res := range get {
// 			expect_request_async, expect_err := cases[i].expect()
// 			assert.NoError(t, expect_err)
// 			assert.EqualValues(t, expect_request_async, res.Request)
// 		}
// 	})
//
// 	t.Run("Generate Request structs into pipeline with one fail case", func(t *testing.T) {
// 		type TestCase struct {
// 			expect http_client.RequestGenerator
// 		}
//
// 		case_post_req_one := func() (http_client.Request, error) {
// 			return http_client.Build(
// 				http_client.Post(),
// 				http_client.Json(
// 					struct {
// 						SomeData string `json:"some_data"`
// 					}{
// 						SomeData: "Hello World",
// 					},
// 				),
// 				http_client.Domain("https://some.domain.local"),
// 				http_client.Path("api/v1/save-message"),
// 			)
// 		}
//
// 		case_get_req_two := func() (http_client.Request, error) {
// 			return http_client.Build(
// 				http_client.Get(),
// 				http_client.Domain("https://some.domain.local"),
// 			)
// 		}
//
// 		case_get_req_three := func() (http_client.Request, error) {
// 			return http_client.Build(
// 				http_client.Get(),
// 				http_client.Domain("https://some.domain-2.local"),
// 				http_client.Path("api/v2/fetch-with"),
// 				http_client.Param(
// 					struct {
// 						ID string
// 					}{
// 						ID: "uuid-1234",
// 					},
// 				),
// 			)
// 		}
//
// 		cases := []TestCase{
// 			{
// 				expect: case_post_req_one,
// 			},
// 			{
// 				expect: case_get_req_two,
// 			},
// 			{
// 				expect: case_get_req_three,
// 			},
// 		}
//
// 		var request_async <-chan http_client.RequestAsync
// 		// done := make(chan struct{})
// 		// defer close(done)
// 		request_async = http_client.GenerateRequests(
// 			// done,
// 			case_post_req_one,
// 			case_get_req_two,
// 			case_get_req_three,
// 		)
//
// 		var get []http_client.RequestAsync
// 		for result := range request_async {
// 			get = append(get, result)
// 		}
//
// 		assert.Equal(t, 2, len(get))
// 		for i, res := range get {
// 			expect_request_async, expect_err := cases[i].expect()
// 			assert.ErrorIs(t, res.Err, expect_err)
// 			assert.EqualValues(t, expect_request_async, res.Request)
// 		}
// 	})
//
// 	// t.Run("Generate Request struct with early fail", func(t *testing.T) {
// 	//
// 	// })
// }
//
// // func TestHttpClientAsyncResolver(t *testing.T) {
// // 	t.Run("Resolve multiple requests with wait() function", func(t *testing.T) {
// //     results, wait := http_client.ResolveAsync()
// //     wait()
// // 	})
// // }
