package cmd

import (
	"fmt"

	"github.com/rokuosan/al/internal/config"
	"github.com/rokuosan/al/internal/runner"
	"github.com/spf13/cobra"
)

func NewRunCmd() *cobra.Command {
	return newRunCmd(staticProvider(), runner.Runner{})
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

			evalCtx, err := buildEvalContext()
			if err != nil {
				return err
			}
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
