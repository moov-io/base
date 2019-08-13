// Copyright 2018 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package admin

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestHealth_healthCheck(t *testing.T) {
	c := &healthCheck{"example", func() error {
		return errors.New("example error")
	}}
	if err := c.Error(); err == nil {
		t.Error("expected error")
	}
}

func TestHealth_processChecks(t *testing.T) {
	checks := []*healthCheck{
		{"good", func() error { return nil }},
		{"bad", func() error { return errors.New("bad") }},
	}
	results := processChecks(checks)
	if len(results) != 2 {
		t.Fatalf("Got %v", results)
	}
	for i := range results {
		if results[i].name == "good" && results[i].err != nil {
			t.Errorf("%q got err=%v", results[i].name, results[i].err)
			continue
		}
		if results[i].name == "bad" && results[i].err.Error() != "bad" {
			t.Errorf("%q got err=%v", results[i].name, results[i].err)
			continue
		}
	}
}

func TestHealth__LiveHTTP(t *testing.T) {
	svc := NewServer(":13993") // hopefully nothing locally has this
	go svc.Listen()
	defer svc.Shutdown()

	// no checks, should be healthy
	resp, err := http.DefaultClient.Get("http://localhost:13993/live")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("bogus HTTP status: %s", resp.Status)
	}
	resp.Body.Close()

	// add a healthy check
	svc.AddLivenessCheck("live-good", func() error {
		return nil
	})
	resp, err = http.DefaultClient.Get("http://localhost:13993/live")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("bogus HTTP status: %s", resp.Status)
	}
	resp.Body.Close()

	// one bad check, should fail
	svc.AddLivenessCheck("live-bad", func() error {
		return errors.New("unhealthy")
	})
	resp, err = http.DefaultClient.Get("http://localhost:13993/live")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("bogus HTTP status: %s", resp.Status)
	}
	defer resp.Body.Close()

	// Read JSON response body
	var checks map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&checks); err != nil {
		t.Fatal(err)
	}
	if len(checks) != 2 {
		t.Errorf("checks: %#v", checks)
	}
	if v := fmt.Sprintf("%v", checks["live-good"]); v != "good" {
		t.Errorf("live-good: %s", v)
	}
	if v := fmt.Sprintf("%v", checks["live-bad"]); v != "unhealthy" {
		t.Errorf("live-bad: %s", v)
	}
}

func TestHealth__ReadyHTTP(t *testing.T) {
	svc := NewServer(":13994") // hopefully nothing locally has this
	go svc.Listen()
	defer svc.Shutdown()

	// no checks, should be healthy
	resp, err := http.DefaultClient.Get("http://localhost:13994/ready")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("bogus HTTP status: %s", resp.Status)
	}
	resp.Body.Close()

	// add a healthy check
	svc.AddReadinessCheck("ready-good", func() error {
		return nil
	})
	resp, err = http.DefaultClient.Get("http://localhost:13994/ready")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("bogus HTTP status: %s", resp.Status)
	}
	resp.Body.Close()

	// one bad check, should fail
	svc.AddReadinessCheck("ready-bad", func() error {
		return errors.New("unhealthy")
	})
	resp, err = http.DefaultClient.Get("http://localhost:13994/ready")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("bogus HTTP status: %s", resp.Status)
	}
	defer resp.Body.Close()

	// Read JSON response body
	var checks map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&checks); err != nil {
		t.Fatal(err)
	}
	if len(checks) != 2 {
		t.Errorf("checks: %#v", checks)
	}
	if v := fmt.Sprintf("%v", checks["ready-good"]); v != "good" {
		t.Errorf("ready-good: %s", v)
	}
	if v := fmt.Sprintf("%v", checks["ready-bad"]); v != "unhealthy" {
		t.Errorf("ready-bad: %s", v)
	}
}

func TestHealth_try(t *testing.T) {
	// happy path, no timeout
	if err := try(func() error { return nil }, 1*time.Second); err != nil {
		t.Error("expected no error")
	}

	// error returned, no timeout
	if err := try(func() error { return errors.New("error") }, 1*time.Second); err == nil {
		t.Error("expected error, got none")
	} else {
		if err.Error() != "error" {
			t.Errorf("got %v", err)
		}
	}

	// timeout
	f := func() error {
		time.Sleep(1 * time.Second)
		return errors.New("after sleep")
	}
	if err := try(f, 10*time.Millisecond); err == nil {
		t.Errorf("expected (timeout) error, got none")
	} else {
		if err != errTimeout {
			t.Errorf("unknown error: %v", err)
		}
	}
}
