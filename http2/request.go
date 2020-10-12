package http2

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/gorilla/mux"
)

type Request struct {
	*http.Request
}

func FormatRequest(r *http.Request) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%v %v %v\n", r.Method, r.URL, r.Proto))
	sb.WriteString(fmt.Sprintf("Host: %v\n", r.Host))

	for name, headers := range r.Header {
		for _, header := range headers {
			sb.WriteString(fmt.Sprintf("%v: %v", name, header))
		}
	}

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		sb.WriteString(r.Form.Encode())
	}

	return sb.String()
}

func (r *Request) Var(key string) string {
	result, ok := mux.Vars(r.Request)[key]
	if !ok {
		panic(errPathVarNotFound{key: key})
	}

	return result
}

func (r *Request) Header(key string) string {
	result := r.Request.Header.Get(key)
	if result == "" {
		panic(errHeaderNotFound{key: key})
	}
	return result
}

func (r *Request) HeaderFound(key string) (string, bool) {
	result := r.Request.Header.Get(key)
	if result == "" {
		return "", false
	}
	return result, true
}

func (r *Request) ParseBody(v interface{}) {
	if r.ContentLength <= 0 {
		return
	}
	err := json.NewDecoder(r.Body).Decode(v)
	if err != nil {
		dump, err := httputil.DumpRequest(r.Request, true)
		if err == nil {
			fmt.Println(string(dump))
		}

		panic(errInvalidJSON)
	}
}
