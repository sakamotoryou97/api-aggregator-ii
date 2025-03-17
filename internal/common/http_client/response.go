package http_client

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/sakamotoryou/api-agg-two/internal/common/helper"
)

type ResponseFunc func(Response) Response

type ResponseExtractor func(*http.Response) ResponseFunc

type Response struct {
	// TODO: In future, will make the ability to filter down the required header
	header       http.Header
	body         io.Reader
	content_type string
	code         int

	resolver []ResponseResolver
}

// func (r Response) OK() bool {
// 	return r.code == http.StatusOK
// }

func ResponseHeader(h http.Header) ResponseFunc {
	return func(resp Response) Response {
		resp.header = h
		return resp
	}
}

func ResponseBody(r io.Reader) ResponseFunc {
	return func(resp Response) Response {
		resp.body = r
		return resp
	}
}

func ResponseContentType(content_type string) ResponseFunc {
	return func(resp Response) Response {
		resp.content_type = content_type
		return resp
	}
}

func ResponseStatusCode(status_code int) ResponseFunc {
	return func(resp Response) Response {
		resp.code = status_code
		return resp
	}
}

func ResponseResolverFunc(fn ResponseResolver) ResponseFunc {
	return func(resp Response) Response {
		resp.resolver = append(resp.resolver, fn)
		return resp
	}
}

func NewResponse(ropts ...ResponseFunc) (Response, error) {
	resp := Response{}

	for _, ropt := range ropts {
		resp = ropt(resp)
	}

	if err := checkResponseOptions(resp); err != nil {
		return Response{}, err
	}

	if err := resp.Resolve(); err != nil {
		return Response{}, err
	}

	return resp, nil
}

func (resp Response) GetHeader() http.Header {
	return resp.header
}

func (resp Response) GetBody() io.Reader {
	return resp.body
}

func (resp Response) GetContentType() string {
	return resp.content_type
}

func (resp Response) GetStatusCode() int {
	return resp.code
}

var (
	ErrHeaderIsEmpty     = errors.New("response header is empty")
	ErrStatusCodeMissing = errors.New("response status code is empty")
)

type responseCheckFunc func(resp Response) error

func checkResponseOptions(resp Response) error {
	checks := []responseCheckFunc{
		checkHeader(),
		checkStatusCode(),
	}

	for _, check := range checks {
		if err := check(resp); err != nil {
			return err
		}
	}

	return nil
}

func checkHeader() responseCheckFunc {
	return func(resp Response) error {
		if resp.header == nil {
			return ErrHeaderIsEmpty
		}

		return nil
	}
}

func checkStatusCode() responseCheckFunc {
	return func(resp Response) error {
		if resp.code == 0 {
			return ErrStatusCodeMissing
		}

		return nil
	}
}

type DecodeBodyFunc func() error

var (
	ErrDataBindInvalid      = errors.New("The data has to be a pointer and not nil")
	ErrDefaultBindNotString = errors.New("The data bind has to be string type")
	ErrResponseBodyNil      = errors.New("Received response body is empty")
)

func (resp Response) decode(v any) error {
	switch resp.content_type {
	case "application/json":
		return resp.DecodeWith(
			func() error {
				if err := json.NewDecoder(resp.body).Decode(v); err != nil {
					return err
				}
				return nil
			},
		)

	default:
		return resp.DecodeWith(
			func() error {
				if !helper.IsString(v) {
					return ErrDefaultBindNotString
				}

				if !helper.IsPointer(v) || helper.IsDeepNil(v) {
					return ErrDataBindInvalid
				}

				buf := bytes.NewBuffer([]byte{})
				buf.ReadFrom(resp.body)

				str_ptr := v.(*string)
				*str_ptr = buf.String()
				return nil
			},
		)
	}
}

func (resp Response) DecodeWith(decode DecodeBodyFunc) error {
	return decode()
}

type ResponseResolver func(Response) error

func (resp Response) Resolve() error {
	for _, resolve := range resp.resolver {
		if err := resolve(resp); err != nil {
			return err
		}
	}

	return nil
}

func BindData(v any) ResponseResolver {
	return func(resp Response) error {
		return resp.decode(v)
	}
}

var (
	ErrBadRequest         = errors.New("bad request")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrNotFound           = errors.New("not found")
	ErrMethodNotAllowed   = errors.New("method not allowed")
	ErrConflict           = errors.New("conflict")
	ErrInternalServer     = errors.New("internal server error")
	ErrNotImplemented     = errors.New("not implemented")
	ErrServiceUnavailable = errors.New("service unavailable")
)

func BindError() ResponseResolver {
	return func(resp Response) error {
		switch resp.code {
		case http.StatusOK, http.StatusCreated, http.StatusNoContent:
			return nil
		case http.StatusBadRequest:
			return ErrBadRequest
		case http.StatusUnauthorized:
			return ErrUnauthorized
		case http.StatusForbidden:
			return ErrForbidden
		case http.StatusNotFound:
			return ErrNotFound
		case http.StatusMethodNotAllowed:
			return ErrMethodNotAllowed
		case http.StatusConflict:
			return ErrConflict
		case http.StatusInternalServerError:
			return ErrInternalServer
		case http.StatusNotImplemented:
			return ErrNotImplemented
		case http.StatusServiceUnavailable:
			return ErrServiceUnavailable
		default:
			return ErrInternalServer
		}
	}
}
