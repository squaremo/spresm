package merge

import (
	"fmt"

	"sigs.k8s.io/kustomize/kyaml/yaml"
	"sigs.k8s.io/kustomize/kyaml/yaml/merge3"
)

// Merge takes three sets of resources -- mine (aka dest), orig (aka
// base, aka older), and yours (aka updated) -- and does a three-way
// merge. It returns an error if the merge has conflicts.
//
// See
// https://www.gnu.org/software/diffutils/manual/html_node/diff3-Merging.html
// for more information about three-way merge.
//
// TODO don't fail utterly when there's conflicts; report the
// conflicts instead, or add conflict markers somehow.
func Merge(mineNodes, origNodes, yoursNodes []*yaml.RNode) ([]*yaml.RNode, error) {
	orig := nodesToMap(origNodes)
	yours := nodesToMap(yoursNodes)

	// Resource set merge algorithm:
	//
	// For each resource (as identified by GVK+namespace/name)
	//
	// - if present in all, then do three-way merge

	// - if only in `mine`, it's introduced locally; keep it
	// - if only in `yours`, it's a new file upstream; keep it
	// - if only in `orig`, it's removed; lose it
	//
	// - if in `orig` and `yours` but not in `mine`, it's been removed
	// locally; if `orig` and `yours` differ, raise a conflict (akin
	// to "file has changes upstream but has been removed locally"),
	// otherwise lose it.

	// - if in `yours` and `mine` but not `orig`, it's added locally
	// _and_ added upstream. Conflict.

	// - if in `orig` and `mine` but not in `yours`, it's removed
	// upstream; if `mine` differs from `orig`, raise a conflict
	// ("removed upstream but changed locally"), otherwise lose it.

	result := []*yaml.RNode{}

	// The following bears a resemblance to
	// https://github.com/kubernetes-sigs/kustomize/blob/master/kyaml/kio/filters/merge3.go#L79,
	// but:
	//
	// - it's operating on sets of nodes, rather the loading from files
	// - it treats change/remove conflicts as conflicts

	for _, mineNode := range mineNodes {
		meta, err := mineNode.GetMeta()
		if err != nil {
			continue // FIXME think about this; ignores anything without meta
		}
		mineId := meta.GetIdentifier()
		origNode, origOk := orig[mineId]
		yoursNode, yoursOk := yours[mineId]
		switch {
		case origOk && yoursOk:
			// present in all three

			// remove from consideration later
			delete(orig, mineId)
			delete(yours, mineId)
			merged, err := merge3.Merge(mineNode, origNode, yoursNode)
			if err != nil {
				return nil, err
			}
			result = append(result, merged)
			break
		case origOk: // and not theirsOk
			// removed in local directory

			// remove from consideration later
			delete(orig, mineId)
			// TODO actually check if they differ; this needs either a
			// walk or a serialisation
			return nil, fmt.Errorf("resource %v is changed from original, but removed in local files", mineId)
		case yoursOk: // and not baseOk
			// added locally and new in generated files -- conflict.
			return nil, fmt.Errorf("resource %v from generated resources conflicts with resource added locally", mineId)
		default: // only in ours
			result = append(result, mineNode)
		}
	}

	// that's all the resources from ours. Now to compare any that are
	// in either or both of base and theirs.
	for origId, _ := range orig {
		_, yoursOk := yours[origId]
		switch {
		case yoursOk:
			// in base and theirs, not in ours.
			delete(yours, origId) // remove from consideration later
			// TODO actually check if it's different.
			return nil, fmt.Errorf("resource %v changed from original, but removed in generated files", origId)
		default:
			// only in base; lose it.
			break
		}
	}

	// lastly, anything left in theirs is not generated, so keep it.
	for _, yoursNode := range yours {
		result = append(result, yoursNode)
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
