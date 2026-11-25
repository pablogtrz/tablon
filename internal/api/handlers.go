package api

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strconv"
	"time"

	"github.com/pablogtrz/tablon/internal/domain"
	"github.com/pablogtrz/tablon/internal/service"
	"github.com/pablogtrz/tablon/internal/version"
)

type listResponse[T any] struct {
	Items      []T    `json:"items"`
	NextCursor string `json:"next_cursor,omitempty"`
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) ready(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.Ready(r.Context()); err != nil {
		writeError(w, http.StatusServiceUnavailable, "not_ready", "database is not ready", nil)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func (s *Server) systemStatus(w http.ResponseWriter, r *http.Request) {
	state := s.idle.State()
	database := "idle"
	if !state.Idle {
		database = "ok"
		if err := s.svc.Ready(r.Context()); err != nil {
			database = "unavailable"
		}
	}

	var idleSince any
	if state.Idle && !state.IdleSince.IsZero() {
		idleSince = state.IdleSince.UTC().Format(time.RFC3339Nano)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"version":      version.Version,
		"auth_enabled": s.cfg.APIToken != "",
		"database":     database,
		"idle": map[string]any{
			"enabled":         state.Enabled,
			"state":           state.Name(),
			"timeout_seconds": int(state.Timeout.Seconds()),
			"last_activity":   state.LastActivity.UTC().Format(time.RFC3339Nano),
			"idle_since":      idleSince,
		},
	})
}

type createProjectRequest struct {
	Name              string `json:"name"`
	Identifier        string `json:"identifier"`
	Description       string `json:"description"`
	Status            string `json:"status"`
	LogoGradient      *int   `json:"logo_gradient"`
	LogoEmoji         string `json:"logo_emoji"`
	LogoGradientFrom  string `json:"logo_gradient_from"`
	LogoGradientTo    string `json:"logo_gradient_to"`
	LogoGradientAngle *int   `json:"logo_gradient_angle"`
}

type updateProjectRequest struct {
	Name              *string `json:"name"`
	Identifier        *string `json:"identifier"`
	Description       *string `json:"description"`
	Status            *string `json:"status"`
	LogoGradient      *int    `json:"logo_gradient"`
	LogoEmoji         *string `json:"logo_emoji"`
	LogoGradientFrom  *string `json:"logo_gradient_from"`
	LogoGradientTo    *string `json:"logo_gradient_to"`
	LogoGradientAngle *int    `json:"logo_gradient_angle"`
}

func (s *Server) listProjects(w http.ResponseWriter, r *http.Request) {
	includeArchived, err := queryBool(r, "include_archived")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_query", err.Error(), nil)
		return
	}
	items, err := s.svc.ListProjects(r.Context(), includeArchived)
	if err != nil {
		handleError(w, s.logger, err)
		return
	}
	writeJSON(w, http.StatusOK, listResponse[domain.ProjectListItem]{Items: items})
}

func (s *Server) createProject(w http.ResponseWriter, r *http.Request) {
	var req createProjectRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeDecodeError(w, err)
		return
	}
	p, err := s.svc.CreateProject(r.Context(), service.CreateProjectInput{
		Name:              req.Name,
		Identifier:        req.Identifier,
		Description:       req.Description,
		Status:            req.Status,
		LogoGradient:      req.LogoGradient,
		LogoEmoji:         req.LogoEmoji,
		LogoGradientFrom:  req.LogoGradientFrom,
		LogoGradientTo:    req.LogoGradientTo,
		LogoGradientAngle: req.LogoGradientAngle,
	})
	if err != nil {
		handleError(w, s.logger, err)
		return
	}
	writeJSON(w, http.StatusCreated, p)
}

