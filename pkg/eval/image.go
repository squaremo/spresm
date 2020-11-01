package eval

import (
	//	"errors"
	"bytes"
	"fmt"
	"os/exec"

	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"

	"github.com/squaremo/spresm/pkg/spec"
)

// debug
// import (
// 	"io"
// 	"os"
// )

// the YAML encoder in kyaml/yaml is an alias for pkg.in/yaml.v3 and
// does not use json struct tags; the ResourceList type in
// kyaml/.../framework doesn't have json _or_ yaml tags.
type ResourceList struct {
	Kind           string      `yaml:"kind"`
	FunctionConfig interface{} `yaml:"functionConfig"`
	// this will always be empty, but what the heck.
	Items []*yaml.RNode `yaml:"items"`
}

// evalImage evaluates a spec with the kind "Image", meaning run an
// image to generate the YAMLs.
func evalImage(s spec.Spec) ([]*yaml.RNode, error) {
	image := s.Source
	tag := s.Version
	imageref := fmt.Sprintf("%s:%s", image, tag)
	args := []string{"run", "--rm", "-i", imageref}
	cmd := exec.Command("docker", args...)

	in := &bytes.Buffer{}
	// to debug: uncomment and use as arg to NewEncoder below
	// tee := io.MultiWriter(in, os.Stderr)
	input := &ResourceList{
		Kind:           "ResourceList",
		FunctionConfig: s.Image.FunctionConfig,
		Items:          []*yaml.RNode{},
	}
	if err := yaml.NewEncoder(in).Encode(input); err != nil {
		return nil, err
	}
	cmd.Stdin = in

	out := &bytes.Buffer{}
	cmd.Stdout = out // no streaming for now (could use `StdoutPipe`)
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	br := &kio.ByteReader{Reader: out}
	return br.Read()
}
