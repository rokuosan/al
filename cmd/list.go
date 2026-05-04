package cmd

import (
	"fmt"
	"text/tabwriter"

	"github.com/rokuosan/al/internal/config"
	"github.com/spf13/cobra"
)

func NewListCmd() *cobra.Command {
	return newListCmd(staticProvider())
}

func newListCmd(provider config.Provider) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available aliases",
		RunE: func(cmd *cobra.Command, args []string) error {
			registry, err := config.BuildRegistry(provider)
			if err != nil {
				return err
			}

			evalCtx, err := buildEvalContext()
			if err != nil {
				return err
			}

			tw := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 4, 2, ' ', 0)
			if _, err := fmt.Fprintln(tw, "NAME\tMODE\tENABLED\tSOURCE\tDESCRIPTION"); err != nil {
				return err
			}

			for _, alias := range registry.Entries() {
				enabled, err := alias.Entry.ConditionOrDefault().Evaluate(evalCtx)
				if err != nil {
					return fmt.Errorf("evaluate alias %q: %w", alias.Entry.Name, err)
				}
				if _, err := fmt.Fprintf(
					tw,
					"%s\t%s\t%t\t%s\t%s\n",
					alias.Entry.Name,
					alias.Entry.ModeOrDefault(),
					enabled,
					alias.Entry.SourcePath,
					alias.Entry.Description,
				); err != nil {
					return err
				}
			}

			return tw.Flush()
		},
	}
}
