package patch

import (
	"fmt"
)

type RemoveOp struct {
	Path Pointer
}

func (op RemoveOp) Apply(doc interface{}) (interface{}, error) {
	tokens := op.Path.Tokens()

	if len(tokens) == 1 {
		return nil, fmt.Errorf("Cannot remove entire document")
	}

	obj := doc
	prevUpdate := func(newObj interface{}) { doc = newObj }

	for i, token := range tokens[1:] {
		isLast := i == len(tokens)-2
		currPath := NewPointer(tokens[:i+2])

		switch typedToken := token.(type) {
		case IndexToken:
			typedObj, ok := obj.([]interface{})
			if !ok {
				return nil, NewOpArrayMismatchTypeErr(currPath, obj)
			}

			idx, err := ArrayIndex{Index: typedToken.Index, Modifiers: typedToken.Modifiers, Array: typedObj, Path: currPath}.Concrete()
			if err != nil {
				return nil, err
			}

			if isLast {
				newAry := []interface{}{}
				newAry = append(newAry, typedObj[:idx]...)
				newAry = append(newAry, typedObj[idx+1:]...)
				prevUpdate(newAry)
			} else {
				obj = typedObj[idx]
				prevUpdate = func(newObj interface{}) { typedObj[idx] = newObj }
			}

		case MatchingIndexToken:
			typedObj, ok := obj.([]interface{})
			if !ok {
				return nil, NewOpArrayMismatchTypeErr(currPath, obj)
			}

			var idxs []int

			for itemIdx, item := range typedObj {
				typedItem, ok := item.(map[interface{}]interface{})
				if ok {
					if typedItem[typedToken.Key] == typedToken.Value {
						idxs = append(idxs, itemIdx)
					}
				}
			}

			if typedToken.Optional && len(idxs) == 0 {
				return doc, nil
			}

			if len(idxs) != 1 {
				return nil, OpMultipleMatchingIndexErr{currPath, idxs}
			}

			idx, err := ArrayIndex{Index: idxs[0], Modifiers: typedToken.Modifiers, Array: typedObj, Path: currPath}.Concrete()
			if err != nil {
				return nil, err
			}

			if isLast {
				newAry := []interface{}{}
				newAry = append(newAry, typedObj[:idx]...)
				newAry = append(newAry, typedObj[idx+1:]...)
				prevUpdate(newAry)
			} else {
				obj = typedObj[idx]
				// no need to change prevUpdate since matching item can only be a map
			}

		case KeyToken:
			typedObj, ok := obj.(map[interface{}]interface{})
			if !ok {
				return nil, NewOpMapMismatchTypeErr(currPath, obj)
			}

			var found bool

			obj, found = typedObj[typedToken.Key]
			if !found {
				if typedToken.Optional {
					return doc, nil
				}

				return nil, OpMissingMapKeyErr{typedToken.Key, currPath, typedObj}
			}

			if isLast {
				delete(typedObj, typedToken.Key)
			} else {
				prevUpdate = func(newObj interface{}) { typedObj[typedToken.Key] = newObj }
			}

		default:
			return nil, OpUnexpectedTokenErr{token, currPath}
		}
	}

	return doc, nil
}
