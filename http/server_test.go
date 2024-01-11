// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package http

import (
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
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

func TestGetSkipAndCount(t *testing.T) {
	r := httptest.NewRequest("GET", "/ping?skip=10&count=20", nil)
	skip, count, exists, err := GetSkipAndCount(r)
	if skip != 10 {
		t.Errorf("skip should be 10. got : %d", skip)
	}
	if count != 20 {
		t.Errorf("count should be 20. got : %d", count)
	}
	if exists != true {
		t.Errorf("exists should be false. got : %t", exists)
	}
	if err != nil {
		t.Error("errors should be nil")
	}
}

func TestGetSkipAndCountReturnsDefaultsWhenNotProvided(t *testing.T) {
	r := httptest.NewRequest("GET", "/ping", nil)
	skip, count, exists, err := GetSkipAndCount(r)
	if skip != 0 {
		t.Errorf("skip should be 0. got : %d", skip)
	}
	if count != 200 {
		t.Errorf("count should be 200. got : %d", count)
	}
	if exists != false {
		t.Errorf("exists should be false. got : %t", exists)
	}
	if err != nil {
		t.Error("errors should be nil")
	}
}

func TestGetSkipAndCountWhenOnlyCountProvided(t *testing.T) {
	r := httptest.NewRequest("GET", "/ping?count=10", nil)
	skip, count, exists, err := GetSkipAndCount(r)
	if skip != 0 {
		t.Errorf("skip should be 0. got : %d", skip)
	}
	if count != 10 {
		t.Errorf("count should be 10. got : %d", count)
	}
	if exists != true {
		t.Errorf("exists should be true. got : %t", exists)
	}
	if err != nil {
		t.Error("errors should be nil")
	}
}

func TestGetSkipAndCountWhenOnlySkipProvidedReturnsDefaultCount(t *testing.T) {
	r := httptest.NewRequest("GET", "/ping?skip=10", nil)
	skip, count, exists, err := GetSkipAndCount(r)
	if skip != 10 {
		t.Errorf("skip should be 10. got : %d", skip)
	}
	if count != 200 {
		t.Errorf("count should be 200. got : %d", count)
	}
	if exists != true {
		t.Errorf("exists should be true. got : %t", exists)
	}
	if err != nil {
		t.Error("errors should be nil")
	}
}

func TestGetCountMaxWhenCountProvidedLargerThanMax(t *testing.T) {
	r := httptest.NewRequest("GET", "/ping?count=10001", nil)
	skip, count, exists, err := GetSkipAndCount(r)
	if skip != 0 {
		t.Errorf("skip should be 0. got : %d", skip)
	}
	if count != 10000 {
		t.Errorf("count should be 200. got : %d", count)
	}
	if exists != true {
		t.Errorf("exists should be true. got : %t", exists)
	}
	if err != nil {
		t.Error("errors should be nil")
	}
}

func TestGetSkipMaxWhenSkipProvidedLargerThanMax(t *testing.T) {
	r := httptest.NewRequest("GET", "/ping?skip=2147483648", nil)
	skip, count, exists, err := GetSkipAndCount(r)
	if skip != math.MaxInt32 {
		t.Errorf("skip should be %d. got : %d", math.MaxInt32, skip)
	}
	if count != 200 {
		t.Errorf("count should be 200. got : %d", count)
	}
	if exists != true {
		t.Errorf("exists should be true. got : %t", exists)
	}
	if err != nil {
		t.Error("error should be nil")
	}
}

func TestGetSkipAndCountErrorParsingCount(t *testing.T) {
	r := httptest.NewRequest("GET", "/ping?count=123abc123", nil)
	skip, count, exists, err := GetSkipAndCount(r)
	if skip != 0 {
		t.Errorf("skip should be 0. got : %d", skip)
	}
	if count != 0 {
		t.Errorf("count should be 0. got : %d", count)
	}
	if exists != true {
		t.Errorf("exists should be true. got : %t", exists)
	}
	if err == nil {
		t.Error("should be an error")
	}
}

func TestGetSkipAndCountErrorParsingSkip(t *testing.T) {
	r := httptest.NewRequest("GET", "/ping?skip=123abc123", nil)
	skip, count, exists, err := GetSkipAndCount(r)
	if skip != 0 {
		t.Errorf("skip should be 0. got : %d", skip)
	}
	if count != 0 {
		t.Errorf("count should be 0. got : %d", count)
	}
	if exists != true {
		t.Errorf("exists should be true. got : %t", exists)
	}
	if err == nil {
		t.Error("should be an error")
	}
}

func TestGetSkipAndCountErrorParsingSkipAndCount(t *testing.T) {
	r := httptest.NewRequest("GET", "/ping?skip=123abc123&count=abc123", nil)
	skip, count, exists, err := GetSkipAndCount(r)
	if skip != 0 {
		t.Errorf("skip should be 0. got : %d", skip)
	}
	if count != 0 {
		t.Errorf("count should be 0. got : %d", count)
	}
	if exists != true {
		t.Errorf("exists should be true. got : %t", exists)
	}
	if err == nil {
		t.Error("should be an error")
	}
}

func TestGetSkipAndCountReturns0IfNegativeValuesPassed(t *testing.T) {
	r := httptest.NewRequest("GET", "/ping?skip=-1&count=-1", nil)
	skip, count, exists, err := GetSkipAndCount(r)
	if skip != 0 {
		t.Errorf("skip should be 0. got : %d", skip)
	}
	if count != 200 {
		t.Errorf("count should be 200. got : %d", count)
	}
	if exists != true {
		t.Errorf("exists should be true. got : %t", exists)
	}
	if err != nil {
		t.Error("errors should be nil")
	}
}

func TestLimitedSkipCount(t *testing.T) {
	r := httptest.NewRequest("GET", "/list?skip=540&count=200", nil)
	skip, count, exists, err := LimitedSkipCount(r, 250, 100)
	require.NoError(t, err)
	require.True(t, exists)
	require.Equal(t, 250, skip)
	require.Equal(t, 100, count)
}

func TestGetOrderBy(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected []OrderBy
		wantErr  bool
	}{
		{
			name:     "valid - missing",
			path:     "/list",
			expected: []OrderBy{},
			wantErr:  false,
		},
		{
			name:     "valid - empty",
			path:     "/list?orderBy=",
			expected: []OrderBy{},
			wantErr:  false,
		},
		{
			name: "valid - single",
			path: "/list?orderBy=createdOn:ascending",
			expected: []OrderBy{
				{
					Name:      "createdOn",
					Direction: Ascending,
				},
			},
			wantErr: false,
		},
		{
			name: "valid - multiple",
			path: "/list?orderBy=createdOn:ascending,updatedOn:descending",
			expected: []OrderBy{
				{
					Name:      "createdOn",
					Direction: Ascending,
				},
				{
					Name:      "updatedOn",
					Direction: Descending,
				},
			},
			wantErr: false,
		},
		{
			name: "valid - short name asc",
			path: "/list?orderBy=createdOn:asc",
			expected: []OrderBy{
				{
					Name:      "createdOn",
					Direction: Ascending,
				},
			},
			wantErr: false,
		},
		{
			name: "valid - short name desc",
			path: "/list?orderBy=createdOn:desc",
			expected: []OrderBy{
				{
					Name:      "createdOn",
					Direction: Descending,
				},
			},
			wantErr: false,
		},
		{
			name: "valid - mixed cases ascending",
			path: "/list?orderBy=createdOn:AsCeNdInG",
			expected: []OrderBy{
				{
					Name:      "createdOn",
					Direction: Ascending,
				},
			},
		},
		{
			name: "valid - mixed cases descending",
			path: "/list?orderBy=createdOn:DeScEnDiNg",
			expected: []OrderBy{
				{
					Name:      "createdOn",
					Direction: Descending,
				},
			},
		},
		{
			name:     "invalid - missing colon",
			path:     "/list?orderBy=createdOndescending",
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "invalid - missing name",
			path:     "/list?orderBy=:ascending",
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "invalid - missing direction",
			path:     "/list?orderBy=createdOn:",
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "invalid - empty name",
			path:     "/list?orderBy=%20%20:ascending",
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "invalid - empty direction",
			path:     "/list?orderBy=createdOn:%20%20",
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "invalid - invalid direction",
			path:     "/list?orderBy=createdOn:invalid",
			expected: nil,
			wantErr:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", test.path, nil)
			orderBy, err := GetOrderBy(r)
			require.Equal(t, test.wantErr, err != nil)
			require.Equal(t, test.expected, orderBy)
		})
	}
}
