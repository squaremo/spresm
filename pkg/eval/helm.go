package eval

import (
	"fmt"
	"path/filepath"
	"strings"

	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/engine"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/kio/kioutil"
	"sigs.k8s.io/kustomize/kyaml/yaml"

	"github.com/squaremo/spresm/pkg/spec"
)

// evalHelmChart evaluates a spec with the kind "HelmChart".
func evalHelmChart(s spec.Spec) ([]*yaml.RNode, error) {
	// The format expected here looks like a regular URL; everything
	// up to the last path element is taken as the repository URL, and
	// the last path element is taken as naming the chart.
	repoAndChartURL := s.Source
	chart, err := ProcureChart(repoAndChartURL, s.Version)
	if err != nil {
		return nil, err
	}

	// Finally we have an actual chart.
	// TODO use values, releaseOptions from the spec.
	values, err := chartutil.ToRenderValues(chart, chartutil.Values{}, chartutil.ReleaseOptions{}, nil)
	if err != nil {
		return nil, fmt.Errorf("could not create values for chart templates: %w", err)
	}

	rendered, err := engine.Render(chart, values)
	if err != nil {
		return nil, fmt.Errorf("failed to render chart: %w", err)
	}

	var result []*yaml.RNode
	basepath := filepath.Join(chart.Name(), "templates")
	for filename, src := range rendered {
		// probably fine hack: ignore anything that's not YAMLish
		if !(strings.HasSuffix(filename, ".yaml") || strings.HasSuffix(filename, ".yml")) {
			continue
		}

		// if this fails, just use the path as-is
		path, err := filepath.Rel(basepath, filename)
		if err != nil {
			path = filename
		}
		br := kio.ByteReader{
			Reader: strings.NewReader(src),
			SetAnnotations: map[string]string{
				kioutil.PathAnnotation: path,
			},
		}
		resources, err := br.Read()
		if err != nil {
			return nil, fmt.Errorf("could not parse output of template %q: %w", filename, err)
		}
		result = append(result, resources...)
	}
	return result, nil
}
