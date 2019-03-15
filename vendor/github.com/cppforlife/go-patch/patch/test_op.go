package patch

import (
	"fmt"
	"reflect"
)

type TestOp struct {
	Path   Pointer
	Value  interface{}
	Absent bool
}

func (op TestOp) Apply(doc interface{}) (interface{}, error) {
	if op.Absent {
		return op.checkAbsence(doc)
	}
	return op.checkValue(doc)
}

func (op TestOp) checkAbsence(doc interface{}) (interface{}, error) {
	_, err := FindOp{Path: op.Path}.Apply(doc)
	if err != nil {
		if typedErr, ok := err.(OpMissingIndexErr); ok {
			if typedErr.Path.String() == op.Path.String() {
				return doc, nil
			}
		}
		if typedErr, ok := err.(OpMissingMapKeyErr); ok {
			if typedErr.Path.String() == op.Path.String() {
				return doc, nil
			}
		}
		return nil, err
	}

	return nil, fmt.Errorf("Expected to not find '%s'", op.Path)
}

func (op TestOp) checkValue(doc interface{}) (interface{}, error) {
	foundVal, err := FindOp{Path: op.Path}.Apply(doc)
	if err != nil {
		return nil, err
	}

	if !reflect.DeepEqual(foundVal, op.Value) {
		return nil, fmt.Errorf("Found value does not match expected value")
	}

	// Return same input document
	return doc, nil
}
