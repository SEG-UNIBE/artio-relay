package logging

import (
	"net/http/httptest"
	"testing"
)

/*
TestGetIPForwardedFor tests that X-Forwarded-For is preferred over X-Real-Ip
*/
func TestGetIPForwardedFor(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("X-Forwarded-For", "1.2.3.4")
	r.Header.Set("X-Real-Ip", "5.6.7.8")

	if ip := getIP(r); ip != "1.2.3.4" {
		t.Fatalf("expected 1.2.3.4 but got %v", ip)
	}
}

/*
TestGetIPRealIP tests that X-Real-Ip is used when X-Forwarded-For is missing
*/
func TestGetIPRealIP(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("X-Real-Ip", "5.6.7.8")

	if ip := getIP(r); ip != "5.6.7.8" {
		t.Fatalf("expected 5.6.7.8 but got %v", ip)
	}
}

/*
TestGetIPNoHeaders tests that an empty string is returned without proxy headers
*/
func TestGetIPNoHeaders(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)

	if ip := getIP(r); ip != "" {
		t.Fatalf("expected an empty ip but got %v", ip)
	}
}
