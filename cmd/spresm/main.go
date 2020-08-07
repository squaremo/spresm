package main

import (
	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use:   "spresm",
		Short: `spresm build|spec`,
	}
	root.AddCommand(
		newBuildCommand(),
		newSpecCommand(),
	)
	root.Execute()
}

func newBuildCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "build",
		Short: `spresm build <dir>`,
	}
}

func newSpecCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "spec",
		Short: `spresm spec`,
	}
}
