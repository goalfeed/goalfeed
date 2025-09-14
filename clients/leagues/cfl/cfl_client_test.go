package cfl

import (
	"encoding/json"
	"testing"
)

func withStubFetchCFL(t *testing.T, payload interface{}) func() {
	oldB := fetchByte
	oldBH := fetchByteWithHeaders
	b, _ := json.Marshal(payload)
	fetchByte = func(url string, ret chan []byte) { ret <- b }
	fetchByteWithHeaders = func(url string, ret chan []byte, headers map[string]string) { ret <- b }
	return func() { fetchByte = oldB; fetchByteWithHeaders = oldBH }
}

func withStubFetchCFLError(t *testing.T) func() {
	oldB := fetchByte
	oldBH := fetchByteWithHeaders
	fetchByte = func(url string, ret chan []byte) { ret <- []byte{} }
	fetchByteWithHeaders = func(url string, ret chan []byte, headers map[string]string) { ret <- []byte{} }
	return func() { fetchByte = oldB; fetchByteWithHeaders = oldBH }
}

func withStubFetchCFLInvalidJSON(t *testing.T) func() {
	oldB := fetchByte
	oldBH := fetchByteWithHeaders
	fetchByte = func(url string, ret chan []byte) { ret <- []byte("invalid json") }
	fetchByteWithHeaders = func(url string, ret chan []byte, headers map[string]string) { ret <- []byte("invalid json") }
	return func() { fetchByte = oldB; fetchByteWithHeaders = oldBH }
}

func TestCFLClient_Schedule(t *testing.T) {
	restore := withStubFetchCFL(t, CFLScheduleResponse{{ID: 1}})
	defer restore()
	c := CFLApiClient{}
	resp := c.GetCFLSchedule()
	if len(resp) != 1 || resp[0].ID != 1 {
		t.Fatalf("unexpected schedule resp: %+v", resp)
	}
}

func TestCFLClient_Schedule_EmptyResponse(t *testing.T) {
	restore := withStubFetchCFLError(t)
	defer restore()
	c := CFLApiClient{}
	resp := c.GetCFLSchedule()
	if len(resp) != 0 {
		t.Fatalf("expected empty response, got: %+v", resp)
	}
}

func TestCFLClient_Schedule_InvalidJSON(t *testing.T) {
	restore := withStubFetchCFLInvalidJSON(t)
	defer restore()
	c := CFLApiClient{}
	resp := c.GetCFLSchedule()
	if len(resp) != 0 {
		t.Fatalf("expected empty response for invalid JSON, got: %+v", resp)
	}
}

func TestCFLClient_Schedule_RetryPath(t *testing.T) {
	// Test the retry path when headers fail but no-headers succeed
	oldB := fetchByte
	oldBH := fetchByteWithHeaders
	defer func() { fetchByte = oldB; fetchByteWithHeaders = oldBH }()

	callCount := 0
	fetchByteWithHeaders = func(url string, ret chan []byte, headers map[string]string) {
		callCount++
		if callCount == 1 {
			ret <- []byte{} // First call fails
		} else {
			ret <- []byte(`[{"id": 1}]`) // Retry succeeds
		}
	}
	fetchByte = func(url string, ret chan []byte) {
		ret <- []byte(`[{"id": 1}]`) // No-headers call succeeds
	}

	c := CFLApiClient{}
	resp := c.GetCFLSchedule()
	if len(resp) != 1 || resp[0].ID != 1 {
		t.Fatalf("unexpected schedule resp after retry: %+v", resp)
	}
}

func TestCFLClient_LiveGame(t *testing.T) {
	restore := withStubFetchCFL(t, CFLLiveGameResponse{Sport: "AmericanFootball"})
	defer restore()
	c := CFLApiClient{}
	resp := c.GetCFLLiveGame("123")
	if resp.Sport != "AmericanFootball" {
		t.Fatalf("unexpected live game resp: %+v", resp)
	}
}

func TestCFLClient_LiveGame_EmptyResponse(t *testing.T) {
	restore := withStubFetchCFLError(t)
	defer restore()
	c := CFLApiClient{}
	resp := c.GetCFLLiveGame("123")
	if resp.Sport != "" {
		t.Fatalf("expected empty response, got: %+v", resp)
	}
}

func TestCFLClient_LiveGame_InvalidJSON(t *testing.T) {
	restore := withStubFetchCFLInvalidJSON(t)
	defer restore()
	c := CFLApiClient{}
	resp := c.GetCFLLiveGame("123")
	if resp.Sport != "" {
		t.Fatalf("expected empty response for invalid JSON, got: %+v", resp)
	}
}

func TestCFLClient_LiveGame_RetryPath(t *testing.T) {
	// Test the retry path when headers fail but no-headers succeed
	oldB := fetchByte
	oldBH := fetchByteWithHeaders
	defer func() { fetchByte = oldB; fetchByteWithHeaders = oldBH }()

	callCount := 0
	fetchByteWithHeaders = func(url string, ret chan []byte, headers map[string]string) {
		callCount++
		if callCount == 1 {
			ret <- []byte{} // First call fails
		} else {
			ret <- []byte(`{"sport": "AmericanFootball"}`) // Retry succeeds
		}
	}
	fetchByte = func(url string, ret chan []byte) {
		ret <- []byte(`{"sport": "AmericanFootball"}`) // No-headers call succeeds
	}

	c := CFLApiClient{}
	resp := c.GetCFLLiveGame("123")
	if resp.Sport != "AmericanFootball" {
		t.Fatalf("unexpected live game resp after retry: %+v", resp)
	}
}

func TestCFLClient_Schedule_JSONUnmarshalError(t *testing.T) {
	// Test the case where JSON unmarshal fails on primary but retry succeeds
	oldB := fetchByte
	oldBH := fetchByteWithHeaders
	defer func() { fetchByte = oldB; fetchByteWithHeaders = oldBH }()

	callCount := 0
	fetchByteWithHeaders = func(url string, ret chan []byte, headers map[string]string) {
		callCount++
		if callCount == 1 {
			ret <- []byte(`invalid json`) // First call fails with invalid JSON
		} else {
			ret <- []byte(`[{"id": 1}]`) // Retry succeeds
		}
	}
	fetchByte = func(url string, ret chan []byte) {
		ret <- []byte(`[{"id": 1}]`) // No-headers call succeeds
	}

	c := CFLApiClient{}
	resp := c.GetCFLSchedule()
	if len(resp) != 1 || resp[0].ID != 1 {
		t.Fatalf("unexpected schedule resp after JSON unmarshal error retry: %+v", resp)
	}
}

func TestCFLClient_Schedule_EmptyResponseRetry(t *testing.T) {
	// Test the case where primary returns empty but retry succeeds
	oldB := fetchByte
	oldBH := fetchByteWithHeaders
	defer func() { fetchByte = oldB; fetchByteWithHeaders = oldBH }()

	callCount := 0
	fetchByteWithHeaders = func(url string, ret chan []byte, headers map[string]string) {
		callCount++
		if callCount == 1 {
			ret <- []byte(`[]`) // First call returns empty array
		} else {
			ret <- []byte(`[{"id": 1}]`) // Retry succeeds
		}
	}
	fetchByte = func(url string, ret chan []byte) {
		ret <- []byte(`[{"id": 1}]`) // No-headers call succeeds
	}

	c := CFLApiClient{}
	resp := c.GetCFLSchedule()
	if len(resp) != 1 || resp[0].ID != 1 {
		t.Fatalf("unexpected schedule resp after empty response retry: %+v", resp)
	}
}
