package model

import "testing"

func TestAliasEntryDefaultsAndValidation(t *testing.T) {
	entry := AliasEntry{
		Name: "gs",
		Run:  "git status --short",
	}

	if err := entry.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if got := entry.ModeOrDefault(); got != ModeAlias {
		t.Fatalf("ModeOrDefault() = %q, want %q", got, ModeAlias)
	}
	if got := entry.RuntimeOrDefault(); got != RuntimeSubshell {
		t.Fatalf("RuntimeOrDefault() = %q, want %q", got, RuntimeSubshell)
	}
	if got := entry.ArgsOrDefault(); got != ArgsAppend {
		t.Fatalf("ArgsOrDefault() = %q, want %q", got, ArgsAppend)
	}
	if ok, err := entry.ConditionOrDefault().Evaluate(EvalContext{}); err != nil || !ok {
		t.Fatalf("ConditionOrDefault().Evaluate() = (%v, %v), want (true, nil)", ok, err)
	}
}

func TestAliasEntryValidateRejectsInvalidEnums(t *testing.T) {
	entry := AliasEntry{
		Name:    "gs",
		Run:     "git status --short",
		Mode:    "nope",
		Runtime: RuntimeSubshell,
		Args:    ArgsAppend,
	}

	if err := entry.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want error")
	}
}

func TestNewEvalContextClonesEnv(t *testing.T) {
	env := map[string]string{"FOO": "bar"}
	ctx := NewEvalContext("/tmp/work", "/tmp/work/app", "zsh", env)

	env["FOO"] = "baz"

	if got := ctx.Env["FOO"]; got != "bar" {
		t.Fatalf("ctx.Env[FOO] = %q, want %q", got, "bar")
	}
	if got := ctx.WorkspaceRoot; got != "/tmp/work" {
		t.Fatalf("WorkspaceRoot = %q, want %q", got, "/tmp/work")
	}
}
