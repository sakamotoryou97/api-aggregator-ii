package response

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/sakamotoryou/api-agg-two/internal/common/helper"
)

type ResponseFunc func(Response) Response

type ResponseExtractor func(*http.Response) ResponseFunc

type Response struct {
	header       http.Header
	body         io.Reader
	content_type string
	code         int

	on_success []SuccessResolver
	on_reject  RejectResolver
}

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

func OnSuccess(fns ...SuccessResolver) ResponseFunc {
	return func(resp Response) Response {
		resp.on_success = fns
		return resp
	}
}

func OnReject(fns ...func(Response) RejectResolver) ResponseFunc {
	return func(resp Response) Response {
		resp.on_reject = createDefaultErrors()

		for _, fn := range fns {
			resp.on_reject = fn(resp)
		}
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
	ErrHeaderIsEmpty          = errors.New("response header is empty")
	ErrStatusCodeMissing      = errors.New("response status code is empty")
	ErrMissingOnSuccessAction = errors.New("response success action is empty")
)

type responseCheckFunc func(resp Response) error

func checkResponseOptions(resp Response) error {
	checks := []responseCheckFunc{
		checkHeader(),
		checkStatusCode(),
		checkSuccessAction(),
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

func checkSuccessAction() responseCheckFunc {
	return func(resp Response) error {
		if resp.on_success == nil {
			return ErrMissingOnSuccessAction
		}

		return nil
	}
}

func Decode(v any) func(Response) ResponseErrorFunc {
	return func(resp Response) ResponseErrorFunc {
		err := resp.decode(v)
		if err != nil {
			return Stack(err.Error())
		}

		return Skip()
	}
}

type ResponseError struct {
	front_message string
	stack_message []string
}

func (r ResponseError) Error() string {
	return r.front_message
}

func (r ResponseError) ReturnIfHasError() error {
	if r.front_message == "" {
		return nil
	}

	return r
}

func SetClient(s string) ResponseErrorFunc {
	return func(r ResponseError) ResponseError {
		if r.front_message == "" {
      r.front_message = s
		} else {
			r.front_message = r.front_message + ";" + s
		}
		return r
	}
}

func Stack(s string) ResponseErrorFunc {
	return func(r ResponseError) ResponseError {
		r.stack_message = append(r.stack_message, s)
		return r
	}
}

// Do nothing on the particular conditions.
func Skip() ResponseErrorFunc {
	return func(r ResponseError) ResponseError {
		return r
	}
}

func (r ResponseError) Log() string {
	return fmt.Sprintf(`
    Front Message: %s,
    Stack Message: %s
    `, r.front_message, func() string {
		str := strings.Builder{}
		for _, s := range r.stack_message {
			str.WriteString(s)
		}
		return str.String()
	}())
}

type RejectResolver struct {
	BadRequest                  func() error
	Unauthorized                func() error
	PaymentRequired             func() error
	Forbidden                   func() error
	NotFound                    func() error
	MethodNotAllowed            func() error
	NotAcceptable               func() error
	ProxyAuthRequired           func() error
	RequestTimeout              func() error
	Conflict                    func() error
	Gone                        func() error
	LengthRequired              func() error
	PreconditionFailed          func() error
	PayloadTooLarge             func() error
	URITooLong                  func() error
	UnsupportedMediaType        func() error
	RangeNotSatisfiable         func() error
	ExpectationFailed           func() error
	Teapot                      func() error
	MisdirectedRequest          func() error
	UnprocessableEntity         func() error
	Locked                      func() error
	FailedDependency            func() error
	TooEarly                    func() error
	UpgradeRequired             func() error
	PreconditionRequired        func() error
	TooManyRequests             func() error
	RequestHeaderFieldsTooLarge func() error
	UnavailableForLegalReasons  func() error
	InternalServerError         func() error
	NotImplemented              func() error
	BadGateway                  func() error
	ServiceUnavailable          func() error
	GatewayTimeout              func() error
}

func createDefaultErrors() RejectResolver {
	return RejectResolver{
		BadRequest:                  func() error { return ResponseError{"Bad request", nil} },
		Unauthorized:                func() error { return ResponseError{"Unauthorized", nil} },
		PaymentRequired:             func() error { return ResponseError{"Payment required", nil} },
		Forbidden:                   func() error { return ResponseError{"Forbidden", nil} },
		NotFound:                    func() error { return ResponseError{"Not found", nil} },
		MethodNotAllowed:            func() error { return ResponseError{"Method not allowed", nil} },
		NotAcceptable:               func() error { return ResponseError{"Not acceptable", nil} },
		ProxyAuthRequired:           func() error { return ResponseError{"Proxy authentication required", nil} },
		RequestTimeout:              func() error { return ResponseError{"Request timeout", nil} },
		Conflict:                    func() error { return ResponseError{"Conflict", nil} },
		Gone:                        func() error { return ResponseError{"Gone", nil} },
		LengthRequired:              func() error { return ResponseError{"Length required", nil} },
		PreconditionFailed:          func() error { return ResponseError{"Precondition failed", nil} },
		PayloadTooLarge:             func() error { return ResponseError{"Payload too large", nil} },
		URITooLong:                  func() error { return ResponseError{"URI too long", nil} },
		UnsupportedMediaType:        func() error { return ResponseError{"Unsupported media type", nil} },
		RangeNotSatisfiable:         func() error { return ResponseError{"Range not satisfiable", nil} },
		ExpectationFailed:           func() error { return ResponseError{"Expectation failed", nil} },
		Teapot:                      func() error { return ResponseError{"I'm a teapot", nil} },
		MisdirectedRequest:          func() error { return ResponseError{"Misdirected request", nil} },
		UnprocessableEntity:         func() error { return ResponseError{"Unprocessable entity", nil} },
		Locked:                      func() error { return ResponseError{"Locked", nil} },
		FailedDependency:            func() error { return ResponseError{"Failed dependency", nil} },
		TooEarly:                    func() error { return ResponseError{"Too early", nil} },
		UpgradeRequired:             func() error { return ResponseError{"Upgrade required", nil} },
		PreconditionRequired:        func() error { return ResponseError{"Precondition required", nil} },
		TooManyRequests:             func() error { return ResponseError{"Too many requests", nil} },
		RequestHeaderFieldsTooLarge: func() error { return ResponseError{"Request header fields too large", nil} },
		UnavailableForLegalReasons:  func() error { return ResponseError{"Unavailable for legal reasons", nil} },
		InternalServerError:         func() error { return ResponseError{"Internal server error", nil} },
		NotImplemented:              func() error { return ResponseError{"Not implemented", nil} },
		BadGateway:                  func() error { return ResponseError{"Bad gateway", nil} },
		ServiceUnavailable:          func() error { return ResponseError{"Service unavailable", nil} },
		GatewayTimeout:              func() error { return ResponseError{"Gateway timeout", nil} },
	}
}

type ResponseErrorFunc func(ResponseError) ResponseError

func OnBadRequest(opts ...func(Response) ResponseErrorFunc) func(Response) RejectResolver {
	return func(resp Response) RejectResolver {
		resp.on_reject.BadRequest = process(resp, opts...)
		return resp.on_reject
	}
}

func OnUnauthorized(opts ...func(Response) ResponseErrorFunc) func(Response) RejectResolver {
	return func(resp Response) RejectResolver {
		resp.on_reject.Unauthorized = process(resp, opts...)
		return resp.on_reject
	}
}

func OnPaymentRequired(opts ...func(Response) ResponseErrorFunc) func(Response) RejectResolver {
	return func(resp Response) RejectResolver {
		resp.on_reject.PaymentRequired = process(resp, opts...)
		return resp.on_reject
	}
}

func OnForbidden(opts ...func(Response) ResponseErrorFunc) func(Response) RejectResolver {
	return func(resp Response) RejectResolver {
		resp.on_reject.Forbidden = process(resp, opts...)
		return resp.on_reject
	}
}

func OnNotFound(opts ...func(Response) ResponseErrorFunc) func(Response) RejectResolver {
	return func(resp Response) RejectResolver {
		resp.on_reject.NotFound = process(resp, opts...)
		return resp.on_reject
	}
}

func OnMethodNotAllowed(opts ...func(Response) ResponseErrorFunc) func(Response) RejectResolver {
	return func(resp Response) RejectResolver {
		resp.on_reject.MethodNotAllowed = process(resp, opts...)
		return resp.on_reject
	}
}

func OnNotAcceptable(opts ...func(Response) ResponseErrorFunc) func(Response) RejectResolver {
	return func(resp Response) RejectResolver {
		resp.on_reject.NotAcceptable = process(resp, opts...)
		return resp.on_reject
	}
}

func OnProxyAuthRequired(opts ...func(Response) ResponseErrorFunc) func(Response) RejectResolver {
	return func(resp Response) RejectResolver {
		resp.on_reject.ProxyAuthRequired = process(resp, opts...)
		return resp.on_reject
	}
}

func OnRequestTimeout(opts ...func(Response) ResponseErrorFunc) func(Response) RejectResolver {
	return func(resp Response) RejectResolver {
		resp.on_reject.RequestTimeout = process(resp, opts...)
		return resp.on_reject
	}
}

func OnConflict(opts ...func(Response) ResponseErrorFunc) func(Response) RejectResolver {
	return func(resp Response) RejectResolver {
		resp.on_reject.Conflict = process(resp, opts...)
		return resp.on_reject
	}
}

func OnGone(opts ...func(Response) ResponseErrorFunc) func(Response) RejectResolver {
	return func(resp Response) RejectResolver {
		resp.on_reject.Gone = process(resp, opts...)
		return resp.on_reject
	}
}

func OnLengthRequired(opts ...func(Response) ResponseErrorFunc) func(Response) RejectResolver {
	return func(resp Response) RejectResolver {
		resp.on_reject.LengthRequired = process(resp, opts...)
		return resp.on_reject
	}
}

func OnPreconditionFailed(opts ...func(Response) ResponseErrorFunc) func(Response) RejectResolver {
	return func(resp Response) RejectResolver {
		resp.on_reject.PreconditionFailed = process(resp, opts...)
		return resp.on_reject
	}
}

func OnPayloadTooLarge(opts ...func(Response) ResponseErrorFunc) func(Response) RejectResolver {
	return func(resp Response) RejectResolver {
		resp.on_reject.PayloadTooLarge = process(resp, opts...)
		return resp.on_reject
	}
}

func OnURITooLong(opts ...func(Response) ResponseErrorFunc) func(Response) RejectResolver {
	return func(resp Response) RejectResolver {
		resp.on_reject.URITooLong = process(resp, opts...)
		return resp.on_reject
	}
}

func OnUnsupportedMediaType(opts ...func(Response) ResponseErrorFunc) func(Response) RejectResolver {
	return func(resp Response) RejectResolver {
		resp.on_reject.UnsupportedMediaType = process(resp, opts...)
		return resp.on_reject
	}
}

func OnRangeNotSatisfiable(opts ...func(Response) ResponseErrorFunc) func(Response) RejectResolver {
	return func(resp Response) RejectResolver {
		resp.on_reject.RangeNotSatisfiable = process(resp, opts...)
		return resp.on_reject
	}
}

func OnExpectationFailed(opts ...func(Response) ResponseErrorFunc) func(Response) RejectResolver {
	return func(resp Response) RejectResolver {
		resp.on_reject.ExpectationFailed = process(resp, opts...)
		return resp.on_reject
	}
}

func OnTeapot(opts ...func(Response) ResponseErrorFunc) func(Response) RejectResolver {
	return func(resp Response) RejectResolver {
		resp.on_reject.Teapot = process(resp, opts...)
		return resp.on_reject
	}
}

func OnMisdirectedRequest(opts ...func(Response) ResponseErrorFunc) func(Response) RejectResolver {
	return func(resp Response) RejectResolver {
		resp.on_reject.MisdirectedRequest = process(resp, opts...)
		return resp.on_reject
	}
}

func OnUnprocessableEntity(opts ...func(Response) ResponseErrorFunc) func(Response) RejectResolver {
	return func(resp Response) RejectResolver {
		resp.on_reject.UnprocessableEntity = process(resp, opts...)
		return resp.on_reject
	}
}

func OnLocked(opts ...func(Response) ResponseErrorFunc) func(Response) RejectResolver {
	return func(resp Response) RejectResolver {
		resp.on_reject.Locked = process(resp, opts...)
		return resp.on_reject
	}
}

func OnFailedDependency(opts ...func(Response) ResponseErrorFunc) func(Response) RejectResolver {
	return func(resp Response) RejectResolver {
		resp.on_reject.FailedDependency = process(resp, opts...)
		return resp.on_reject
	}
}

func OnTooEarly(opts ...func(Response) ResponseErrorFunc) func(Response) RejectResolver {
	return func(resp Response) RejectResolver {
		resp.on_reject.TooEarly = process(resp, opts...)
		return resp.on_reject
	}
}

func OnUpgradeRequired(opts ...func(Response) ResponseErrorFunc) func(Response) RejectResolver {
	return func(resp Response) RejectResolver {
		resp.on_reject.UpgradeRequired = process(resp, opts...)
		return resp.on_reject
	}
}

func OnPreconditionRequired(opts ...func(Response) ResponseErrorFunc) func(Response) RejectResolver {
	return func(resp Response) RejectResolver {
		resp.on_reject.PreconditionRequired = process(resp, opts...)
		return resp.on_reject
	}
}

func OnTooManyRequests(opts ...func(Response) ResponseErrorFunc) func(Response) RejectResolver {
	return func(resp Response) RejectResolver {
		resp.on_reject.TooManyRequests = process(resp, opts...)
		return resp.on_reject
	}
}

func OnRequestHeaderFieldsTooLarge(opts ...func(Response) ResponseErrorFunc) func(Response) RejectResolver {
	return func(resp Response) RejectResolver {
		resp.on_reject.RequestHeaderFieldsTooLarge = process(resp, opts...)
		return resp.on_reject
	}
}

func OnUnavailableForLegalReasons(opts ...func(Response) ResponseErrorFunc) func(Response) RejectResolver {
	return func(resp Response) RejectResolver {
		resp.on_reject.UnavailableForLegalReasons = process(resp, opts...)
		return resp.on_reject
	}
}

func OnInternalServerError(opts ...func(Response) ResponseErrorFunc) func(Response) RejectResolver {
	return func(resp Response) RejectResolver {
		resp.on_reject.InternalServerError = process(resp, opts...)
		return resp.on_reject
	}
}

func OnNotImplemented(opts ...func(Response) ResponseErrorFunc) func(Response) RejectResolver {
	return func(resp Response) RejectResolver {
		resp.on_reject.NotImplemented = process(resp, opts...)
		return resp.on_reject
	}
}

func OnBadGateway(opts ...func(Response) ResponseErrorFunc) func(Response) RejectResolver {
	return func(resp Response) RejectResolver {
		resp.on_reject.BadGateway = process(resp, opts...)
		return resp.on_reject
	}
}

func OnServiceUnavailable(opts ...func(Response) ResponseErrorFunc) func(Response) RejectResolver {
	return func(resp Response) RejectResolver {
		resp.on_reject.ServiceUnavailable = process(resp, opts...)
		return resp.on_reject
	}
}

func OnGatewayTimeout(opts ...func(Response) ResponseErrorFunc) func(Response) RejectResolver {
	return func(resp Response) RejectResolver {
		resp.on_reject.GatewayTimeout = process(resp, opts...)
		return resp.on_reject
	}
}

func SetClientError(front_message string) func(r Response) ResponseErrorFunc {
	return func(r Response) ResponseErrorFunc {
		return SetClient(front_message)
	}
}

func process(resp Response, actions ...func(Response) ResponseErrorFunc) func() error {
	var err ResponseError
	return func() error {
		for _, run := range actions {
			err_action := run(resp)
			err = err_action(err)
		}

		return err.ReturnIfHasError()
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
		if err := json.NewDecoder(resp.body).Decode(v); err != nil {
			return err
		}
		return nil

	default:
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
	}
}

func (resp Response) Success() bool {
	return resp.GetStatusCode() == http.StatusOK ||
		resp.GetStatusCode() == http.StatusCreated ||
		resp.GetStatusCode() == http.StatusNoContent
}

type ResponseResolver func(Response) ResponseErrorFunc

func (resp Response) Resolve() error {
	if resp.Success() {
		return resp.ResolveSuccess()
	}

	return resp.ResolveReject()
}

func (resp Response) ResolveSuccess() error {
	var err ResponseError
	for _, run := range resp.on_success {
		err_action := run(resp)
		err = err_action(err)
	}

	return err.ReturnIfHasError()
}

func (resp Response) ResolveReject() error {
	code := resp.GetStatusCode()
	switch code {
	case http.StatusBadRequest:
		return resp.on_reject.BadRequest()
	case http.StatusUnauthorized:
		return resp.on_reject.Unauthorized()
	case http.StatusPaymentRequired:
		return resp.on_reject.PaymentRequired()
	case http.StatusForbidden:
		return resp.on_reject.Forbidden()
	case http.StatusNotFound:
		return resp.on_reject.NotFound()
	case http.StatusMethodNotAllowed:
		return resp.on_reject.MethodNotAllowed()
	case http.StatusNotAcceptable:
		return resp.on_reject.NotAcceptable()
	case http.StatusProxyAuthRequired:
		return resp.on_reject.ProxyAuthRequired()
	case http.StatusRequestTimeout:
		return resp.on_reject.RequestTimeout()
	case http.StatusConflict:
		return resp.on_reject.Conflict()
	case http.StatusGone:
		return resp.on_reject.Gone()
	case http.StatusLengthRequired:
		return resp.on_reject.LengthRequired()
	case http.StatusPreconditionFailed:
		return resp.on_reject.PreconditionFailed()
	case http.StatusRequestEntityTooLarge:
		return resp.on_reject.PayloadTooLarge()
	case http.StatusRequestURITooLong:
		return resp.on_reject.URITooLong()
	case http.StatusUnsupportedMediaType:
		return resp.on_reject.UnsupportedMediaType()
	case http.StatusRequestedRangeNotSatisfiable:
		return resp.on_reject.RangeNotSatisfiable()
	case http.StatusExpectationFailed:
		return resp.on_reject.ExpectationFailed()
	case http.StatusTeapot:
		return resp.on_reject.Teapot()
	case http.StatusMisdirectedRequest:
		return resp.on_reject.MisdirectedRequest()
	case http.StatusUnprocessableEntity:
		return resp.on_reject.UnprocessableEntity()
	case http.StatusLocked:
		return resp.on_reject.Locked()
	case http.StatusFailedDependency:
		return resp.on_reject.FailedDependency()
	case http.StatusTooEarly:
		return resp.on_reject.TooEarly()
	case http.StatusUpgradeRequired:
		return resp.on_reject.UpgradeRequired()
	case http.StatusPreconditionRequired:
		return resp.on_reject.PreconditionRequired()
	case http.StatusTooManyRequests:
		return resp.on_reject.TooManyRequests()
	case http.StatusRequestHeaderFieldsTooLarge:
		return resp.on_reject.RequestHeaderFieldsTooLarge()
	case http.StatusUnavailableForLegalReasons:
		return resp.on_reject.UnavailableForLegalReasons()
	case http.StatusInternalServerError:
		return resp.on_reject.InternalServerError()
	case http.StatusNotImplemented:
		return resp.on_reject.NotImplemented()
	case http.StatusBadGateway:
		return resp.on_reject.BadGateway()
	case http.StatusServiceUnavailable:
		return resp.on_reject.ServiceUnavailable()
	case http.StatusGatewayTimeout:
		return resp.on_reject.GatewayTimeout()
	default:
		return ResponseError{"Oops! Something went wrong.", []string{fmt.Sprintf("%d Unknown Error", code)}}
	}
}
