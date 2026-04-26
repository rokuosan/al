package cmd

import (
	"io"

	"github.com/spf13/cobra"
)

func NewRootCmd(stdout, stderr io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "al",
		Short:         "Contextual aliases for your shell",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)

	cmd.AddCommand(
		NewRunCmd(),
		NewListCmd(),
		NewDoctorCmd(),
		NewInitCmd(),
	)

	return cmd
}
