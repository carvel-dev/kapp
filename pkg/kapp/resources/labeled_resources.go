package resources

import (
	"fmt"
	"strings"
	"sync"

	"k8s.io/apimachinery/pkg/labels"
)

type OwnershipLabelModsFunc func(kvs map[string]string) []StringMapAppendMod
type LabelScopingModsFunc func(kvs map[string]string) []StringMapAppendMod

type LabeledResources struct {
	labelSelector       labels.Selector
	identifiedResources IdentifiedResources
}

func NewLabeledResources(labelSelector labels.Selector, identifiedResources IdentifiedResources) *LabeledResources {
	return &LabeledResources{labelSelector, identifiedResources}
}

func (a *LabeledResources) Prepare(resources []Resource, olmFunc OwnershipLabelModsFunc, lsmFunc LabelScopingModsFunc) error {
	labelKey, labelVal, err := NewSimpleLabel(a.labelSelector).KV()
	if err != nil {
		return err
	}

	for _, res := range resources {
		assocLabel := NewAssociationLabel(res)
		ownershipLabels := map[string]string{
			labelKey:         labelVal,
			assocLabel.Key(): assocLabel.Value(),
		}

		for _, t := range olmFunc(ownershipLabels) {
			err := t.Apply(res)
			if err != nil {
				return err
			}
		}

		scopingLabels := map[string]string{
			labelKey: labelVal,
		}

		for _, t := range lsmFunc(scopingLabels) {
			err := t.Apply(res)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (a *LabeledResources) GetAssociated(resource Resource) ([]Resource, error) {
	return a.identifiedResources.List(NewAssociationLabel(resource).AsSelector())
}

func (a *LabeledResources) All() ([]Resource, error) {
	resources, err := a.identifiedResources.List(a.labelSelector)
	if err != nil {
		return nil, err
	}

	return resources, nil
}

// AllAndMatching returns set of all labeled resources
// plus resources that match newResources.
// Returns errors if non-labeled resources were labeled
// with a different value.
func (a *LabeledResources) AllAndMatching(newResources []Resource) ([]Resource, error) {
	resources, err := a.All()
	if err != nil {
		return nil, err
	}

	nonLabeledResources, err := a.findNonLabeledResources(resources, newResources)
	if err != nil {
		return nil, err
	}

	if len(nonLabeledResources) > 0 {
		err := a.checkResourceOwnership(nonLabeledResources)
		if err != nil {
			return nil, err
		}
	}

	return append(resources, nonLabeledResources...), nil
}

func (a *LabeledResources) checkResourceOwnership(resources []Resource) error {
	labelKey, labelVal, err := NewSimpleLabel(a.labelSelector).KV()
	if err != nil {
		return err
	}

	var errs []error

	for _, res := range resources {
		if val, found := res.Labels()[labelKey]; found {
			if val != labelVal {
				errMsg := "Resource '%s' is associated with a different label value '%s=%s'"
				errs = append(errs, fmt.Errorf(errMsg, res.Description(), labelKey, val))
			}
		}
	}

	if len(errs) > 0 {
		var msgs []string
		for _, err := range errs {
			msgs = append(msgs, "- "+err.Error())
		}
		return fmt.Errorf("Ownership errors:\n%s", strings.Join(msgs, "\n"))
	}

	return nil
}

func (a *LabeledResources) findNonLabeledResources(labeledResources, newResources []Resource) ([]Resource, error) {
	var foundResources []Resource
	rsMap := map[string]struct{}{}

	for _, res := range labeledResources {
		rsMap[NewUniqueResourceKey(res).String()] = struct{}{}
	}

	var wg sync.WaitGroup
	errCh := make(chan error, len(newResources))
	resCh := make(chan Resource, len(newResources))

	for _, res := range newResources {
		res := res // copy

		if _, found := rsMap[NewUniqueResourceKey(res).String()]; !found {
			wg.Add(1)
			go func() {
				defer func() { wg.Done() }()

				exists, err := a.identifiedResources.Exists(res)
				if err != nil {
					errCh <- err
					return
				}

				if exists {
					clusterRes, err := a.identifiedResources.Get(res)
					if err != nil {
						errCh <- err
						return
					}
					resCh <- clusterRes
				}
			}()
		}
	}

	wg.Wait()
	close(errCh)
	close(resCh)

	for err := range errCh {
		return nil, err
	}
	for res := range resCh {
		foundResources = append(foundResources, res)
	}

	return foundResources, nil
}
