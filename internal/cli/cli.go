package cli

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/pablogtrz/tablon/internal/domain"
)

type globals struct {
	url      string
	urlSet   bool
	token    string
	tokenSet bool
	timeout  string
	json     bool
	compact  bool
	version  bool
}

type usageError struct {
	msg     string
	printed bool
}

func (e *usageError) Error() string { return e.msg }

type commandContext struct {
	client  *Client
	in      io.Reader
	out     io.Writer
	errOut  io.Writer
	json    bool
	compact bool
}

func Run(args []string, env func(string) string, stdin io.Reader, stdout, stderr io.Writer) int {
	g, rest, err := parseGlobals(args)
	if err != nil {
		fmt.Fprintln(stderr, "tablon:", err)
		return 2
	}
	if g.version {
		_ = printVersion(stdout, g.json)
		return 0
	}
	if len(rest) == 0 || rest[0] == "-h" || rest[0] == "--help" {
		printGlobalHelp(stdout)
		return 0
	}
	if rest[0] == "help" {
		printGroupHelp(stdout, rest[1:])
		return 0
	}
	if rest[0] == "version" {
		_ = printVersion(stdout, g.json)
		return 0
	}

	baseURL := g.url
	if !g.urlSet {
		baseURL = env("TABLON_CLI_URL")
	}
	token := g.token
	if !g.tokenSet {
		token = env("TABLON_CLI_TOKEN")
	}
	timeout := 15 * time.Second
	if g.timeout != "" {
		timeout, err = time.ParseDuration(g.timeout)
		if err != nil {
			fmt.Fprintln(stderr, "tablon: --timeout must be a duration such as 15s")
			return 2
		}
	}

	client, err := NewClient(baseURL, token, timeout)
	if err != nil {
		fmt.Fprintln(stderr, "tablon:", err)
		return 2
	}

	c := &commandContext{
		client:  client,
		in:      stdin,
		out:     stdout,
		errOut:  stderr,
		json:    g.json,
		compact: g.compact,
	}
	if err := c.dispatch(rest); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		var ue *usageError
		if errors.As(err, &ue) {
			if !ue.printed && ue.msg != "" {
				fmt.Fprintln(stderr, "tablon:", ue.msg)
			}
			return 2
		}
		fmt.Fprintln(stderr, "tablon:", err)
		return 1
	}
	return 0
}

func parseGlobals(args []string) (globals, []string, error) {
	var g globals
	rest := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--json":
			g.json = true
		case arg == "--compact":
			g.compact = true
		case arg == "--version":
			g.version = true
		case arg == "--":
			rest = append(rest, args[i+1:]...)
			return validateGlobals(g, rest)
		case arg == "--url":
			i++
			if i >= len(args) {
				return g, nil, errors.New("--url needs a value")
			}
			g.url, g.urlSet = args[i], true
		case strings.HasPrefix(arg, "--url="):
			g.url, g.urlSet = strings.TrimPrefix(arg, "--url="), true
		case arg == "--token":
			i++
			if i >= len(args) {
				return g, nil, errors.New("--token needs a value")
			}
			g.token, g.tokenSet = args[i], true
		case strings.HasPrefix(arg, "--token="):
			g.token, g.tokenSet = strings.TrimPrefix(arg, "--token="), true
		case arg == "--timeout":
			i++
			if i >= len(args) {
				return g, nil, errors.New("--timeout needs a value")
			}
			g.timeout = args[i]
		case strings.HasPrefix(arg, "--timeout="):
			g.timeout = strings.TrimPrefix(arg, "--timeout=")
		default:
			rest = append(rest, arg)
		}
	}
	return validateGlobals(g, rest)
}

func validateGlobals(g globals, rest []string) (globals, []string, error) {
	if g.json && g.compact {
		return g, nil, errors.New("--json and --compact cannot be used together")
	}
	return g, rest, nil
}

func (c *commandContext) dispatch(args []string) error {
	switch args[0] {
	case "project":
		return c.project(args[1:])
	case "epic":
		return c.epic(args[1:])
	case "issue":
		return c.issue(args[1:])
	case "comment":
		return c.comment(args[1:])
	default:
		return &usageError{msg: fmt.Sprintf("unknown command %q", args[0])}
	}
}

