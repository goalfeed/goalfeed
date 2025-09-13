package utils

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetString(t *testing.T) {
	// Create a test server that returns a simple string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	}))
	defer server.Close()

	// Test successful response
	ret := make(chan string)
	go GetString(server.URL, ret)

	select {
	case result := <-ret:
		assert.Equal(t, "test response", result)
	case <-time.After(5 * time.Second):
		t.Fatal("GetString timed out")
	}
}

func TestGetByte(t *testing.T) {
	// Test successful response
	t.Run("SuccessfulResponse", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("test response"))
		}))
		defer server.Close()

		ret := make(chan []byte)
		go GetByte(server.URL, ret)

		select {
		case result := <-ret:
			assert.Equal(t, []byte("test response"), result)
		case <-time.After(5 * time.Second):
			t.Fatal("GetByte timed out")
		}
	})

	// Test network error
	t.Run("NetworkError", func(t *testing.T) {
		ret := make(chan []byte)
		go GetByte("http://nonexistent.example.com", ret)

		select {
		case result := <-ret:
			assert.Equal(t, []byte{}, result)
		case <-time.After(5 * time.Second):
			t.Fatal("GetByte timed out")
		}
	})

	// Test server error response
	t.Run("ServerError", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("server error"))
		}))
		defer server.Close()

		ret := make(chan []byte)
		go GetByte(server.URL, ret)

		select {
		case result := <-ret:
			assert.Equal(t, []byte("server error"), result)
		case <-time.After(5 * time.Second):
			t.Fatal("GetByte timed out")
		}
	})

	// Test empty response
	t.Run("EmptyResponse", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		ret := make(chan []byte)
		go GetByte(server.URL, ret)

		select {
		case result := <-ret:
			assert.Equal(t, []byte{}, result)
		case <-time.After(5 * time.Second):
			t.Fatal("GetByte timed out")
		}
	})
}

func TestGetByteWithHeaders(t *testing.T) {
	// Success case
	t.Run("Success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("X-Test") != "1" {
				t.Fatalf("expected header set")
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ok"))
		}))
		defer server.Close()

		ret := make(chan []byte)
		go GetByteWithHeaders(server.URL, ret, map[string]string{"X-Test": "1"})
		select {
		case result := <-ret:
			assert.Equal(t, []byte("ok"), result)
		case <-time.After(5 * time.Second):
			t.Fatal("GetByteWithHeaders timed out")
		}
	})

	// Error case
	t.Run("Error", func(t *testing.T) {
		ret := make(chan []byte)
		go GetByteWithHeaders("http://127.0.0.1:0", ret, map[string]string{"X-Test": "1"})
		select {
		case result := <-ret:
			assert.Equal(t, 0, len(result))
		case <-time.After(5 * time.Second):
			t.Fatal("GetByteWithHeaders timed out")
		}
	})

	// Test with multiple headers
	t.Run("MultipleHeaders", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authorization") != "Bearer token" {
				t.Fatalf("expected Authorization header")
			}
			if r.Header.Get("Content-Type") != "application/json" {
				t.Fatalf("expected Content-Type header")
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("headers received"))
		}))
		defer server.Close()

		headers := map[string]string{
			"Authorization": "Bearer token",
			"Content-Type":  "application/json",
		}

		ret := make(chan []byte)
		go GetByteWithHeaders(server.URL, ret, headers)
		select {
		case result := <-ret:
			assert.Equal(t, []byte("headers received"), result)
		case <-time.After(5 * time.Second):
			t.Fatal("GetByteWithHeaders timed out")
		}
	})

	// Test with empty headers
	t.Run("EmptyHeaders", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("no headers"))
		}))
		defer server.Close()

		ret := make(chan []byte)
		go GetByteWithHeaders(server.URL, ret, nil)
		select {
		case result := <-ret:
			assert.Equal(t, []byte("no headers"), result)
		case <-time.After(5 * time.Second):
			t.Fatal("GetByteWithHeaders timed out")
		}
	})

	// Test server error response
	t.Run("ServerError", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("server error"))
		}))
		defer server.Close()

		ret := make(chan []byte)
		go GetByteWithHeaders(server.URL, ret, nil)
		select {
		case result := <-ret:
			assert.Equal(t, []byte("server error"), result)
		case <-time.After(5 * time.Second):
			t.Fatal("GetByteWithHeaders timed out")
		}
	})
}
