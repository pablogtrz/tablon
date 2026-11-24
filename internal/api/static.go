package api

import (
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"path"
	"strings"
)

const (
	frontendRoot = "frontend/dist"
	indexFile    = "index.html"

	cacheImmutable = "public, max-age=31536000, immutable"
	cacheNoCache   = "no-cache"
	cacheShort     = "public, max-age=3600"
	cacheNoStore   = "no-store"
)

// staticFiles serves the embedded SPA with history fallback.
type staticFiles struct {
	fs       fs.FS
	hasIndex bool
}

func newStaticFiles(assets fs.FS, logger *slog.Logger) (*staticFiles, error) {
	dist, err := fs.Sub(assets, frontendRoot)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", frontendRoot, err)
	}

	info, err := fs.Stat(dist, indexFile)
	hasIndex := err == nil && !info.IsDir()
	if !hasIndex && logger != nil {
		logger.Warn("frontend assets are not built; HTML routes will return 503",
			"hint", "pnpm --dir frontend build")
	}
	return &staticFiles{fs: dist, hasIndex: hasIndex}, nil
}

// securityHeaders applies conservative headers to every response. The CSP is
// intentionally compatible with CodeMirror and Svelte inline styles.
func (s *Server) securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		h.Set("Content-Security-Policy",
			"default-src 'self'; "+
				"img-src 'self' data: https:; "+
				"style-src 'self' 'unsafe-inline'; "+
				"script-src 'self'; "+
				"connect-src 'self'; "+
				"font-src 'self'; "+
				"object-src 'none'; "+
				"base-uri 'self'; "+
				"form-action 'self'; "+
				"frame-ancestors 'none'")
		next.ServeHTTP(w, r)
	})
}

func (s *Server) serveAssets(w http.ResponseWriter, r *http.Request) {
	rel := r.PathValue("path")
	if rel == "" || !fs.ValidPath(rel) {
		http.NotFound(w, r)
		return
	}
	if !s.static.serveFile(w, r, path.Join("assets", rel), cacheImmutable) {
		http.NotFound(w, r)
	}
}

func (s *Server) serveSPA(w http.ResponseWriter, r *http.Request) {
	rel := strings.TrimPrefix(r.URL.Path, "/")

	// Unknown API routes stay JSON. Known API routes are matched by more
	// specific patterns before this catch-all.
	if rel == "api" || strings.HasPrefix(rel, "api/") {
		writeError(w, http.StatusNotFound, "not_found", "resource not found", nil)
		return
	}
	if rel == "assets" || strings.HasPrefix(rel, "assets/") {
		http.NotFound(w, r)
		return
	}

	if rel == "" {
		s.static.serveIndex(w, r)
		return
	}

	cache := cacheShort
	switch {
	case rel == indexFile:
		cache = cacheNoCache
	case strings.HasPrefix(rel, "assets/"):
		cache = cacheImmutable
	case path.Ext(rel) != "":
		// A missing file with an extension is a real 404, not a client route.
		if !s.static.serveFile(w, r, rel, cache) {
			http.NotFound(w, r)
		}
		return
	}

	if s.static.serveFile(w, r, rel, cache) {
		return
	}
	if path.Ext(rel) != "" {
		http.NotFound(w, r)
		return
	}
	s.static.serveIndex(w, r)
}

func (s *staticFiles) serveFile(w http.ResponseWriter, r *http.Request, name, cache string) bool {
	if s == nil || s.fs == nil || !fs.ValidPath(name) {
		return false
	}
	info, err := fs.Stat(s.fs, name)
	if err != nil || info.IsDir() {
		return false
	}
	w.Header().Set("Cache-Control", cache)
	http.ServeFileFS(w, r, s.fs, name)
	return true
}

func (s *staticFiles) serveIndex(w http.ResponseWriter, r *http.Request) {
	if !s.hasIndex {
		w.Header().Set("Cache-Control", cacheNoStore)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprint(w, "Tablon frontend is not built.\nRun: pnpm --dir frontend build\n")
		return
	}
	w.Header().Set("Cache-Control", cacheNoCache)
	http.ServeFileFS(w, r, s.fs, indexFile)
}