func (c *commandContext) flags(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(c.errOut)
	fs.Usage = func() {
		fmt.Fprintln(c.out, usageFor(name))
	}
	return fs
}

func parseFlags(fs *flag.FlagSet, args []string) error {
	_, err := parseCommand(fs, args)
	return err
}

// splitArgs separates flags from positionals so flags may appear before or
// after positional arguments (for example "issue comment PROJ-1 --body -").
func splitArgs(fs *flag.FlagSet, args []string) ([]string, []string) {
	flags := make([]string, 0, len(args))
	positionals := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			positionals = append(positionals, args[i+1:]...)
			break
		}
		if !strings.HasPrefix(arg, "-") || arg == "-" {
			positionals = append(positionals, arg)
			continue
		}
		flags = append(flags, arg)
		name := strings.TrimLeft(arg, "-")
		if strings.Contains(name, "=") {
			continue
		}
		f := fs.Lookup(name)
		if f == nil {
			continue
		}
		if boolFlag, ok := f.Value.(interface{ IsBoolFlag() bool }); ok && boolFlag.IsBoolFlag() {
			continue
		}
		if i+1 < len(args) {
			i++
			flags = append(flags, args[i])
		}
	}
	return flags, positionals
}

func parseCommand(fs *flag.FlagSet, args []string) ([]string, error) {
	flagArgs, positionals := splitArgs(fs, args)
	if err := fs.Parse(flagArgs); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil, flag.ErrHelp
		}
		return nil, &usageError{printed: true}
	}
	return positionals, nil
}

func requireArgs(fs *flag.FlagSet, args []string, count int, usage string) error {
	if len(args) != count {
		fs.Usage()
		return &usageError{printed: true, msg: usage}
	}
	return nil
}

func (c *commandContext) output(raw []byte, plain func() error) error {
	if c.json {
		return printJSON(c.out, raw)
	}
	return plain()
}

func (c *commandContext) do(method, path string, body any, headers map[string]string) ([]byte, error) {
	return c.client.Do(context.Background(), method, path, body, headers)
}

func decodePage[T any](raw []byte) ([]T, error) {
	var page struct {
		Items []T `json:"items"`
	}
	if err := json.Unmarshal(raw, &page); err != nil {
		return nil, err
	}
	return page.Items, nil
}

func idempotencyHeader(key string) (map[string]string, error) {
	if key == "" {
		var err error
		key, err = newIdempotencyKey()
		if err != nil {
			return nil, err
		}
	}
	return map[string]string{"Idempotency-Key": key}, nil
}

func newIdempotencyKey() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func issuePath(ref string) string   { return "/api/v1/issues/" + url.PathEscape(ref) }
func projectPath(ref string) string { return "/api/v1/projects/" + url.PathEscape(ref) }
func epicPath(ref string) string    { return "/api/v1/epics/" + url.PathEscape(ref) }
func commentPath(ref string) string { return "/api/v1/comments/" + url.PathEscape(ref) }

func withQuery(path string, values url.Values) string {
	if len(values) == 0 {
		return path
	}
	return path + "?" + values.Encode()
}

func provided(fs *flag.FlagSet) map[string]bool {
	set := make(map[string]bool)
	fs.Visit(func(f *flag.Flag) {
		set[f.Name] = true
	})
	return set
}

func printGlobalHelp(w io.Writer) {
	fmt.Fprint(w, `tablon - lightweight CLI for Tablon

Usage:
  tablon [--json|--compact] [--version] [--url URL] [--token TOKEN] [--timeout DURATION] <command> [flags]

Commands:
  project   list, get, create, update
  epic      list, get, create, update
  issue     list, get, create, update, comments, comment, claim, release
  comment   update, delete
  version   print build version

Environment:
  TABLON_CLI_URL    default http://127.0.0.1:8080
  TABLON_CLI_TOKEN  optional bearer token

Output:
  --json       raw API JSON
  --compact    minimal one-line output for lists
  default      readable tables and fields

Exit codes:
  0 success, 1 runtime/API/network, 2 usage/config
`)
}

