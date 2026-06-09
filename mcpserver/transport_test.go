package mcpserver

import (
	"net/http"
	"testing"
)

func TestStartHTTPRequiresToken(t *testing.T) {
	srv := New(sampleEngine(), Options{})
	if _, _, err := srv.StartHTTP(HTTPConfig{Addr: "127.0.0.1:0"}); err == nil {
		t.Error("HTTP transport must require a token")
	}
}

func TestHTTPRejectsMissingToken(t *testing.T) {
	srv := New(sampleEngine(), Options{})
	addr, stop, err := srv.StartHTTP(HTTPConfig{Addr: "127.0.0.1:0", Token: "secret"})
	if err != nil {
		t.Fatal(err)
	}
	defer stop()

	// No Authorization header -> rejected (SC-010).
	req, _ := http.NewRequest("POST", "http://"+addr, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401 without a token", resp.StatusCode)
	}

	// Wrong token -> rejected.
	req2, _ := http.NewRequest("POST", "http://"+addr, nil)
	req2.Header.Set("Authorization", "Bearer wrong")
	resp2, err := http.DefaultClient.Do(req2)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp2.Body.Close() }()
	if resp2.StatusCode != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401 with a wrong token", resp2.StatusCode)
	}
}

func TestIsLoopback(t *testing.T) {
	cases := map[string]bool{
		"127.0.0.1:7777": true,
		"localhost:80":   true,
		"[::1]:7777":     true,
		"0.0.0.0:7777":   false,
		"192.168.1.5:80": false,
	}
	for addr, want := range cases {
		if got := isLoopback(addr); got != want {
			t.Errorf("isLoopback(%q) = %v, want %v", addr, got, want)
		}
	}
}
