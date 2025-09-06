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