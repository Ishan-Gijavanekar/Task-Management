package gateway

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
)

func proxyApp(t *testing.T, upstream http.HandlerFunc, limit int64, compose bool) *fiber.App {
	t.Helper()
	server := httptest.NewServer(upstream)
	t.Cleanup(server.Close)
	base, _ := url.Parse(server.URL)
	proxy, err := NewProxy(server.Client(), base, base, time.Second, limit)
	if err != nil {
		t.Fatal(err)
	}
	app := fiber.New()
	app.All("/*", func(c *fiber.Ctx) error {
		c.Locals("requestId", "req-1")
		return proxy.Forward(c, Route{Service: UserService, ComposeUser: compose}, "actor-1")
	})
	return app
}
func TestProxyForwardsContextHeadersAndComposesMe(t *testing.T) {
	app := proxyApp(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/users/actor-1" || r.Header.Get("X-Request-ID") != "req-1" || r.Header.Get("Authorization") != "Bearer token" {
			t.Errorf("unexpected request %s headers=%v", r.URL.Path, r.Header)
		}
		w.Header().Add("Set-Cookie", "refresh=x; HttpOnly")
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"id":"actor-1"}`)
	}, 1024, true)
	req := httptest.NewRequest("GET", "http://gateway/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer token")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 || !strings.Contains(string(body), `"user"`) || len(resp.Header.Values("Set-Cookie")) != 1 {
		t.Fatalf("status=%d body=%s headers=%v", resp.StatusCode, body, resp.Header)
	}
}
func TestProxyNormalizesErrors(t *testing.T) {
	app := proxyApp(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(409)
		_, _ = io.WriteString(w, `{"error":{"code":"VERSION_CONFLICT","message":"reload","fields":{"version":"stale","unsafe":12}}}`)
	}, 1024, false)
	resp, err := app.Test(httptest.NewRequest("PATCH", "http://gateway/api/v1/users/x", nil))
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	text := string(body)
	if resp.StatusCode != 409 || !strings.Contains(text, `"code":"version_conflict"`) || !strings.Contains(text, `"requestId":"req-1"`) || strings.Contains(text, "unsafe") {
		t.Fatalf("%d %s", resp.StatusCode, text)
	}
}
func TestProxyBoundsResponsesAndTimeouts(t *testing.T) {
	app := proxyApp(t, func(w http.ResponseWriter, r *http.Request) { _, _ = io.WriteString(w, "12345") }, 4, false)
	resp, _ := app.Test(httptest.NewRequest("GET", "http://gateway/api/v1/users/x", nil))
	if resp.StatusCode != 502 {
		t.Fatalf("got %d", resp.StatusCode)
	}
}
