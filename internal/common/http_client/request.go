package http_client

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/sakamotoryou/api-agg-two/internal/common/helper"
)

var (
	ErrJsonEncode           = errors.New("request body has fail to encoded to json.")
	ErrUnknownContentEncode = errors.New("request body has fail to encoded as content type is unknown.")
)

type RawRequestData struct {
	param any
	body  any
}

func (d RawRequestData) GetBody() any {
	return d.body
}

func (d RawRequestData) GetParam() any {
	return d.param
}

func (d RawRequestData) encodeParam() string {
	s := strings.Builder{}
	struct_val := reflect.ValueOf(d.param)
	t := reflect.VisibleFields(reflect.TypeOf(d.param))
	for i, v := range t {
		if i != 0 {
			s.WriteString("&")
		}

		var deref_data reflect.Value
		switch struct_val.Field(i).Kind() {
		case reflect.Pointer:
			deref_data = struct_val.Field(i).Elem()
		default:
			deref_data = struct_val.Field(i)
		}

		s.WriteString(strings.ToLower(v.Name))
		s.WriteString("=")
		s.WriteString(fmt.Sprintf("%v", deref_data))
	}

	return s.String()
}

func (d RawRequestData) encodeBody(content_type string) (string, error) {
	switch content_type {
	case "application/json":
		encoded_json, err := json.Marshal(d.body)
		if err != nil {
			return "", fmt.Errorf("%w:%v", ErrJsonEncode, err)
		}
		return string(encoded_json), nil
	}

	return "", ErrUnknownContentEncode
}

type RawRequestDataSetter func(RawRequestData) RawRequestData

func SetRequestParam(data any) RawRequestDataSetter {
	return func(d RawRequestData) RawRequestData {
		d.param = data
		return d
	}
}

func SetRequestBody(data any) RawRequestDataSetter {
	return func(d RawRequestData) RawRequestData {
		d.body = data
		return d
	}
}

type requestHeader struct {
	key   string
	value string
}
type requestHeaders []requestHeader

func SetRequestHeader(key, value string) requestHeader {
	return requestHeader{key, value}
}

func (h requestHeaders) Get(key string) requestHeader {
	for _, header := range h {
		if header.key == key {
			return header
		}
	}

	return requestHeader{}
}

type Request struct {
	domain   string
	path     string
	method   string
	header   requestHeaders
	raw_data RawRequestData

	// Processed request
	url   string
	param string
	body  string
}

func (r Request) GetDomain() string {
	return r.domain
}

func (r Request) GetPath() string {
	return r.path
}

func (r Request) GetMethod() string {
	return r.method
}

func (r Request) GetRawRequestData() RawRequestData {
	return r.raw_data
}

func (r Request) GetContentType() string {
	return r.header.Get("Content-Type").value
}

func (r Request) EncodeUrl() string {
	s := strings.Builder{}
	s.WriteString(r.domain)
	s.WriteString(r.path)
	s.WriteString(r.param)
	return s.String()
}

func (r Request) EncodeParam() string {
	if helper.IsInterfaceNil(r.raw_data.param) {
		return ""
	}
	return r.raw_data.encodeParam()
}

func (r Request) EncodeBody() (string, error) {
	if helper.IsInterfaceNil(r.raw_data.body) {
		return "", nil
	}
	return r.raw_data.encodeBody(r.GetContentType())
}

func (r Request) GetHeader(key string) string {
	for _, value := range r.header {
		if value.key == key {
			return value.value
		}
	}

	return ""
}

func (r Request) GetUrl() string {
	return r.url
}

func (r Request) GetParam() string {
	return r.param
}

func (r Request) GetBody() string {
	return r.body
}

func (r Request) GetBodyReader() io.Reader {
	if r.body == "" {
		return nil
	}

	return strings.NewReader(r.GetBody())
}

type RequestOptions func(Request) Request

func Domain(domain string) RequestOptions {
	return func(o Request) Request {
		o.domain = domain
		return o
	}
}

func Path(path string) RequestOptions {
	return func(o Request) Request {
		o.path = path
		return o
	}
}

func Get() RequestOptions {
	return func(o Request) Request {
		o.method = "GET"
		return o
	}
}

func Post() RequestOptions {
	return func(o Request) Request {
		o.method = "POST"
		return o
	}
}

func Put() RequestOptions {
	return func(o Request) Request {
		o.method = "PUT"
		return o
	}
}

func Delete() RequestOptions {
	return func(o Request) Request {
		o.method = "DELETE"
		return o
	}
}

func Patch() RequestOptions {
	return func(o Request) Request {
		o.method = "PATCH"
		return o
	}
}

