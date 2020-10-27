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
