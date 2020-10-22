package eval

import (
	"fmt"
	"net/url"
	"strings"

	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/engine"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/kio/kioutil"
	"sigs.k8s.io/kustomize/kyaml/yaml"

	"github.com/squaremo/spresm/pkg/spec"
)

// [DEBUG]
import (
	yamlv3 "gopkg.in/yaml.v3"
	"os"
)

// evalHelmChart evaluates a spec with the kind "HelmChart".
func evalHelmChart(s spec.Spec) ([]*yaml.RNode, error) {
	// The format expected here looks like a regular URL; everything
	// up to the last path element is taken as the repository URL, and
	// the last path element is taken as naming the chart.
	repoAndChartUrl := s.Source
	u, err := url.Parse(repoAndChartUrl)
	if err != nil {
		return nil, fmt.Errorf("unable to parse chart URL: %w", err)
	}
	pathElements := strings.Split(u.Path, "/")
	if len(pathElements) < 1 {
		return nil, fmt.Errorf("path of chart URL must include at least one element, naming the chart")
	}
	chartName := pathElements[len(pathElements)-1]
	u.Path = u.Path[:len(u.Path)-len(chartName)]

	// Now we expect the URL to be the Helm repository; do the magic to fetch the index.
	// TODO: Respect the local cache, rather than downloading every time.
	certFile, keyFile, caFile := "", "", ""
	providers := getter.Providers{
		getter.Provider{
			Schemes: []string{"http", "https"},
			New:     getter.NewHTTPGetter,
		},
	}
	// ^ cargo culted from fluxcd/source-controller, I can't find
	// where this is done in Helm itself.

	downloadUrl, err := repo.FindChartInRepoURL(u.String(), chartName, s.Version, certFile, keyFile, caFile, providers)
	if err != nil {
		return nil, fmt.Errorf("could not find chart download URL: %w", err)
	}

	get, err := providers.ByScheme(u.Scheme)
	if err != nil {
		return nil, fmt.Errorf("could not find how to download chart: %w", err)
	}

	buf, err := get.Get(downloadUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to download chart: %w", err)
	}
	chart, err := loader.LoadArchive(buf)
	if err != nil {
		return nil, fmt.Errorf("could not load downlaoded chart: %w", err)
	}

	// Finally we have an actual chart.
	// TODO use values, releaseOptions from the spec.
	values, err := chartutil.ToRenderValues(chart, chartutil.Values{}, chartutil.ReleaseOptions{}, nil)
	if err != nil {
		return nil, fmt.Errorf("could not create values for chart templates: %w", err)
	}

	// [DEBUG]
	yamlv3.NewEncoder(os.Stderr).Encode(values)

	rendered, err := engine.Render(chart, values)
	if err != nil {
		return nil, fmt.Errorf("failed to render chart: %w", err)
	}

	var result []*yaml.RNode
	for filename, src := range rendered {
		// probably fine hack: ignore anything that's not YAMLish
		if !(strings.HasSuffix(filename, ".yaml") || strings.HasSuffix(filename, ".yml")) {
			continue
		}
		br := kio.ByteReader{
			Reader: strings.NewReader(src),
			SetAnnotations: map[string]string{
				kioutil.PathAnnotation: filename,
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
