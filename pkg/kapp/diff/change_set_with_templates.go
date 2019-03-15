package diff

import (
	"fmt"
	"sort"
	"strconv"

	ctlconf "github.com/k14s/kapp/pkg/kapp/config"
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
)

const (
	templateAnnKey        = "kapp.k14s.io/versioned" // Value is ignored
	templateNumVersAnnKey = "kapp.k14s.io/num-versions"
)

type ChangeSetWithTemplates struct {
	existingRs, newRs []ctlres.Resource
	rules             []ctlconf.TemplateRule
	opts              ChangeSetOpts
	changeFactory     ChangeFactory
}

func NewChangeSetWithTemplates(existingRs, newRs []ctlres.Resource,
	rules []ctlconf.TemplateRule, opts ChangeSetOpts, changeFactory ChangeFactory) *ChangeSetWithTemplates {

	return &ChangeSetWithTemplates{existingRs, newRs, rules, opts, changeFactory}
}

func (d ChangeSetWithTemplates) Calculate() ([]Change, error) {
	existingRs := newTemplateResources(d.existingRs)
	existingRsByTemplate := d.groupResourcesByTemplate(existingRs.Template)

	newRs := newTemplateResources(d.newRs)
	allChanges := []Change{}

	d.assignNewNames(newRs, existingRsByTemplate)

	// First try to calculate changes will update references on all resources
	// (which includes templated and non-templated resources)
	_, _, err := d.addChanges(newRs, existingRsByTemplate)
	if err != nil {
		return nil, err
	}

	// Since there might have been circular dependencies;
	// second try catches ones that werent changed during first run
	addChanges, alreadyAdded, err := d.addChanges(newRs, existingRsByTemplate)
	if err != nil {
		return nil, err
	}

	allChanges = append(allChanges, addChanges...)

	keepAndDeleteChanges, err := d.keepAndDeleteChanges(existingRsByTemplate, alreadyAdded)
	if err != nil {
		return nil, err
	}

	allChanges = append(allChanges, keepAndDeleteChanges...)

	nonTemplateChangeSet := NewChangeSet(
		existingRs.NonTemplate, newRs.NonTemplate, d.opts, d.changeFactory)

	nonTemplateChanges, err := nonTemplateChangeSet.Calculate()
	if err != nil {
		return nil, err
	}

	allChanges = append(allChanges, nonTemplateChanges...)

	return allChanges, nil
}

func (d ChangeSetWithTemplates) groupResourcesByTemplate(rs []ctlres.Resource) map[string][]ctlres.Resource {
	result := map[string][]ctlres.Resource{}

	groupByTemplateFunc := func(res ctlres.Resource) string {
		if _, found := res.Annotations()[templateAnnKey]; found {
			return TemplateResource{res, nil}.UniqTemplateKey().String()
		}
		panic("Expected to find template annotation on resource")
	}

	for resKey, subRs := range (GroupResources{rs, groupByTemplateFunc}).Resources() {
		sort.Slice(subRs, func(i, j int) bool {
			return TemplateResource{subRs[i], nil}.Version() < TemplateResource{subRs[j], nil}.Version()
		})
		result[resKey] = subRs
	}

	return result
}

func (d ChangeSetWithTemplates) assignNewNames(
	newRs templateResources, existingRsByTemplate map[string][]ctlres.Resource) {

	// TODO name isnt used during diffing, should it?
	for _, newRes := range newRs.Template {
		newTemplateRes := TemplateResource{newRes, nil}
		newResKey := newTemplateRes.UniqTemplateKey().String()

		if existingRs, found := existingRsByTemplate[newResKey]; found {
			existingRes := existingRs[len(existingRs)-1]
			newTemplateRes.SetTemplatedName(TemplateResource{existingRes, nil}.Version() + 1)
		} else {
			newTemplateRes.SetTemplatedName(1)
		}
	}
}

