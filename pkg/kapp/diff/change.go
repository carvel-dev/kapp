package diff

import (
	"strings"

	"github.com/cppforlife/go-patch/patch"
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
	"gopkg.in/yaml.v2"
)

type ChangeOp string

const (
	ChangeOpAdd    ChangeOp = "add"
	ChangeOpDelete ChangeOp = "delete"
	ChangeOpUpdate ChangeOp = "update"
	ChangeOpKeep   ChangeOp = "keep" // unchanged
)

var (
	allChangeOps = []ChangeOp{ChangeOpAdd, ChangeOpDelete, ChangeOpUpdate, ChangeOpKeep}
)

type Change interface {
	NewOrExistingResource() ctlres.Resource
	NewResource() ctlres.Resource
	ExistingResource() ctlres.Resource
	AppliedResource() ctlres.Resource

	Op() ChangeOp
	TextDiff() TextDiff
	OpsDiff() OpsDiff

	IsIgnored() bool
	IgnoredReason() string
}

type ChangeImpl struct {
	existingRes, newRes ctlres.Resource

	// appliedRes is an unmodified copy of what's being applied
	appliedRes ctlres.Resource

	textDiff *TextDiff
	opsDiff  *OpsDiff
}

var _ Change = &ChangeImpl{}

func NewChange(existingRes, newRes, appliedRes ctlres.Resource) *ChangeImpl {
	if existingRes == nil && newRes == nil {
		panic("Expected either existingRes or newRes be non-nil")
	}

	if existingRes != nil {
		existingRes = existingRes.DeepCopy()
	}
	if newRes != nil {
		newRes = newRes.DeepCopy()
	}
	if appliedRes != nil {
		appliedRes = appliedRes.DeepCopy()
	}

	return &ChangeImpl{existingRes: existingRes, newRes: newRes, appliedRes: appliedRes}
}

func (d *ChangeImpl) NewOrExistingResource() ctlres.Resource {
	if d.newRes != nil {
		return d.newRes
	}
	if d.existingRes != nil {
		return d.existingRes
	}
	panic("Not possible")
}

func (d *ChangeImpl) NewResource() ctlres.Resource      { return d.newRes }
func (d *ChangeImpl) ExistingResource() ctlres.Resource { return d.existingRes }
func (d *ChangeImpl) AppliedResource() ctlres.Resource  { return d.appliedRes }

func (d *ChangeImpl) Op() ChangeOp {
	if d.existingRes == nil {
		return ChangeOpAdd
	}

	if d.newRes == nil {
		return ChangeOpDelete
	}

	if d.TextDiff().HasChanges() {
		return ChangeOpUpdate
	}

	return ChangeOpKeep
}

func (d *ChangeImpl) IsIgnored() bool { return d.isIgnoredTransient() }

func (d *ChangeImpl) isIgnoredTransient() bool {
	return d.existingRes != nil && d.newRes == nil && d.existingRes.Transient()
}

func (d *ChangeImpl) IgnoredReason() string {
	if d.isIgnoredTransient() {
		return "cluster managed"
	}
	return ""
}

func (d *ChangeImpl) TextDiff() TextDiff {
	// diff is called very often, so memoize
	if d.textDiff != nil {
		return *d.textDiff
	}

	textDiff := d.calculateTextDiff()
	d.textDiff = &textDiff

	return *d.textDiff
}

func (d *ChangeImpl) OpsDiff() OpsDiff {
	if d.opsDiff != nil {
		return *d.opsDiff
	}

	opsDiff := d.calculateOpsDiff()
	d.opsDiff = &opsDiff

	return *d.opsDiff
}

func (d *ChangeImpl) calculateTextDiff() TextDiff {
	existingLines := []string{}
	newLines := []string{}

	if d.existingRes != nil {
		existingBytes, err := d.existingRes.AsYAMLBytes()
		if err != nil {
			panic("yamling existingRes") // TODO panic
		}
		existingLines = strings.Split(string(existingBytes), "\n")
	}

	if d.newRes != nil {
		newBytes, err := d.newRes.AsYAMLBytes()
		if err != nil {
			panic("yamling newRes") // TODO panic
		}
		newLines = strings.Split(string(newBytes), "\n")
	} else if d.IsIgnored() {
		newLines = existingLines // show as no changes
	}

	return NewTextDiff(existingLines, newLines)
}

func (d *ChangeImpl) calculateOpsDiff() OpsDiff {
	var existingObj interface{}
	var newObj interface{}

	if d.existingRes != nil {
		existingBytes, err := d.existingRes.AsYAMLBytes()
		if err != nil {
			panic("yamling existingRes") // TODO panic
		}

		err = yaml.Unmarshal(existingBytes, &existingObj)
		if err != nil {
			panic("unyamling existingRes") // TODO panic
		}
	}

	if d.newRes != nil {
		newBytes, err := d.newRes.AsYAMLBytes()
		if err != nil {
			panic("yamling newRes") // TODO panic
		}

		err = yaml.Unmarshal(newBytes, &newObj)
		if err != nil {
			panic("unyamling newRes") // TODO panic
		}
	} else if d.IsIgnored() {
		newObj = existingObj // show as no changes
	}

	return OpsDiff(patch.Diff{Left: existingObj, Right: newObj}.Calculate())
}
