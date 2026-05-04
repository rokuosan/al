package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/rokuosan/al/internal/config"
	"github.com/rokuosan/al/internal/model"
	"github.com/rokuosan/al/internal/runner"
)

func TestRunCmdRunsMatchingAlias(t *testing.T) {
	command := newRunCmd(config.StaticProvider{
		Configs: []config.LoadedConfig{
			{
				Scope: model.ScopeGlobal,
				Config: config.Config{
					Aliases: map[string]config.AliasConfig{
						"hello": {Run: "printf 'hello\\n'"},
					},
				},
			},
		},
	}, runner.Runner{})

	var stdout, stderr bytes.Buffer
	command.SetOut(&stdout)
	command.SetErr(&stderr)
	command.SetArgs([]string{"hello"})

	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if got := stdout.String(); got != "hello\n" {
		t.Fatalf("stdout = %q, want %q", got, "hello\n")
	}
}

func TestRunCmdRejectsDisabledAlias(t *testing.T) {
	command := newRunCmd(config.StaticProvider{
		Configs: []config.LoadedConfig{
			{
				Scope: model.ScopeGlobal,
				Config: config.Config{
					Aliases: map[string]config.AliasConfig{
						"hello": {
							Run: "printf 'hello\\n'",
							When: config.WhenConfig{
								Shell: []string{"bash"},
							},
						},
					},
				},
			},
		},
	}, runner.Runner{})

	command.SetArgs([]string{"hello"})

	err := command.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "disabled") {
		t.Fatalf("error = %v, want disabled error", err)
	}
}
