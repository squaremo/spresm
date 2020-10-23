package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/squaremo/spresm/pkg/spec"
)

func newImportImageCommand() *cobra.Command {
	flags := &importImageFlags{}
	cmd := &cobra.Command{
		Use:   "image <dir> --image",
		Short: `import a container image as a package`,
		RunE:  flags.importImage,
	}
	flags.init(cmd)
	return cmd
}

type importImageFlags struct {
	image, tag string
}

func (flags *importImageFlags) init(cmd *cobra.Command) {
	cmd.Flags().StringVar(&flags.image, "image", "", "image repo (not including tag) for container to import")
	cmd.Flags().StringVar(&flags.tag, "tag", "", "tag of image for container to import")
}

func (flags *importImageFlags) importImage(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("expected exactly one argument, the directory in which to put the package files")
	}
	dir := args[0]
	if flags.image == "" {
		return fmt.Errorf("need an image ref (supply this with --image)")
	}

	if err := ensurePackageDirectory(dir); err != nil {
		return nil
	}

	// create spec file
	var s spec.Spec
	s.Init(spec.ImageKind)
	s.Source = flags.image
	s.Version = flags.tag

	// TODO figure out some scheme for giving the user a default
	// function config to edit

	return writePackage(dir, s)
}
