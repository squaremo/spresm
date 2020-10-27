package spec

const APIVersion = "spresm.squaremo.dev/v1alpha1"

type Spec struct {
	APIVersion string `json:"apiVersion"`
	Kind       Kind   `json:"kind"`

	// the upstream source; might be an image repository, or a git URL
	Source string `json:"source"`
	// the version of the source that's to be evaluated
	Version string `json:"version"`

	// kind-specific bits
	Helm *HelmArgs `json:"helm,omitempty"`
}

type Kind string

const (
	ImageKind Kind = "Image"
	ChartKind Kind = "HelmChart"
	GitKind   Kind = "Git"
)

func (s *Spec) Init(k Kind) {
	s.APIVersion = APIVersion
	s.Kind = k
}

func (s *Spec) Config() interface{} {
	switch s.Kind {
	case ChartKind:
		return s.Helm
	default: // TODO: other kinds
		return nil
	}
}

type HelmArgs struct {
	Values  map[string]interface{} `json:"values"`
	Release struct {
		Name string `json:"name"`
	} `json:"release"`
}
