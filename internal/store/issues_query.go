package store

import (
	"encoding/base64"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/pablogtrz/tablon/internal/domain"
)

type issueCursor struct {
	Value string `json:"v"`
	ID    string `json:"id"`
}

func encodeIssueCursor(c issueCursor) string {
	body, _ := json.Marshal(c)
	return base64.RawURLEncoding.EncodeToString(body)
}

func decodeIssueCursor(value string) (issueCursor, error) {
	body, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return issueCursor{}, domain.Invalid("cursor", "is invalid")
	}
	var c issueCursor
	if err := json.Unmarshal(body, &c); err != nil || c.ID == "" {
		return issueCursor{}, domain.Invalid("cursor", "is invalid")
	}
	return c, nil
}

func issueSortExpr(sort string) (expr string, isPriority bool, err error) {
	switch sort {
	case "", "created_at":
		return "created_at", false, nil
	case "updated_at":
		return "updated_at", false, nil
	case "identifier":
		return "identifier", false, nil
	case "priority":
		return `CASE priority
			WHEN 'urgent' THEN 4
			WHEN 'high' THEN 3
			WHEN 'medium' THEN 2
			WHEN 'low' THEN 1
			ELSE 0
		END`, true, nil
	default:
		return "", false, domain.Invalid("sort", "must be created_at, updated_at, identifier or priority")
	}
}

func issueSortValue(i domain.Issue, sort string) string {
	switch sort {
	case "updated_at":
		return formatTime(i.UpdatedAt)
	case "identifier":
		return i.Identifier
	case "priority":
		return strconv.Itoa(priorityRank(i.Priority))
	default:
		return formatTime(i.CreatedAt)
	}
}

func priorityRank(p domain.Priority) int {
	switch p {
	case domain.PriorityUrgent:
		return 4
	case domain.PriorityHigh:
		return 3
	case domain.PriorityMedium:
		return 2
	case domain.PriorityLow:
		return 1
	default:
		return 0
	}
}

func join(parts []string, sep string) string {
	return strings.Join(parts, sep)
}
