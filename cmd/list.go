package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available aliases",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("list is not implemented yet")
		},
	}
}
