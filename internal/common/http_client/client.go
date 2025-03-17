package http_client

import (
	"errors"
	"fmt"
	"net/http"
)

type ClientRequestFunc func() (Request, error)
type ClientResponseFunc func(*http.Response) (Response, error)

type Client struct {
	Request  Request
	Response Response
}

func BuildClient(
	request_build func() (Request, error),
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

func sender(req Request) (*http.Response, error) {
	http_req, err := http.NewRequest(
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
	reqFunc func() (Request, error),
	respFunc func(*http.Response) (Response, error),
	sendFunc func(Request) (*http.Response, error),
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
	reqFunc func() (Request, error),
	respFunc func(*http.Response) (Response, error),
) (Client, error) {
	return Send(reqFunc, respFunc, sender)
}

func PrepareRequest(reqFunc ...RequestOptions) func() (Request, error) {
	return func() (Request, error) {
		return Build(reqFunc...)
	}
}

func ResponseFuncBodyBind(v any) func(*http.Response) ResponseFunc {
	return func(resp *http.Response) ResponseFunc {
		return ResponseResolverFunc(
			BindData(v),
		)
	}
}

func PrepareResponse(respFunc ...func(*http.Response) ResponseFunc) func(*http.Response) (Response, error) {
	return func(resp *http.Response) (Response, error) {
		resp_opts := []ResponseFunc{
			ResponseHeader(resp.Header),
			ResponseBody(resp.Body),
			ResponseContentType(resp.Header.Get("Content-Type")),
			ResponseStatusCode(resp.StatusCode),
		}
		for _, fn := range respFunc {
			resp_opts = append(resp_opts, fn(resp))
		}
		return NewResponse(
			resp_opts...,
		)
	}
}
