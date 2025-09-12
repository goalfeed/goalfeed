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

func TestCFLClient_Schedule(t *testing.T) {
	restore := withStubFetchCFL(t, CFLScheduleResponse{{ID: 1}})
	defer restore()
	c := CFLApiClient{}
	resp := c.GetCFLSchedule()
	if len(resp) != 1 || resp[0].ID != 1 {
		t.Fatalf("unexpected schedule resp: %+v", resp)
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
