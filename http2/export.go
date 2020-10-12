package http2

import "net/http"

var (
	MethodGet    = http.MethodGet
	MethodPut    = http.MethodPut
	MethodPost   = http.MethodPost
	MethodDelete = http.MethodDelete
)

var (
	StatusOK                  = http.StatusOK
	StatusNoContent           = http.StatusNoContent
	StatusBadRequest          = http.StatusBadRequest
	StatusInternalServerError = http.StatusInternalServerError
)
