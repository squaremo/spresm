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
		Use:   "update",
		Short: `update a previously imported package from its upstream`,
	}
}

func newEvalCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "eval",
		Short: `(re)evaluate a previously imported package`,
	}
}

func newBuildCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "build",
		Short: `create an image that can be used as a package`,
	}
}
