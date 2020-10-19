package main

import (
	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use:   "spresm",
		Short: `spresm import|update|eval|build`,
	}
	root.AddCommand(
		newImportCommand(),
		newUpdateCommand(),
		newEvalCommand(),
		newBuildCommand(),
	)
	root.Execute()
}

func newImportCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "import",
		Short: `import a package from a git repository, chart or image`,
	}
}

func newUpdateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "update <dir>",
		Short: `update the package in <dir> according to its spec file`,
		RunE:  updateCmd,
	}
}

func newEvalCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "eval <dir>",
		Short: `evaluate the spec file in <dir> and show the output`,
		RunE:  evalCmd,
	}
}

func newBuildCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "build",
		Short: `create an image that can be used as a package`,
	}
}