// func Head() RequestOptions {
// 	return func(o Request) Request {
// 		o.method = "HEAD"
// 		return o
// 	}
// }
//
// func Options() RequestOptions {
// 	return func(o Request) Request {
// 		o.method = "OPTIONS"
// 		return o
// 	}
// }

func Header(headers ...requestHeader) RequestOptions {
	return func(o Request) Request {
		var h []requestHeader
		if headers == nil {
			h = make([]requestHeader, 0)
		} else {
			h = o.header
		}

		for _, value := range headers {
			h = append(h, value)
		}

		o.header = h
		return o
	}
}

func Param(param any) RequestOptions {
	return func(o Request) Request {
		newParam := SetRequestParam(param)
		o.raw_data.param = newParam(o.raw_data).param
		return o
	}
}

func Body(body any) RequestOptions {
	return func(o Request) Request {
		newParam := SetRequestBody(body)
		o.raw_data.body = newParam(o.raw_data).body
		return o
	}
}

func Json(body any) RequestOptions {
	return func(o Request) Request {
		o = Header(
			SetRequestHeader("Content-Type", "application/json"),
		)(o)

		if body != nil {
			o = Body(body)(o)
		}

		return o
	}
}

func Prepare(opts ...RequestOptions) func() (Request, error) {
	return func() (Request, error) {
		return Build(opts...)
	}
}

func Build(opts ...RequestOptions) (Request, error) {
	req := Request{}
	for _, setOption := range opts {
		req = setOption(req)
	}

	if err := checkOptions(req); err != nil {
		return Request{}, err
	}

	if result, err := process(req); err != nil {
		return Request{}, err
	} else {
		req = result()
	}

	return req, nil
}

var (
	ErrDomainNameEmpty                 = errors.New("empty domain name")
	ErrPathEmpty                       = errors.New("empty url path")
	ErrRequestMethodEmpty              = errors.New("empty Request method")
	ErrUnexpectNilAssignedResponseData = errors.New("expect type struct")
	ErrRequestParamInvalidType         = errors.New("expect request parameter to be struct")
	ErrRequestBodyInvalidType          = errors.New("expect request body to be struct")
)

func checkOptions(req Request) error {
	checks := []checkFunc{
		checkDomain(),
		checkPath(),
		checkMethod(),
		checkRequest(),
	}

	for _, check := range checks {
		if err := check(req); err != nil {
			return err
		}
	}

	return nil
}

type checkFunc func(req Request) error

func checkDomain() checkFunc {
	return func(req Request) error {
		if ok := helper.IsStringEmpty(req.domain); ok {
			return ErrDomainNameEmpty
		}

		return nil
	}
}

func checkPath() checkFunc {
	return func(req Request) error {
		if ok := helper.IsStringEmpty(req.path); ok {
			return ErrPathEmpty
		}

		return nil
	}
}

func checkMethod() checkFunc {
	return func(req Request) error {
		if ok := helper.IsStringEmpty(req.method); ok {
			return ErrRequestMethodEmpty
		}

		return nil
	}
}

func checkRequest() checkFunc {
	return func(req Request) error {
		if err := requestDataValidate(req); err != nil {
			return err
		}

		return nil
	}
}

func requestDataValidate(req Request) error {
	if ok := helper.IsInterfaceNil(req.GetRawRequestData().GetParam()); !ok {
		if reflect.TypeOf(req.raw_data.param).Kind() != reflect.Struct {
			return ErrRequestParamInvalidType
		}
	}

	if ok := helper.IsInterfaceNil(req.GetRawRequestData().GetBody()); !ok {
		if reflect.TypeOf(req.raw_data.body).Kind() != reflect.Struct {
			return ErrRequestBodyInvalidType
		}
	}

	return nil
}

func process(req Request) (func() Request, error) {
	req.url = req.EncodeUrl()
	req.param = req.EncodeParam()

	body, err := req.EncodeBody()
	if err != nil {
		return nil, fmt.Errorf("Process request body: %w", err)
	}

	req.body = body
	return func() Request {
		return req
	}, nil
}

// Can be considered a repo layer...
// Possible move on future stage
type BulkRequest struct {
	requests []Request
	errors   []error
}

type RequestBulkBuilder func() (Request, error)

func BulkBuild(builds ...RequestBulkBuilder) BulkRequest {
  requests := make([]Request, len(builds))
  errors := make([]error, len(builds))

  for i, build := range builds {
    requests[i], errors[i] = build()
  }

	return BulkRequest{
    requests: requests,
    errors: errors,
  }
}

func (br BulkRequest) HasError() bool {
  for _, v := range br.errors {
    if v != nil {
      return true
    }
  }

  return false
}

func (br BulkRequest) Errors() []error {
  return br.errors
}

func (br BulkRequest) Requests() []Request {
  return br.requests
} 
