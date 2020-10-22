package eval

import (
	"errors"

	"sigs.k8s.io/kustomize/kyaml/yaml"

	"github.com/squaremo/spresm/pkg/spec"
)

var ErrNotImplemented = errors.New("not implemented")

// Eval takes a spec and runs it, to produce the YAML output. The
// output is in a kyaml/kio collection, so that it can be output to
// disk, further transformed, or merged with other output.
func Eval(s spec.Spec) ([]*yaml.RNode, error) {
	switch s.Kind {
	case spec.ImageKind:
		return evalImage(s)
	case spec.ChartKind:
		return evalHelmChart(s)
	default:
		return nil, ErrNotImplemented
	}
}