func printGroupHelp(w io.Writer, args []string) {
	if len(args) == 0 {
		printGlobalHelp(w)
		return
	}
	name := args[0]
	switch name {
	case "project":
		fmt.Fprintln(w, "Usage: tablon project list|get|create|update [flags]")
	case "epic":
		fmt.Fprintln(w, "Usage: tablon epic list|get|create|update [flags]")
	case "issue":
		fmt.Fprintln(w, "Usage: tablon issue list|get|create|update|comments|comment|claim|release [flags]")
	case "comment":
		fmt.Fprintln(w, "Usage: tablon comment update|delete [flags]")
	default:
		fmt.Fprintf(w, "Unknown command %q\n", name)
	}
}

func usageFor(name string) string {
	switch name {
	case "project list":
		return "Usage: tablon project list [--include-archived]"
	case "project get":
		return "Usage: tablon project get <ref>"
	case "project create":
		return "Usage: tablon project create --name <name> [--identifier ID] [--description TEXT] [--status STATUS]"
	case "project update":
		return "Usage: tablon project update <ref> [--name TEXT] [--description TEXT] [--status STATUS]"
	case "epic list":
		return "Usage: tablon epic list --project <ref>"
	case "epic get":
		return "Usage: tablon epic get <id>"
	case "epic create":
		return "Usage: tablon epic create --project <ref> --title TEXT [--description TEXT] [--status STATUS]"
	case "epic update":
		return "Usage: tablon epic update <id> [--title TEXT] [--description TEXT] [--status STATUS]"
	case "issue list":
		return "Usage: tablon issue list --project <ref> [--status] [--priority] [--epic] [--q] [--sort] [--order] [--limit] [--cursor]"
	case "issue get":
		return "Usage: tablon issue get <ref>"
	case "issue create":
		return "Usage: tablon issue create --project <ref> --title TEXT [--description TEXT] [--status] [--priority] [--epic] [--agent] [--idempotency-key]"
	case "issue update":
		return "Usage: tablon issue update <ref> [--title TEXT] [--description TEXT] [--status] [--priority] [--epic] [--clear-epic]"
	case "issue comments":
		return "Usage: tablon issue comments <ref> [--limit N] [--cursor CURSOR]"
	case "issue comment":
		return "Usage: tablon issue comment <ref> --body TEXT [--agent ID] [--idempotency-key]"
	case "issue claim":
		return "Usage: tablon issue claim <ref> --agent <id>"
	case "issue release":
		return "Usage: tablon issue release <ref> --agent <id>"
	case "comment update":
		return "Usage: tablon comment update <comment-id> --body TEXT"
	case "comment delete":
		return "Usage: tablon comment delete <comment-id>"
	default:
		return "Usage: tablon --help"
	}
}

