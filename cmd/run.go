package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewRunCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "run <name> [args...]",
		Short: "Run an alias explicitly",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("run is not implemented yet")
		},
	}
}