func (d ChangeSetWithTemplates) addChanges(
	newRs templateResources, existingRsByTemplate map[string][]ctlres.Resource) (
	[]Change, map[string]ctlres.Resource, error) {

	changes := []Change{}
	alreadyAdded := map[string]ctlres.Resource{}

	for _, newRes := range newRs.Template {
		newResKey := TemplateResource{newRes, nil}.UniqTemplateKey().String()
		usedRes := newRes

		if existingRs, found := existingRsByTemplate[newResKey]; found {
			existingRes := existingRs[len(existingRs)-1]

			// Calculate update change to determine if anything changed
			updateChange, err := d.newChange(existingRes, newRes)
			if err != nil {
				return nil, nil, err
			}

			switch updateChange.Op() {
			case ChangeOpUpdate:
				changes = append(changes, d.newAddChangeFromUpdateChange(newRes, updateChange))
			case ChangeOpKeep:
				// Use latest copy of resource to update affected resources
				usedRes = existingRes
			default:
				panic(fmt.Sprintf("Unexpected change op %s", updateChange.Op()))
			}
		} else {
			// Since there no existing resource, create change for new resource
			addChange, err := d.newChange(nil, newRes)
			if err != nil {
				return nil, nil, err
			}
			changes = append(changes, addChange)
		}

		// Update both templates and non-templates
		tplRes := TemplateResource{usedRes, d.rules}

		err := tplRes.UpdateAffected(newRs.NonTemplate)
		if err != nil {
			return nil, nil, err
		}

		err = tplRes.UpdateAffected(newRs.Template)
		if err != nil {
			return nil, nil, err
		}

		alreadyAdded[newResKey] = newRes
	}

	return changes, alreadyAdded, nil
}

func (d ChangeSetWithTemplates) newAddChangeFromUpdateChange(
	newRes ctlres.Resource, updateChange Change) Change {

	// Use update's diffs but create a change for new resource
	addChange := NewChangePrecalculated(nil, newRes, newRes)
	// TODO private field access
	addChange.op = ChangeOpAdd
	addChange.textDiff = updateChange.TextDiff()
	addChange.opsDiff = updateChange.OpsDiff()
	return addChange
}

func (d ChangeSetWithTemplates) keepAndDeleteChanges(
	existingRsByTemplate map[string][]ctlres.Resource,
	alreadyAdded map[string]ctlres.Resource) ([]Change, error) {

	changes := []Change{}

	// Find existing resources that were not already diffed (not in new set of resources)
	for existingResKey, existingRs := range existingRsByTemplate {
		numToKeep := 0

		if newRes, found := alreadyAdded[existingResKey]; found {
			var err error
			numToKeep, err = d.numOfResourcesToKeep(newRes)
			if err != nil {
				return nil, err
			}
		}
		if numToKeep > len(existingRs) {
			numToKeep = len(existingRs)
		}

		// Create changes to delete all or extra resources
		for _, existingRes := range existingRs[0 : len(existingRs)-numToKeep] {
			change, err := d.newChange(existingRes, nil)
			if err != nil {
				return nil, err
			}
			changes = append(changes, change)
		}

		// Create changes that "keep" resources
		for _, existingRes := range existingRs[len(existingRs)-numToKeep:] {
			changes = append(changes, d.newKeepChange(existingRes))
		}
	}

	return changes, nil
}

func (d ChangeSetWithTemplates) newKeepChange(existingRes ctlres.Resource) Change {
	// Use update's diffs but create a change for new resource
	addChange := NewChangePrecalculated(existingRes, nil, nil)
	// TODO private field access
	addChange.op = ChangeOpKeep
	return addChange
}

func (ChangeSetWithTemplates) numOfResourcesToKeep(res ctlres.Resource) (int, error) {
	// TODO get rid of arbitrary cut off
	numToKeep := 5

	if numToKeepAnn, found := res.Annotations()[templateNumVersAnnKey]; found {
		var err error
		numToKeep, err = strconv.Atoi(numToKeepAnn)
		if err != nil {
			return 0, fmt.Errorf("Expected annotation '%s' value to be an integer", templateNumVersAnnKey)
		}
		if numToKeep < 1 {
			return 0, fmt.Errorf("Expected annotation '%s' value to be a >= 1", templateNumVersAnnKey)
		}
	} else {
		numToKeep = 5
	}

	return numToKeep, nil
}

func (d ChangeSetWithTemplates) newChange(existingRes, newRes ctlres.Resource) (Change, error) {
	changeFactoryFunc := d.changeFactory.NewExactChange
	if d.opts.AgainstLastApplied {
		changeFactoryFunc = d.changeFactory.NewChangeAgainstLastApplied
	}
	return changeFactoryFunc(existingRes, newRes)
}

type templateResources struct {
	Template    []ctlres.Resource
	NonTemplate []ctlres.Resource
}

func newTemplateResources(rs []ctlres.Resource) templateResources {
	var result templateResources
	for _, res := range rs {
		if _, found := res.Annotations()[templateAnnKey]; found {
			result.Template = append(result.Template, res)
		} else {
			result.NonTemplate = append(result.NonTemplate, res)
		}
	}
	return result
}
