package spec

const APIVersion = "spresm.squaremo.dev/v1alpha1"

// Spec is a specification for generating configuration.
type Spec struct {
	APIVersion string `json:"apiVersion" yaml:"apiVersion"`
	Kind       Kind   `json:"kind" yaml:"kind"`

	// the upstream source; might be an image repository, or a git URL
	Source string `json:"source",yaml:"source"`
	// the version of the source that's to be evaluated
	Version string `json:"version" yaml:"version"`

	// kind-specific bits
	// +optional
	Helm  *HelmArgs  `json:"helm,omitempty" yaml:"helm,omitempty"`
	Image *ImageArgs `json:"image,omitempty" yaml:"image,omitempty"`
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
	switch k {
	case ChartKind:
		s.Helm = &HelmArgs{}
	case ImageKind:
		s.Image = &ImageArgs{}
	default:
		// git not supported yet
	}
}

type HelmArgs struct {
	Release struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
	} `json:"release"`
	Values map[string]interface{} `json:"values"`
}

type ImageArgs struct {
	FunctionConfig interface{} `json:"functionConfig" yaml:"functionConfig"`
}
