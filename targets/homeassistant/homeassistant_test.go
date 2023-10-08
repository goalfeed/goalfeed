package homeassistant

import (
	"goalfeed/models"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestSendEvent(t *testing.T) {
	// Mock Home Assistant server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected method POST; got %s", r.Method)
		}
		if r.URL.String() != "/core/api/events/goal" {
			t.Errorf("Expected URL /core/api/events/goal; got %s", r.URL.String())
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Set environment variables to mock running inside Home Assistant add-on environment
	os.Setenv("SUPERVISOR_API", server.URL)
	os.Setenv("SUPERVISOR_TOKEN", "test-token")

	// Create a test event
	event := models.Event{
		// Fill in the event fields as needed for your test
	}

	// Call SendEvent
	SendEvent(event)

	// Reset environment variables (cleanup)
	os.Unsetenv("SUPERVISOR_API")
	os.Unsetenv("SUPERVISOR_TOKEN")
}
