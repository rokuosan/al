package runner

import (
	"bytes"
	"strings"
	"testing"

	"github.com/rokuosan/al/internal/model"
)

func TestRunAppendsArgs(t *testing.T) {
	var stdout bytes.Buffer

	r := Runner{Stdout: &stdout, Stderr: &stdout}
	entry := model.AliasEntry{
		Name: "hello",
		Run:  "echo hello world",
	}

	if err := r.Run(entry, []string{"again"}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if got := stdout.String(); got != "hello world again\n" {
		t.Fatalf("stdout = %q, want %q", got, "hello world again\n")
	}
}

func TestRunRejectsCurrentShellRuntime(t *testing.T) {
	r := Runner{}
	entry := model.AliasEntry{
		Name:    "jump",
		Run:     "cd /tmp",
		Runtime: model.RuntimeCurrentShell,
	}

	err := r.Run(entry, nil)
	if err == nil {
		t.Fatal("Run() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "current-shell runtime") {
		t.Fatalf("error = %v, want current-shell runtime error", err)
	}
}

func TestRunRejectsExtraArgs(t *testing.T) {
	r := Runner{}
	entry := model.AliasEntry{
		Name: "hello",
		Run:  "printf hello",
		Args: model.ArgsReject,
	}

	err := r.Run(entry, []string{"again"})
	if err == nil {
		t.Fatal("Run() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "does not accept extra args") {
		t.Fatalf("error = %v, want args rejection", err)
	}
}
