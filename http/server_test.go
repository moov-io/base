// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/gorilla/mux"
)

func truncate(s string) string {
	if utf8.RuneCountInString(s) > maxHeaderLength {
		return s[:maxHeaderLength]
	}
	return s
}

func TestHTTP__AddCORSHandler(t *testing.T) {
	router := mux.NewRouter()
	w := httptest.NewRecorder()
	r := httptest.NewRequest("OPTIONS", "https://api.moov.io/v1/auth/ping", nil)
	r.Header.Set("Origin", "https://moov.io")

	AddCORSHandler(router)
	router.ServeHTTP(w, r)
	w.Flush()

	if w.Code != 200 {
		t.Errorf("got %d", w.Code)
	}
	if v := w.Header().Get("Access-Control-Allow-Origin"); v != "https://moov.io" {
		t.Errorf("got %q", v)
	}
	headers := []string{
		"Access-Control-Allow-Methods",
		"Access-Control-Allow-Headers",
		"Access-Control-Allow-Credentials",
	}
	for i := range headers {
		v := w.Header().Get(headers[i])
		if v == "" {
			t.Errorf("%s's value is an empty string", headers[i])
		}
	}
}

func TestHTTP__emptyOrigin(t *testing.T) {
	router := mux.NewRouter()
	w := httptest.NewRecorder()
	r := httptest.NewRequest("OPTIONS", "https://api.moov.io/v1/auth/ping", nil)
	r.Header.Set("Origin", "")

	AddCORSHandler(router)
	router.ServeHTTP(w, r)
	w.Flush()

	if w.Code != http.StatusBadRequest {
		t.Errorf("got %d", w.Code)
	}
}

func TestHTTP__Problem(t *testing.T) {
	w := httptest.NewRecorder()
	Problem(w, errors.New("problem X"))
	w.Flush()

	// check http response
	if w.Code != http.StatusBadRequest {
		t.Errorf("got %d", w.Code)
	}
	v := w.Header().Get("Content-Type")
	if !strings.Contains(v, "application/json") {
		t.Errorf("got %s", v)
	}

	type resp struct {
		Error string `json:"error"`
	}
	var response resp
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Error(err)
	}
	if response.Error != "problem X" {
		t.Errorf("got %q", response.Error)
	}

	// nil error, respond http.StatusOK
	w = httptest.NewRecorder()
	Problem(w, nil)
	w.Flush()

	if w.Code != http.StatusOK {
		t.Errorf("got %d", w.Code)
	}
}

func TestHTTP_InternalError(t *testing.T) {
	w := httptest.NewRecorder()
	where := InternalError(w, errors.New("problem Y"))
	w.Flush()

	if w.Code != http.StatusInternalServerError {
		t.Errorf("got %d", w.Code)
	}

	if !strings.HasPrefix(where, "server_test.go") { // This will always be this file's name
		t.Errorf("got %s", where)
	}
}

func TestHTTP__GetRequestID(t *testing.T) {
	r := httptest.NewRequest("GET", "/ping", nil)
	r.Header.Set("x-request-id", "requestID")

	if requestID := GetRequestID(r); requestID != "requestID" {
		t.Errorf("got %s", requestID)
	}
}

func TestHTTP__GetUserID(t *testing.T) {
	r := httptest.NewRequest("GET", "/ping", nil)
	r.Header.Set("x-user-id", "userID")

	if userID := GetUserID(r); userID != "userID" {
		t.Errorf("got %s", userID)
	}

	r = httptest.NewRequest("GET", "/ping", nil)
	r.Header.Set("x-user", "other")

	if userID := GetUserID(r); userID != "other" {
		t.Errorf("got %s", userID)
	}
}

func TestHTTP__truncate(t *testing.T) {
	s1 := "1234567890123456789012345678901234567890" // 40 characters
	s2 := truncate(s1)
	if s1 == s2 {
		t.Errorf("strings shouldn't match")
	}
	if n := utf8.RuneCountInString(s2); n != maxHeaderLength {
		t.Errorf("s2 length is %d", n)
	}

	s1, s2 = "12345", truncate("12345")
	if s1 != s2 {
		t.Errorf("strings should match: s1=%s s2=%s", s1, s2)
	}
}