func (c *commandContext) project(args []string) error {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		printGroupHelp(c.out, []string{"project"})
		return nil
	}
	switch args[0] {
	case "list":
		fs := c.flags("project list")
		includeArchived := fs.Bool("include-archived", false, "include archived projects")
		if _, err := parseCommand(fs, args[1:]); err != nil {
			return err
		}
		path := "/api/v1/projects"
		if *includeArchived {
			path += "?include_archived=true"
		}
		raw, err := c.do("GET", path, nil, nil)
		if err != nil {
			return err
		}
		items, err := decodePage[domain.ProjectListItem](raw)
		if err != nil {
			return err
		}
		return c.output(raw, func() error { return printProjects(c.out, items, c.compact) })
	case "get":
		fs := c.flags("project get")
		positionals, err := parseCommand(fs, args[1:])
		if err != nil {
			return err
		}
		if err := requireArgs(fs, positionals, 1, "project get needs a project ref"); err != nil {
			return err
		}
		raw, err := c.do("GET", projectPath(positionals[0]), nil, nil)
		if err != nil {
			return err
		}
		var p domain.Project
		if err := json.Unmarshal(raw, &p); err != nil {
			return err
		}
		return c.output(raw, func() error { return printProject(c.out, p, c.compact) })
	case "create":
		fs := c.flags("project create")
		name := fs.String("name", "", "project name")
		identifier := fs.String("identifier", "", "project identifier")
		description := fs.String("description", "", "project description")
		status := fs.String("status", "", "planning, active, archived or done")
		if _, err := parseCommand(fs, args[1:]); err != nil {
			return err
		}
		if strings.TrimSpace(*name) == "" {
			return &usageError{msg: "project create needs --name"}
		}
		body := map[string]any{"name": *name}
		if *identifier != "" {
			body["identifier"] = *identifier
		}
		if *description != "" {
			body["description"] = *description
		}
		if *status != "" {
			body["status"] = *status
		}
		raw, err := c.do("POST", "/api/v1/projects", body, nil)
		if err != nil {
			return err
		}
		var p domain.Project
		if err := json.Unmarshal(raw, &p); err != nil {
			return err
		}
		return c.output(raw, func() error { return printProject(c.out, p, c.compact) })
	case "update":
		fs := c.flags("project update")
		name := fs.String("name", "", "project name")
		description := fs.String("description", "", "project description")
		status := fs.String("status", "", "planning, active, archived or done")
		positionals, err := parseCommand(fs, args[1:])
		if err != nil {
			return err
		}
		if err := requireArgs(fs, positionals, 1, "project update needs a project ref"); err != nil {
			return err
		}
		set := provided(fs)
		body := map[string]any{}
		if set["name"] {
			body["name"] = *name
		}
		if set["description"] {
			body["description"] = *description
		}
		if set["status"] {
			body["status"] = *status
		}
		if len(body) == 0 {
			return &usageError{msg: "project update needs at least one field"}
		}
		raw, err := c.do("PATCH", projectPath(positionals[0]), body, nil)
		if err != nil {
			return err
		}
		var p domain.Project
		if err := json.Unmarshal(raw, &p); err != nil {
			return err
		}
		return c.output(raw, func() error { return printProject(c.out, p, c.compact) })
	default:
		return &usageError{msg: fmt.Sprintf("unknown project command %q", args[0])}
	}
}
func (c *commandContext) epic(args []string) error {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		printGroupHelp(c.out, []string{"epic"})
		return nil
	}
	switch args[0] {
	case "list":
		fs := c.flags("epic list")
		project := fs.String("project", "", "project ref")
		if _, err := parseCommand(fs, args[1:]); err != nil {
			return err
		}
		if *project == "" {
			return &usageError{msg: "epic list needs --project"}
		}
		raw, err := c.do("GET", projectPath(*project)+"/epics", nil, nil)
		if err != nil {
			return err
		}
		items, err := decodePage[domain.Epic](raw)
		if err != nil {
			return err
		}
		return c.output(raw, func() error { return printEpics(c.out, items, c.compact) })
	case "get":
		fs := c.flags("epic get")
		positionals, err := parseCommand(fs, args[1:])
		if err != nil {
			return err
		}
		if err := requireArgs(fs, positionals, 1, "epic get needs an epic id"); err != nil {
			return err
		}
		raw, err := c.do("GET", epicPath(positionals[0]), nil, nil)
		if err != nil {
			return err
		}
		var e domain.Epic
		if err := json.Unmarshal(raw, &e); err != nil {
			return err
		}
		return c.output(raw, func() error { return printEpic(c.out, e, c.compact) })
	case "create":
		fs := c.flags("epic create")
		project := fs.String("project", "", "project ref")
		title := fs.String("title", "", "epic title")
		description := fs.String("description", "", "epic description")
		status := fs.String("status", "", "backlog, in-progress or done")
		if _, err := parseCommand(fs, args[1:]); err != nil {
			return err
		}
		if *project == "" || strings.TrimSpace(*title) == "" {
			return &usageError{msg: "epic create needs --project and --title"}
		}
		body := map[string]any{"title": *title}
		if *description != "" {
			body["description"] = *description
		}
		if *status != "" {
			body["status"] = *status
		}
		raw, err := c.do("POST", projectPath(*project)+"/epics", body, nil)
		if err != nil {
			return err
		}
		var e domain.Epic
		if err := json.Unmarshal(raw, &e); err != nil {
			return err
		}
		return c.output(raw, func() error { return printEpic(c.out, e, c.compact) })
	case "update":
		fs := c.flags("epic update")
		title := fs.String("title", "", "epic title")
		description := fs.String("description", "", "epic description")
		status := fs.String("status", "", "backlog, in-progress or done")
		positionals, err := parseCommand(fs, args[1:])
		if err != nil {
			return err
		}
		if err := requireArgs(fs, positionals, 1, "epic update needs an epic id"); err != nil {
			return err
		}
		set := provided(fs)
		body := map[string]any{}
		if set["title"] {
			body["title"] = *title
		}
		if set["description"] {
			body["description"] = *description
		}
		if set["status"] {
			body["status"] = *status
		}
		if len(body) == 0 {
			return &usageError{msg: "epic update needs at least one field"}
		}
		raw, err := c.do("PATCH", epicPath(positionals[0]), body, nil)
		if err != nil {
			return err
		}
		var e domain.Epic
		if err := json.Unmarshal(raw, &e); err != nil {
			return err
		}
		return c.output(raw, func() error { return printEpic(c.out, e, c.compact) })
	default:
		return &usageError{msg: fmt.Sprintf("unknown epic command %q", args[0])}
	}
}

