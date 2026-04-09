package taskfile

import (
	"strings"
)

func Render(doc *Document) []byte {
	var b strings.Builder
	b.WriteString("# Task\n\n")
	b.WriteString("<!-- taskmd:version 1 -->\n\n")
	renderSection(&b, StatusTodo, doc.Todo)
	b.WriteString("\n")
	renderSection(&b, StatusInProgress, doc.InProgress)
	b.WriteString("\n")
	renderSection(&b, StatusDone, doc.Done)
	return []byte(strings.TrimRight(b.String(), "\n") + "\n")
}

func renderSection(b *strings.Builder, status Status, tasks []*Task) {
	b.WriteString("## ")
	b.WriteString(status.Heading())
	b.WriteString("\n")
	for _, task := range tasks {
		b.WriteString("\n")
		b.WriteString("### ")
		b.WriteString(task.ID)
		b.WriteString(" - ")
		b.WriteString(task.Title)
		b.WriteString("\n")
		b.WriteString("- priority: ")
		b.WriteString(string(task.Priority))
		b.WriteString("\n")
		if strings.TrimSpace(task.Assignee) != "" {
			b.WriteString("- assignee: ")
			b.WriteString(task.Assignee)
			b.WriteString("\n")
		}
		b.WriteString("- labels: ")
		b.WriteString(strings.Join(task.Labels, ", "))
		b.WriteString("\n")
		b.WriteString("- created: ")
		b.WriteString(task.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))
		b.WriteString("\n")
		b.WriteString("- updated: ")
		b.WriteString(task.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"))
		b.WriteString("\n")
		if strings.TrimSpace(task.Notes) == "" {
			continue
		}
		b.WriteString("\n")
		b.WriteString("#### Notes\n")
		b.WriteString(task.Notes)
		b.WriteString("\n")
	}
}
