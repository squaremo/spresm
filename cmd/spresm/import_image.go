package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

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
	s.Image = &spec.ImageArgs{}
	// In the absence of indications otherwise, use a ConfigMap. This
	// is the convention for kyaml functions.
	s.Image.FunctionConfig = map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "ConfigMap",
		"data":       map[string]string{},
	}

	valuesReader, err := editConfig(s.Image)
	if err != nil {
		return err
	}

	if err := yaml.NewDecoder(valuesReader).Decode(s.Image); err != nil {
		return fmt.Errorf("unable to re-read config after editing: %w", err)
	}

	return writePackage(dir, s)
}
