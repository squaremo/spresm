package spec

type Spec struct {
	APIVersion string `json:"apiVersion"`
	Kind       Kind   `json:"kind"`

	// the upstream source; might be an image repository, or a git URL
	Source string `json:"source"`
	// the version of the source that's to be evaluated
	Version string `json:"version"`
}

type Kind string

const (
	ImageKind Kind = "Image"
	ChartKind Kind = "HelmChart"
	GitKind   Kind = "Git"
)