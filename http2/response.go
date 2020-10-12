package http2

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func Ok(responseBody interface{}) Response {
	return jsonResponse(http.StatusOK, responseBody)
}

func OkEmpty() Response {
	return newResponse(http.StatusNoContent, nil)
}

func Error(err error) Response {
	data := map[string]interface{}{
		"error": err.Error(),
	}

	return jsonResponse(http.StatusBadRequest, data)
}

func Errorf(message string, parts ...interface{}) Response {
	return Error(fmt.Errorf(message, parts...))
}

func NotFound() Response {
	return newResponse(http.StatusNotFound, nil)
}

func InternalServerError() Response {
	return newResponse(http.StatusInternalServerError, nil)
}

func jsonResponse(code int, data interface{}) Response {
	b, err := json.Marshal(data)
	if err != nil {
		return newResponse(400, nil)
	}

	response := newResponse(code, b)
	response.headers["Content-Type"] = "application/json; charset=utf-8"

	return response
}

func newResponse(code int, body []byte) Response {
	resp := Response{
		headers: make(map[string]string),
		code:    code,
		body:    body,
	}

	return resp
}

// type Response http.Response
type Response struct {
	headers map[string]string
	code    int
	body    []byte
}

func (r *Response) close(w http.ResponseWriter) {
	w.WriteHeader(r.code)
	_, err := w.Write(r.body)
	if err != nil {
		panic(errInvalidJSON)
	}

	for key, val := range r.headers {
		w.Header().Set(key, val)
	}
}
