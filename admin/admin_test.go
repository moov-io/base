// Copyright 2020 The Moov Authors
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
	resp, err := http.DefaultClient.Get("http://localhost:13983/metrics")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("bogus HTTP status code: %s", resp.Status)
	}
	resp.Body.Close()

	// Check always on pprof endpoint
	resp, err = http.DefaultClient.Get("http://localhost:13983/debug/pprof/cmdline")
	if err != nil {
		t.Fatal(err)
	}
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

	req, err := http.NewRequest("GET", "http://localhost:13984/special-path", nil)
	if err != nil {
		t.Fatal(err)
	}
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

func TestAdmin__fullAddress(t *testing.T) {
	svc := NewServer("127.0.0.1:13985")
	go svc.Listen()
	defer svc.Shutdown()

	resp, err := http.DefaultClient.Get("http://localhost:13985/metrics")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("bogus HTTP status code: %s", resp.Status)
	}
	resp.Body.Close()
}

func TestAdmin__AddVersionHandler(t *testing.T) {
	svc := NewServer(":0")
	go svc.Listen()
	defer svc.Shutdown()

	svc.AddVersionHandler("v0.1.0")

	req, err := http.NewRequest("GET", "http://"+svc.BindAddr()+"/version", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("bogus HTTP status: %d", resp.StatusCode)
	}
	bs, _ := ioutil.ReadAll(resp.Body)
	if v := string(bs); v != "v0.1.0" {
		t.Errorf("got %s", v)
	}
}

func TestAdmin__Listen(t *testing.T) {
	svc := &Server{}
	if err := svc.Listen(); err != nil {
		t.Error("expected no error")
	}

	svc = nil
	if err := svc.Listen(); err != nil {
		t.Error("expected no error")
	}
}

func TestAdmin__BindAddr(t *testing.T) {
	svc := NewServer(":0")

	svc.AddHandler("/test/ping", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	go svc.Listen()
	defer svc.Shutdown()

	if v := svc.BindAddr(); v == ":0" {
		t.Errorf("BindAddr: %v", v)
	}

	resp, err := http.DefaultClient.Get("http://" + svc.BindAddr() + "/test/ping")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("bogus HTTP status code: %d", resp.StatusCode)
	}
}
