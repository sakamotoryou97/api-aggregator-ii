package client_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sakamotoryou/api-agg-two/internal/common/http_client/Entity/request"
	"github.com/sakamotoryou/api-agg-two/internal/common/http_client/Entity/response"
	"github.com/sakamotoryou/api-agg-two/internal/common/http_client/Service/client"
	"github.com/stretchr/testify/assert"
)

func TestRequestAndResponseOptions(t *testing.T) {
	t.Run("Declare request and response and become Client struct", func(t *testing.T) {
		type Data struct {
			SomeData string `json:"some_data"`
		}

		mux := http.NewServeMux()
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

		var returnData Data
		_, err := client.SendDefault(
			// Request prepare statement
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

			// Response resolver
			client.PrepareResponse(
        response.OnSuccess(
          response.Decode(&returnData),
        ),
        response.OnReject(),
			),
		)

		assert.NoError(t, err)
		assert.EqualValues(t, Data{"Hello World Appended"}, returnData)
	})
}
