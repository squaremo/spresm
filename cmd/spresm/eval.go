package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"sigs.k8s.io/kustomize/kyaml/kio"

	"github.com/squaremo/spresm/pkg/eval"
	"github.com/squaremo/spresm/pkg/spec"
)

// TODO move to somewhere else
const Spresmfile = "Spresmfile"

func evalCmd(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("eval expects exactly one argument")
	}
	dir := args[0]

	spec, err := getSpec(dir)
	if err != nil {
		return err
	}

	nodes, err := eval.Eval(spec)
	if err != nil {
		return fmt.Errorf("unable to evaluate spec file in %s: %w", dir, err)
	}

	// FIXME print out for now
	out := &kio.ByteWriter{Writer: os.Stdout}
	if err := out.Write(nodes); err != nil {
		return err
	}

	return nil
}

func getSpec(dir string) (spec.Spec, error) {
	var spec spec.Spec

	specpath := filepath.Join(dir, Spresmfile)
	specfile, err := os.Open(specpath)
	if err != nil {
		return spec, fmt.Errorf("expected to find spec file %s: %w", specpath, err)
	}
	defer specfile.Close()

	if err := yaml.NewDecoder(specfile).Decode(&spec); err != nil {
		return spec, fmt.Errorf("unable to decode spec file at %s: %w", specpath, err)
	}
	return spec, nil
}
