package preview

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/ferdikt/taskmd-cli/internal/service"
	"github.com/ferdikt/taskmd-cli/internal/taskfile"
	"github.com/yuin/goldmark"
)

//go:embed template.html
var pageTemplate string

type Server struct {
	service *service.Service
	path    string
}

type sectionView struct {
	Name  string
	Key   string
	Tasks []taskView
}

type taskView struct {
	ID         string
	Title      string
	Priority   string
	PriorityUI string
	Status     string
	Assignee   string
	Labels     []string
	CreatedAt  string
	UpdatedAt  string
	NotesHTML  template.HTML
	OpenURL    string
	Selected   bool
}

type detailView struct {
	ID         string
	Title      string
	Status     string
	Priority   string
	PriorityUI string
	Assignee   string
	Labels     []string
	CreatedAt  string
	UpdatedAt  string
	NotesHTML  template.HTML
}

type pageData struct {
	Title           string
	ProjectName     string
	ProjectKey      string
	ProjectInitials string
	Path            string
	GeneratedAt     string
	CurrentView     string
	ViewTitle       string
	ViewDescription string
	SearchQuery     string
	LabelFilter     string
	AssigneeFilter  string
	AvailableLabels []string
	AvailablePeople []string
	HasFilters      bool
	ActiveFilters   int
	AllTasksCount   int
	TotalTasks      int
	TodoCount       int
	InProgressCount int
	DoneCount       int
	ListURL         string
	BoardURL        string
	ClearURL        string
	CloseIssueURL   string
	SelectedTask    *detailView
	Sections        []sectionView
}

type apiResponse struct {
	Path        string           `json:"path"`
	GeneratedAt time.Time        `json:"generated_at"`
	Version     int              `json:"version"`
	Todo        []*taskfile.Task `json:"todo"`
	InProgress  []*taskfile.Task `json:"in_progress"`
	Done        []*taskfile.Task `json:"done"`
}

