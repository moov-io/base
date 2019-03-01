// Copyright 2018 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package admin

import (
	"io/ioutil"
	"net/http"
	"testing"
)

func TestAdmin__pprof(t *testing.T) {
	svc := NewServer(":13983") // hopefully nothing locally has this
	go svc.Listen()
	defer svc.Shutdown()

	// Check for Prometheus metrics endpoint
	resp, _ := http.DefaultClient.Get("http://localhost:13983/metrics")
	if resp.StatusCode != http.StatusOK {
		t.Errorf("bogus HTTP status code: %s", resp.Status)
	}
	resp.Body.Close()

	// Check always on pprof endpoint
	resp, _ = http.DefaultClient.Get("http://localhost:13983/debug/pprof/cmdline")
	if resp.StatusCode != http.StatusOK {
		t.Errorf("bogus HTTP status code: %s", resp.Status)
	}
	resp.Body.Close()
}

func TestAdmin__AddHandler(t *testing.T) {
	svc := NewServer(":13984")
	go svc.Listen()
	defer svc.Shutdown()

	special := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/special-path" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("special"))
	}
	svc.AddHandler("/special-path", special)

	req, _ := http.NewRequest("GET", "http://localhost:13984/special-path", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("bogus HTTP status: %d", resp.StatusCode)
	}
	bs, _ := ioutil.ReadAll(resp.Body)
	if v := string(bs); v != "special" {
		t.Errorf("response was %q", v)
	}
}
