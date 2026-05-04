package condition

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/rokuosan/al/internal/model"
)

// Table is the table-form condition evaluator shared by config-backed aliases.
type Table struct {
	Git        bool
	Inside     []string
	Exists     []string
	ExistsAny  []string
	Env        map[string]string
	HasCommand []string
	OS         []string
	Shell      []string
}

func (t Table) Evaluate(ctx model.EvalContext) (bool, error) {
	if t.Git && !ctx.InGitRepo {
		return false, nil
	}
	if !matchInside(ctx, t.Inside) {
		return false, nil
	}
	ok, err := pathsExist(ctx.WorkspaceRoot, t.Exists)
	if err != nil || !ok {
		return ok, err
	}
	ok, err = anyPathExists(ctx.WorkspaceRoot, t.ExistsAny)
	if err != nil || !ok {
		return ok, err
	}
	if !matchEnv(ctx.Env, t.Env) {
		return false, nil
	}
	ok, err = commandsExist(t.HasCommand)
	if err != nil || !ok {
		return ok, err
	}
	if !matchOne(ctx.OS, t.OS) {
		return false, nil
	}
	if !matchOne(ctx.Shell, t.Shell) {
		return false, nil
	}
	return true, nil
}

func matchInside(ctx model.EvalContext, patterns []string) bool {
	if len(patterns) == 0 {
		return true
	}
	rel, err := filepath.Rel(ctx.WorkspaceRoot, ctx.CurrentDir)
	if err != nil {
		return false
	}
	rel = filepath.Clean(rel)
	for _, pattern := range patterns {
		pattern = filepath.Clean(pattern)
		if rel == pattern {
			return true
		}
		if strings.HasPrefix(rel, pattern+string(filepath.Separator)) {
			return true
		}
	}
	return false
}

func pathsExist(root string, paths []string) (bool, error) {
	for _, path := range paths {
		exists, err := pathExists(filepath.Join(root, path))
		if err != nil {
			return false, err
		}
		if !exists {
			return false, nil
		}
	}
	return true, nil
}

func anyPathExists(root string, paths []string) (bool, error) {
	if len(paths) == 0 {
		return true, nil
	}
	for _, path := range paths {
		exists, err := pathExists(filepath.Join(root, path))
		if err != nil {
			return false, err
		}
		if exists {
			return true, nil
		}
	}
	return false, nil
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errorsIsNotExist(err) {
		return false, nil
	}
	return false, fmt.Errorf("stat %q: %w", path, err)
}

func matchEnv(env, required map[string]string) bool {
	for key, want := range required {
		if got := env[key]; got != want {
			return false
		}
	}
	return true
}

func commandsExist(commands []string) (bool, error) {
	for _, command := range commands {
		if _, err := exec.LookPath(command); err != nil {
			if errorsIsNotExist(err) {
				return false, nil
			}
			return false, fmt.Errorf("look up command %q: %w", command, err)
		}
	}
	return true, nil
}

func matchOne(actual string, allowed []string) bool {
	if len(allowed) == 0 {
		return true
	}
	return slices.Contains(allowed, actual)
}

func errorsIsNotExist(err error) bool {
	return errors.Is(err, fs.ErrNotExist) || errors.Is(err, exec.ErrNotFound)
}
