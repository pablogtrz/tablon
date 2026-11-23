package service

import (
	"context"
	"fmt"
	"hash/fnv"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/pablogtrz/tablon/internal/domain"
	"github.com/pablogtrz/tablon/internal/store"
)

var identifierPattern = regexp.MustCompile(`^[A-Z][A-Z0-9]{1,9}$`)

type Service struct {
	store *store.Store
	now   func() time.Time
	newID func() string
}

func New(st *store.Store) *Service {
	return &Service{
		store: st,
		now:   time.Now,
		newID: func() string { return uuid.Must(uuid.NewV7()).String() },
	}
}

func (s *Service) Ready(ctx context.Context) error {
	return s.store.Ping(ctx)
}

type CreateProjectInput struct {
	Name         string
	Identifier   string
	Description  string
	Status       string
	LogoGradient *int
	LogoEmoji    string
}

type UpdateProjectInput struct {
	Name         *string
	Identifier   *string
	Description  *string
	Status       *string
	LogoGradient *int
	LogoEmoji    *string
}

func (s *Service) CreateProject(ctx context.Context, in CreateProjectInput) (domain.Project, error) {
	if err := validText("name", in.Name, 1, 120); err != nil {
		return domain.Project{}, err
	}
	if err := validText("description", in.Description, 0, 20000); err != nil {
		return domain.Project{}, err
	}
	if err := validEmoji(in.LogoEmoji); err != nil {
		return domain.Project{}, err
	}

	status := domain.ProjectStatus(strings.TrimSpace(in.Status))
	if status == "" {
		status = domain.ProjectPlanning
	}
	if !status.Valid() {
		return domain.Project{}, domain.Invalid("status", "must be planning, active, archived or done")
	}

	id := s.newID()
	now := s.now().UTC()

	identifier := strings.ToUpper(strings.TrimSpace(in.Identifier))
	if identifier == "" {
		var err error
		identifier, err = s.availableIdentifier(ctx, deriveIdentifier(in.Name))
		if err != nil {
			return domain.Project{}, err
		}
	} else {
		if !identifierPattern.MatchString(identifier) {
			return domain.Project{}, domain.Invalid("identifier", "must be 2-10 uppercase letters or digits and start with a letter")
		}
		exists, err := s.store.IdentifierExists(ctx, identifier)
		if err != nil {
			return domain.Project{}, err
		}
		if exists {
			return domain.Project{}, domain.ErrConflict
		}
	}

	gradient := gradientFromID(id)
	if in.LogoGradient != nil {
		if *in.LogoGradient < 0 || *in.LogoGradient > 7 {
			return domain.Project{}, domain.Invalid("logo_gradient", "must be between 0 and 7")
		}
		gradient = *in.LogoGradient
	}

	p := domain.Project{
		ID:           id,
		Identifier:   identifier,
		Name:         strings.TrimSpace(in.Name),
		Description:  in.Description,
		Status:       status,
		LogoGradient: gradient,
		LogoEmoji:    strings.TrimSpace(in.LogoEmoji),
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.store.CreateProject(ctx, p); err != nil {
		return domain.Project{}, err
	}
	return p, nil
}

func (s *Service) ListProjects(ctx context.Context, includeArchived bool) ([]domain.ProjectListItem, error) {
	return s.store.ListProjects(ctx, includeArchived)
}

func (s *Service) GetProject(ctx context.Context, ref string) (domain.Project, error) {
	return s.store.GetProject(ctx, ref)
}

func (s *Service) UpdateProject(ctx context.Context, ref string, in UpdateProjectInput) (domain.Project, error) {
	p, err := s.store.GetProject(ctx, ref)
	if err != nil {
		return domain.Project{}, err
	}

	if in.Identifier != nil {
		identifier := strings.ToUpper(strings.TrimSpace(*in.Identifier))
		if identifier != "" && identifier != p.Identifier {
			return domain.Project{}, domain.Invalid("identifier", "is immutable")
		}
	}
	if in.Name != nil {
		if err := validText("name", *in.Name, 1, 120); err != nil {
			return domain.Project{}, err
		}
		p.Name = strings.TrimSpace(*in.Name)
	}
	if in.Description != nil {
		if err := validText("description", *in.Description, 0, 20000); err != nil {
			return domain.Project{}, err
		}
		p.Description = *in.Description
	}
	if in.Status != nil {
		status := domain.ProjectStatus(strings.TrimSpace(*in.Status))
		if !status.Valid() {
			return domain.Project{}, domain.Invalid("status", "must be planning, active, archived or done")
		}
		p.Status = status
	}
	if in.LogoGradient != nil {
		if *in.LogoGradient < 0 || *in.LogoGradient > 7 {
			return domain.Project{}, domain.Invalid("logo_gradient", "must be between 0 and 7")
		}
		p.LogoGradient = *in.LogoGradient
	}
	if in.LogoEmoji != nil {
		if err := validEmoji(*in.LogoEmoji); err != nil {
			return domain.Project{}, err
		}
		p.LogoEmoji = strings.TrimSpace(*in.LogoEmoji)
	}

	p.UpdatedAt = s.now().UTC()
	if err := s.store.UpdateProject(ctx, p); err != nil {
		return domain.Project{}, err
	}
	return p, nil
}

func (s *Service) DeleteProject(ctx context.Context, ref string) error {
	return s.store.DeleteProject(ctx, ref)
}

type CreateEpicInput struct {
	Title       string
	Description string
	Status      string
}

type UpdateEpicInput struct {
	Title       *string
	Description *string
	Status      *string
}

func (s *Service) CreateEpic(ctx context.Context, projectRef string, in CreateEpicInput) (domain.Epic, error) {
	project, err := s.store.GetProject(ctx, projectRef)
	if err != nil {
		return domain.Epic{}, err
	}
	if err := validText("title", in.Title, 1, 200); err != nil {
		return domain.Epic{}, err
	}
	if err := validText("description", in.Description, 0, 20000); err != nil {
		return domain.Epic{}, err
	}
	status := domain.EpicStatus(strings.TrimSpace(in.Status))
	if status == "" {
		status = domain.EpicBacklog
	}
	if !status.Valid() {
		return domain.Epic{}, domain.Invalid("status", "must be backlog, in-progress or done")
	}

	now := s.now().UTC()
	e := domain.Epic{
		ID:          s.newID(),
		ProjectID:   project.ID,
		Title:       strings.TrimSpace(in.Title),
		Description: in.Description,
		Status:      status,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.store.CreateEpic(ctx, e); err != nil {
		return domain.Epic{}, err
	}
	return e, nil
}

func (s *Service) ListEpics(ctx context.Context, projectRef string) ([]domain.Epic, error) {
	project, err := s.store.GetProject(ctx, projectRef)
	if err != nil {
		return nil, err
	}
	return s.store.ListEpics(ctx, project.ID)
}

func (s *Service) GetEpic(ctx context.Context, ref string) (domain.Epic, error) {
	return s.store.GetEpic(ctx, ref)
}

func (s *Service) UpdateEpic(ctx context.Context, ref string, in UpdateEpicInput) (domain.Epic, error) {
	e, err := s.store.GetEpic(ctx, ref)
	if err != nil {
		return domain.Epic{}, err
	}
	if in.Title != nil {
		if err := validText("title", *in.Title, 1, 200); err != nil {
			return domain.Epic{}, err
		}
		e.Title = strings.TrimSpace(*in.Title)
	}
	if in.Description != nil {
		if err := validText("description", *in.Description, 0, 20000); err != nil {
			return domain.Epic{}, err
		}
		e.Description = *in.Description
	}
	if in.Status != nil {
		status := domain.EpicStatus(strings.TrimSpace(*in.Status))
		if !status.Valid() {
			return domain.Epic{}, domain.Invalid("status", "must be backlog, in-progress or done")
		}
		e.Status = status
	}
	e.UpdatedAt = s.now().UTC()
	if err := s.store.UpdateEpic(ctx, e); err != nil {
		return domain.Epic{}, err
	}
	return e, nil
}

func (s *Service) DeleteEpic(ctx context.Context, ref string) error {
	return s.store.DeleteEpic(ctx, ref)
}

type CreateIssueInput struct {
	Title       string
	Description string
	Status      string
	Priority    string
	EpicID      *string
	AgentID     *string
}

type UpdateIssueInput struct {
	Title       *string
	Description *string
	Status      *string
	Priority    *string
	EpicID      *string
	EpicIDSet   bool
}

func (s *Service) CreateIssue(ctx context.Context, projectRef string, in CreateIssueInput) (domain.Issue, error) {
	project, err := s.store.GetProject(ctx, projectRef)
	if err != nil {
		return domain.Issue{}, err
	}
	if err := validText("title", in.Title, 1, 200); err != nil {
		return domain.Issue{}, err
	}
	if err := validText("description", in.Description, 0, 20000); err != nil {
		return domain.Issue{}, err
	}
	status := domain.IssueStatus(strings.TrimSpace(in.Status))
	if status == "" {
		status = domain.IssueBacklog
	}
	if !status.Valid() {
		return domain.Issue{}, domain.Invalid("status", "is not a valid issue status")
	}
	priority := domain.Priority(strings.TrimSpace(in.Priority))
	if priority == "" {
		priority = domain.PriorityNone
	}
	if !priority.Valid() {
		return domain.Issue{}, domain.Invalid("priority", "must be urgent, high, medium, low or none")
	}

	epicID, err := s.validateEpic(ctx, project.ID, in.EpicID)
	if err != nil {
		return domain.Issue{}, err
	}
	agentID, err := validAgent(in.AgentID)
	if err != nil {
		return domain.Issue{}, err
	}

	now := s.now().UTC()
	i := domain.Issue{
		ID:          s.newID(),
		ProjectID:   project.ID,
		EpicID:      epicID,
		Title:       strings.TrimSpace(in.Title),
		Description: in.Description,
		Status:      status,
		Priority:    priority,
		AgentID:     agentID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	return s.store.CreateIssue(ctx, i)
}

func (s *Service) ListIssues(ctx context.Context, projectRef string, q domain.IssueQuery) (domain.Page[domain.Issue], error) {
	project, err := s.store.GetProject(ctx, projectRef)
	if err != nil {
		return domain.Page[domain.Issue]{}, err
	}
	if q.Status != "" && !domain.IssueStatus(q.Status).Valid() {
		return domain.Page[domain.Issue]{}, domain.Invalid("status", "is not a valid issue status")
	}
	if q.Priority != "" && !domain.Priority(q.Priority).Valid() {
		return domain.Page[domain.Issue]{}, domain.Invalid("priority", "must be urgent, high, medium, low or none")
	}
	if q.Order == "" {
		q.Order = "desc"
	}
	if q.Order != "asc" && q.Order != "desc" {
		return domain.Page[domain.Issue]{}, domain.Invalid("order", "must be asc or desc")
	}
	if _, _, err := issueSort(q.Sort); err != nil {
		return domain.Page[domain.Issue]{}, err
	}
	if q.Limit <= 0 {
		q.Limit = 50
	}
	if q.Limit > 200 {
		q.Limit = 200
	}
	q.Search = strings.TrimSpace(q.Search)
	if utf8.RuneCountInString(q.Search) > 200 {
		return domain.Page[domain.Issue]{}, domain.Invalid("q", "must be at most 200 characters")
	}
	return s.store.ListIssues(ctx, project.ID, q)
}

func (s *Service) GetIssue(ctx context.Context, ref string) (domain.Issue, error) {
	return s.store.GetIssue(ctx, ref)
}

func (s *Service) UpdateIssue(ctx context.Context, ref string, in UpdateIssueInput) (domain.Issue, error) {
	i, err := s.store.GetIssue(ctx, ref)
	if err != nil {
		return domain.Issue{}, err
	}
	if in.Title != nil {
		if err := validText("title", *in.Title, 1, 200); err != nil {
			return domain.Issue{}, err
		}
		i.Title = strings.TrimSpace(*in.Title)
	}
	if in.Description != nil {
		if err := validText("description", *in.Description, 0, 20000); err != nil {
			return domain.Issue{}, err
		}
		i.Description = *in.Description
	}
	if in.Status != nil {
		status := domain.IssueStatus(strings.TrimSpace(*in.Status))
		if !status.Valid() {
			return domain.Issue{}, domain.Invalid("status", "is not a valid issue status")
		}
		i.Status = status
	}
	if in.Priority != nil {
		priority := domain.Priority(strings.TrimSpace(*in.Priority))
		if !priority.Valid() {
			return domain.Issue{}, domain.Invalid("priority", "must be urgent, high, medium, low or none")
		}
		i.Priority = priority
	}
	if in.EpicIDSet {
		epicID, err := s.validateEpic(ctx, i.ProjectID, in.EpicID)
		if err != nil {
			return domain.Issue{}, err
		}
		i.EpicID = epicID
	}

	i.UpdatedAt = s.now().UTC()
	if err := s.store.UpdateIssue(ctx, i); err != nil {
		return domain.Issue{}, err
	}
	return i, nil
}

func (s *Service) UpdateIssueStatus(ctx context.Context, ref, status string) (domain.Issue, error) {
	return s.UpdateIssue(ctx, ref, UpdateIssueInput{Status: &status})
}

func (s *Service) DeleteIssue(ctx context.Context, ref string) error {
	return s.store.DeleteIssue(ctx, ref)
}

func (s *Service) ClaimIssue(ctx context.Context, ref, agentID string) (domain.Issue, error) {
	agentID = strings.TrimSpace(agentID)
	if agentID == "" || utf8.RuneCountInString(agentID) > 100 {
		return domain.Issue{}, domain.Invalid("agent_id", "must be 1-100 characters")
	}
	return s.store.ClaimIssue(ctx, ref, agentID, s.now().UTC())
}

func (s *Service) ReleaseIssue(ctx context.Context, ref, agentID string) (domain.Issue, error) {
	agentID = strings.TrimSpace(agentID)
	if agentID == "" || utf8.RuneCountInString(agentID) > 100 {
		return domain.Issue{}, domain.Invalid("agent_id", "must be 1-100 characters")
	}
	return s.store.ReleaseIssue(ctx, ref, agentID, s.now().UTC())
}

func (s *Service) validateEpic(ctx context.Context, projectID string, epicID *string) (*string, error) {
	if epicID == nil {
		return nil, nil
	}
	id := strings.TrimSpace(*epicID)
	if id == "" {
		return nil, nil
	}
	e, err := s.store.GetEpic(ctx, id)
	if err != nil {
		if err == domain.ErrNotFound {
			return nil, domain.Invalid("epic_id", "does not exist")
		}
		return nil, err
	}
	if e.ProjectID != projectID {
		return nil, domain.Invalid("epic_id", "does not belong to this project")
	}
	return &e.ID, nil
}

func (s *Service) availableIdentifier(ctx context.Context, base string) (string, error) {
	if !identifierPattern.MatchString(base) {
		base = "PROJ"
	}
	for n := 0; n < 100; n++ {
		candidate := base
		if n > 0 {
			suffix := strconv.Itoa(n + 1)
			maxBase := 10 - len(suffix)
			if maxBase < 2 {
				maxBase = 2
			}
			trimmed := base
			if len(trimmed) > maxBase {
				trimmed = trimmed[:maxBase]
			}
			candidate = trimmed + suffix
		}
		exists, err := s.store.IdentifierExists(ctx, candidate)
		if err != nil {
			return "", err
		}
		if !exists {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("could not allocate a project identifier")
}

func deriveIdentifier(name string) string {
	var b strings.Builder
	for _, r := range strings.ToUpper(strings.TrimSpace(name)) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		}
		if b.Len() == 8 {
			break
		}
	}
	value := b.String()
	if value == "" {
		return "PROJ"
	}
	if !unicode.IsLetter(rune(value[0])) {
		value = "P" + value
	}
	if len(value) < 2 {
		value += "X"
	}
	if len(value) > 8 {
		value = value[:8]
	}
	return value
}

func gradientFromID(id string) int {
	h := fnv.New32a()
	_, _ = h.Write([]byte(id))
	return int(h.Sum32() % 8)
}

func validText(field, value string, min, max int) error {
	n := utf8.RuneCountInString(strings.TrimSpace(value))
	if n < min {
		return domain.Invalid(field, "is required")
	}
	if n > max {
		return domain.Invalid(field, fmt.Sprintf("must be at most %d characters", max))
	}
	return nil
}

func validEmoji(value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	if utf8.RuneCountInString(value) > 16 || len(value) > 64 {
		return domain.Invalid("logo_emoji", "must be a short emoji")
	}
	return nil
}

func validAgent(value *string) (*string, error) {
	if value == nil {
		return nil, nil
	}
	agent := strings.TrimSpace(*value)
	if agent == "" {
		return nil, nil
	}
	if utf8.RuneCountInString(agent) > 100 {
		return nil, domain.Invalid("agent_id", "must be at most 100 characters")
	}
	return &agent, nil
}

func issueSort(sort string) (string, bool, error) {
	switch sort {
	case "", "created_at", "updated_at", "identifier", "priority":
		return sort, false, nil
	default:
		return "", false, domain.Invalid("sort", "must be created_at, updated_at, identifier or priority")
	}
}
