package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init <shell>",
		Short: "Print shell integration code",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("init is not implemented yet")
		},
	}
}
