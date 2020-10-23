package main

import (
	"github.com/spf13/cobra"
)

func newImportCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import",
		Short: `import a package from a git repository, chart or image`,
	}
	cmd.AddCommand(
		newImportHelmChartCommand(),
	)
	return cmd
}