func New(service *service.Service, path string) *Server {
	return &Server{service: service, path: path}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/api/document", s.handleAPI)
	return mux
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	doc, err := s.service.Document(s.path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tpl, err := template.New("preview").Parse(pageTemplate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	currentView := normalizeView(r.URL.Query().Get("view"))
	filters := readFilters(r)
	availableLabels, availablePeople := filterOptions(doc)
	todoTasks := filterTasks(doc.Todo, filters)
	inProgressTasks := filterTasks(doc.InProgress, filters)
	doneTasks := filterTasks(doc.Done, filters)
	selectedTask := findTaskByID(filters.IssueID, todoTasks, inProgressTasks, doneTasks)
	projectName, projectKey, projectInitials := projectIdentity(s.path)
	viewTitle, viewDescription := viewCopy(currentView)
	data := pageData{
		Title:           "taskmd preview",
		ProjectName:     projectName,
		ProjectKey:      projectKey,
		ProjectInitials: projectInitials,
		Path:            s.path,
		GeneratedAt:     formatTimestampLong(time.Now()),
		CurrentView:     currentView,
		ViewTitle:       viewTitle,
		ViewDescription: viewDescription,
		SearchQuery:     filters.Query,
		LabelFilter:     filters.Label,
		AssigneeFilter:  filters.Assignee,
		AvailableLabels: availableLabels,
		AvailablePeople: availablePeople,
		HasFilters:      filters.activeCount() > 0,
		ActiveFilters:   filters.activeCount(),
		AllTasksCount:   len(doc.AllTasks()),
		TotalTasks:      len(todoTasks) + len(inProgressTasks) + len(doneTasks),
		TodoCount:       len(todoTasks),
		InProgressCount: len(inProgressTasks),
		DoneCount:       len(doneTasks),
		ListURL:         buildViewURL("list", filters),
		BoardURL:        buildViewURL("board", filters),
		ClearURL:        buildViewURL(currentView, previewFilters{}),
		CloseIssueURL: buildViewURL(currentView, previewFilters{
			Query:    filters.Query,
			Label:    filters.Label,
			Assignee: filters.Assignee,
		}),
		SelectedTask: buildDetailView(selectedTask),
		Sections: []sectionView{
			buildSection(taskfile.StatusTodo, todoTasks, currentView, filters),
			buildSection(taskfile.StatusInProgress, inProgressTasks, currentView, filters),
			buildSection(taskfile.StatusDone, doneTasks, currentView, filters),
		},
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type previewFilters struct {
	Query    string
	Label    string
	Assignee string
	IssueID  string
}

func (s *Server) handleAPI(w http.ResponseWriter, r *http.Request) {
	doc, err := s.service.Document(s.path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	resp := apiResponse{
		Path:        s.path,
		GeneratedAt: time.Now(),
		Version:     doc.Version,
		Todo:        doc.Todo,
		InProgress:  doc.InProgress,
		Done:        doc.Done,
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	_ = encoder.Encode(resp)
}

func buildSection(status taskfile.Status, tasks []*taskfile.Task, currentView string, filters previewFilters) sectionView {
	views := make([]taskView, 0, len(tasks))
	for _, task := range tasks {
		views = append(views, taskView{
			ID:         task.ID,
			Title:      task.Title,
			Priority:   strings.ToUpper(string(task.Priority)),
			PriorityUI: priorityClass(task.Priority),
			Status:     status.Heading(),
			Assignee:   task.Assignee,
			Labels:     task.Labels,
			CreatedAt:  formatTimestampShort(task.CreatedAt),
			UpdatedAt:  formatTimestampShort(task.UpdatedAt),
			NotesHTML:  renderMarkdown(task.Notes),
			OpenURL: buildViewURL(currentView, previewFilters{
				Query:    filters.Query,
				Label:    filters.Label,
				Assignee: filters.Assignee,
				IssueID:  task.ID,
			}),
			Selected: filters.IssueID == task.ID,
		})
	}
	return sectionView{
		Name:  status.Heading(),
		Key:   string(status),
		Tasks: views,
	}
}

func priorityClass(priority taskfile.Priority) string {
	switch priority {
	case taskfile.PriorityP1:
		return "p1"
	case taskfile.PriorityP2:
		return "p2"
	case taskfile.PriorityP3:
		return "p3"
	default:
		return "none"
	}
}

func renderMarkdown(value string) template.HTML {
	if strings.TrimSpace(value) == "" {
		return ""
	}
	var buf bytes.Buffer
	if err := goldmark.Convert([]byte(value), &buf); err != nil {
		return template.HTML(template.HTMLEscapeString(value))
	}
	return template.HTML(buf.String())
}

func URL(port int) string {
	return fmt.Sprintf("http://127.0.0.1:%d", port)
}

func normalizeView(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "board":
		return "board"
	case "backlog", "list":
		return "list"
	default:
		return "list"
	}
}

func projectIdentity(path string) (name, key, initials string) {
	root := filepath.Base(filepath.Dir(filepath.Dir(path)))
	if strings.TrimSpace(root) == "" || root == "." || root == string(filepath.Separator) {
		root = "Project"
	}
	name = root

	parts := strings.FieldsFunc(root, func(r rune) bool {
		return !(r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z' || r >= '0' && r <= '9')
	})

	var keyBuilder strings.Builder
	var initialsBuilder strings.Builder
	for _, part := range parts {
		if part == "" {
			continue
		}
		upper := strings.ToUpper(part)
		if initialsBuilder.Len() < 2 {
			initialsBuilder.WriteByte(upper[0])
		}
		keyBuilder.WriteString(upper)
	}
	if keyBuilder.Len() == 0 {
		keyBuilder.WriteString("TASK")
	}
	if initialsBuilder.Len() == 0 {
		initialsBuilder.WriteString(keyBuilder.String()[:1])
	}
	key = keyBuilder.String()
	if len(key) > 8 {
		key = key[:8]
	}
	initials = initialsBuilder.String()
	return name, key, initials
}

func viewCopy(view string) (title, description string) {
	switch view {
	case "board":
		return "Board", "Kanban board grouped by status, close to the working view teams use in Jira."
	default:
		return "List", "Structured issue list grouped by status for planning, triage, and assignment."
	}
}

func readFilters(r *http.Request) previewFilters {
	return previewFilters{
		Query:    strings.TrimSpace(r.URL.Query().Get("q")),
		Label:    strings.ToLower(strings.TrimSpace(r.URL.Query().Get("label"))),
		Assignee: taskfile.NormalizeAssignee(r.URL.Query().Get("assignee")),
		IssueID:  strings.TrimSpace(r.URL.Query().Get("issue")),
	}
}

func (f previewFilters) activeCount() int {
	count := 0
	if f.Query != "" {
		count++
	}
	if f.Label != "" {
		count++
	}
	if f.Assignee != "" {
		count++
	}
	return count
}

func filterOptions(doc *taskfile.Document) ([]string, []string) {
	labelSeen := map[string]struct{}{}
	personSeen := map[string]struct{}{}
	labels := make([]string, 0)
	people := make([]string, 0)
	for _, task := range doc.AllTasks() {
		for _, label := range task.Labels {
			if _, ok := labelSeen[label]; ok {
				continue
			}
			labelSeen[label] = struct{}{}
			labels = append(labels, label)
		}
		if task.Assignee != "" {
			key := strings.ToLower(task.Assignee)
			if _, ok := personSeen[key]; !ok {
				personSeen[key] = struct{}{}
				people = append(people, task.Assignee)
			}
		}
	}
	sort.Strings(labels)
	sort.Slice(people, func(i, j int) bool {
		return strings.ToLower(people[i]) < strings.ToLower(people[j])
	})
	return labels, people
}

func filterTasks(tasks []*taskfile.Task, filters previewFilters) []*taskfile.Task {
	if filters.activeCount() == 0 {
		return tasks
	}
	filtered := make([]*taskfile.Task, 0, len(tasks))
	query := strings.ToLower(filters.Query)
	for _, task := range tasks {
		if filters.Label != "" && !containsLabel(task.Labels, filters.Label) {
			continue
		}
		if filters.Assignee != "" && !strings.EqualFold(task.Assignee, filters.Assignee) {
			continue
		}
		if query != "" && !matchesQuery(task, query) {
			continue
		}
		filtered = append(filtered, task)
	}
	return filtered
}

func containsLabel(labels []string, target string) bool {
	for _, label := range labels {
		if label == target {
			return true
		}
	}
	return false
}

func matchesQuery(task *taskfile.Task, query string) bool {
	fields := []string{
		task.ID,
		task.Title,
		task.Assignee,
		task.Notes,
		strings.Join(task.Labels, " "),
	}
	for _, field := range fields {
		if strings.Contains(strings.ToLower(field), query) {
			return true
		}
	}
	return false
}

func formatTimestampShort(ts time.Time) string {
	return ts.Format("02 Jan 15:04")
}

func formatTimestampLong(ts time.Time) string {
	return ts.Format("02 Jan 2006, 15:04")
}

func buildViewURL(view string, filters previewFilters) string {
	values := url.Values{}
	values.Set("view", view)
	if filters.Query != "" {
		values.Set("q", filters.Query)
	}
	if filters.Label != "" {
		values.Set("label", filters.Label)
	}
	if filters.Assignee != "" {
		values.Set("assignee", filters.Assignee)
	}
	if filters.IssueID != "" {
		values.Set("issue", filters.IssueID)
	}
	return "/?" + values.Encode()
}

func findTaskByID(id string, sections ...[]*taskfile.Task) *taskfile.Task {
	if id == "" {
		return nil
	}
	for _, tasks := range sections {
		for _, task := range tasks {
			if task.ID == id {
				return task
			}
		}
	}
	return nil
}

func buildDetailView(task *taskfile.Task) *detailView {
	if task == nil {
		return nil
	}
	return &detailView{
		ID:         task.ID,
		Title:      task.Title,
		Status:     task.Status.Heading(),
		Priority:   strings.ToUpper(string(task.Priority)),
		PriorityUI: priorityClass(task.Priority),
		Assignee:   task.Assignee,
		Labels:     task.Labels,
		CreatedAt:  formatTimestampLong(task.CreatedAt),
		UpdatedAt:  formatTimestampLong(task.UpdatedAt),
		NotesHTML:  renderMarkdown(task.Notes),
	}
}
