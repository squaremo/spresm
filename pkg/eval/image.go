package eval

import (
	//	"errors"
	"bytes"
	"fmt"
	"os/exec"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"

	"github.com/squaremo/spresm/pkg/spec"
)

// evalImage evaluates a spec with the kind "Image", meaning run an
// image to generate the YAMLs.
func evalImage(s spec.Spec) ([]*yaml.RNode, error) {
	image := s.Source
	tag := s.Version
	imageref := fmt.Sprintf("%s:%s", image, tag)
	args := []string{"run", "--rm", "-i", imageref}
	cmd := exec.Command("docker", args...)

	in := &bytes.Buffer{}
	input := &framework.ResourceList{}
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
