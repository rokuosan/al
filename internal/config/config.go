package config

import (
	"fmt"
	"maps"
	"slices"

	"github.com/rokuosan/al/internal/condition"
	"github.com/rokuosan/al/internal/model"
)

// Config is the file-format-agnostic top-level config shape.
type Config struct {
	Version  int
	Settings Settings
	Aliases  map[string]AliasConfig
}

// Settings holds global config behavior toggles.
type Settings struct {
	ConflictPolicy string
}

// AliasConfig is the config-layer representation of one alias definition.
type AliasConfig struct {
	Run         string
	Description string
	Mode        model.Mode
	Runtime     model.Runtime
	Shell       string
	Workdir     string
	Args        model.ArgsMode
	Override    bool
	When        WhenConfig
}

// WhenConfig is the config-layer representation of alias activation conditions.
type WhenConfig struct {
	Expr       string
	Git        bool
	Inside     []string
	Exists     []string
	ExistsAny  []string
	Env        map[string]string
	HasCommand []string
	OS         []string
	Shell      []string
}

// Empty reports whether the config contains no condition clauses.
func (w WhenConfig) Empty() bool {
	return w.Expr == "" &&
		!w.Git &&
		len(w.Inside) == 0 &&
		len(w.Exists) == 0 &&
		len(w.ExistsAny) == 0 &&
		len(w.Env) == 0 &&
		len(w.HasCommand) == 0 &&
		len(w.OS) == 0 &&
		len(w.Shell) == 0
}

// LoadedConfig associates a config value with its source metadata.
type LoadedConfig struct {
	Config Config
	Scope  model.Scope
	Path   string
}

// Provider returns config layers without constraining how they are sourced.
type Provider interface {
	Load() ([]LoadedConfig, error)
}

// StaticProvider is a temporary provider for hard-coded config during bring-up.
type StaticProvider struct {
	Configs []LoadedConfig
}

func (p StaticProvider) Load() ([]LoadedConfig, error) {
	return slices.Clone(p.Configs), nil
}

// Normalize converts a config layer into validated model entries.
func (c Config) Normalize(sourcePath string) ([]model.AliasEntry, error) {
	names := make([]string, 0, len(c.Aliases))
	for name := range c.Aliases {
		names = append(names, name)
	}
	slices.Sort(names)

	entries := make([]model.AliasEntry, 0, len(names))
	for _, name := range names {
		alias := c.Aliases[name]
		entry := model.AliasEntry{
			Name:        name,
			Run:         alias.Run,
			Description: alias.Description,
			Mode:        alias.Mode,
			Runtime:     alias.Runtime,
			Shell:       alias.Shell,
			Workdir:     alias.Workdir,
			Args:        alias.Args,
			Override:    alias.Override,
			Condition:   normalizeCondition(alias.When),
			SourcePath:  sourcePath,
		}
		if err := entry.Validate(); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

// BuildRegistry loads every config layer and normalizes them into a model registry.
func BuildRegistry(provider Provider) (*model.Registry, error) {
	loaded, err := provider.Load()
	if err != nil {
		return nil, err
	}

	var workspaceEntries []model.AliasEntry
	var globalEntries []model.AliasEntry

	for _, layer := range loaded {
		entries, err := layer.Config.Normalize(layer.Path)
		if err != nil {
			return nil, err
		}
		switch layer.Scope {
		case model.ScopeWorkspace:
			workspaceEntries = append(workspaceEntries, entries...)
		case model.ScopeGlobal:
			globalEntries = append(globalEntries, entries...)
		default:
			return nil, fmt.Errorf("unsupported config scope %q", layer.Scope)
		}
	}

	return model.NewRegistry(workspaceEntries, globalEntries)
}

func normalizeCondition(when WhenConfig) model.Condition {
	if when.Empty() {
		return model.TrueCondition{}
	}
	return condition.Table{
		Git:        when.Git,
		Inside:     slices.Clone(when.Inside),
		Exists:     slices.Clone(when.Exists),
		ExistsAny:  slices.Clone(when.ExistsAny),
		Env:        cloneEnv(when.Env),
		HasCommand: slices.Clone(when.HasCommand),
		OS:         slices.Clone(when.OS),
		Shell:      slices.Clone(when.Shell),
	}
}

func cloneEnv(env map[string]string) map[string]string {
	if env == nil {
		return nil
	}
	cloned := make(map[string]string, len(env))
	maps.Copy(cloned, env)
	return cloned
}
