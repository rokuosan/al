package condition

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/rokuosan/al/internal/model"
)

func TestTableEvaluateMatchesAllClauses(t *testing.T) {
	root := t.TempDir()
	appDir := filepath.Join(root, "apps", "web")
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "package.json"), []byte("{}"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	cond := Table{
		Git:        true,
		Inside:     []string{"apps/web"},
		Exists:     []string{"package.json"},
		ExistsAny:  []string{"missing.txt", "package.json"},
		Env:        map[string]string{"APP_ENV": "dev"},
		HasCommand: []string{"sh"},
		OS:         []string{runtime.GOOS},
		Shell:      []string{"zsh"},
	}

	ok, err := cond.Evaluate(model.EvalContext{
		WorkspaceRoot: root,
		CurrentDir:    appDir,
		Shell:         "zsh",
		Env:           map[string]string{"APP_ENV": "dev"},
		OS:            runtime.GOOS,
		InGitRepo:     true,
	})
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if !ok {
		t.Fatal("Evaluate() = false, want true")
	}
}

func TestTableEvaluateReturnsFalseWhenClauseDoesNotMatch(t *testing.T) {
	cond := Table{
		Shell: []string{"bash"},
	}

	ok, err := cond.Evaluate(model.EvalContext{Shell: "zsh"})
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if ok {
		t.Fatal("Evaluate() = true, want false")
	}
}

func TestTableEvaluateInsideMatchesDescendantPath(t *testing.T) {
	ok, err := Table{Inside: []string{"apps/web"}}.Evaluate(model.EvalContext{
		WorkspaceRoot: "/repo",
		CurrentDir:    "/repo/apps/web/src",
	})
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if !ok {
		t.Fatal("Evaluate() = false, want true")
	}
}
