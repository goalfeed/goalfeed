package homeassistant

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/dns/dnsmessage"
)

// fakeDNS answers A queries for names whose first label is in hosts, and
// returns an empty NOERROR for everything else, so a test can make a
// single-label LAN-looking name resolve to any address it likes, the way a
// DNS search domain or a rebinding attacker could.
func fakeDNS(t *testing.T, hosts map[string]net.IP) *net.Resolver {
	t.Helper()
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { pc.Close() })
	go func() {
		buf := make([]byte, 1500)
		for {
			n, from, err := pc.ReadFrom(buf)
			if err != nil {
				return
			}
			var p dnsmessage.Parser
			hdr, err := p.Start(buf[:n])
			if err != nil {
				continue
			}
			q, err := p.Question()
			if err != nil {
				continue
			}
			b := dnsmessage.NewBuilder(nil, dnsmessage.Header{ID: hdr.ID, Response: true, Authoritative: true})
			_ = b.StartQuestions()
			_ = b.Question(q)
			_ = b.StartAnswers()
			label := strings.SplitN(q.Name.String(), ".", 2)[0]
			if ip, ok := hosts[label]; ok && q.Type == dnsmessage.TypeA {
				var a [4]byte
				copy(a[:], ip.To4())
				_ = b.AResource(dnsmessage.ResourceHeader{Name: q.Name, Class: dnsmessage.ClassINET, TTL: 1}, dnsmessage.AResource{A: a})
			}
			msg, _ := b.Finish()
			_, _ = pc.WriteTo(msg, from)
		}
	}()
	return &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, _, _ string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, "udp", pc.LocalAddr().String())
		},
	}
}

func withHAConfig(t *testing.T, url string, allowRemote bool) {
	t.Helper()
	os.Unsetenv("SUPERVISOR_API")
	os.Unsetenv("SUPERVISOR_TOKEN")
	viper.Set("home_assistant.url", url)
	viper.Set("home_assistant.access_token", "super-secret-token")
	viper.Set("home_assistant.allow_remote_url", allowRemote)
	t.Cleanup(func() {
		viper.Set("home_assistant.url", "")
		viper.Set("home_assistant.access_token", "")
		viper.Set("home_assistant.allow_remote_url", false)
	})
}

func TestGuardDialedAddress(t *testing.T) {
	for _, ok := range []string{"127.0.0.1:8123", "[::1]:8123", "192.168.1.5:8123", "10.0.0.2:443", "172.30.32.2:80"} {
		assert.NoError(t, guardDialedAddress("tcp", ok, nil), ok)
	}
	for _, bad := range []string{"8.8.8.8:8123", "169.254.169.254:80", "[fe80::1]:8123", "0.0.0.0:8123", "[::ffff:8.8.8.8]:443", "nonsense"} {
		assert.Error(t, guardDialedAddress("tcp", bad, nil), bad)
	}
}

// The headline case: a name that passes the string check (single label, so
// "local") but resolves to a public address is refused at dial time, before
// any connection is made, so the token never leaves.
func TestHAClient_HostnameResolvingToPublicIPIsRefusedAtDial(t *testing.T) {
	old := haResolver
	haResolver = fakeDNS(t, map[string]net.IP{"evilbox": net.ParseIP("8.8.8.8")})
	t.Cleanup(func() { haResolver = old })
	withHAConfig(t, "http://evilbox:8123", false)

	require.NoError(t, validateOutboundHAURL("http://evilbox:8123"), "the name alone looks local; only the dial check can catch this")
	ok, _, msg := CheckConnection(2 * time.Second)
	assert.False(t, ok)
	assert.Contains(t, msg, "8.8.8.8 is not a loopback or private address")
}

func TestHAClient_HostnameResolvingToLoopbackStillWorks(t *testing.T) {
	var gotAuth atomic.Value
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth.Store(r.Header.Get("Authorization"))
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	_, port, _ := net.SplitHostPort(strings.TrimPrefix(srv.URL, "http://"))

	old := haResolver
	haResolver = fakeDNS(t, map[string]net.IP{"nasbox": net.ParseIP("127.0.0.1")})
	t.Cleanup(func() { haResolver = old })
	withHAConfig(t, "http://nasbox:"+port, false)

	ok, _, msg := CheckConnection(2 * time.Second)
	assert.True(t, ok, msg)
	assert.Equal(t, "Bearer super-secret-token", gotAuth.Load())
}

func TestHAClient_LoopbackLiteralAndLocalhostWork(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }))
	t.Cleanup(srv.Close)
	_, port, _ := net.SplitHostPort(strings.TrimPrefix(srv.URL, "http://"))

	for _, u := range []string{"http://127.0.0.1:" + port, "http://localhost:" + port} {
		withHAConfig(t, u, false)
		ok, _, msg := CheckConnection(2 * time.Second)
		assert.True(t, ok, "%s: %s", u, msg)
	}
}

// With a proxy configured in the environment, the guarded client must still
// connect directly: a dial check that only saw the proxy's address would let
// a private proxy forward the token anywhere.
func TestHAClient_GuardedTransportIgnoresProxy(t *testing.T) {
	tr := newHATransport(true)
	assert.Nil(t, tr.Proxy)
	assert.NotNil(t, newHATransport(false).Proxy, "the opted-out transport keeps the environment proxy")
}

func TestHAClient_AllowRemoteSkipsDialCheck(t *testing.T) {
	withHAConfig(t, "http://example.invalid", true)
	assert.Same(t, haHTTPClient(0).Transport, haHTTPClient(time.Second).Transport, "transports are shared, not per request")
	viper.Set("home_assistant.allow_remote_url", false)
	guarded := haHTTPClient(0).Transport
	viper.Set("home_assistant.allow_remote_url", true)
	assert.NotSame(t, guarded, haHTTPClient(0).Transport)
}

// Every request carrying the HA token must use haHTTPClient, or it skips
// the dial check. An end-to-end test cannot prove that here (a sandbox that
// blocks 8.8.8.8 makes an unguarded send fail too), so this scans the
// package source: the only http.Client literal or DefaultClient use allowed
// is the one inside haHTTPClient.
func TestHASenders_AllUseGuardedClient(t *testing.T) {
	files, err := filepath.Glob("*.go")
	require.NoError(t, err)
	fset := token.NewFileSet()
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, f, nil, 0)
		require.NoError(t, err)
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Name.Name == "haHTTPClient" {
				continue
			}
			ast.Inspect(fn, func(n ast.Node) bool {
				switch x := n.(type) {
				case *ast.CompositeLit:
					if sel, ok := x.Type.(*ast.SelectorExpr); ok && sel.Sel.Name == "Client" {
						t.Errorf("%s: %s builds its own http.Client; use haHTTPClient", fset.Position(x.Pos()), fn.Name.Name)
					}
				case *ast.SelectorExpr:
					if id, ok := x.X.(*ast.Ident); ok && id.Name == "http" &&
						(x.Sel.Name == "DefaultClient" || x.Sel.Name == "Get" || x.Sel.Name == "Post" || x.Sel.Name == "PostForm" || x.Sel.Name == "Head") {
						t.Errorf("%s: %s uses http.%s; use haHTTPClient", fset.Position(x.Pos()), fn.Name.Name, x.Sel.Name)
					}
				}
				return true
			})
		}
	}
}
