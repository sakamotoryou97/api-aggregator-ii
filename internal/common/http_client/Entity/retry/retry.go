package retry

import (
	"errors"
	"time"
)

var (
	ErrExceedMaxRetry         = errors.New("exceed max retry")
	ErrInvalidMissingValidate = errors.New("invalid/missing validate option")
)

type Retry struct {
	Next         func() bool
	ValidateCode func(int)
}

type RetryOptions func(Retry) Retry

func NextOption(fn func() bool) RetryOptions {
	return func(r Retry) Retry {
		r.Next = fn
		return r
	}
}

func ValidateOption(fn func(int)) RetryOptions {
	return func(r Retry) Retry {
		r.ValidateCode = fn
		return r
	}
}

func New(opts ...RetryOptions) (Retry, error) {
	r := Retry{}

	for _, opt := range opts {
		r = opt(r)
	}
  
  // TODO: add validation on retry option

	return r, nil
}

func Default() []RetryOptions {
	return Simple(1, 0, func(code int) bool { return true })
}

func Simple(max_retries int, interval time.Duration, validate_opt func(int) bool) []RetryOptions {
	s := struct {
		interval    time.Duration
		max_retries int
		cur_retries int
	}{
		interval:    interval,
		max_retries: max_retries,
		cur_retries: 0,
	}

	return []RetryOptions{
		NextOption(func() bool {
			if s.cur_retries == s.max_retries {
				return false
			}

			if s.cur_retries > 0 {
				time.Sleep(s.interval)
			}

			s.cur_retries++
			return true
		}),
		ValidateOption(func(code int) {
			if validate_opt(code) {
				s.cur_retries = s.max_retries
			}
		}),
	}
}
