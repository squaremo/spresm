package eval

import (
	"fmt"
	"net/url"
	"strings"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
)

func ProcureChart(repoAndChartURL, version string) (*chart.Chart, error) {
	u, err := url.Parse(repoAndChartURL)
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

	downloadUrl, err := repo.FindChartInRepoURL(u.String(), chartName, version, certFile, keyFile, caFile, providers)
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
		return nil, fmt.Errorf("could not load downloaded chart archive: %w", err)
	}

	return chart, nil
}
