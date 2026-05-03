package config

import (
	"testing"

	"github.com/rokuosan/al/internal/model"
)

func TestConfigNormalizeAppliesDefaultsAndSourcePath(t *testing.T) {
	cfg := Config{
		Aliases: map[string]AliasConfig{
			"gs": {Run: "git status --short"},
		},
	}

	entries, err := cfg.Normalize(".al.toml")
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("len(entries) = %d, want 1", len(entries))
	}

	entry := entries[0]
	if entry.Name != "gs" {
		t.Fatalf("entry.Name = %q, want %q", entry.Name, "gs")
	}
	if entry.SourcePath != ".al.toml" {
		t.Fatalf("entry.SourcePath = %q, want %q", entry.SourcePath, ".al.toml")
	}
	if entry.ModeOrDefault() != model.ModeAlias {
		t.Fatalf("entry.ModeOrDefault() = %q, want %q", entry.ModeOrDefault(), model.ModeAlias)
	}
	if ok, err := entry.ConditionOrDefault().Evaluate(model.EvalContext{}); err != nil || !ok {
		t.Fatalf("entry.ConditionOrDefault().Evaluate() = (%v, %v), want (true, nil)", ok, err)
	}
}

func TestConfigNormalizeUsesDeferredConditionForNonEmptyWhen(t *testing.T) {
	cfg := Config{
		Aliases: map[string]AliasConfig{
			"web": {
				Run: "pnpm dev",
				When: WhenConfig{
					Inside: []string{"apps/web"},
				},
			},
		},
	}

	entries, err := cfg.Normalize(".al.toml")
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}

	if _, ok := entries[0].Condition.(DeferredCondition); !ok {
		t.Fatalf("entries[0].Condition = %T, want DeferredCondition", entries[0].Condition)
	}
}

func TestBuildRegistryPrefersWorkspaceLayer(t *testing.T) {
	registry, err := BuildRegistry(StaticProvider{
		Configs: []LoadedConfig{
			{
				Scope: model.ScopeGlobal,
				Path:  "/Users/me/.config/al/config.toml",
				Config: Config{
					Aliases: map[string]AliasConfig{
						"gs": {Run: "git status"},
					},
				},
			},
			{
				Scope: model.ScopeWorkspace,
				Path:  "/repo/.al.toml",
				Config: Config{
					Aliases: map[string]AliasConfig{
						"gs": {Run: "git status --short"},
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("BuildRegistry() error = %v", err)
	}

	got, ok := registry.Resolve("gs")
	if !ok {
		t.Fatal("Resolve(gs) ok = false, want true")
	}
	if got.Scope != model.ScopeWorkspace {
		t.Fatalf("Resolve(gs) scope = %q, want %q", got.Scope, model.ScopeWorkspace)
	}
	if got.Entry.Run != "git status --short" {
		t.Fatalf("Resolve(gs) run = %q, want %q", got.Entry.Run, "git status --short")
	}
}

func TestBuildRegistryRejectsUnknownScope(t *testing.T) {
	_, err := BuildRegistry(StaticProvider{
		Configs: []LoadedConfig{
			{
				Scope: "custom",
				Config: Config{
					Aliases: map[string]AliasConfig{
						"gs": {Run: "git status"},
					},
				},
			},
		},
	})
	if err == nil {
		t.Fatal("BuildRegistry() error = nil, want error")
	}
}
