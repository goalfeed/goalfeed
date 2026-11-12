package utils

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

// ValidateHomeAssistantURL checks whether rawURL is safe to use as a Home
// Assistant base URL - i.e. safe to attach the long-lived HA access token to
// as a Bearer credential.
//
// It exists so that a poisoned or hand-edited home_assistant.url (whether set
// via the runtime config API, an environment variable, or a directly edited
// config.yaml) cannot cause Goalfeed to send its access token to an
// attacker-controlled host. Rules:
//
//  1. rawURL must parse as an absolute URL with scheme "http" or "https" only
//     (no file:, gopher:, javascript:, etc).
//  2. It must have a non-empty host.
//  3. Unless allowRemote is true, the host must resolve to something clearly
//     local:
//     - an IP literal that is loopback or private (RFC 1918 / RFC 4193);
//     link-local is refused, see isPrivateOrLocalIP, or
//     - a hostname that is clearly local: "localhost", "homeassistant",
//     "homeassistant.local", "supervisor", anything ending in ".local",
//     or a bare single-label hostname (no dots) that inet_aton could not
//     read as a numeric address - i.e. not a public hostname.
//
// This is a check on the name only. The Home Assistant HTTP client also
// checks the address it actually dials (IsAllowedHomeAssistantIP), which is
// what covers DNS search domains and rebinding.
//
// Public/routable IPs and public hostnames are rejected unless allowRemote is
// true, in which case only the scheme/host checks (1 and 2) still apply - the
// caller has explicitly opted in via home_assistant.allow_remote_url for a
// genuinely remote HA instance.
func ValidateHomeAssistantURL(rawURL string, allowRemote bool) error {
	trimmed := strings.TrimSpace(rawURL)
	if trimmed == "" {
		return fmt.Errorf("home assistant url is empty")
	}
	u, err := url.Parse(trimmed)
	if err != nil {
		return fmt.Errorf("could not parse home assistant url: %w", err)
	}
	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("home assistant url must use http or https, got %q", u.Scheme)
	}
	host := u.Hostname()
	if host == "" {
		return fmt.Errorf("home assistant url has no host")
	}

	if allowRemote {
		return nil
	}

	if isLocalHomeAssistantHost(host) {
		return nil
	}

	return fmt.Errorf(
		"home assistant url %q does not look like a private/local address; "+
			"set home_assistant.allow_remote_url: true if you intentionally run a remote Home Assistant instance",
		trimmed,
	)
}

// isLocalHomeAssistantHost reports whether host (already extracted from a
// parsed URL, so no scheme/port/brackets) is clearly local.
func isLocalHomeAssistantHost(host string) bool {
	lower := strings.ToLower(host)

	switch lower {
	case "localhost", "homeassistant", "homeassistant.local", "supervisor":
		return true
	}
	if strings.HasSuffix(lower, ".local") {
		return true
	}

	if ip := net.ParseIP(host); ip != nil {
		return isPrivateOrLocalIP(ip)
	}

	// Not an IP literal. A bare single-label hostname (no dots, e.g. an mDNS
	// or NetBIOS name on the LAN) is treated as local; anything with a dot
	// that isn't covered above is assumed to be a public DNS name.
	if strings.Contains(lower, ".") {
		return false
	}
	// ...except a single label that inet_aton would read as an IPv4 address.
	// net.ParseIP rejects "134744072", "0x08080808" and "010", but a cgo /
	// getaddrinfo resolver turns them into 8.8.8.8, 8.8.8.8 and 0.0.0.8, so
	// treating them as LAN names would let the token reach a public IP.
	return !looksLikeNumericIPv4(lower)
}

// looksLikeNumericIPv4 reports whether a single label could be parsed by
// inet_aton as an address: all decimal (or octal) digits, or a 0x prefix.
func looksLikeNumericIPv4(label string) bool {
	if strings.HasPrefix(label, "0x") {
		return true
	}
	if label == "" {
		return false
	}
	for _, r := range label {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// IsAllowedHomeAssistantIP reports whether ip is an address Goalfeed may
// connect to with the Home Assistant token when remote URLs are not allowed:
// loopback or private only. It is checked against the address actually being
// dialled, after DNS, so it holds even when a hostname resolves somewhere
// its name did not suggest.
func IsAllowedHomeAssistantIP(ip net.IP) bool {
	return isPrivateOrLocalIP(ip)
}

func isPrivateOrLocalIP(ip net.IP) bool {
	// Link-local (169.254.0.0/16, fe80::/10) is deliberately NOT accepted.
	// A Home Assistant instance is never legitimately reached at a link-local
	// address, and the range contains the cloud instance metadata endpoint
	// (169.254.169.254) - the canonical SSRF target. Allowing it would let a
	// poisoned config send the HA access token to the metadata service of any
	// cloud host Goalfeed runs on.
	return ip.IsLoopback() || ip.IsPrivate()
}