func (c *commandContext) issue(args []string) error {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		printGroupHelp(c.out, []string{"issue"})
		return nil
	}
	switch args[0] {
	case "list":
		return c.issueList(args[1:])
	case "get":
		fs := c.flags("issue get")
		positionals, err := parseCommand(fs, args[1:])
		if err != nil {
			return err
		}
		if err := requireArgs(fs, positionals, 1, "issue get needs an issue ref"); err != nil {
			return err
		}
		raw, err := c.do("GET", issuePath(positionals[0]), nil, nil)
		if err != nil {
			return err
		}
		var issue domain.Issue
		if err := json.Unmarshal(raw, &issue); err != nil {
			return err
		}
		return c.output(raw, func() error { return printIssue(c.out, issue, c.compact) })
	case "create":
		return c.issueCreate(args[1:])
	case "update":
		return c.issueUpdate(args[1:])
	case "comments":
		return c.issueComments(args[1:])
	case "comment":
		return c.issueComment(args[1:])
	case "claim", "release":
		return c.issueAgent(args[0], args[1:])
	default:
		return &usageError{msg: fmt.Sprintf("unknown issue command %q", args[0])}
	}
}

func (c *commandContext) issueList(args []string) error {
	fs := c.flags("issue list")
	project := fs.String("project", "", "project ref")
	status := fs.String("status", "", "status filter")
	priority := fs.String("priority", "", "priority filter")
	epic := fs.String("epic", "", "epic id filter")
	search := fs.String("q", "", "search title and description")
	sortBy := fs.String("sort", "", "created_at, updated_at, identifier or priority")
	order := fs.String("order", "", "asc or desc")
	limit := fs.Int("limit", 0, "page size")
	cursor := fs.String("cursor", "", "page cursor")
	if err := parseFlags(fs, args); err != nil {
		return err
	}
	if *project == "" {
		return &usageError{msg: "issue list needs --project"}
	}
	values := url.Values{}
	if *status != "" {
		values.Set("status", *status)
	}
	if *priority != "" {
		values.Set("priority", *priority)
	}
	if *epic != "" {
		values.Set("epic_id", *epic)
	}
	if *search != "" {
		values.Set("q", *search)
	}
	if *sortBy != "" {
		values.Set("sort", *sortBy)
	}
	if *order != "" {
		values.Set("order", *order)
	}
	if *limit > 0 {
		values.Set("limit", fmt.Sprint(*limit))
	}
	if *cursor != "" {
		values.Set("cursor", *cursor)
	}
	raw, err := c.do("GET", withQuery(projectPath(*project)+"/issues", values), nil, nil)
	if err != nil {
		return err
	}
	items, err := decodePage[domain.Issue](raw)
	if err != nil {
		return err
	}
	return c.output(raw, func() error { return printIssues(c.out, items, c.compact) })
}

