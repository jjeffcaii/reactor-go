package hc_test

import (
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/hc"
	"github.com/jjeffcaii/reactor-go/mono"
	"github.com/jjeffcaii/reactor-go/scheduler"
	"github.com/jjeffcaii/reactor-go/tuple"
)

var httpBinUrl = "https://httpbin.org/anything"

func TestClient_Get(t *testing.T) {
	v, err := hc.Get("http://ipecho.net", hc.WithStringResponse()).Block(context.Background())
	assert.NoError(t, err, "request failed")
	assert.NotEmpty(t, v.(string), "should not be empty")

	cli := hc.NewClient(http.DefaultClient, hc.WithStringResponse())

	v, err = mono.
		Zip(
			cli.Get("http://ipecho.net/plain").SubscribeOn(scheduler.Parallel()),
			cli.Get("http://ipecho.net/json").SubscribeOn(scheduler.Parallel()),
			cli.Get("http://ipecho.net/xml").SubscribeOn(scheduler.Parallel()),
		).
		Block(context.Background())
	assert.NoError(t, err)

	tup := v.(tuple.Tuple)

	tup.ForEachWithIndex(func(v reactor.Any, e error, i int) bool {
		t.Logf("#%02d: value=%v, error=%v", i, v, e)
		return true
	})

	_, err = tup.Last()
	assert.Error(t, err)
}

type HttpBin struct {
	Method  string            `json:"method"`
	Url     string            `json:"url"`
	Headers map[string]string `json:"headers"`
}

func TestWithJSONBody(t *testing.T) {
	fakeHeaderKey := "X-Fake-Header"
	fakeHeaderValue := "fake value"
	v, err := hc.DefaultClient.
		Get(httpBinUrl, hc.WithRequestHijack(func(req *http.Request) {
			req.Header.Set(fakeHeaderKey, fakeHeaderValue)
		}), hc.WithJSONResponse(func() interface{} {
			return &HttpBin{}
		})).
		Block(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, http.MethodGet, v.(*HttpBin).Method)
	assert.Equal(t, fakeHeaderValue, v.(*HttpBin).Headers[fakeHeaderKey])
}

func TestRequests(t *testing.T) {
	validate := func(v interface{}, err error) {
		assert.NoError(t, err)
		_, ok := v.(*http.Response)
		assert.True(t, ok)
	}

	for _, next := range []func(string, string, io.Reader, ...hc.Option) mono.Mono{
		hc.Post, hc.Put, hc.Patch,
	} {
		v, err := next(httpBinUrl, "application/json", nil).Block(context.Background())
		validate(v, err)
	}
	validate(hc.Get(httpBinUrl).Block(context.Background()))
	validate(hc.Delete(httpBinUrl).Block(context.Background()))
}

func TestClient_Do(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, httpBinUrl, nil)
	v, err := hc.Do(req, hc.WithStringResponse()).Block(context.Background())
	assert.NoError(t, err, "should not return error")
	assert.NotEmpty(t, v, "should not be empty")
}

func TestDo_Failed(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1:8080/not-exists-path", nil)
	_, err := hc.Do(req).Block(context.Background())
	assert.Error(t, err)
}
