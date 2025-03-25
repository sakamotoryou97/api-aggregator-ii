package async_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sakamotoryou/api-agg-two/internal/common/http_client/Entity/request"
	"github.com/sakamotoryou/api-agg-two/internal/common/http_client/Entity/response"
	"github.com/sakamotoryou/api-agg-two/internal/common/http_client/Service/async"
	"github.com/sakamotoryou/api-agg-two/internal/common/http_client/Service/client"
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
		result := async.Do(
			async.PrepareClientGenerator(
				client.PrepareRequest(
					request.Get(),
					request.Domain(server.URL),
					request.Path("/api/v1/recommended-response"),
				),

				client.PrepareResponse(
					response.OnSuccess(
						response.Decode(&recommended_response_data),
					),
				),
			),
			async.PrepareClientGenerator(
				client.PrepareRequest(
					request.Get(),
					request.Domain(server.URL),
					request.Path("/api/v1/additional-response"),
				),

				client.PrepareResponse(
					response.OnSuccess(
						response.Decode(&additional_response_data),
					),
				),
			),
			async.PrepareClientGenerator(
				client.PrepareRequest(
					request.Post(),
					request.Json(
						Data{
							SomeData: "Hello World",
						},
					),
					request.Domain(server.URL),
					request.Path("/api/v1/save-message"),
				),

				client.PrepareResponse(
					response.OnSuccess(
						response.Decode(&returnData),
					),
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