func (c *commandContext) issueCreate(args []string) error {
	fs := c.flags("issue create")
	project := fs.String("project", "", "project ref")
	title := fs.String("title", "", "issue title")
	description := fs.String("description", "", "issue description")
	status := fs.String("status", "", "issue status")
	priority := fs.String("priority", "", "issue priority")
	epic := fs.String("epic", "", "epic id")
	agent := fs.String("agent", "", "agent id")
	key := fs.String("idempotency-key", "", "idempotency key")
	if err := parseFlags(fs, args); err != nil {
		return err
	}
	if *project == "" || strings.TrimSpace(*title) == "" {
		return &usageError{msg: "issue create needs --project and --title"}
	}
	body := map[string]any{"title": *title}
	if *description != "" {
		body["description"] = *description
	}
	if *status != "" {
		body["status"] = *status
	}
	if *priority != "" {
		body["priority"] = *priority
	}
	if *epic != "" {
		body["epic_id"] = *epic
	}
	if *agent != "" {
		body["agent_id"] = *agent
	}
	headers, err := idempotencyHeader(*key)
	if err != nil {
		return err
	}
	raw, err := c.do("POST", projectPath(*project)+"/issues", body, headers)
	if err != nil {
		return err
	}
	var issue domain.Issue
	if err := json.Unmarshal(raw, &issue); err != nil {
		return err
	}
	return c.output(raw, func() error { return printIssue(c.out, issue, c.compact) })
}

func (c *commandContext) issueUpdate(args []string) error {
	fs := c.flags("issue update")
	title := fs.String("title", "", "issue title")
	description := fs.String("description", "", "issue description")
	status := fs.String("status", "", "issue status")
	priority := fs.String("priority", "", "issue priority")
	epic := fs.String("epic", "", "epic id")
	clearEpic := fs.Bool("clear-epic", false, "remove the epic")
	positionals, err := parseCommand(fs, args)
	if err != nil {
		return err
	}
	if err := requireArgs(fs, positionals, 1, "issue update needs an issue ref"); err != nil {
		return err
	}
	if *epic != "" && *clearEpic {
		return &usageError{msg: "--epic and --clear-epic cannot be used together"}
	}
	set := provided(fs)
	body := map[string]any{}
	if set["title"] {
		body["title"] = *title
	}
	if set["description"] {
		body["description"] = *description
	}
	if set["status"] {
		body["status"] = *status
	}
	if set["priority"] {
		body["priority"] = *priority
	}
	if set["epic"] {
		body["epic_id"] = *epic
	}
	if set["clear-epic"] {
		body["epic_id"] = nil
	}
	if len(body) == 0 {
		return &usageError{msg: "issue update needs at least one field"}
	}
	raw, err := c.do("PATCH", issuePath(positionals[0]), body, nil)
	if err != nil {
		return err
	}
	var issue domain.Issue
	if err := json.Unmarshal(raw, &issue); err != nil {
		return err
	}
	return c.output(raw, func() error { return printIssue(c.out, issue, c.compact) })
}

func (c *commandContext) issueComments(args []string) error {
	fs := c.flags("issue comments")
	limit := fs.Int("limit", 0, "page size")
	cursor := fs.String("cursor", "", "page cursor")
	positionals, err := parseCommand(fs, args)
	if err != nil {
		return err
	}
	if err := requireArgs(fs, positionals, 1, "issue comments needs an issue ref"); err != nil {
		return err
	}
	values := url.Values{}
	if *limit > 0 {
		values.Set("limit", fmt.Sprint(*limit))
	}
	if *cursor != "" {
		values.Set("cursor", *cursor)
	}
	raw, err := c.do("GET", withQuery(issuePath(positionals[0])+"/comments", values), nil, nil)
	if err != nil {
		return err
	}
	items, err := decodePage[domain.Comment](raw)
	if err != nil {
		return err
	}
	return c.output(raw, func() error { return printComments(c.out, items, c.compact) })
}

