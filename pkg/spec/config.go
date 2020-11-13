package spec

import (
	"io"

	"gopkg.in/yaml.v3"
)

// Return the current configuration
func (s *Spec) Config() interface{} {
	switch s.Kind {
	case ChartKind:
		return s.Helm
	case ImageKind:
		return s.Image
	default: // TODO: other kinds
		return nil
	}
}

// Set the configuration from an input source, and return error if it
// cannot be read.
func (s *Spec) ReadConfig(reader io.Reader) error {
	switch s.Kind {
	case ChartKind:
		s.Helm = &HelmArgs{}
		return yaml.NewDecoder(reader).Decode(s.Helm)
	case ImageKind:
		s.Image = &ImageArgs{}
		return yaml.NewDecoder(reader).Decode(s.Image)
	default: // TODO: other kinds
		return nil
	}
}
