package cmd

import (
	"fmt"
	"os"

	"github.com/rokuosan/al/internal/config"
	"github.com/rokuosan/al/internal/model"
	"github.com/rokuosan/al/internal/runner"
	"github.com/spf13/cobra"
)

func NewRunCmd() *cobra.Command {
	return newRunCmd(config.StaticProvider{
		Configs: []config.LoadedConfig{
			{
				Scope: model.ScopeGlobal,
				Path:  "static",
				Config: config.Config{
					Aliases: map[string]config.AliasConfig{
						"hello": {
							Run:         "printf 'hello\\n'",
							Description: "Temporary hard-coded alias for bring-up",
						},
					},
				},
			},
		},
	}, runner.Runner{})
}

func newRunCmd(provider config.Provider, commandRunner runner.Runner) *cobra.Command {
	return &cobra.Command{
		Use:   "run <name> [args...]",
		Short: "Run an alias explicitly",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			registry, err := config.BuildRegistry(provider)
			if err != nil {
				return err
			}

			resolved, ok := registry.Resolve(args[0])
			if !ok {
				return fmt.Errorf("alias %q not found", args[0])
			}

			wd, err := os.Getwd()
			if err != nil {
				return err
			}

			evalCtx := model.NewEvalContext(wd, wd, runner.DefaultShellName(), envMap())
			enabled, err := resolved.Entry.ConditionOrDefault().Evaluate(evalCtx)
			if err != nil {
				return fmt.Errorf("evaluate alias %q: %w", resolved.Entry.Name, err)
			}
			if !enabled {
				return fmt.Errorf("alias %q is disabled in the current context", resolved.Entry.Name)
			}

			commandRunner.Stdout = cmd.OutOrStdout()
			commandRunner.Stderr = cmd.ErrOrStderr()
			return commandRunner.Run(resolved.Entry, args[1:])
		},
	}
}

func envMap() map[string]string {
	env := make(map[string]string)
	for _, pair := range os.Environ() {
		for i := 0; i < len(pair); i++ {
			if pair[i] != '=' {
				continue
			}
			env[pair[:i]] = pair[i+1:]
			break
		}
	}
	return env
}
