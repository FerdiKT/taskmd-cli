package preview

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ferdikt/taskmd-cli/internal/service"
)

func TestPreviewRendersHTMLAndAPI(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "docs", "Task.md")

	svc := service.New()
	if err := svc.Init(path, false); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Add(path, service.AddInput{
		Title:    "Initialize parser",
		Priority: "p1",
		Assignee: "main-agent",
		Labels:   []string{"cli", "preview"},
		Notes:    "Render **Markdown** notes.",
	}); err != nil {
		t.Fatal(err)
	}

	server := New(svc, path)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	for _, fragment := range []string{"Initialize parser", "Projects", "Search issues", "/api/document", "List", "main-agent"} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("expected HTML to contain %q", fragment)
		}
	}

	filteredReq := httptest.NewRequest(http.MethodGet, "/?assignee=other-agent", nil)
	filteredRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(filteredRec, filteredReq)
	if filteredRec.Code != http.StatusOK {
		t.Fatalf("expected filtered view 200, got %d", filteredRec.Code)
	}
	if !strings.Contains(filteredRec.Body.String(), "0 / 1 issues") {
		t.Fatalf("expected filtered issue count, got %s", filteredRec.Body.String())
	}

	detailReq := httptest.NewRequest(http.MethodGet, "/?issue=T001", nil)
	detailRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(detailRec, detailReq)
	if detailRec.Code != http.StatusOK {
		t.Fatalf("expected detail view 200, got %d", detailRec.Code)
	}
	if !strings.Contains(detailRec.Body.String(), "Description") || !strings.Contains(detailRec.Body.String(), "Created") {
		t.Fatalf("expected detail drawer content, got %s", detailRec.Body.String())
	}

	boardReq := httptest.NewRequest(http.MethodGet, "/?view=board", nil)
	boardRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(boardRec, boardReq)
	if boardRec.Code != http.StatusOK {
		t.Fatalf("expected board view 200, got %d", boardRec.Code)
	}
	if !strings.Contains(boardRec.Body.String(), "Kanban board grouped by status") {
		t.Fatalf("expected board view copy, got %s", boardRec.Body.String())
	}

	apiReq := httptest.NewRequest(http.MethodGet, "/api/document", nil)
	apiRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(apiRec, apiReq)
	if apiRec.Code != http.StatusOK {
		t.Fatalf("expected API 200, got %d", apiRec.Code)
	}
	if !strings.Contains(apiRec.Body.String(), `"title": "Initialize parser"`) {
		t.Fatalf("expected API body to include task JSON, got %s", apiRec.Body.String())
	}

	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
}
