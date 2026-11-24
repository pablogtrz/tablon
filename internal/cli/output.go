package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/pablogtrz/tablon/internal/domain"
)

func printJSON(w io.Writer, raw []byte) error {
	if len(raw) == 0 {
		return nil
	}
	var out bytes.Buffer
	if err := json.Indent(&out, raw, "", "  "); err != nil {
		return err
	}
	_, err := fmt.Fprintln(w, out.String())
	return err
}

func printProjects(w io.Writer, items []domain.ProjectListItem, compact bool) error {
	if compact {
		for _, p := range items {
			fmt.Fprintf(w, "%s | %s | %s\n", p.Identifier, p.Status, singleLine(p.Name))
		}
		return nil
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "ID\tSTATUS\tNAME")
	for _, p := range items {
		fmt.Fprintf(tw, "%s\t%s\t%s\n", p.Identifier, p.Status, singleLine(p.Name))
	}
	return tw.Flush()
}

func printProject(w io.Writer, p domain.Project, compact bool) error {
	if compact {
		fmt.Fprintf(w, "%s | %s | %s\n", p.Identifier, p.Status, singleLine(p.Name))
		return nil
	}
	fmt.Fprintf(w, "ID: %s\nName: %s\nStatus: %s\nDescription: %s\n",
		p.Identifier, singleLine(p.Name), p.Status, singleLine(p.Description))
	return nil
}

func printEpics(w io.Writer, items []domain.Epic, compact bool) error {
	if compact {
		for _, e := range items {
			fmt.Fprintf(w, "%s | %s | %s\n", e.ID, e.Status, singleLine(e.Title))
		}
		return nil
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "ID\tSTATUS\tTITLE")
	for _, e := range items {
		fmt.Fprintf(tw, "%s\t%s\t%s\n", e.ID, e.Status, singleLine(e.Title))
	}
	return tw.Flush()
}

func printEpic(w io.Writer, e domain.Epic, compact bool) error {
	if compact {
		fmt.Fprintf(w, "%s | %s | %s\n", e.ID, e.Status, singleLine(e.Title))
		return nil
	}
	fmt.Fprintf(w, "ID: %s\nTitle: %s\nStatus: %s\nDescription: %s\n",
		e.ID, singleLine(e.Title), e.Status, singleLine(e.Description))
	return nil
}

func printIssues(w io.Writer, items []domain.Issue, compact bool) error {
	if compact {
		for _, i := range items {
			fmt.Fprintf(w, "%s | %s | %s | %s\n", i.Identifier, i.Status, i.Priority, singleLine(i.Title))
		}
		return nil
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "ID\tSTATUS\tPRIORITY\tTITLE")
	for _, i := range items {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", i.Identifier, i.Status, i.Priority, singleLine(i.Title))
	}
	return tw.Flush()
}

func printIssue(w io.Writer, i domain.Issue, compact bool) error {
	if compact {
		fmt.Fprintf(w, "%s | %s | %s | %s\n", i.Identifier, i.Status, i.Priority, singleLine(i.Title))
		return nil
	}
	epic := "-"
	if i.EpicID != nil {
		epic = *i.EpicID
	}
	agent := "-"
	if i.AgentID != nil {
		agent = *i.AgentID
	}
	fmt.Fprintf(w, "ID: %s\nTitle: %s\nStatus: %s\nPriority: %s\nEpic: %s\nAgent: %s\nDescription: %s\n",
		i.Identifier, singleLine(i.Title), i.Status, i.Priority, epic, agent, singleLine(i.Description))
	return nil
}

func printComments(w io.Writer, items []domain.Comment, compact bool) error {
	if compact {
		for _, c := range items {
			agent := "-"
			if c.AgentID != nil {
				agent = *c.AgentID
			}
			fmt.Fprintf(w, "%s | %s | %s | %s\n", c.ID, c.CreatedAt.Format("2006-01-02T15:04:05Z"), agent, singleLine(c.Content))
		}
		return nil
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "ID\tCREATED\tAGENT\tCONTENT")
	for _, c := range items {
		agent := "-"
		if c.AgentID != nil {
			agent = *c.AgentID
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", c.ID, c.CreatedAt.Format("2006-01-02T15:04:05Z"), agent, singleLine(c.Content))
	}
	return tw.Flush()
}

func printComment(w io.Writer, c domain.Comment, compact bool) error {
	if compact {
		agent := "-"
		if c.AgentID != nil {
			agent = *c.AgentID
		}
		fmt.Fprintf(w, "%s | %s | %s | %s\n", c.ID, c.CreatedAt.Format("2006-01-02T15:04:05Z"), agent, singleLine(c.Content))
		return nil
	}
	agent := "-"
	if c.AgentID != nil {
		agent = *c.AgentID
	}
	fmt.Fprintf(w, "ID: %s\nCreated: %s\nAgent: %s\nContent: %s\n",
		c.ID, c.CreatedAt.Format("2006-01-02T15:04:05Z"), agent, singleLine(c.Content))
	return nil
}

func singleLine(value string) string {
	value = strings.ReplaceAll(value, "\r\n", "\n")
	value = strings.ReplaceAll(value, "\n", "\\n")
	return strings.TrimSpace(value)
}
