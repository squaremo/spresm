package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"sigs.k8s.io/kustomize/kyaml/kio"

	"github.com/squaremo/spresm/pkg/eval"
	"github.com/squaremo/spresm/pkg/spec"
)

const (
	defaultEditor = "vi"
)

func newImportCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import",
		Short: `import a package from a git repository, chart or image`,
	}
	cmd.AddCommand(
		newImportHelmChartCommand(),
		newImportImageCommand(),
	)
	return cmd
}

func ensurePackageDirectory(dir string) error {
	dirstat, err := os.Stat(dir)
	switch {
	case os.IsNotExist(err):
		if err := os.MkdirAll(dir, os.FileMode(0777)); err != nil {
			return fmt.Errorf("failed to create directory %s for import: %w", dir, err)
		}
	case err != nil:
		return fmt.Errorf("error trying to establish import directory %s: %w", dir, err)
	case !dirstat.IsDir():
		return fmt.Errorf("expected %s to be a directory or not exist yet")
	default:
		// exists already, and is a directory
		d, err := os.Open(dir)
		if err != nil {
			return fmt.Errorf("error reading contents of %s: %w", dir, err)
		}
		contents, err := d.Readdir(0)
		if err != nil {
			return fmt.Errorf("error reading contents of %s: %w", dir, err)
		}
		if len(contents) > 0 {
			return fmt.Errorf("directory %s exists but is not empty; cannot import into a directory which already has files", dir)
		}
	}
	return nil
}

func writeSpec(dir string, s spec.Spec) (string, error) {
	specPath := filepath.Join(dir, Spresmfile)
	buf := &bytes.Buffer{}
	err := yaml.NewEncoder(buf).Encode(s)
	if err == nil {
		err = ioutil.WriteFile(specPath, buf.Bytes(), os.FileMode(0600))
	}
	if err != nil {
		return "", fmt.Errorf("failed to encode and write spec to %s: %w", specPath, err)
	}
	return specPath, nil
}

func writePackage(dir string, s spec.Spec) error {
	specPath, err := writeSpec(dir, s)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "spec file written to %s\n", specPath)

	// eval the spec, to render the chart into the directory. TODO
	// stick it in pkg somewhere.
	resources, err := eval.Eval(s)
	writer := kio.LocalPackageWriter{PackagePath: dir}
	if err := writer.Write(resources); err != nil {
		return fmt.Errorf("problem writing to the directory %s/: %w", dir, err)
	}
	fmt.Fprintf(os.Stderr, "spec evaluated to %s/\n", dir)
	return nil
}

func editConfig(initialConfig interface{}) (io.Reader, error) {
	// present the chart default values for editing
	tmpvalues, err := ioutil.TempFile("", "spresm-edit")
	if err != nil {
		return nil, fmt.Errorf("could not create temp file for editing config: %w", err)
	}
	defer os.Remove(tmpvalues.Name())

	if err := yaml.NewEncoder(tmpvalues).Encode(initialConfig); err != nil {
		return nil, fmt.Errorf("failed to write config to file for editing: %w", err)
	}
	tmpvalues.Close()

	c := exec.Command("sh", "-c", editCommand(tmpvalues.Name()))
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin
	fmt.Fprintf(os.Stderr, "opening config file %s for editing ...\n", tmpvalues.Name())
	if err := c.Run(); err != nil {
		return nil, fmt.Errorf("error editing config: %w", err)
	}
	fmt.Fprintf(os.Stderr, "... done.\n")
	valuesBytes, err := ioutil.ReadFile(tmpvalues.Name())
	if err != nil {
		return nil, fmt.Errorf("unable to re-read config file %s after editing: %w", tmpvalues.Name(), err)
	}
	return bytes.NewBuffer(valuesBytes), nil
}

func editCommand(s string) string {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = defaultEditor
	}
	return editor + " " + s
}
