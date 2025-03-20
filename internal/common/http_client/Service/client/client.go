package client

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/sakamotoryou/api-agg-two/internal/common/http_client/Entity/request"
	"github.com/sakamotoryou/api-agg-two/internal/common/http_client/Entity/response"
)

type (
	ClientRequestFunc  func() (request.Request, error)
	ClientResponseFunc func(*http.Response) (response.Response, error)
)

type Client struct {
	Request  request.Request
	Response response.Response
}

func BuildClient(
	request_build func() (request.Request, error),
) (Client, error) {
	req, err := request_build()
	if err != nil {
		return Client{}, err
	}

	return Client{
		Request: req,
	}, nil
}

var (
	ErrNewRequest = errors.New("Initialize new request fail")
	ErrRequestDo  = errors.New("On sending request fail")
)

func sender(req request.Request) (*http.Response, error) { http_req, err := http.NewRequest(
		req.GetMethod(),
		req.GetUrl(),
		req.GetBodyReader(),
	)
	if err != nil {
		return nil, fmt.Errorf("%w:%v", ErrNewRequest, err)
	}

	client := http.Client{}
	http_resp, err := client.Do(http_req)
	if err != nil {
		return nil, fmt.Errorf("%w:%v", ErrRequestDo, err)
	}

	return http_resp, nil
}

func Send(
	reqFunc func() (request.Request, error),
	respFunc func(*http.Response) (response.Response, error),
	sendFunc func(request.Request) (*http.Response, error),
) (Client, error) {
	req, err := reqFunc()
	if err != nil {
		return Client{}, fmt.Errorf("%w %v", err, "at reqFunc()")
	}

	http_resp, err := sendFunc(req)
	if err != nil {
		return Client{}, fmt.Errorf("%w %v", err, "at sendFunc()")
	}

	resp, err := respFunc(http_resp)
	if err != nil {
		return Client{}, fmt.Errorf("%w %v", err, "at respFunc()")
	}

	return Client{req, resp}, nil
}

func SendDefault(
	reqFunc func() (request.Request, error),
	respFunc func(*http.Response) (response.Response, error),
) (Client, error) {
	return Send(reqFunc, respFunc, sender)
}

func PrepareRequest(reqFunc ...request.RequestOptions) func() (request.Request, error) {
	return func() (request.Request, error) {
		return request.Build(reqFunc...)
	}
}

func PrepareResponse(respFunc ...response.ResponseFunc) func(*http.Response) (response.Response, error) {
	return func(resp *http.Response) (response.Response, error) {
		resp_opts := []response.ResponseFunc{
			response.ResponseHeader(resp.Header),
			response.ResponseBody(resp.Body),
			response.ResponseContentType(resp.Header.Get("Content-Type")),
			response.ResponseStatusCode(resp.StatusCode),
		}
		for _, fn := range respFunc {
			resp_opts = append(resp_opts, fn)
		}
		return response.NewResponse(
			resp_opts...,
		)
	}
}
