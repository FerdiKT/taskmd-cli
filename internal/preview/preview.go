package preview

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
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
	Labels     []string
	CreatedAt  string
	UpdatedAt  string
	NotesHTML  template.HTML
}

type pageData struct {
	Title           string
	Path            string
	GeneratedAt     string
	TotalTasks      int
	TodoCount       int
	InProgressCount int
	DoneCount       int
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

	data := pageData{
		Title:           "taskmd preview",
		Path:            s.path,
		GeneratedAt:     time.Now().Format("2006-01-02 15:04:05 -07:00"),
		TotalTasks:      len(doc.AllTasks()),
		TodoCount:       len(doc.Todo),
		InProgressCount: len(doc.InProgress),
		DoneCount:       len(doc.Done),
		Sections: []sectionView{
			buildSection(taskfile.StatusTodo, doc.Todo),
			buildSection(taskfile.StatusInProgress, doc.InProgress),
			buildSection(taskfile.StatusDone, doc.Done),
		},
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
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

func buildSection(status taskfile.Status, tasks []*taskfile.Task) sectionView {
	views := make([]taskView, 0, len(tasks))
	for _, task := range tasks {
		views = append(views, taskView{
			ID:         task.ID,
			Title:      task.Title,
			Priority:   strings.ToUpper(string(task.Priority)),
			PriorityUI: priorityClass(task.Priority),
			Labels:     task.Labels,
			CreatedAt:  task.CreatedAt.Format("2006-01-02 15:04"),
			UpdatedAt:  task.UpdatedAt.Format("2006-01-02 15:04"),
			NotesHTML:  renderMarkdown(task.Notes),
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
