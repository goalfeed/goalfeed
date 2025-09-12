package homeassistant

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestResolveHANoConfig(t *testing.T) {
	os.Unsetenv("SUPERVISOR_API")
	os.Unsetenv("SUPERVISOR_TOKEN")
	url, token, source := ResolveHA()
	if url != "" || token != "" || source != "unset" {
		t.Fatalf("expected unset, got url=%q token=%q source=%q", url, token, source)
	}
}

func TestCheckConnectionOK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	os.Setenv("SUPERVISOR_API", srv.URL)
	os.Setenv("SUPERVISOR_TOKEN", "t")
	defer os.Unsetenv("SUPERVISOR_API")
	defer os.Unsetenv("SUPERVISOR_TOKEN")
	ok, source, msg := CheckConnection(1 * time.Second)
	if !ok || source != "env" || msg != "OK" {
		t.Fatalf("unexpected: ok=%v source=%s msg=%s", ok, source, msg)
	}
}
