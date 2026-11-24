package api_test

import (
	"io"
	"net/http"
	"strings"
	"testing"
	"testing/fstest"
)

func staticAssets() fstest.MapFS {
	return fstest.MapFS{
		"frontend/dist/index.html": &fstest.MapFile{
			Data: []byte("<!doctype html><html><body>tablon</body></html>"),
		},
		"frontend/dist/favicon.svg": &fstest.MapFile{
			Data: []byte("<svg></svg>"),
		},
		"frontend/dist/assets/app.js": &fstest.MapFile{
			Data: []byte("console.log('tablon')"),
		},
	}
}

func get(t *testing.T, a *testAPI, path string) (*http.Response, string) {
	t.Helper()
	res, err := a.srv.Client().Get(a.srv.URL + path)
	if err != nil {
		t.Fatalf("GET %s: %v", path, err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return res, string(body)
}

func TestStaticServesIndexAndDeepLinks(t *testing.T) {
	a := newTestAPIWithAssets(t, "", staticAssets())

	for _, path := range []string{"/", "/projects/TABLON", "/projects/TABLON/issues/TABLON-1", "/settings"} {
		res, body := get(t, a, path)
		if res.StatusCode != http.StatusOK {
			t.Fatalf("GET %s status = %d", path, res.StatusCode)
		}
		if !strings.Contains(body, "tablon") {
			t.Fatalf("GET %s did not serve index: %q", path, body)
		}
		if got := res.Header.Get("Cache-Control"); got != "no-cache" {
			t.Fatalf("GET %s Cache-Control = %q", path, got)
		}
		if got := res.Header.Get("Content-Security-Policy"); got == "" {
			t.Fatalf("GET %s missing CSP", path)
		}
	}
}

func TestStaticServesHashedAssets(t *testing.T) {
	a := newTestAPIWithAssets(t, "", staticAssets())

	res, body := get(t, a, "/assets/app.js")
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", res.StatusCode)
	}
	if body != "console.log('tablon')" {
		t.Fatalf("body = %q", body)
	}
	if got := res.Header.Get("Cache-Control"); got != "public, max-age=31536000, immutable" {
		t.Fatalf("Cache-Control = %q", got)
	}
	if got := res.Header.Get("Content-Type"); !strings.Contains(got, "javascript") {
		t.Fatalf("Content-Type = %q", got)
	}

	res, _ = get(t, a, "/favicon.svg")
	if res.StatusCode != http.StatusOK {
		t.Fatalf("favicon status = %d", res.StatusCode)
	}
	if got := res.Header.Get("Cache-Control"); got != "public, max-age=3600" {
		t.Fatalf("favicon Cache-Control = %q", got)
	}
}

func TestStaticDoesNotMaskMissingFilesOrAPI(t *testing.T) {
	a := newTestAPIWithAssets(t, "", staticAssets())

	res, body := get(t, a, "/assets/missing.js")
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("missing asset status = %d", res.StatusCode)
	}
	if strings.Contains(body, "tablon") {
		t.Fatalf("missing asset served HTML: %q", body)
	}

	res, body = get(t, a, "/missing.js")
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("missing file status = %d", res.StatusCode)
	}
	if strings.Contains(body, "tablon") {
		t.Fatalf("missing file served HTML: %q", body)
	}

	res, body = get(t, a, "/api/v1/not-a-route")
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("unknown API status = %d", res.StatusCode)
	}
	if !strings.Contains(body, `"code":"not_found"`) {
		t.Fatalf("unknown API body = %q", body)
	}
	if got := res.Header.Get("Content-Type"); !strings.HasPrefix(got, "application/json") {
		t.Fatalf("unknown API Content-Type = %q", got)
	}
	if got := res.Header.Get("Cache-Control"); got != "no-store" {
		t.Fatalf("unknown API Cache-Control = %q", got)
	}
}

func TestStaticWithoutEmbeddedFSReturns503(t *testing.T) {
	a := newTestAPIWithAssets(t, "", fstest.MapFS{})

	res, body := get(t, a, "/")
	if res.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", res.StatusCode)
	}
	if !strings.Contains(body, "frontend is not built") {
		t.Fatalf("body = %q", body)
	}
}

func TestStaticMissingIndexReturns503(t *testing.T) {
	assets := fstest.MapFS{
		"frontend/dist/.gitkeep": &fstest.MapFile{Data: []byte{}},
	}
	a := newTestAPIWithAssets(t, "", assets)

	res, body := get(t, a, "/")
	if res.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", res.StatusCode)
	}
	if !strings.Contains(body, "frontend is not built") {
		t.Fatalf("body = %q", body)
	}
	if got := res.Header.Get("Cache-Control"); got != "no-store" {
		t.Fatalf("Cache-Control = %q", got)
	}
}

func TestStaticSecurityHeaders(t *testing.T) {
	a := newTestAPIWithAssets(t, "", staticAssets())

	res, _ := get(t, a, "/")
	for _, header := range []string{
		"X-Content-Type-Options",
		"X-Frame-Options",
		"Referrer-Policy",
		"Permissions-Policy",
		"Content-Security-Policy",
	} {
		if got := res.Header.Get(header); got == "" {
			t.Fatalf("missing %s", header)
		}
	}
}
