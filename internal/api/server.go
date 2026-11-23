package api

import (
	"crypto/subtle"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/pablogtrz/tablon/internal/config"
	"github.com/pablogtrz/tablon/internal/service"
)

type Server struct {
	cfg    config.Config
	svc    *service.Service
	logger *slog.Logger
	mux    *http.ServeMux
}

func New(cfg config.Config, svc *service.Service, logger *slog.Logger) http.Handler {
	s := &Server{
		cfg:    cfg,
		svc:    svc,
		logger: logger,
		mux:    http.NewServeMux(),
	}
	s.routes()
	return s.recover(s.logRequests(s.auth(s.mux)))
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /health", s.health)
	s.mux.HandleFunc("GET /ready", s.ready)
	s.mux.HandleFunc("GET /api/v1/system/status", s.systemStatus)

	s.mux.HandleFunc("GET /api/v1/projects", s.listProjects)
	s.mux.HandleFunc("POST /api/v1/projects", s.createProject)
	s.mux.HandleFunc("GET /api/v1/projects/{projectRef}", s.getProject)
	s.mux.HandleFunc("PATCH /api/v1/projects/{projectRef}", s.updateProject)
	s.mux.HandleFunc("DELETE /api/v1/projects/{projectRef}", s.deleteProject)

	s.mux.HandleFunc("GET /api/v1/projects/{projectRef}/epics", s.listEpics)
	s.mux.HandleFunc("POST /api/v1/projects/{projectRef}/epics", s.createEpic)
	s.mux.HandleFunc("GET /api/v1/epics/{epicRef}", s.getEpic)
	s.mux.HandleFunc("PATCH /api/v1/epics/{epicRef}", s.updateEpic)
	s.mux.HandleFunc("DELETE /api/v1/epics/{epicRef}", s.deleteEpic)

	s.mux.HandleFunc("GET /api/v1/projects/{projectRef}/issues", s.listIssues)
	s.mux.HandleFunc("POST /api/v1/projects/{projectRef}/issues", s.createIssue)
	s.mux.HandleFunc("GET /api/v1/issues/{issueRef}", s.getIssue)
	s.mux.HandleFunc("PATCH /api/v1/issues/{issueRef}", s.updateIssue)
	s.mux.HandleFunc("DELETE /api/v1/issues/{issueRef}", s.deleteIssue)
	s.mux.HandleFunc("PATCH /api/v1/issues/{issueRef}/status", s.updateIssueStatus)
	s.mux.HandleFunc("POST /api/v1/issues/{issueRef}/claim", s.claimIssue)
	s.mux.HandleFunc("POST /api/v1/issues/{issueRef}/release", s.releaseIssue)
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
