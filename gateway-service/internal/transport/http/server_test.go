package http

import (
	"io"
	"log/slog"
	stdhttp "net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/Ishan-Gijavanekar/gateway-service/internal/auth"
	"github.com/Ishan-Gijavanekar/gateway-service/internal/config"
	"github.com/Ishan-Gijavanekar/gateway-service/internal/gateway"
	"github.com/golang-jwt/jwt/v5"
)

func testServer(t *testing.T, upstream stdhttp.HandlerFunc) *Server {
	t.Helper()
	backend := httptest.NewServer(upstream)
	t.Cleanup(backend.Close)
	base, _ := url.Parse(backend.URL)
	proxy, err := gateway.NewProxy(backend.Client(), base, base, time.Second, 1024)
	if err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{HTTPPort: "0", PublicOrigin: "https://app.example.com", UserServiceURL: base, TaskServiceURL: base, JWTIssuer: "issuer", JWTKeys: map[string]string{"current": "secret"}, BodyLimit: 64, ResponseLimit: 1024, UpstreamTimeout: time.Second, ShutdownTimeout: time.Second}
	server, err := New(cfg, proxy, auth.New(cfg.JWTKeys, cfg.JWTIssuer), slog.New(slog.NewTextHandler(io.Discard, nil)), []byte(`{"openapi":"3.1.0"}`))
	if err != nil {
		t.Fatal(err)
	}
	return server
}
func accessToken(t *testing.T) string {
	t.Helper()
	claims := auth.Claims{UserID: "user-1", Role: "user", RegisteredClaims: jwt.RegisteredClaims{Issuer: "issuer", Subject: "user-1", ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))}}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token.Header["kid"] = "current"
	raw, err := token.SignedString([]byte("secret"))
	if err != nil {
		t.Fatal(err)
	}
	return raw
}
func TestProtectedRoutingAuthenticationAndRequestID(t *testing.T) {
	server := testServer(t, func(w stdhttp.ResponseWriter, r *stdhttp.Request) { _, _ = io.WriteString(w, `{"ok":true}`) })
	request := httptest.NewRequest("GET", "http://gateway/api/v1/workspaces/w/tasks", nil)
	response, _ := server.App().Test(request)
	if response.StatusCode != 401 {
		t.Fatalf("got %d", response.StatusCode)
	}
	request = httptest.NewRequest("GET", "http://gateway/api/v1/workspaces/w/tasks", nil)
	request.Header.Set("Authorization", "Bearer "+accessToken(t))
	request.Header.Set("X-Request-ID", "client-id")
	response, _ = server.App().Test(request)
	if response.StatusCode != 200 || response.Header.Get("X-Request-ID") != "client-id" {
		t.Fatalf("status=%d headers=%v", response.StatusCode, response.Header)
	}
}
func TestRouteContentTypeBodyLimitAndCORS(t *testing.T) {
	server := testServer(t, func(w stdhttp.ResponseWriter, r *stdhttp.Request) { w.WriteHeader(204) })
	request := httptest.NewRequest("POST", "http://gateway/api/v1/auth/login", strings.NewReader(`x`))
	request.Header.Set("Content-Type", "text/plain")
	response, _ := server.App().Test(request)
	if response.StatusCode != 415 {
		t.Fatalf("got %d", response.StatusCode)
	}
	request = httptest.NewRequest("POST", "http://gateway/api/v1/auth/login", strings.NewReader(strings.Repeat("x", 65)))
	request.Header.Set("Content-Type", "application/json")
	response, err := server.App().Test(request)
	if err == nil && (response == nil || response.StatusCode != 413) {
		t.Fatalf("body limit response=%v error=%v", response, err)
	}
	request = httptest.NewRequest("OPTIONS", "http://gateway/api/v1/auth/login", nil)
	request.Header.Set("Origin", "https://app.example.com")
	request.Header.Set("Access-Control-Request-Method", "POST")
	response, _ = server.App().Test(request)
	if response.Header.Get("Access-Control-Allow-Origin") != "https://app.example.com" {
		t.Fatalf("cors headers=%v", response.Header)
	}
}
func TestUnknownRouteAndInvalidRequestID(t *testing.T) {
	server := testServer(t, func(w stdhttp.ResponseWriter, r *stdhttp.Request) {})
	request := httptest.NewRequest("GET", "http://gateway/api/v1/internal/admin", nil)
	request.Header.Set("X-Request-ID", "bad id")
	response, _ := server.App().Test(request)
	body, _ := io.ReadAll(response.Body)
	if response.StatusCode != 404 || response.Header.Get("X-Request-ID") == "bad id" || !strings.Contains(string(body), "requestId") {
		t.Fatalf("status=%d id=%q body=%s", response.StatusCode, response.Header.Get("X-Request-ID"), body)
	}
}
