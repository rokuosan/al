package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Inspect configuration and environment",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("doctor is not implemented yet")
		},
	}
}
