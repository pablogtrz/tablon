package api

import (
	"crypto/subtle"
	"io/fs"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/pablogtrz/tablon/internal/config"
	"github.com/pablogtrz/tablon/internal/events"
	"github.com/pablogtrz/tablon/internal/idle"
	"github.com/pablogtrz/tablon/internal/service"
)

type Server struct {
	cfg     config.Config
	svc     *service.Service
	hub     *events.Hub
	logger  *slog.Logger
	mux     *http.ServeMux
	static  *staticFiles
	idle    *idle.Controller
	openAPI []byte
}

func New(cfg config.Config, svc *service.Service, hub *events.Hub, idleCtl *idle.Controller, logger *slog.Logger, assets fs.FS, openAPI []byte) http.Handler {
	if idleCtl == nil {
		idleCtl = idle.New(false, 0, logger, nil, nil)
	}
	s := &Server{
		cfg:     cfg,
		svc:     svc,
		hub:     hub,
		logger:  logger,
		mux:     http.NewServeMux(),
		idle:    idleCtl,
		openAPI: openAPI,
	}
	if assets != nil {
		static, err := newStaticFiles(assets, logger)
		if err != nil {
			logger.Warn("static assets unavailable; build with -tags embed", "error", err)
			s.static = &staticFiles{}
		} else {
			s.static = static
		}
	}
	s.routes()
	return s.recover(s.logRequests(s.securityHeaders(s.auth(s.activity(s.mux)))))
}

func (s *Server) routes() {
	for _, route := range s.apiRoutes() {
		s.mux.HandleFunc(route.Method+" "+route.Pattern, route.Handler)
	}

	if len(s.openAPI) > 0 {
		s.mux.HandleFunc("GET /openapi.yaml", s.serveOpenAPI)
	}

	if s.static != nil {
		s.mux.HandleFunc("GET /assets/{path...}", s.serveAssets)
		s.mux.HandleFunc("GET /{path...}", s.serveSPA)
	}
}

// activity marks meaningful API traffic. Health probes and the status endpoint
// deliberately do not reactivate idle mode.
func (s *Server) activity(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/api/v1/") || r.URL.Path == "/api/v1/system/status" {
			next.ServeHTTP(w, r)
			return
		}
		done := s.idle.Begin()
		defer done()
		next.ServeHTTP(w, r)
	})
}

func (s *Server) auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.cfg.APIToken == "" || !strings.HasPrefix(r.URL.Path, "/api/v1/") {
			next.ServeHTTP(w, r)
			return
		}
		expected := "Bearer " + s.cfg.APIToken
		got := r.Header.Get("Authorization")
		if subtle.ConstantTimeCompare([]byte(got), []byte(expected)) != 1 {
			writeError(w, http.StatusUnauthorized, "unauthorized", "missing or invalid bearer token", nil)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &responseRecorder{ResponseWriter: w}
		next.ServeHTTP(rec, r)
		status := rec.status
		if status == 0 {
			status = http.StatusOK
		}
		s.logger.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", status,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}

func (s *Server) recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if value := recover(); value != nil {
				s.logger.Error("panic recovered", "panic", value, "path", r.URL.Path)
				writeError(w, http.StatusInternalServerError, "internal_error", "internal server error", nil)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

type responseRecorder struct {
	http.ResponseWriter
	status int
}

func (r *responseRecorder) WriteHeader(status int) {
	if r.status == 0 {
		r.status = status
	}
	r.ResponseWriter.WriteHeader(status)
}

func (r *responseRecorder) Write(body []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	return r.ResponseWriter.Write(body)
}

func (r *responseRecorder) Flush() {
	if flusher, ok := r.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (r *responseRecorder) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}
