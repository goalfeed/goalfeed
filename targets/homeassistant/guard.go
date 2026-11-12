package homeassistant

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"syscall"
	"time"

	"goalfeed/config"
	"goalfeed/utils"
)

// validateOutboundHAURL is called immediately before any outbound request to
// Home Assistant is built, at every call site that would otherwise attach the
// configured long-lived access token as a Bearer header. It exists so that a
// poisoned or hand-edited home_assistant.url - set via the runtime config API,
// an environment variable, or a directly edited config.yaml - cannot cause
// Goalfeed to leak that token off the local network, even if it slipped past
// validation elsewhere (or was never validated at all).
//
// See utils.ValidateHomeAssistantURL for the exact rules. The
// home_assistant.allow_remote_url config key opts out of the private-address
// requirement for a genuinely remote HA instance; scheme and host validation
// always apply regardless.
func validateOutboundHAURL(rawURL string) error {
	allowRemote := config.GetBool("home_assistant.allow_remote_url")
	if err := utils.ValidateHomeAssistantURL(rawURL, allowRemote); err != nil {
		return fmt.Errorf("refusing to send Home Assistant request: %w", err)
	}
	return nil
}

// haResolver is the resolver the guarded dialer uses; nil means the system
// default. Tests replace it to make a hostname resolve to a chosen address.
var haResolver *net.Resolver

// guardDialedAddress is the net.Dialer Control hook for Home Assistant
// connections. It runs after DNS resolution with the literal IP about to be
// connected to, so it refuses a public address however the hostname got
// there: a DNS search domain completing a single-label name, a resolver that
// applies inet_aton, or DNS rebinding between validation and the request.
// Redirects are dialled through the same hook.
func guardDialedAddress(network, address string, _ syscall.RawConn) error {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return fmt.Errorf("refusing to send Home Assistant request: cannot parse dial address %q: %w", address, err)
	}
	ip := net.ParseIP(host)
	if ip == nil || !utils.IsAllowedHomeAssistantIP(ip) {
		return fmt.Errorf("refusing to send Home Assistant request: %s is not a loopback or private address; "+
			"set home_assistant.allow_remote_url: true if you intentionally run a remote Home Assistant instance", host)
	}
	return nil
}

func newHATransport(guarded bool) *http.Transport {
	dialer := &net.Dialer{Timeout: 30 * time.Second, KeepAlive: 30 * time.Second}
	t := &http.Transport{
		DialContext:           dialer.DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          20,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	if guarded {
		dialer.Control = guardDialedAddress
		// No proxy: the dial check would see only the proxy's address, and a
		// private proxy would happily forward the token to a public host.
		t.Proxy = nil
		t.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			d := *dialer
			d.Resolver = haResolver
			return d.DialContext(ctx, network, addr)
		}
	} else {
		t.Proxy = http.ProxyFromEnvironment
	}
	return t
}

var (
	haTransportOnce                       sync.Once
	haGuardedTransport, haRemoteTransport *http.Transport
)

// haHTTPClient returns the HTTP client for every request that carries the
// Home Assistant token. Unless home_assistant.allow_remote_url is set, its
// connections are checked at dial time by guardDialedAddress. The two
// transports are shared so connections are pooled as before.
func haHTTPClient(timeout time.Duration) *http.Client {
	haTransportOnce.Do(func() {
		haGuardedTransport = newHATransport(true)
		haRemoteTransport = newHATransport(false)
	})
	t := haGuardedTransport
	if config.GetBool("home_assistant.allow_remote_url") {
		t = haRemoteTransport
	}
	return &http.Client{Timeout: timeout, Transport: t}
}
