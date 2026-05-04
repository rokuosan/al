package model

import (
	"fmt"
	"maps"
	"path/filepath"
	"runtime"
)

type Mode string

const (
	// ModeAlias exposes an entry as a top-level shell alias/function.
	ModeAlias Mode = "alias"
	// ModeAbbr exposes an entry through shell abbreviation support when available.
	ModeAbbr Mode = "abbr"
	// ModeCommandOnly keeps an entry available via `al run` without top-level exposure.
	ModeCommandOnly Mode = "command-only"
)

func (m Mode) Validate() error {
	switch m {
	case ModeAlias, ModeAbbr, ModeCommandOnly:
		return nil
	default:
		return fmt.Errorf("invalid mode %q", m)
	}
}

type Runtime string

const (
	// RuntimeSubshell runs the command in a child shell.
	RuntimeSubshell Runtime = "subshell"
	// RuntimeCurrentShell runs the command in the caller's current shell context.
	RuntimeCurrentShell Runtime = "current-shell"
)

func (r Runtime) Validate() error {
	switch r {
	case RuntimeSubshell, RuntimeCurrentShell:
		return nil
	default:
		return fmt.Errorf("invalid runtime %q", r)
	}
}

type ArgsMode string

const (
	// ArgsAppend appends extra CLI args after the configured command.
	ArgsAppend ArgsMode = "append"
	// ArgsIgnore discards extra CLI args.
	ArgsIgnore ArgsMode = "ignore"
	// ArgsReject returns an error when extra CLI args are passed.
	ArgsReject ArgsMode = "reject"
)

func (m ArgsMode) Validate() error {
	switch m {
	case ArgsAppend, ArgsIgnore, ArgsReject:
		return nil
	default:
		return fmt.Errorf("invalid args mode %q", m)
	}
}

type Condition interface {
	Evaluate(EvalContext) (bool, error)
}

// TrueCondition is the default condition for always-enabled aliases.
type TrueCondition struct{}

func (TrueCondition) Evaluate(EvalContext) (bool, error) {
	return true, nil
}

// AliasEntry is the normalized internal representation of one alias definition.
type AliasEntry struct {
	Name        string
	Run         string
	Description string
	Mode        Mode
	Runtime     Runtime
	Shell       string
	Workdir     string
	Args        ArgsMode
	Override    bool
	Condition   Condition
	SourcePath  string
}

func (e AliasEntry) Validate() error {
	if e.Name == "" {
		return fmt.Errorf("alias name is required")
	}
	if e.Run == "" {
		return fmt.Errorf("alias %q: run is required", e.Name)
	}
	if err := e.Mode.WithDefault().Validate(); err != nil {
		return fmt.Errorf("alias %q: %w", e.Name, err)
	}
	if err := e.Runtime.WithDefault().Validate(); err != nil {
		return fmt.Errorf("alias %q: %w", e.Name, err)
	}
	if err := e.Args.WithDefault().Validate(); err != nil {
		return fmt.Errorf("alias %q: %w", e.Name, err)
	}
	return nil
}

func (e AliasEntry) ModeOrDefault() Mode {
	return e.Mode.WithDefault()
}

func (e AliasEntry) RuntimeOrDefault() Runtime {
	return e.Runtime.WithDefault()
}

func (e AliasEntry) ArgsOrDefault() ArgsMode {
	return e.Args.WithDefault()
}

func (e AliasEntry) ConditionOrDefault() Condition {
	if e.Condition == nil {
		return TrueCondition{}
	}
	return e.Condition
}

func (m Mode) WithDefault() Mode {
	if m == "" {
		return ModeAlias
	}
	return m
}

func (r Runtime) WithDefault() Runtime {
	if r == "" {
		return RuntimeSubshell
	}
	return r
}

func (m ArgsMode) WithDefault() ArgsMode {
	if m == "" {
		return ArgsAppend
	}
	return m
}

// EvalContext carries the runtime state used when evaluating alias conditions.
type EvalContext struct {
	WorkspaceRoot string
	CurrentDir    string
	Shell         string
	Env           map[string]string
	OS            string
	InGitRepo     bool
}

func NewEvalContext(workspaceRoot, currentDir, shell string, env map[string]string) EvalContext {
	return EvalContext{
		WorkspaceRoot: filepath.Clean(workspaceRoot),
		CurrentDir:    filepath.Clean(currentDir),
		Shell:         shell,
		Env:           cloneEnv(env),
		OS:            runtime.GOOS,
	}
}

func cloneEnv(env map[string]string) map[string]string {
	if env == nil {
		return map[string]string{}
	}
	cloned := make(map[string]string, len(env))
	maps.Copy(cloned, env)
	return cloned
}
