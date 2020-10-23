package main

import (
	"bytes"
	"fmt"
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

func newImportCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import",
		Short: `import a package from a git repository, chart or image`,
	}
	cmd.AddCommand(
		newImportHelmChartCommand(),
	)
	return cmd
}

func newImportHelmChartCommand() *cobra.Command {
	flags := &importHelmChartFlags{}
	cmd := &cobra.Command{
		Use:   "helm <dir> --chart <chart URL> --version <version>",
		Short: `import a Helm chart as a package`,
		RunE:  flags.importHelmChart,
	}
	flags.init(cmd)
	return cmd
}

type importHelmChartFlags struct {
	chartURL, version string
}

func (flags *importHelmChartFlags) init(cmd *cobra.Command) {
	cmd.Flags().StringVar(&flags.chartURL, "chart", "", "URL for chart, including the repository; e.g., https://charts.fluxcd.io/flux")
	cmd.Flags().StringVar(&flags.version, "version", "", "version of chart to use")
}

func (flags *importHelmChartFlags) importHelmChart(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("expected exactly one argument, the directory in which to put the package files")
	}
	dir := args[0]
	if flags.chartURL == "" || flags.version == "" {
		return fmt.Errorf("need both chart URL (--chart) and version (--version) flags")
	}

	// TODO make sure the target directory doesn't exist, or at least
	// is empty (?)
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

	// create spec file
	var s spec.Spec
	s.Init(spec.ChartKind)
	s.Source = flags.chartURL
	s.Version = flags.version

	// get the chart
	chart, err := eval.ProcureChart(flags.chartURL, flags.version)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "chart found %q\n", chart.Name())

	// present the chart default values for editing
	tmpvalues, err := ioutil.TempFile("", "spresm-import")
	if err != nil {
		return fmt.Errorf("could not create temp file for editing values: %w", err)
	}
	defer os.Remove(tmpvalues.Name())

	if err := yaml.NewEncoder(tmpvalues).Encode(chart.Values); err != nil {
		return fmt.Errorf("failed to write values to file for editing: %w", err)
	}
	tmpvalues.Close()

	// FIXME deal with no env entry for EDITOR
	c := exec.Command("sh", "-c", "$EDITOR "+tmpvalues.Name())
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin
	fmt.Fprintf(os.Stderr, "opening values file %s for editing ...\n", tmpvalues.Name())
	if err := c.Run(); err != nil {
		return fmt.Errorf("error editing values: %w", err)
	}
	fmt.Fprintf(os.Stderr, "... done.\n")
	valuesBytes, err := ioutil.ReadFile(tmpvalues.Name())
	if err != nil {
		return fmt.Errorf("unable to re-read values file %s after editing: %w", tmpvalues.Name(), err)
	}

	s.Helm = &spec.HelmArgs{}
	if err := yaml.NewDecoder(bytes.NewBuffer(valuesBytes)).Decode(&s.Helm.Values); err != nil {
		return fmt.Errorf("unable to re-read values from file %s after editing: %w", tmpvalues.Name(), err)
	}
	s.Helm.Release.Name = filepath.Base(dir)

	// read the values and put into the spec (and respect them
	// when evaluating it)

	specPath := filepath.Join(dir, Spresmfile)
	buf := &bytes.Buffer{}
	err = yaml.NewEncoder(buf).Encode(s)
	if err == nil {
		err = ioutil.WriteFile(specPath, buf.Bytes(), os.FileMode(0600))
	}
	if err != nil {
		return fmt.Errorf("failed to encode and write spec to %s: %w", specPath, err)
	}
	fmt.Fprintf(os.Stderr, "spec file written to %s\n", specPath)

	// TODO eval (update, really) the spec, to render the chart into
	// the directory. This bit will be in common with other import
	// commands, so stick it in pkg somewhere.
	resources, err := eval.Eval(s)
	writer := kio.LocalPackageWriter{PackagePath: dir}
	if err := writer.Write(resources); err != nil {
		return fmt.Errorf("problem writing to the directory %s: %w", dir, err)
	}
	fmt.Fprintf(os.Stderr, "spec evaluated to %s\n", dir)
	return nil
}
