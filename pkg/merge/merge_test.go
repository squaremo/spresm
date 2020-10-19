package merge

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func testMerge(t *testing.T, oursSrc, baseSrc, theirsSrc, expected string) {
	var ours, base, theirs []*yaml.RNode

	parse := func(src string, slice *[]*yaml.RNode) {
		reader := kio.ByteReader{Reader: bytes.NewBuffer([]byte(src))}
		nodes, err := reader.Read()
		assert.NoError(t, err)
		*slice = nodes
	}
	parse(oursSrc, &ours)
	parse(baseSrc, &base)
	parse(theirsSrc, &theirs)

	merged, err := Merge(ours, base, theirs)
	assert.NoError(t, err)

	buf := &bytes.Buffer{}
	writer := kio.ByteWriter{Writer: buf}
	assert.NoError(t, writer.Write(merged))
	assert.Equal(t, strings.TrimSpace(expected), strings.TrimSpace(buf.String()))
}

// Merging nothings results in .. nothing.
func TestMergeEmpty(t *testing.T) {
	testMerge(t, "# ours", "# base", "# theirs", "")
}

// Merging something that hasn't changed gets the same thing back.
func TestMergeNoChange(t *testing.T) {
	src := `
apiVersion: v1
kind: Service
meta:
  name: foo
---
apiVersion: apps/v1
kind: Deployment
meta:
  name: bar
  namespace: app
`
	testMerge(t, src, src, src, src)
}

// A local change and a change in generated resources are both
// preserved.
func TestMergeLocalOnly(t *testing.T) {
	base := `
apiVersion: apps/v1
kind: Deployment
meta:
  name: bar
  namespace: app
spec:
  replicas: 1
`
	local := `
apiVersion: apps/v1
kind: Deployment
meta:
  name: bar
  namespace: app
spec:
  replicas: 4
`
	updated := `
apiVersion: apps/v1
kind: Deployment
meta:
  name: bar
  namespace: app
spec:
  replicas: 1
  strategy: Replace
`
	merged := `
apiVersion: apps/v1
kind: Deployment
meta:
  name: bar
  namespace: app
spec:
  replicas: 4
  strategy: Replace
`

	testMerge(t, updated, base, local, merged)
}

// A resource that's (supposedly) generated by the new spec is kept.
func TestMergeNew(t *testing.T) {
	local := `
apiVersion: v1
kind: Service
meta:
  name: foo
`
	base := `
apiVersion: v1
kind: Service
meta:
  name: foo
`
	updated := `
apiVersion: v1
kind: Service
meta:
  name: foo
---
apiVersion: apps/v1
kind: Deployment
meta:
  name: bar
  namespace: app
`
	testMerge(t, updated, base, local, updated)
}

// A resource that's only in base (previously generated) is discarded.
func TestRemoved(t *testing.T) {
	base := `
apiVersion: v1
kind: Service
meta:
  name: foo
---
apiVersion: apps/v1
kind: Deployment
meta:
  name: bar
  namespace: app
`
	local := `
apiVersion: v1
kind: Service
meta:
  name: foo
`
	updated := `
apiVersion: v1
kind: Service
meta:
  name: foo
`
	testMerge(t, updated, base, local, updated)
}
