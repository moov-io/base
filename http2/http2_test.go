package http2_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	gohttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	http "github.com/moov-io/base/http2"
	"github.com/moov-io/base/log"
)

func TestSuite(t *testing.T) {
	suite.Run(t, new(Suite))
}

type Suite struct {
	suite.Suite
	router *http.Router
}

func (s *Suite) SetupTest() {
	logger := log.NewDefaultLogger()
	s.router = http.NewRouter(logger)
}

func (s *Suite) TestSimplePing() {
	pingHandler := func(req http.Request) http.Response {
		return http.Ok(nil)
	}
	s.router.SetRoute("/ping").Get(pingHandler)
	s.router.Build()

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	recorder := httptest.NewRecorder()
	s.router.ServeHTTP(recorder, req)

	resp := recorder.Result()

	s.Equal(http.StatusOK, resp.StatusCode)
}

func (s *Suite) TestPathVar() {
	tests := []struct {
		name               string
		pathVarKey         string
		expectedStatusCode int
		expectedError      string
	}{
		{
			name:               "success: path var exists",
			pathVarKey:         "customerID",
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "fail",
			pathVarKey:         "foobar",
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      "path variable 'foobar' not found in request URL",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			handler := func(req http.Request) http.Response {
				req.Var(tc.pathVarKey)
				return http.Ok(nil)
			}
			s.router.SetRoute("/customers/{customerID}").Get(handler)
			s.router.Build()

			url := fmt.Sprintf("/customers/%s", tc.pathVarKey)
			req := httptest.NewRequest(http.MethodGet, url, nil)
			recorder := httptest.NewRecorder()
			s.router.ServeHTTP(recorder, req)
			resp := recorder.Result()

			s.Equal(tc.expectedStatusCode, resp.StatusCode)

			if tc.expectedError != "" {
				var got struct {
					Error string `json:"error"`
				}
				body, err := ioutil.ReadAll(resp.Body)
				s.NoError(err)

				err = json.Unmarshal(body, &got)
				s.NoError(err)

				s.Equal(tc.expectedError, got.Error)
			}
		})
	}
}

func GetCustomers(req http.Request) http.Response {
	req.ParseBody("")
	res := map[string]string{"hi": "woah"}
	// _ = req.GetVar("some-var")
	// req.GetHeader("h")

	return http.Ok(res)
}

func CreateCustomer(req http.Request) http.Response {
	return http.OkEmpty()
}

func NewFixture(t *testing.T, router *http.Router) Fixture {
	return Fixture{
		assert: require.New(t),
	}
}

type Fixture struct {
	assert  *require.Assertions
	handler gohttp.Handler
}

type Request struct {
	assert  *require.Assertions
	inner   *gohttp.Request
	handler gohttp.Handler
}

func (r *Request) SetBody(data interface{}) *Request {
	b, err := json.Marshal(data)
	if err != nil {
		r.assert.FailNow("marshal into JSON: \n%v", data)
	}
	r.inner.Body = ioutil.NopCloser(bytes.NewBuffer(b))
	r.inner.Header.Set("Content-Type", "application/json")
	return r
}

func (r *Request) SetHeader(key string, val string) *Request {
	r.inner.Header.Set(key, val)
	return r
}

func (r *Request) Do() *Response {
	recorder := httptest.NewRecorder()
	r.handler.ServeHTTP(recorder, r.inner)

	return &Response{
		assert: r.assert,
		inner:  recorder.Result(),
	}
}

func (f *Fixture) Get(url string) *Request {
	return f.call(http.MethodGet, url)
}

func (f *Fixture) Post(url string) *Request {
	return f.call(http.MethodPost, url)
}

func (f *Fixture) call(method string, url string) *Request {
	req := httptest.NewRequest(method, url, nil)
	return &Request{
		assert: f.assert,
		inner:  req,
	}
}

type Response struct {
	assert *require.Assertions
	inner  *gohttp.Response
}

type ErrorBody struct {
	Error string `json:"error"`
}

func (r *Response) ParseBody(obj interface{}) *Response {
	r.assert.NotEmpty(r.inner.Body)

	b, err := ioutil.ReadAll(r.inner.Body)
	r.assert.NoError(err)
	defer r.inner.Body.Close()

	err = json.Unmarshal(b, obj)
	r.assert.NoError(err)

	return r
}

func (r *Response) HasCode(statusCode int) *Response {
	r.assert.Equal(statusCode, r.inner.StatusCode)
	return r
}

func (r *Response) HasEmptyBody() *Response {
	b, err := ioutil.ReadAll(r.inner.Body)
	r.assert.NoError(err)
	r.assert.Empty(b)
	return r
}

func (r *Response) HasHeader(key string, val string) *Response {
	got := r.inner.Header.Get(key)
	r.assert.NotEmpty(got)
	r.assert.Equal(got, val)
	return r
}

func (r *Response) Headers() map[string]string {
	res := make(map[string]string)
	for k, v := range r.inner.Header {
		if len(v) == 1 {
			res[k] = v[0]
		} else {
			res[k] = strings.Join(v, ",")
		}
	}
	return res
}

func n(t *testing.T) {
	router := http.NewRouter(log.NewNopLogger())
	fixture := NewFixture(t, router)
	req := fixture.Get("/customers")
	res := req.Do()
	res.HasCode(http.StatusOK)

	r := Response{}
	r.HasCode(200)
	var body ErrorBody
	r.ParseBody(&body)
}

func printRequest(req *gohttp.Request) {
	panic("implement!")
}

func printResponse(req *gohttp.Response) {
	panic("implement!")
}
