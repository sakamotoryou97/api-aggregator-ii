package http_client_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sakamotoryou/api-agg-two/internal/common/http_client"
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
		_, err := http_client.SendDefault(
			// Request prepare statement
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

			// Response resolver
			http_client.PrepareResponse(
				http_client.ResponseFuncBodyBind(&returnData),
			),
		)

		assert.NoError(t, err)
		assert.EqualValues(t, Data{"Hello World Appended"}, returnData)
	})
}
