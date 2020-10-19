package merge

import (
	"fmt"

	"sigs.k8s.io/kustomize/kyaml/yaml"
	"sigs.k8s.io/kustomize/kyaml/yaml/merge3"
)

// Merge takes three sets of resources -- ours (aka dest), base (aka
// original), and theirs (aka updated) -- and does a three-way
// merge. It returns an error if the merge has conflicts.
//
// TODO don't fail utterly when there's conflicts; report the
// conflicts instead, or add conflict markers somehow.
func Merge(oursNodes, baseNodes, theirsNodes []*yaml.RNode) ([]*yaml.RNode, error) {
	base := nodesToMap(baseNodes)
	theirs := nodesToMap(theirsNodes)

	// Resource set merge algorithm:
	//
	// For each resource (as identified by GVK+namespace/name)
	//
	// - if present in all, then do three-way merge

	// - if only in theirs, it's not generated; keep it
	// - if only in ours, it's a new generated file; keep it
	// - if only in base, it's removed; lose it
	//
	// - if in base and theirs but not ours, it's been removed from
	// generation; if base and theirs differ, raise a conflict (akin
	// to "file has changes upstream but has been removed locally"),
	// otherwise lose it.

	// - if in theirs and ours but not base, it's added locally _and_
	// added to generated files. Conflict.

	// - if in base and ours but not theirs, it's removed upstream; if
	// ours differs from base, raise a conflict ("removed upstream but
	// changed locally"), otherwise lose it.

	result := []*yaml.RNode{}

	for _, oursNode := range oursNodes {
		meta, err := oursNode.GetMeta()
		if err != nil {
			continue // FIXME think about this; ignores anything without meta
		}
		oursId := meta.GetIdentifier()
		baseNode, baseOk := base[oursId]
		theirsNode, theirsOk := theirs[oursId]
		switch {
		case baseOk && theirsOk:
			// present in all three

			// remove from consideration later
			delete(base, oursId)
			delete(theirs, oursId)
			merged, err := merge3.Merge(oursNode, baseNode, theirsNode)
			if err != nil {
				return nil, err
			}
			result = append(result, merged)
			break
		case baseOk: // and not theirsOk
			// removed in local directory

			// remove from consideration later
			delete(base, oursId)
			// TODO actually check if they differ; this needs either a
			// walk or a serialisation
			return nil, fmt.Errorf("resource %v is changed from original, but removed in local files", oursId)
		case theirsOk: // and not baseOk
			// added locally and new in generated files -- conflict.
			return nil, fmt.Errorf("resource %v from generated resources conflicts with resource added locally", oursId)
		default: // only in ours
			result = append(result, oursNode)
		}
	}

	// that's all the resources from ours. Now to compare any that are
	// in either or both of base and theirs.
	for baseId, _ := range base {
		_, theirsOk := theirs[baseId]
		switch {
		case theirsOk:
			// in base and theirs, not in ours.
			delete(theirs, baseId) // remove from consideration later
			// TODO actually check if it's different.
			return nil, fmt.Errorf("resource %v changed from original, but removed in generated files", baseId)
		default:
			// only in base; lose it.
			break
		}
	}

	// lastly, anything left in theirs is not generated, so keep it.
	for _, theirsNode := range theirs {
		result = append(result, theirsNode)
	}

	return result, nil
}

func nodesToMap(nodes []*yaml.RNode) map[yaml.ResourceIdentifier]*yaml.RNode {
	mapped := make(map[yaml.ResourceIdentifier]*yaml.RNode)
	// FIXME deal with duplicates
	for _, node := range nodes {
		meta, err := node.GetMeta()
		// FIXME think about this: it'll exclude anything it couldn't
		// treat as a resource.
		if err == nil {
			mapped[meta.GetIdentifier()] = node
		}
	}
	return mapped
}
