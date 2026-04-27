package model

import "fmt"

type Scope string

const (
	// ScopeWorkspace is the repository-local config layer.
	ScopeWorkspace Scope = "workspace"
	// ScopeGlobal is the user-level config layer.
	ScopeGlobal Scope = "global"
)

// RegisteredAlias is a resolved alias paired with its originating config scope.
type RegisteredAlias struct {
	Entry AliasEntry
	Scope Scope
}

// Registry stores aliases from each config layer and resolves them by precedence.
type Registry struct {
	workspace map[string]AliasEntry
	global    map[string]AliasEntry
}

func NewRegistry(workspace, global []AliasEntry) (*Registry, error) {
	r := &Registry{
		workspace: make(map[string]AliasEntry, len(workspace)),
		global:    make(map[string]AliasEntry, len(global)),
	}

	for _, entry := range workspace {
		if err := r.add(ScopeWorkspace, entry); err != nil {
			return nil, err
		}
	}
	for _, entry := range global {
		if err := r.add(ScopeGlobal, entry); err != nil {
			return nil, err
		}
	}

	return r, nil
}

func (r *Registry) add(scope Scope, entry AliasEntry) error {
	if err := entry.Validate(); err != nil {
		return err
	}

	var target map[string]AliasEntry
	switch scope {
	case ScopeWorkspace:
		target = r.workspace
	case ScopeGlobal:
		target = r.global
	default:
		return fmt.Errorf("unknown scope %q", scope)
	}

	if _, exists := target[entry.Name]; exists {
		return fmt.Errorf("duplicate %s alias %q", scope, entry.Name)
	}
	target[entry.Name] = entry
	return nil
}

func (r *Registry) Resolve(name string) (RegisteredAlias, bool) {
	if entry, ok := r.workspace[name]; ok {
		return RegisteredAlias{Entry: entry, Scope: ScopeWorkspace}, true
	}
	if entry, ok := r.global[name]; ok {
		return RegisteredAlias{Entry: entry, Scope: ScopeGlobal}, true
	}
	return RegisteredAlias{}, false
}

func (r *Registry) Entries() []RegisteredAlias {
	seen := make(map[string]struct{}, len(r.workspace)+len(r.global))
	entries := make([]RegisteredAlias, 0, len(r.workspace)+len(r.global))

	for _, entry := range r.workspace {
		entries = append(entries, RegisteredAlias{Entry: entry, Scope: ScopeWorkspace})
		seen[entry.Name] = struct{}{}
	}
	for _, entry := range r.global {
		if _, exists := seen[entry.Name]; exists {
			continue
		}
		entries = append(entries, RegisteredAlias{Entry: entry, Scope: ScopeGlobal})
	}

	return entries
}
