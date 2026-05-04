package runner

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/rokuosan/al/internal/model"
)

// Runner executes normalized alias entries.
type Runner struct {
	Stdout io.Writer
	Stderr io.Writer
}

// Run evaluates argument handling and executes a matching alias entry.
func (r Runner) Run(entry model.AliasEntry, args []string) error {
	if entry.RuntimeOrDefault() == model.RuntimeCurrentShell {
		return fmt.Errorf("alias %q requires current-shell runtime and cannot be run via `al run` yet", entry.Name)
	}

	command, err := buildCommand(entry, args)
	if err != nil {
		return err
	}

	shell := resolveShell(entry)
	cmd := exec.Command(shell, "-c", command)
	cmd.Stdout = r.Stdout
	cmd.Stderr = r.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = os.Environ()

	if entry.Workdir != "" {
		cmd.Dir = entry.Workdir
	}

	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func buildCommand(entry model.AliasEntry, args []string) (string, error) {
	switch entry.ArgsOrDefault() {
	case model.ArgsAppend:
		if len(args) == 0 {
			return entry.Run, nil
		}
		quoted := make([]string, 0, len(args))
		for _, arg := range args {
			quoted = append(quoted, shellQuote(arg))
		}
		return entry.Run + " " + strings.Join(quoted, " "), nil
	case model.ArgsIgnore:
		return entry.Run, nil
	case model.ArgsReject:
		if len(args) > 0 {
			return "", fmt.Errorf("alias %q does not accept extra args", entry.Name)
		}
		return entry.Run, nil
	default:
		return "", fmt.Errorf("alias %q has unsupported args mode %q", entry.Name, entry.Args)
	}
}

func resolveShell(entry model.AliasEntry) string {
	if entry.Shell != "" {
		return entry.Shell
	}
	if shell := os.Getenv("SHELL"); shell != "" {
		return shell
	}
	return "sh"
}

func shellQuote(arg string) string {
	if arg == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(arg, "'", `'\''`) + "'"
}

// DefaultShellName returns the current shell name for condition evaluation.
func DefaultShellName() string {
	shell := os.Getenv("SHELL")
	if shell == "" {
		return ""
	}
	return filepath.Base(shell)
}
