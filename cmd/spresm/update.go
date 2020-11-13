package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
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
}

func (flags *updateFlags) init(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&flags.edit, "edit", false, "present the package config for editing before updating")
	cmd.Flags().BoolVar(&flags.overwrite, "overwrite", false, "overwrite files rather than attempting a 3-way merge")
	cmd.Flags().StringVar(&flags.version, "version", "", "change the package version to this value")
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
		config := updatedSpec.Config()
		configReader, err := editConfig(config)
		if err != nil {
			return err
		}
		if err := yaml.NewDecoder(configReader).Decode(config); err != nil {
			return fmt.Errorf("unable to re-read config after editing: %w", err)
		}
	}

	if flags.overwrite {
		updated, err := eval.Eval(updatedSpec)
		if err != nil {
			return fmt.Errorf("could not eval local spec: %w", err)
		}
		destW := kio.LocalPackageReadWriter{
			PackagePath:   dir,
			NoDeleteFiles: false, // just to be explicit
		}
		if err = destW.Write(updated); err != nil {
			return fmt.Errorf("failed to write files back to directory %s: %w", dir, err)
		}

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
			return fmt.Errorf("expected git repo at %s: %w", dir, err)
		}
		origSpec, err := getSpecFromGitHead(repo, filepath.Join(dir, Spresmfile))
		if err != nil {
			return fmt.Errorf("could not get spec from git repo: %w", err)
		}

		updated, err := eval.Eval(updatedSpec)
		if err != nil {
			return fmt.Errorf("could not eval local spec: %w", err)
		}
		orig, err := eval.Eval(origSpec)
		if err != nil {
			return fmt.Errorf("could not eval local spec: %w", err)
		}

		destRW := kio.LocalPackageReadWriter{
			PackagePath: dir,
		}
		dest, err := destRW.Read()
		if err != nil {
			return fmt.Errorf("could not parse local files: %w", err)
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

func getSpecFromGitHead(repo *git.Repository, path string) (spec.Spec, error) {
	var spec spec.Spec

	headRef, err := repo.Head()
	if err != nil {
		return spec, fmt.Errorf("could not reasolve HEAD ref: %w", err)
	}
	headCommit, err := repo.CommitObject(headRef.Hash())
	if err != nil {
		return spec, fmt.Errorf("could not get commit from HEAD ref: %w", err)
	}

	tree, err := headCommit.Tree()
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
