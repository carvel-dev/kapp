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
		if !reflect.DeepEqual(left, right) {
			return []Op{
				TestOp{Path: NewPointer(tokens), Value: left},
				ReplaceOp{Path: NewPointer(tokens), Value: right},
			}
		}
	}

	return []Op{}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
