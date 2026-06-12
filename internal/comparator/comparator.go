package comparator

import (
	"slices"
	"strings"

	"github.com/simplesvet/svcomp/pkg/types"
)

type Comparator struct{}

func New() *Comparator {
	return &Comparator{}
}

func IndexByKey(objects []types.DBObject) map[string]types.DBObject {
	index := make(map[string]types.DBObject, len(objects))
	for _, obj := range objects {
		index[obj.Key()] = obj
	}
	return index
}

func (c *Comparator) Compare(source, target []types.DBObject) []types.DiffResult {
	sourceIndex := IndexByKey(source)
	targetIndex := IndexByKey(target)

	results := make([]types.DiffResult, 0, len(source)+len(target))

	for key, srcObj := range sourceIndex {
		tgtObj, exists := targetIndex[key]
		if !exists {
			results = append(results, types.DiffResult{Object: srcObj, Action: types.ActionCreate})
			continue
		}
		if normalizeDefinition(srcObj.Definition) != normalizeDefinition(tgtObj.Definition) {
			results = append(results, types.DiffResult{Object: srcObj, Action: types.ActionUpdate})
		}
		delete(targetIndex, key)
	}

	for _, tgtObj := range targetIndex {
		results = append(results, types.DiffResult{Object: tgtObj, Action: types.ActionDelete})
	}

	slices.SortFunc(results, func(a, b types.DiffResult) int {
		if kindOrder(a.Object.Kind) != kindOrder(b.Object.Kind) {
			if kindOrder(a.Object.Kind) < kindOrder(b.Object.Kind) {
				return -1
			}
			return 1
		}
		if a.Object.Name == b.Object.Name {
			if actionOrder(a.Action) < actionOrder(b.Action) {
				return -1
			}
			if actionOrder(a.Action) > actionOrder(b.Action) {
				return 1
			}
			return 0
		}
		if a.Object.Name < b.Object.Name {
			return -1
		}
		return 1
	})

	return results
}

func normalizeDefinition(s string) string {
	return strings.TrimSpace(strings.Join(strings.Fields(s), " "))
}

func kindOrder(kind types.ObjectKind) int {
	for i, k := range types.OrderedKinds {
		if k == kind {
			return i
		}
	}
	return len(types.OrderedKinds) + 1
}

func actionOrder(action types.DiffAction) int {
	switch action {
	case types.ActionDelete:
		return 0
	case types.ActionUpdate:
		return 1
	case types.ActionCreate:
		return 2
	default:
		return 3
	}
}