func (c *commandContext) issueComment(args []string) error {
	fs := c.flags("issue comment")
	body := fs.String("body", "", "comment body, or - for stdin")
	agent := fs.String("agent", "", "agent id")
	key := fs.String("idempotency-key", "", "idempotency key")
	positionals, err := parseCommand(fs, args)
	if err != nil {
		return err
	}
	if err := requireArgs(fs, positionals, 1, "issue comment needs an issue ref"); err != nil {
		return err
	}
	content, err := c.commentBody(*body)
	if err != nil {
		return err
	}
	payload := map[string]any{"content": content}
	if *agent != "" {
		payload["agent_id"] = *agent
	}
	headers, err := idempotencyHeader(*key)
	if err != nil {
		return err
	}
	raw, err := c.do("POST", issuePath(positionals[0])+"/comments", payload, headers)
	if err != nil {
		return err
	}
	var comment domain.Comment
	if err := json.Unmarshal(raw, &comment); err != nil {
		return err
	}
	return c.output(raw, func() error { return printComment(c.out, comment, c.compact) })
}

func (c *commandContext) commentBody(value string) (string, error) {
	if value == "-" {
		raw, err := io.ReadAll(io.LimitReader(c.in, 1<<20))
		if err != nil {
			return "", err
		}
		value = strings.TrimRight(string(raw), "\r\n")
	}
	if strings.TrimSpace(value) == "" {
		return "", &usageError{msg: "comment body cannot be empty"}
	}
	return value, nil
}

func (c *commandContext) issueAgent(action string, args []string) error {
	fs := c.flags("issue " + action)
	agent := fs.String("agent", "", "agent id")
	positionals, err := parseCommand(fs, args)
	if err != nil {
		return err
	}
	if err := requireArgs(fs, positionals, 1, "issue "+action+" needs an issue ref"); err != nil {
		return err
	}
	if strings.TrimSpace(*agent) == "" {
		return &usageError{msg: "issue " + action + " needs --agent"}
	}
	raw, err := c.do("POST", issuePath(positionals[0])+"/"+action, map[string]any{"agent_id": *agent}, nil)
	if err != nil {
		return err
	}
	var issue domain.Issue
	if err := json.Unmarshal(raw, &issue); err != nil {
		return err
	}
	return c.output(raw, func() error {
		if c.compact {
			fmt.Fprintf(c.out, "%s | %s\n", issue.Identifier, *agent)
			return nil
		}
		return printIssue(c.out, issue, false)
	})
}

func (c *commandContext) comment(args []string) error {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		printGroupHelp(c.out, []string{"comment"})
		return nil
	}
	switch args[0] {
	case "update":
		fs := c.flags("comment update")
		body := fs.String("body", "", "comment body, or - for stdin")
		positionals, err := parseCommand(fs, args[1:])
		if err != nil {
			return err
		}
		if err := requireArgs(fs, positionals, 1, "comment update needs a comment id"); err != nil {
			return err
		}
		content, err := c.commentBody(*body)
		if err != nil {
			return err
		}
		raw, err := c.do("PATCH", commentPath(positionals[0]), map[string]any{"content": content}, nil)
		if err != nil {
			return err
		}
		var comment domain.Comment
		if err := json.Unmarshal(raw, &comment); err != nil {
			return err
		}
		return c.output(raw, func() error { return printComment(c.out, comment, c.compact) })
	case "delete":
		fs := c.flags("comment delete")
		positionals, err := parseCommand(fs, args[1:])
		if err != nil {
			return err
		}
		if err := requireArgs(fs, positionals, 1, "comment delete needs a comment id"); err != nil {
			return err
		}
		raw, err := c.do("DELETE", commentPath(positionals[0]), nil, nil)
		if err != nil {
			return err
		}
		return c.output(raw, func() error {
			fmt.Fprintln(c.out, "deleted")
			return nil
		})
	default:
		return &usageError{msg: fmt.Sprintf("unknown comment command %q", args[0])}
	}
}
