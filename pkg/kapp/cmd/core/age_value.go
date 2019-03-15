package core

import (
	"time"

	uitable "github.com/cppforlife/go-cli-ui/ui/table"
	"k8s.io/apimachinery/pkg/util/duration"
)

type ValueAge struct {
	T time.Time
}

var _ uitable.Value = ValueAge{}

func NewValueAge(t time.Time) ValueAge { return ValueAge{T: t} }

func (t ValueAge) String() string {
	if t.T.IsZero() {
		return ""
	}
	return duration.ShortHumanDuration(time.Now().Sub(t.T))
}

func (t ValueAge) Value() uitable.Value { return t }

func (t ValueAge) Compare(other uitable.Value) int {
	otherT := other.(ValueAge).T
	switch {
	case t.T.Equal(otherT):
		return 0
	case t.T.Before(otherT):
		return -1
	default:
		return 1
	}
}
