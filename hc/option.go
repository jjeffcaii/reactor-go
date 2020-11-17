package hc

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
)

type BodyParser func(*http.Response) (interface{}, error)

// WithStringResponse decodes the response body as string.
func WithStringResponse() Option {
	return WithResponseBodyParser(func(res *http.Response) (interface{}, error) {
		defer func() {
			if res.Body != nil {
				_ = res.Body.Close()
			}
		}()

		if !isSuccessful(res.StatusCode) {
			b, err := ioutil.ReadAll(res.Body)
			if err == nil {
				err = errors.New(string(b))
			}
			return nil, err
		}

		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		return string(b), nil
	})
}

// WithJSONResponse decodes the response body as json.
func WithJSONResponse(gen func() (ptr interface{})) Option {
	return func(o *option) {
		o.bodyParser = func(res *http.Response) (interface{}, error) {
			defer func() {
				if res.Body != nil {
					_ = res.Body.Close()
				}
			}()

			if !isSuccessful(res.StatusCode) {
				b, err := ioutil.ReadAll(res.Body)
				if err != nil {
					return nil, err
				}
				return nil, errors.New(string(b))
			}

			value := gen()

			if err := json.NewDecoder(res.Body).Decode(value); err != nil {
				return nil, err
			}
			return value, nil
		}
	}
}

// WithRequestHijack can be used to customize the http request.
func WithRequestHijack(h func(*http.Request)) Option {
	return func(o *option) {
		o.reqChain = append(o.reqChain, h)
	}
}

// WithResponseBodyParser set the response body parser.
func WithResponseBodyParser(bp BodyParser) Option {
	return func(o *option) {
		o.bodyParser = bp
	}
}

type option struct {
	bodyParser BodyParser
	reqChain   []func(*http.Request)
}

func isSuccessful(code int) bool {
	return code > 199 && code < 300
}
