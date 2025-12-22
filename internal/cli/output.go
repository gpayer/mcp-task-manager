package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/gpayer/mcp-task-manager/internal/task"
)

// FormatTaskDetail formats a single task for human-readable output
func FormatTaskDetail(t *task.Task) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Task #%d\n", t.ID))
	sb.WriteString(fmt.Sprintf("Title:       %s\n", t.Title))
	sb.WriteString(fmt.Sprintf("Status:      %s\n", t.Status))
	sb.WriteString(fmt.Sprintf("Priority:    %s\n", t.Priority))
	sb.WriteString(fmt.Sprintf("Type:        %s\n", t.Type))
	sb.WriteString(fmt.Sprintf("Created:     %s\n", t.CreatedAt.Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("Updated:     %s\n", t.UpdatedAt.Format("2006-01-02 15:04:05")))
	if t.Description != "" {
		sb.WriteString(fmt.Sprintf("\nDescription:\n%s\n", t.Description))
	}
	return sb.String()
}

// FormatTaskTable formats a list of tasks as a table
func FormatTaskTable(tasks []*task.Task) string {
	if len(tasks) == 0 {
		return "No tasks found."
	}

	var sb strings.Builder
	w := tabwriter.NewWriter(&sb, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tTitle\tStatus\tPriority\tType")
	for _, t := range tasks {
		title := t.Title
		if len(title) > 40 {
			title = title[:37] + "..."
		}
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n", t.ID, title, t.Status, t.Priority, t.Type)
	}
	w.Flush()
	return sb.String()
}

// FormatMessage formats a simple message
func FormatMessage(msg string, id int) string {
	return msg
}

// FormatJSON writes a value as JSON to the writer
func FormatJSON(w io.Writer, v interface{}) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// FormatJSONMessage writes a message with ID as JSON
func FormatJSONMessage(w io.Writer, msg string, id int) error {
	return FormatJSON(w, map[string]interface{}{
		"message": msg,
		"id":      id,
	})
}
