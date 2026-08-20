package utils

import "testing"

func TestValidateHomeAssistantURL_AllowsPrivateAndLocalHosts(t *testing.T) {
	cases := []string{
		"http://127.0.0.1:8123",
		"http://192.168.1.5:8123",
		"http://10.0.0.2:8123",
		"http://172.16.5.9:8123",
		"https://[::1]:8123",
		"http://homeassistant:8123",
		"http://homeassistant.local:8123",
		"http://supervisor/core",
		"http://myhassbox:8123", // bare single-label LAN hostname
		"http://localhost:8123",
	}
	for _, raw := range cases {
		if err := ValidateHomeAssistantURL(raw, false); err != nil {
			t.Errorf("expected %q to be allowed as a private/local url, got error: %v", raw, err)
		}
	}
}

func TestValidateHomeAssistantURL_RejectsExternalURL(t *testing.T) {
	cases := []string{
		"http://1.2.3.4:8123",
		"http://example.com:8123",
		"https://ha.duckdns.org",
		"http://attacker.example",
	}
	for _, raw := range cases {
		if err := ValidateHomeAssistantURL(raw, false); err == nil {
			t.Errorf("expected %q to be rejected as a public/external url by default", raw)
		}
	}
}

func TestValidateHomeAssistantURL_RejectsNonHTTPScheme(t *testing.T) {
	cases := []string{
		"file:///etc/passwd",
		"gopher://127.0.0.1:70/",
		"javascript:alert(1)",
		"ftp://127.0.0.1/",
	}
	for _, raw := range cases {
		// Even with allowRemote true, only http/https schemes are ever valid.
		if err := ValidateHomeAssistantURL(raw, false); err == nil {
			t.Errorf("expected %q to be rejected for non-http(s) scheme", raw)
		}
		if err := ValidateHomeAssistantURL(raw, true); err == nil {
			t.Errorf("expected %q to be rejected for non-http(s) scheme even with allowRemote=true", raw)
		}
	}
}

func TestValidateHomeAssistantURL_AllowRemoteOptOutStillEnforcesSchemeAndHost(t *testing.T) {
	// A genuinely remote HA instance is allowed only when the operator opts in.
	if err := ValidateHomeAssistantURL("https://ha.example.com:8123", true); err != nil {
		t.Fatalf("expected remote url to be allowed with allowRemote=true, got: %v", err)
	}
	// But scheme and host requirements are never relaxed.
	if err := ValidateHomeAssistantURL("ftp://ha.example.com", true); err == nil {
		t.Fatal("expected non-http(s) scheme to still be rejected with allowRemote=true")
	}
	if err := ValidateHomeAssistantURL("http:///no-host", true); err == nil {
		t.Fatal("expected missing host to still be rejected with allowRemote=true")
	}
}

func TestValidateHomeAssistantURL_RejectsEmptyAndUnparsable(t *testing.T) {
	if err := ValidateHomeAssistantURL("", false); err == nil {
		t.Fatal("expected empty url to be rejected")
	}
	if err := ValidateHomeAssistantURL("   ", false); err == nil {
		t.Fatal("expected whitespace-only url to be rejected")
	}
	if err := ValidateHomeAssistantURL("http://[::1", false); err == nil {
		t.Fatal("expected malformed url to be rejected")
	}
}

// TestValidateHomeAssistantURL_RejectsKnownExfiltrationPayloads pins the
// specific payloads from the vulnerability this validator was written to close:
// an unauthenticated POST to /api/homeassistant/config could point Goalfeed at
// an attacker host, after which every outbound call attached the real Home
// Assistant long-lived access token as a Bearer credential.
//
// 169.254.169.254 is included deliberately: link-local addresses are rejected
// because that range holds the cloud instance metadata endpoint.
func TestValidateHomeAssistantURL_RejectsKnownExfiltrationPayloads(t *testing.T) {
	for _, raw := range []string{
		"http://attacker.example.com:8123",
		"https://evil.io/api",
		"http://8.8.8.8:8123",
		"file:///etc/passwd",
		"javascript:alert(1)",
		"gopher://attacker.example.com",
		"http://169.254.169.254/latest/meta-data/",
		"http://[fe80::1]:8123",
	} {
		if err := ValidateHomeAssistantURL(raw, false); err == nil {
			t.Errorf("exfiltration payload was accepted: %s", raw)
		}
	}
}

// TestValidateHomeAssistantURL_AcceptsRealWorldLocalDeployments guards against
// a validator so strict it breaks ordinary self-hosted installs.
func TestValidateHomeAssistantURL_AcceptsRealWorldLocalDeployments(t *testing.T) {
	for _, raw := range []string{
		"http://homeassistant.local:8123",
		"http://192.168.1.50:8123",
		"http://10.0.0.5:8123",
		"http://172.16.4.2:8123",
		"http://127.0.0.1:8123",
		"http://supervisor/core",
		"http://hassio:8123",
	} {
		if err := ValidateHomeAssistantURL(raw, false); err != nil {
			t.Errorf("legitimate local deployment URL rejected: %s -> %v", raw, err)
		}
	}
}
