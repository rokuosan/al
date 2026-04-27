package model

import "testing"

func TestRegistryResolvePrefersWorkspace(t *testing.T) {
	workspace := []AliasEntry{
		{Name: "gs", Run: "git status --short"},
	}
	global := []AliasEntry{
		{Name: "gs", Run: "git status"},
		{Name: "gc", Run: "git commit"},
	}

	registry, err := NewRegistry(workspace, global)
	if err != nil {
		t.Fatalf("NewRegistry() error = %v", err)
	}

	got, ok := registry.Resolve("gs")
	if !ok {
		t.Fatal("Resolve(gs) ok = false, want true")
	}
	if got.Scope != ScopeWorkspace {
		t.Fatalf("Resolve(gs) scope = %q, want %q", got.Scope, ScopeWorkspace)
	}
	if got.Entry.Run != "git status --short" {
		t.Fatalf("Resolve(gs) run = %q, want %q", got.Entry.Run, "git status --short")
	}
}

func TestRegistryEntriesIncludeWorkspaceOverridesOnce(t *testing.T) {
	workspace := []AliasEntry{
		{Name: "gs", Run: "git status --short"},
	}
	global := []AliasEntry{
		{Name: "gs", Run: "git status"},
		{Name: "gc", Run: "git commit"},
	}

	registry, err := NewRegistry(workspace, global)
	if err != nil {
		t.Fatalf("NewRegistry() error = %v", err)
	}

	entries := registry.Entries()
	if len(entries) != 2 {
		t.Fatalf("len(Entries()) = %d, want 2", len(entries))
	}
}

func TestRegistryRejectsDuplicatesWithinScope(t *testing.T) {
	_, err := NewRegistry([]AliasEntry{
		{Name: "gs", Run: "git status"},
		{Name: "gs", Run: "git status --short"},
	}, nil)
	if err == nil {
		t.Fatal("NewRegistry() error = nil, want error")
	}
}
