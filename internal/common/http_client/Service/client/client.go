package client

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/sakamotoryou/api-agg-two/internal/common/http_client/Entity/request"
	"github.com/sakamotoryou/api-agg-two/internal/common/http_client/Entity/response"
	"github.com/sakamotoryou/api-agg-two/internal/common/http_client/Entity/retry"
)

type (
	BeforeClientRequest func() (request.Request, error)
	OnClientRequest     func() (retry.Retry, error)
	AfterClientRequest  func(*http.Response) (response.Response, error)
)

type Client struct {
	Name  string
	Event Event
}

type Event struct {
	OnResolveBefore BeforeClientRequest
	OnClientRequest OnClientRequest
	OnResolveAfter  AfterClientRequest
}

func New(name string) Client {
	return Client{
		Name: name,
	}
}

func (c Client) Register(
	resolve_before BeforeClientRequest,
	resolve_on_req OnClientRequest,
	resolve_after AfterClientRequest,
) Client {
	e := Event{
		resolve_before, resolve_on_req, resolve_after,
	}
	c.Event = e
	return c
}

var (
	ErrNewRequest = errors.New("Initialize new request fail")
	ErrRequestDo  = errors.New("On sending request fail")
)

func do(req request.Request) (*http.Response, error) {
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

func (c Client) Send() error {
	req, err := c.Event.OnResolveBefore()
	if err != nil {
		return err
	}

	retry, err := c.Event.OnClientRequest()
	if err != nil {
		return err
	}

	var result *http.Response
	for retry.Next() {
		result, err = do(req)
		if err != nil {
			return err
		}

		retry.ValidateCode(result.StatusCode)
	}

	resp, err := c.Event.OnResolveAfter(result)
	if err != nil {
		return err
	}

	if err := resp.Resolve(); err != nil {
		return err
	}

	return nil
}

func BeforeDoRequest(reqFunc ...request.RequestOptions) BeforeClientRequest {
	return func() (request.Request, error) {
		return request.Build(reqFunc...)
	}
}

func OnDoRequest(retryFunc ...retry.RetryOptions) OnClientRequest {
	return func() (retry.Retry, error) {
		return retry.New(
      retryFunc...,
    )
	}
}

func AfterDoRequest(respFunc ...response.ResponseFunc) AfterClientRequest {
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
