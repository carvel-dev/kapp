package patch

import (
	"fmt"
	"reflect"
	"sort"

	"gopkg.in/yaml.v2"
)

type Diff struct {
	Left      interface{}
	Right     interface{}
	Unchecked bool
}

func (d Diff) Calculate() Ops {
	ops := d.calculate(d.Left, d.Right, []Token{RootToken{}})
	if !d.Unchecked {
		return ops
	}

	newOps := []Op{}
	for _, op := range ops {
		if _, ok := op.(TestOp); !ok {
			newOps = append(newOps, op)
		}
	}
	return newOps
}

func (d Diff) calculate(left, right interface{}, tokens []Token) []Op {
	switch typedLeft := left.(type) {
	case map[string]interface{}:
		if typedRight, ok := right.(map[string]interface{}); ok {
			ops := []Op{}
			var allKeys []string
			for k := range typedLeft {
				allKeys = append(allKeys, k)
			}
			for k := range typedRight {
				if _, found := typedLeft[k]; !found {
					allKeys = append(allKeys, k)
				}
			}
			sort.SliceStable(allKeys, func(i, j int) bool {
				return allKeys[i] < allKeys[j]
			})
			for _, k := range allKeys {
				newTokens := append([]Token{}, tokens...)
				if leftVal, found := typedLeft[k]; found {
					newTokens = append(newTokens, KeyToken{Key: k})
					if rightVal, found := typedRight[k]; found {
						ops = append(ops, d.calculate(leftVal, rightVal, newTokens)...)
					} else { // remove existing
						ops = append(ops,
							TestOp{Path: NewPointer(newTokens), Value: leftVal},
							RemoveOp{Path: NewPointer(newTokens)},
						)
					}
				} else { // add new
					testOpTokens := append([]Token{}, newTokens...)
					testOpTokens = append(testOpTokens, KeyToken{Key: k})
					newTokens = append(newTokens, KeyToken{Key: k, Optional: true})
					ops = append(ops,
						TestOp{Path: NewPointer(testOpTokens), Absent: true},
						ReplaceOp{Path: NewPointer(newTokens), Value: typedRight[k]},
					)
				}
			}
			return ops
		}
		return []Op{
			TestOp{Path: NewPointer(tokens), Value: left},
			ReplaceOp{Path: NewPointer(tokens), Value: right},
		}

	case map[interface{}]interface{}:
		if typedRight, ok := right.(map[interface{}]interface{}); ok {
			ops := []Op{}
			var allKeys []interface{}
			for k, _ := range typedLeft {
				allKeys = append(allKeys, k)
			}
			for k, _ := range typedRight {
				if _, found := typedLeft[k]; !found {
					allKeys = append(allKeys, k)
				}
			}
			sort.SliceStable(allKeys, func(i, j int) bool {
				iBs, _ := yaml.Marshal(allKeys[i])
				jBs, _ := yaml.Marshal(allKeys[j])
				return string(iBs) < string(jBs)
			})
			for _, k := range allKeys {
				newTokens := append([]Token{}, tokens...)
				if leftVal, found := typedLeft[k]; found {
					newTokens = append(newTokens, KeyToken{Key: fmt.Sprintf("%s", k)})
					if rightVal, found := typedRight[k]; found {
						ops = append(ops, d.calculate(leftVal, rightVal, newTokens)...)
					} else { // remove existing
						ops = append(ops,
							TestOp{Path: NewPointer(newTokens), Value: leftVal},
							RemoveOp{Path: NewPointer(newTokens)},
						)
					}
				} else { // add new
					testOpTokens := append([]Token{}, newTokens...)
					testOpTokens = append(testOpTokens, KeyToken{Key: fmt.Sprintf("%s", k)})
					newTokens = append(newTokens, KeyToken{Key: fmt.Sprintf("%s", k), Optional: true})
					ops = append(ops,
						TestOp{Path: NewPointer(testOpTokens), Absent: true},
						ReplaceOp{Path: NewPointer(newTokens), Value: typedRight[k]},
					)
				}
			}
			return ops
		}
		return []Op{
			TestOp{Path: NewPointer(tokens), Value: left},
			ReplaceOp{Path: NewPointer(tokens), Value: right},
		}

	case []interface{}:
		if typedRight, ok := right.([]interface{}); ok {
			ops := []Op{}
			actualIndex := 0
			for i := 0; i < max(len(typedLeft), len(typedRight)); i++ {
				newTokens := append([]Token{}, tokens...)
				switch {
				case i >= len(typedRight): // remove existing
					newTokens = append(newTokens, IndexToken{Index: actualIndex})
					ops = append(ops,
						TestOp{Path: NewPointer(newTokens), Value: typedLeft[i]}, // capture actual value at index
						RemoveOp{Path: NewPointer(newTokens)},
					)
					// keep actualIndex the same
				case i >= len(typedLeft): // add new
					testOpTokens := append([]Token{}, newTokens...)
					testOpTokens = append(testOpTokens, IndexToken{Index: i}) // use actual index
					newTokens = append(newTokens, AfterLastIndexToken{})
					ops = append(ops,
						TestOp{Path: NewPointer(testOpTokens), Absent: true},
						ReplaceOp{Path: NewPointer(newTokens), Value: typedRight[i]},
					)
					actualIndex++
				default:
					newTokens = append(newTokens, IndexToken{Index: actualIndex})
					ops = append(ops, d.calculate(typedLeft[i], typedRight[i], newTokens)...)
					actualIndex++
				}
			}
			return ops
		}
		return []Op{
			TestOp{Path: NewPointer(tokens), Value: left},
			ReplaceOp{Path: NewPointer(tokens), Value: right},
		}

	default:
		if !reflect.DeepEqual(jsonToYAMLValue(left), jsonToYAMLValue(right)) {
			return []Op{
				TestOp{Path: NewPointer(tokens), Value: left},
				ReplaceOp{Path: NewPointer(tokens), Value: right},
			}
		}
	}

	return []Op{}
}

// The Go JSON library doesn't try to pick the right number type (int, float,
// etc.) when unmarshalling to interface{}, it just picks float64
// universally
func jsonToYAMLValue(j interface{}) interface{} {
	switch j := j.(type) {
	case float64:
		// replicate the logic in https://github.com/go-yaml/yaml/blob/51d6538a90f86fe93ac480b35f37b2be17fef232/resolve.go#L151
		if i64 := int64(j); j == float64(i64) {
			if i := int(i64); i64 == int64(i) {
				return i
			}
			return i64
		}
		if ui64 := uint64(j); j == float64(ui64) {
			return ui64
		}
		return j
	case int64:
		if i := int(j); j == int64(i) {
			return i
		}
		return j
	}
	return j
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
