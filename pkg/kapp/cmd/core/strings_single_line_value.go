package core

import (
	"strings"

	uitable "github.com/cppforlife/go-cli-ui/ui/table"
)

type ValueStringsSingleLine struct {
	S []string
}

func NewValueStringsSingleLine(s []string) ValueStringsSingleLine { return ValueStringsSingleLine{S: s} }

func (t ValueStringsSingleLine) String() string       { return strings.Join(t.S, ", ") }
func (t ValueStringsSingleLine) Value() uitable.Value { return t }

func (t ValueStringsSingleLine) Compare(other uitable.Value) int {
	otherS := other.(ValueStringsSingleLine).S
	switch {
	case len(t.S) == len(otherS):
		return 0
	case len(t.S) < len(otherS):
		return -1
	default:
		return 1
	}
}
