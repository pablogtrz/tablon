package api

import "net/http"

type apiRoute struct {
	Method  string
	Pattern string
	Handler http.HandlerFunc
}

// apiRoutes is the single source of truth for API route registration and the
// OpenAPI contract test. Static/SPA routes are intentionally not listed here.
func (s *Server) apiRoutes() []apiRoute {
	return []apiRoute{
		{http.MethodGet, "/health", s.health},
		{http.MethodGet, "/ready", s.ready},
		{http.MethodGet, "/api/v1/system/status", s.systemStatus},

		{http.MethodGet, "/api/v1/projects", s.listProjects},
		{http.MethodPost, "/api/v1/projects", s.createProject},
		{http.MethodGet, "/api/v1/projects/{projectRef}", s.getProject},
		{http.MethodPatch, "/api/v1/projects/{projectRef}", s.updateProject},
		{http.MethodDelete, "/api/v1/projects/{projectRef}", s.deleteProject},

		{http.MethodGet, "/api/v1/projects/{projectRef}/epics", s.listEpics},
		{http.MethodPost, "/api/v1/projects/{projectRef}/epics", s.createEpic},
		{http.MethodGet, "/api/v1/epics/{epicRef}", s.getEpic},
		{http.MethodPatch, "/api/v1/epics/{epicRef}", s.updateEpic},
		{http.MethodDelete, "/api/v1/epics/{epicRef}", s.deleteEpic},

		{http.MethodGet, "/api/v1/projects/{projectRef}/issues", s.listIssues},
		{http.MethodPost, "/api/v1/projects/{projectRef}/issues", s.createIssue},
		{http.MethodGet, "/api/v1/issues/{issueRef}", s.getIssue},
		{http.MethodPatch, "/api/v1/issues/{issueRef}", s.updateIssue},
		{http.MethodDelete, "/api/v1/issues/{issueRef}", s.deleteIssue},
		{http.MethodPatch, "/api/v1/issues/{issueRef}/status", s.updateIssueStatus},
		{http.MethodPost, "/api/v1/issues/{issueRef}/claim", s.claimIssue},
		{http.MethodPost, "/api/v1/issues/{issueRef}/release", s.releaseIssue},

		{http.MethodGet, "/api/v1/projects/{projectRef}/events", s.projectEvents},

		{http.MethodGet, "/api/v1/issues/{issueRef}/comments", s.listComments},
		{http.MethodPost, "/api/v1/issues/{issueRef}/comments", s.createComment},
		{http.MethodPatch, "/api/v1/comments/{commentRef}", s.updateComment},
		{http.MethodDelete, "/api/v1/comments/{commentRef}", s.deleteComment},
	}
}
