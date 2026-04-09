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
	for _, fragment := range []string{"Initialize parser", "taskmd preview", "/api/document"} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("expected HTML to contain %q", fragment)
		}
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
