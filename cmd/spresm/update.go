package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"

	"github.com/squaremo/spresm/pkg/eval"
	"github.com/squaremo/spresm/pkg/merge"
	"github.com/squaremo/spresm/pkg/spec"
)

func newUpdateCommand() *cobra.Command {
	flags := &updateFlags{}
	cmd := &cobra.Command{
		Use:   "update <dir>",
		Short: `update the package in <dir> according to its spec file`,
		RunE:  flags.run,
	}
	flags.init(cmd)
	return cmd
}

type updateFlags struct {
	edit      bool   // edit the spec
	version   string // change the version
	overwrite bool   // overwrite the files in the local dir, rather than merging
	base      string // use this ref for the base revision when merging
}

func (flags *updateFlags) init(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&flags.edit, "edit", false, "present the package config for editing before updating")
	cmd.Flags().BoolVar(&flags.overwrite, "overwrite", false, "overwrite files rather than attempting a 3-way merge")
	cmd.Flags().StringVar(&flags.version, "version", "", "change the package version to this value")
	cmd.Flags().StringVar(&flags.base, "base", "HEAD", "use this git ref for the base revision when merging")
}

func (flags *updateFlags) run(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("update expected exactly one argument")
	}
	dir := args[0]

	// get the spec as it is in the file system
	updatedSpec, err := getSpec(dir)
	if err != nil {
		return err
	}

	writeBackSpec := false

	if flags.version != "" {
		writeBackSpec = true
		updatedSpec.Version = flags.version
	}

	// if --edit, extract the package config and present it for
	// editing.
	if flags.edit {
		writeBackSpec = true
		configReader, err := editConfig(updatedSpec.Config())
		if err != nil {
			return err
		}
		if err := updatedSpec.ReadConfig(configReader); err != nil {
			return fmt.Errorf("unable to re-read config after editing: %w", err)
		}
	}

	destRW := kio.LocalPackageReadWriter{
		PackagePath: dir,
		// Just to be explicit. If overwriting, we want to delete
		// files that no longer feature in the output. If merging,
		// we'll be deciding for each resource whether it stays or
		// goes in the merged results; so again, if there's nothing
		// left in a file it can be deleted.
		NoDeleteFiles: false,
	}
	dest, err := destRW.Read()
	if err != nil {
		return fmt.Errorf("could not parse local files: %w", err)
	}

	if flags.overwrite {
		updated, err := eval.Eval(updatedSpec)
		if err != nil {
			return fmt.Errorf("could not eval local spec: %w", err)
		}
		if err = destRW.Write(updated); err != nil {
			return fmt.Errorf("failed to write files back to directory %s: %w", dir, err)
		}
		// fall through to writing the spec back
	} else {

		// This will do a three way merge between:
		//  - the resources in the working directory (`dest`)
		//  - the resources as previously defined (`orig`)
		//  - the resources as defined by the updated spec (`updated`)

		// get the spec as it is in HEAD
		repo, err := git.PlainOpenWithOptions(dir, &git.PlainOpenOptions{
			DetectDotGit: true,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, `
If this is not a git repo, use --overwrite to overwrite
files rather than merging.
`)
			return fmt.Errorf("expected git repo at %s: %w", dir, err)
		}
		origSpec, err := getSpecFromGitRef(repo, flags.base, filepath.Join(dir, Spresmfile))
		if err != nil {
			fmt.Fprintf(os.Stderr, `
Ref %q does not exist; if there is no spec
committed, you can use --overwrite to overwrite
files rather than merging.
`)
			return fmt.Errorf("could not get spec from git repo ref %q: %w", flags.base, err)
		}

		updated, err := eval.Eval(updatedSpec)
		if err != nil {
			return fmt.Errorf("could not eval local spec: %w", err)
		}
		orig, err := eval.Eval(origSpec)
		if err != nil {
			return fmt.Errorf("could not eval base spec: %w", err)
		}

		merged, err := merge.Merge(dest, orig, updated)
		if err != nil {
			return err
		}
		if err = destRW.Write(merged); err != nil {
			return fmt.Errorf("failed to write merged files back to working directory: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Merged files written to %s\n", dir)
	}

	if writeBackSpec {
		if specPath, err := writeSpec(dir, updatedSpec); err != nil {
			return err
		} else {
			fmt.Fprintf(os.Stderr, "Updated spec file written to %s\n", specPath)
		}
	}
	return nil
}

func getSpecFromGitRef(repo *git.Repository, ref, path string) (spec.Spec, error) {
	var spec spec.Spec

	baseRef, err := repo.Reference(plumbing.ReferenceName(ref), true)
	if err != nil {
		return spec, fmt.Errorf("could not resolve base ref %q: %w", ref, err)
	}
	baseCommit, err := repo.CommitObject(baseRef.Hash())
	if err != nil {
		return spec, fmt.Errorf("could not get commit from HEAD ref: %w", err)
	}

	tree, err := baseCommit.Tree()
	if err != nil {
		return spec, fmt.Errorf("could not obtain tree for HEAD commit: %w", err)
	}

	// TODO does this resolve subtrees if the path has separators?
	specFile, err := tree.File(path)
	if err != nil {
		return spec, fmt.Errorf("could not find spec file in tree object: %w", err)
	}
	reader, err := specFile.Blob.Reader()
	if err != nil {
		return spec, fmt.Errorf("could not open spec file blob: %w")
	}
	defer reader.Close()

	if err = yaml.NewDecoder(reader).Decode(&spec); err != nil {
		return spec, fmt.Errorf("unable to decode spec file: %w", err)
	}
	return spec, nil
}
