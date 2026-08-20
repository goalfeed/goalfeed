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
//     - an IP literal that is loopback, link-local, or private
//     (RFC 1918 / RFC 4193), or
//     - a hostname that is clearly local: "localhost", "homeassistant",
//     "homeassistant.local", "supervisor", anything ending in ".local",
//     or a bare single-label hostname (no dots) - i.e. not a public
//     hostname.
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
	return !strings.Contains(lower, ".")
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
