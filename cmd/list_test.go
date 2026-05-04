package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/rokuosan/al/internal/config"
	"github.com/rokuosan/al/internal/model"
)

func TestListCmdPrintsAliases(t *testing.T) {
	command := newListCmd(config.StaticProvider{
		Configs: []config.LoadedConfig{
			{
				Scope: model.ScopeGlobal,
				Path:  "static",
				Config: config.Config{
					Aliases: map[string]config.AliasConfig{
						"hello": {
							Run:         "printf 'hello\\n'",
							Description: "Says hello",
						},
					},
				},
			},
		},
	})

	var stdout, stderr bytes.Buffer
	command.SetOut(&stdout)
	command.SetErr(&stderr)

	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	out := stdout.String()
	for _, want := range []string{"NAME", "hello", "alias", "true", "static", "Says hello"} {
		if !strings.Contains(out, want) {
			t.Fatalf("stdout = %q, want substring %q", out, want)
		}
	}
}

func TestListCmdMarksDisabledAliases(t *testing.T) {
	command := newListCmd(config.StaticProvider{
		Configs: []config.LoadedConfig{
			{
				Scope: model.ScopeGlobal,
				Path:  "static",
				Config: config.Config{
					Aliases: map[string]config.AliasConfig{
						"hello": {
							Run: "printf 'hello\\n'",
							When: config.WhenConfig{
								Shell: []string{"definitely-not-the-current-shell"},
							},
						},
					},
				},
			},
		},
	})

	var stdout bytes.Buffer
	command.SetOut(&stdout)

	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(stdout.String(), "false") {
		t.Fatalf("stdout = %q, want disabled marker", stdout.String())
	}
}
