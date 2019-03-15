package core

import (
	"fmt"

	uitable "github.com/cppforlife/go-cli-ui/ui/table"
)

type ValueUnknownBool struct {
	B *bool
}

var _ uitable.Value = ValueUnknownBool{}

func NewValueUnknownBool(b *bool) ValueUnknownBool { return ValueUnknownBool{B: b} }

func (t ValueUnknownBool) String() string {
	if t.B != nil {
		return fmt.Sprintf("%t", *t.B)
	}
	return ""
}

func (t ValueUnknownBool) Value() uitable.Value            { return t }
func (t ValueUnknownBool) Compare(other uitable.Value) int { panic("Never called") }
