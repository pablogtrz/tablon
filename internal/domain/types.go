package domain

import "time"

type ProjectStatus string

const (
	ProjectPlanning ProjectStatus = "planning"
	ProjectActive   ProjectStatus = "active"
	ProjectArchived ProjectStatus = "archived"
	ProjectDone     ProjectStatus = "done"
)

func (s ProjectStatus) Valid() bool {
	switch s {
	case ProjectPlanning, ProjectActive, ProjectArchived, ProjectDone:
		return true
	default:
		return false
	}
}

type EpicStatus string

const (
	EpicBacklog    EpicStatus = "backlog"
	EpicInProgress EpicStatus = "in-progress"
	EpicDone       EpicStatus = "done"
)

func (s EpicStatus) Valid() bool {
	switch s {
	case EpicBacklog, EpicInProgress, EpicDone:
		return true
	default:
		return false
	}
}

type IssueStatus string

const (
	IssueBacklog    IssueStatus = "backlog"
	IssueTodo       IssueStatus = "todo"
	IssueInProgress IssueStatus = "in-progress"
	IssueInReview   IssueStatus = "in-review"
	IssueDone       IssueStatus = "done"
	IssueCanceled   IssueStatus = "canceled"
)

func (s IssueStatus) Valid() bool {
	switch s {
	case IssueBacklog, IssueTodo, IssueInProgress, IssueInReview, IssueDone, IssueCanceled:
		return true
	default:
		return false
	}
}

type Priority string

const (
	PriorityUrgent Priority = "urgent"
	PriorityHigh   Priority = "high"
	PriorityMedium Priority = "medium"
	PriorityLow    Priority = "low"
	PriorityNone   Priority = "none"
)

func (p Priority) Valid() bool {
	switch p {
	case PriorityUrgent, PriorityHigh, PriorityMedium, PriorityLow, PriorityNone:
		return true
	default:
		return false
	}
}

type Project struct {
	ID           string        `json:"id"`
	Identifier   string        `json:"identifier"`
	Name         string        `json:"name"`
	Description  string        `json:"description"`
	Status       ProjectStatus `json:"status"`
	LogoGradient int           `json:"logo_gradient"`
	LogoEmoji    string        `json:"logo_emoji"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
}

type ProjectListItem struct {
	Project
	EpicCount  int `json:"epic_count"`
	IssueCount int `json:"issue_count"`
	DoneCount  int `json:"done_count"`
}

type Epic struct {
	ID          string     `json:"id"`
	ProjectID   string     `json:"project_id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      EpicStatus `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type Issue struct {
	ID          string      `json:"id"`
	Identifier  string      `json:"identifier"`
	ProjectID   string      `json:"project_id"`
	EpicID      *string     `json:"epic_id"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Status      IssueStatus `json:"status"`
	Priority    Priority    `json:"priority"`
	AgentID     *string     `json:"agent_id"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

type Comment struct {
	ID        string    `json:"id"`
	IssueID   string    `json:"issue_id"`
	Content   string    `json:"content"`
	AgentID   *string   `json:"agent_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type IssueQuery struct {
	Status   string
	Priority string
	EpicID   string
	Search   string
	Sort     string
	Order    string
	Limit    int
	Cursor   string
}

type Page[T any] struct {
	Items      []T    `json:"items"`
	NextCursor string `json:"next_cursor,omitempty"`
}
