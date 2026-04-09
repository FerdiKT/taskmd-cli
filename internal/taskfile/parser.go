package taskfile

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

type ParseError struct {
	Line int
	Msg  string
}

func (e *ParseError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("line %d: %s", e.Line, e.Msg)
	}
	return e.Msg
}

func Parse(data []byte) (*Document, error) {
	if len(bytes.TrimSpace(data)) == 0 {
		return NewDocument(), nil
	}

	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	source := []byte(strings.Join(lines, "\n"))
	parser := goldmark.DefaultParser()
	docAST := parser.Parse(text.NewReader(source))

	doc := NewDocument()
	doc.Version = extractVersion(lines)
	if doc.Version == 0 {
		doc.Version = CurrentVersion
	}

	type headingInfo struct {
		Level int
		Title string
		Line  int
	}

	headings := []headingInfo{}
	seenIDs := map[string]int{}
	err := ast.Walk(docAST, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		heading, ok := n.(*ast.Heading)
		if !ok {
			return ast.WalkContinue, nil
		}
		line, _ := lineRange(heading, source)
		headings = append(headings, headingInfo{
			Level: heading.Level,
			Title: strings.TrimSpace(string(heading.Text(source))),
			Line:  line,
		})
		return ast.WalkContinue, nil
	})
	if err != nil {
		return nil, err
	}

	if len(headings) == 0 || headings[0].Level != 1 || headings[0].Title != "Task" {
		return nil, &ParseError{Line: 1, Msg: `expected "# Task" heading`}
	}

	for i := 0; i < len(headings); i++ {
		heading := headings[i]
		if heading.Level != 2 {
			continue
		}

		status, err := statusFromHeading(heading.Title)
		if err != nil {
			return nil, &ParseError{Line: heading.Line, Msg: err.Error()}
		}

		sectionEnd := len(lines) + 1
		for j := i + 1; j < len(headings); j++ {
			if headings[j].Level <= 2 {
				sectionEnd = headings[j].Line
				break
			}
		}

		for j := i + 1; j < len(headings); j++ {
			if headings[j].Level <= 2 && headings[j].Line >= sectionEnd {
				break
			}
			if headings[j].Level != 3 {
				continue
			}
			if headings[j].Line >= sectionEnd {
				break
			}

			taskEnd := sectionEnd
			for k := j + 1; k < len(headings); k++ {
				if headings[k].Level <= 3 {
					taskEnd = headings[k].Line
					break
				}
			}

			task, err := parseTaskBlock(lines, headings[j].Title, status, headings[j].Line, taskEnd)
			if err != nil {
				return nil, err
			}
			if line, ok := seenIDs[task.ID]; ok {
				return nil, &ParseError{Line: headings[j].Line, Msg: fmt.Sprintf("duplicate task id %s (first seen at line %d)", task.ID, line)}
			}
			seenIDs[task.ID] = headings[j].Line
			doc.AppendTask(task)
		}
	}

	return doc, nil
}

func parseTaskBlock(lines []string, headingTitle string, status Status, startLine, endLine int) (*Task, error) {
	id, title, ok := strings.Cut(headingTitle, " - ")
	if !ok {
		return nil, &ParseError{Line: startLine, Msg: `task heading must look like "### T001 - Title"`}
	}
	id = strings.TrimSpace(id)
	title = strings.TrimSpace(title)
	if id == "" || title == "" {
		return nil, &ParseError{Line: startLine, Msg: "task heading is missing id or title"}
	}

	task := &Task{
		ID:       id,
		Title:    title,
		Status:   status,
		Priority: PriorityNone,
		Assignee: "",
		Labels:   []string{},
	}

	block := []string{}
	if startLine < len(lines) {
		block = append(block, lines[startLine:endLine-1]...)
	}

	metaDone := false
	notesMode := false
	notesLines := []string{}

	for offset, raw := range block {
		lineNo := startLine + offset + 1
		line := strings.TrimRight(raw, "\r")
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			if notesMode {
				notesLines = append(notesLines, "")
			}
			continue
		}
		if trimmed == "#### Notes" {
			notesMode = true
			metaDone = true
			continue
		}
		if notesMode {
			notesLines = append(notesLines, line)
			continue
		}
		if !strings.HasPrefix(trimmed, "- ") {
			return nil, &ParseError{Line: lineNo, Msg: "expected metadata bullet or Notes block"}
		}
		keyValue := strings.TrimSpace(strings.TrimPrefix(trimmed, "- "))
		key, value, ok := strings.Cut(keyValue, ":")
		if !ok {
			return nil, &ParseError{Line: lineNo, Msg: "metadata bullet must use key: value format"}
		}
		key = strings.TrimSpace(strings.ToLower(key))
		value = strings.TrimSpace(value)

		switch key {
		case "priority":
			priority, err := ParsePriority(value)
			if err != nil {
				return nil, &ParseError{Line: lineNo, Msg: err.Error()}
			}
			task.Priority = priority
		case "assignee":
			task.Assignee = NormalizeAssignee(value)
		case "labels":
			if value == "" {
				task.Labels = []string{}
			} else {
				task.Labels = NormalizeLabels(strings.Split(value, ","))
			}
		case "created":
			ts, err := time.Parse(time.RFC3339, value)
			if err != nil {
				return nil, &ParseError{Line: lineNo, Msg: fmt.Sprintf("invalid created timestamp %q", value)}
			}
			task.CreatedAt = ts
		case "updated":
			ts, err := time.Parse(time.RFC3339, value)
			if err != nil {
				return nil, &ParseError{Line: lineNo, Msg: fmt.Sprintf("invalid updated timestamp %q", value)}
			}
			task.UpdatedAt = ts
		default:
			return nil, &ParseError{Line: lineNo, Msg: fmt.Sprintf("unknown metadata field %q", key)}
		}
		metaDone = true
	}

	if !metaDone {
		return nil, &ParseError{Line: startLine, Msg: "task is missing metadata bullets"}
	}
	if task.CreatedAt.IsZero() {
		return nil, &ParseError{Line: startLine, Msg: "task is missing created timestamp"}
	}
	if task.UpdatedAt.IsZero() {
		return nil, &ParseError{Line: startLine, Msg: "task is missing updated timestamp"}
	}
	task.Notes = strings.TrimSpace(strings.Join(notesLines, "\n"))
	return task, nil
}

func lineRange(node ast.Node, source []byte) (int, int) {
	lines := node.Lines()
	if lines.Len() == 0 {
		return 0, 0
	}
	first := lines.At(0)
	last := lines.At(lines.Len() - 1)
	start := bytes.Count(source[:first.Start], []byte("\n")) + 1
	end := bytes.Count(source[:last.Stop], []byte("\n")) + 1
	return start, end
}

func statusFromHeading(title string) (Status, error) {
	switch strings.TrimSpace(title) {
	case "Todo":
		return StatusTodo, nil
	case "In Progress":
		return StatusInProgress, nil
	case "Done":
		return StatusDone, nil
	default:
		return "", fmt.Errorf("unknown section %q", title)
	}
}

func extractVersion(lines []string) int {
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "<!-- taskmd:version ") && strings.HasSuffix(trimmed, "-->") {
			value := strings.TrimSuffix(strings.TrimPrefix(trimmed, "<!-- taskmd:version "), " -->")
			var version int
			if _, err := fmt.Sscanf(value, "%d", &version); err == nil {
				return version
			}
		}
	}
	return 0
}

func Validate(data []byte) error {
	_, err := Parse(data)
	return err
}

var ErrTaskNotFound = errors.New("task not found")