func (s *Server) getProject(w http.ResponseWriter, r *http.Request) {
	p, err := s.svc.GetProject(r.Context(), r.PathValue("projectRef"))
	if err != nil {
		handleError(w, s.logger, err)
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (s *Server) updateProject(w http.ResponseWriter, r *http.Request) {
	var req updateProjectRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeDecodeError(w, err)
		return
	}
	p, err := s.svc.UpdateProject(r.Context(), r.PathValue("projectRef"), service.UpdateProjectInput{
		Name:              req.Name,
		Identifier:        req.Identifier,
		Description:       req.Description,
		Status:            req.Status,
		LogoGradient:      req.LogoGradient,
		LogoEmoji:         req.LogoEmoji,
		LogoGradientFrom:  req.LogoGradientFrom,
		LogoGradientTo:    req.LogoGradientTo,
		LogoGradientAngle: req.LogoGradientAngle,
	})
	if err != nil {
		handleError(w, s.logger, err)
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (s *Server) deleteProject(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteProject(r.Context(), r.PathValue("projectRef")); err != nil {
		handleError(w, s.logger, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type createEpicRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

type updateEpicRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Status      *string `json:"status"`
}

func (s *Server) listEpics(w http.ResponseWriter, r *http.Request) {
	items, err := s.svc.ListEpics(r.Context(), r.PathValue("projectRef"))
	if err != nil {
		handleError(w, s.logger, err)
		return
	}
	writeJSON(w, http.StatusOK, listResponse[domain.Epic]{Items: items})
}

func (s *Server) createEpic(w http.ResponseWriter, r *http.Request) {
	var req createEpicRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeDecodeError(w, err)
		return
	}
	e, err := s.svc.CreateEpic(r.Context(), r.PathValue("projectRef"), service.CreateEpicInput{
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
	})
	if err != nil {
		handleError(w, s.logger, err)
		return
	}
	writeJSON(w, http.StatusCreated, e)
}

func (s *Server) getEpic(w http.ResponseWriter, r *http.Request) {
	e, err := s.svc.GetEpic(r.Context(), r.PathValue("epicRef"))
	if err != nil {
		handleError(w, s.logger, err)
		return
	}
	writeJSON(w, http.StatusOK, e)
}

func (s *Server) updateEpic(w http.ResponseWriter, r *http.Request) {
	var req updateEpicRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeDecodeError(w, err)
		return
	}
	e, err := s.svc.UpdateEpic(r.Context(), r.PathValue("epicRef"), service.UpdateEpicInput{
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
	})
	if err != nil {
		handleError(w, s.logger, err)
		return
	}
	writeJSON(w, http.StatusOK, e)
}

func (s *Server) deleteEpic(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteEpic(r.Context(), r.PathValue("epicRef")); err != nil {
		handleError(w, s.logger, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type createIssueRequest struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Status      string  `json:"status"`
	Priority    string  `json:"priority"`
	EpicID      *string `json:"epic_id"`
	AgentID     *string `json:"agent_id"`
}

type updateIssueRequest struct {
	Title       *string        `json:"title"`
	Description *string        `json:"description"`
	Status      *string        `json:"status"`
	Priority    *string        `json:"priority"`
	EpicID      nullableString `json:"epic_id"`
}

type issueStatusRequest struct {
	Status string `json:"status"`
}

type agentRequest struct {
	AgentID string `json:"agent_id"`
}

func (s *Server) listIssues(w http.ResponseWriter, r *http.Request) {
	limit := 0
	if raw := r.URL.Query().Get("limit"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 {
			writeError(w, http.StatusBadRequest, "invalid_query", "limit must be a positive integer", nil)
			return
		}
		limit = value
	}
	epicID := r.URL.Query().Get("epic_id")
	if epicID == "" {
		epicID = r.URL.Query().Get("epic")
	}
	page, err := s.svc.ListIssues(r.Context(), r.PathValue("projectRef"), domain.IssueQuery{
		Status:   r.URL.Query().Get("status"),
		Priority: r.URL.Query().Get("priority"),
		EpicID:   epicID,
		Search:   r.URL.Query().Get("q"),
		Sort:     r.URL.Query().Get("sort"),
		Order:    r.URL.Query().Get("order"),
		Limit:    limit,
		Cursor:   r.URL.Query().Get("cursor"),
	})
	if err != nil {
		handleError(w, s.logger, err)
		return
	}
	writeJSON(w, http.StatusOK, page)
}

func (s *Server) createIssue(w http.ResponseWriter, r *http.Request) {
	var req createIssueRequest
	raw, err := decodeJSONWithBody(w, r, &req)
	if err != nil {
		writeDecodeError(w, err)
		return
	}
	i, _, err := s.svc.CreateIssueIdempotent(
		r.Context(),
		r.PathValue("projectRef"),
		service.CreateIssueInput{
			Title:       req.Title,
			Description: req.Description,
			Status:      req.Status,
			Priority:    req.Priority,
			EpicID:      req.EpicID,
			AgentID:     req.AgentID,
		},
		r.Header.Get("Idempotency-Key"),
		requestFingerprint(r, raw),
	)
	if err != nil {
		handleError(w, s.logger, err)
		return
	}
	writeJSON(w, http.StatusCreated, i)
}

func requestFingerprint(r *http.Request, raw []byte) string {
	sum := sha256.Sum256([]byte(r.Method + "\n" + r.URL.Path + "\n" + string(raw)))
	return hex.EncodeToString(sum[:])
}

func (s *Server) getIssue(w http.ResponseWriter, r *http.Request) {
	i, err := s.svc.GetIssue(r.Context(), r.PathValue("issueRef"))
	if err != nil {
		handleError(w, s.logger, err)
		return
	}
	writeJSON(w, http.StatusOK, i)
}

func (s *Server) updateIssue(w http.ResponseWriter, r *http.Request) {
	var req updateIssueRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeDecodeError(w, err)
		return
	}
	i, err := s.svc.UpdateIssue(r.Context(), r.PathValue("issueRef"), service.UpdateIssueInput{
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		Priority:    req.Priority,
		EpicIDSet:   req.EpicID.Set,
		EpicID:      req.EpicID.Value,
	})
	if err != nil {
		handleError(w, s.logger, err)
		return
	}
	writeJSON(w, http.StatusOK, i)
}

func (s *Server) deleteIssue(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteIssue(r.Context(), r.PathValue("issueRef")); err != nil {
		handleError(w, s.logger, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) updateIssueStatus(w http.ResponseWriter, r *http.Request) {
	var req issueStatusRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeDecodeError(w, err)
		return
	}
	i, err := s.svc.UpdateIssueStatus(r.Context(), r.PathValue("issueRef"), req.Status)
	if err != nil {
		handleError(w, s.logger, err)
		return
	}
	writeJSON(w, http.StatusOK, i)
}

func (s *Server) claimIssue(w http.ResponseWriter, r *http.Request) {
	var req agentRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeDecodeError(w, err)
		return
	}
	i, err := s.svc.ClaimIssue(r.Context(), r.PathValue("issueRef"), req.AgentID)
	if err != nil {
		handleError(w, s.logger, err)
		return
	}
	writeJSON(w, http.StatusOK, i)
}

func (s *Server) releaseIssue(w http.ResponseWriter, r *http.Request) {
	var req agentRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeDecodeError(w, err)
		return
	}
	i, err := s.svc.ReleaseIssue(r.Context(), r.PathValue("issueRef"), req.AgentID)
	if err != nil {
		handleError(w, s.logger, err)
		return
	}
	writeJSON(w, http.StatusOK, i)
}

type createCommentRequest struct {
	Content string  `json:"content"`
	AgentID *string `json:"agent_id"`
}

type updateCommentRequest struct {
	Content *string `json:"content"`
}

func (s *Server) listComments(w http.ResponseWriter, r *http.Request) {
	limit := 0
	if raw := r.URL.Query().Get("limit"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 {
			writeError(w, http.StatusBadRequest, "invalid_query", "limit must be a positive integer", nil)
			return
		}
		limit = value
	}
	page, err := s.svc.ListComments(
		r.Context(),
		r.PathValue("issueRef"),
		limit,
		r.URL.Query().Get("cursor"),
	)
	if err != nil {
		handleError(w, s.logger, err)
		return
	}
	writeJSON(w, http.StatusOK, page)
}

func (s *Server) createComment(w http.ResponseWriter, r *http.Request) {
	var req createCommentRequest
	raw, err := decodeJSONWithBody(w, r, &req)
	if err != nil {
		writeDecodeError(w, err)
		return
	}
	c, _, err := s.svc.CreateCommentIdempotent(
		r.Context(),
		r.PathValue("issueRef"),
		service.CreateCommentInput{Content: req.Content, AgentID: req.AgentID},
		r.Header.Get("Idempotency-Key"),
		requestFingerprint(r, raw),
	)
	if err != nil {
		handleError(w, s.logger, err)
		return
	}
	writeJSON(w, http.StatusCreated, c)
}

func (s *Server) updateComment(w http.ResponseWriter, r *http.Request) {
	var req updateCommentRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeDecodeError(w, err)
		return
	}
	c, err := s.svc.UpdateComment(r.Context(), r.PathValue("commentRef"), service.UpdateCommentInput{
		Content: req.Content,
	})
	if err != nil {
		handleError(w, s.logger, err)
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func (s *Server) deleteComment(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteComment(r.Context(), r.PathValue("commentRef")); err != nil {
		handleError(w, s.logger, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func queryBool(r *http.Request, key string) (bool, error) {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return false, nil
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		return false, &queryError{key: key}
	}
	return value, nil
}

type queryError struct {
	key string
}

func (e *queryError) Error() string {
	return e.key + " must be true or false"
}
