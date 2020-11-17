package hc

import (
	"context"
	"io"
	"net/http"

	"github.com/jjeffcaii/reactor-go/mono"
)

// DefaultClient is the default Client.
var DefaultClient = NewClient(http.DefaultClient)

// Option is the type of http client options.
type Option func(*option)

// Mono is alias of mono.Mono.
type Mono = mono.Mono

// Client is a reactor style http client which returns a Mono.
type Client interface {
	// Do sends an HTTP request and returns a response Mono.
	Do(req *http.Request, opts ...Option) Mono
	// Get sends a GET request and returns a response Mono.
	Get(url string, opts ...Option) Mono
	// Delete sends a DELETE request and returns a response Mono.
	Delete(url string, opts ...Option) Mono
	// Post sends a POST request and returns a response Mono.
	Post(url string, contentType string, body io.Reader, opts ...Option) Mono
	// Put sends a PUT request and returns a response Mono.
	Put(url string, contentType string, body io.Reader, opts ...Option) Mono
	// Patch sends a PATCH request and returns a response Mono.
	Patch(url string, contentType string, body io.Reader, opts ...Option) Mono
}

// Do sends an HTTP request and returns a response Mono.
func Do(req *http.Request, opts ...Option) Mono {
	return DefaultClient.Do(req, opts...)
}

// Get sends a GET request and returns a response Mono.
func Get(url string, opts ...Option) Mono {
	return DefaultClient.Get(url, opts...)
}

// Delete sends a DELETE request and returns a response Mono.
func Delete(url string, opts ...Option) Mono {
	return DefaultClient.Delete(url, opts...)
}

// Post sends a POST request and returns a response Mono.
func Post(url string, contentType string, body io.Reader, opts ...Option) Mono {
	return DefaultClient.Post(url, contentType, body, opts...)
}

// Put sends a PUT request and returns a response Mono.
func Put(url string, contentType string, body io.Reader, opts ...Option) Mono {
	return DefaultClient.Put(url, contentType, body, opts...)
}

// Patch sends a PATCH request and returns a response Mono.
func Patch(url string, contentType string, body io.Reader, opts ...Option) Mono {
	return DefaultClient.Patch(url, contentType, body, opts...)
}

// NewClient creates a new Client.
func NewClient(cli *http.Client, defaultOptions ...Option) Client {
	return &client{
		c:    cli,
		opts: defaultOptions,
	}
}

var _ Client = (*client)(nil)

type client struct {
	c    *http.Client
	opts []Option
}

func (h *client) Do(req *http.Request, opts ...Option) Mono {
	gen := func(ctx context.Context) (*http.Request, error) {
		return req.WithContext(ctx), nil
	}
	return h.execute(gen, opts...)
}

func (h *client) Get(url string, opts ...Option) Mono {
	return h.customExecute(http.MethodGet, url, "", nil, opts...)
}

func (h *client) Delete(url string, opts ...Option) Mono {
	return h.customExecute(http.MethodDelete, url, "", nil, opts...)
}

func (h *client) Post(url string, contentType string, body io.Reader, opts ...Option) Mono {
	return h.customExecute(http.MethodPost, url, contentType, body, opts...)
}

func (h *client) Put(url string, contentType string, body io.Reader, opts ...Option) Mono {
	return h.customExecute(http.MethodPut, url, contentType, body, opts...)
}

func (h *client) Patch(url string, contentType string, body io.Reader, opts ...Option) Mono {
	return h.customExecute(http.MethodPatch, url, contentType, body, opts...)
}

func (h *client) customExecute(method string, url string, contentType string, body io.Reader, opts ...Option) Mono {
	return h.execute(func(ctx context.Context) (*http.Request, error) {
		req, err := http.NewRequestWithContext(ctx, method, url, body)
		if err != nil {
			return nil, err
		}
		if len(contentType) > 0 {
			req.Header.Set("Content-Type", contentType)
		}
		return req, nil
	}, opts...)
}

func (h *client) execute(gen func(ctx context.Context) (*http.Request, error), opts ...Option) Mono {
	return mono.Create(func(ctx context.Context, s mono.Sink) {
		var (
			opt *option
			req *http.Request
			err error
		)

		// init option
		if len(opts)+len(h.opts) > 0 {
			opt = new(option)
			for _, it := range h.opts {
				it(opt)
			}
			for _, it := range opts {
				it(opt)
			}
		}

		req, err = gen(ctx)
		if err != nil {
			s.Error(err)
			return
		}

		if opt != nil && len(opt.reqChain) > 0 {
			for _, it := range opt.reqChain {
				it(req)
			}
		}

		var res *http.Response
		res, err = h.c.Do(req)
		if err != nil {
			s.Error(err)
			return
		}

		if opt == nil || opt.bodyParser == nil {
			s.Success(res)
			return
		}

		var body interface{}
		body, err = opt.bodyParser(res)
		if err != nil {
			s.Error(err)
			return
		}

		s.Success(body)
	})
}
