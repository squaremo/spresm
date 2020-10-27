package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/squaremo/spresm/pkg/eval"
	"github.com/squaremo/spresm/pkg/spec"
)

func newImportHelmChartCommand() *cobra.Command {
	flags := &importHelmChartFlags{}
	cmd := &cobra.Command{
		Use:   "helm <dir> --chart <chart URL> --version <version>",
		Short: `import a Helm chart as a package`,
		RunE:  flags.run,
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

func (flags *importHelmChartFlags) run(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("expected exactly one argument, the directory in which to put the package files")
	}
	dir := args[0]
	if flags.chartURL == "" || flags.version == "" {
		return fmt.Errorf("need both chart URL (--chart) and version (--version) flags")
	}

	if err := ensurePackageDirectory(dir); err != nil {
		return err
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

	s.Helm = &spec.HelmArgs{Values: chart.Values}
	s.Helm.Release.Name = filepath.Base(dir)

	valuesReader, err := editConfig(s.Helm)
	if err != nil {
		return err
	}

	if err := yaml.NewDecoder(valuesReader).Decode(s.Helm); err != nil {
		return fmt.Errorf("unable to re-read config after editing: %w", err)
	}

	return writePackage(dir, s)
}
