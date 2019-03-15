package clusterapply

import (
	ctldiff "github.com/k14s/kapp/pkg/kapp/diff"
)

var SplitChangesRestFunc func(ctldiff.Change) bool = nil // does not match any changes as it may be placed in the middle

type SplitChanges struct {
	changes    []ctldiff.Change
	matchFuncs []func(ctldiff.Change) bool
}

func (r SplitChanges) ChangesByFunc() [][]ctldiff.Change {
	result := [][]ctldiff.Change{}
	restIndex := -1

	for i := 0; i < len(r.matchFuncs); i++ {
		if r.matchFuncs[i] == nil {
			restIndex = i
		}
		result = append(result, []ctldiff.Change{})
	}

	if restIndex == -1 {
		panic("SplitChanges configuration requires use of SplitChangesRestFunc")
	}

	for _, res := range r.changes {
		var matched bool
		for i, matchFunc := range r.matchFuncs {
			if matchFunc != nil && matchFunc(res) {
				result[i] = append(result[i], res)
				matched = true
				break
			}
		}
		if !matched {
			result[restIndex] = append(result[restIndex], res)
		}
	}

	return result
}
