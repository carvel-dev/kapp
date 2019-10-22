package resources

import (
	"fmt"
	"strings"
	"sync"

	"github.com/k14s/kapp/pkg/kapp/logger"
	"k8s.io/apimachinery/pkg/labels"
)

const (
	disableLabelScopingAnnKey = "kapp.k14s.io/disable-label-scoping" // valid value is ''
)

type OwnershipLabelModsFunc func(kvs map[string]string) []StringMapAppendMod
type LabelScopingModsFunc func(kvs map[string]string) []StringMapAppendMod

type LabeledResources struct {
	labelSelector       labels.Selector
	identifiedResources IdentifiedResources
	logger              logger.Logger
}

func NewLabeledResources(labelSelector labels.Selector,
	identifiedResources IdentifiedResources, logger logger.Logger) *LabeledResources {

	return &LabeledResources{labelSelector, identifiedResources, logger.NewPrefixed("LabeledResources")}
}

func (a *LabeledResources) Prepare(resources []Resource, olmFunc OwnershipLabelModsFunc,
	lsmFunc LabelScopingModsFunc, additionalLabels map[string]string) error {

	defer a.logger.DebugFunc("Prepare").Finish()

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

		for k, v := range additionalLabels {
			ownershipLabels[k] = v
		}

		for _, t := range olmFunc(ownershipLabels) {
			err := t.Apply(res)
			if err != nil {
				return err
			}
		}

		// Scope labels on all resources except ones that explicitly opt out
		disableLSVal, disableLabelScoping := res.Annotations()[disableLabelScopingAnnKey]
		if disableLabelScoping {
			if disableLSVal != "" {
				return fmt.Errorf("Expected annotation '%s' on resource '%s' to have value ''",
					disableLabelScopingAnnKey, res.Description())
			}
		} else {
			for _, t := range lsmFunc(map[string]string{labelKey: labelVal}) {
				err := t.Apply(res)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (a *LabeledResources) GetAssociated(resource Resource) ([]Resource, error) {
	defer a.logger.DebugFunc("GetAssociated").Finish()
	return a.identifiedResources.List(NewAssociationLabel(resource).AsSelector())
}

func (a *LabeledResources) All() ([]Resource, error) {
	defer a.logger.DebugFunc("All").Finish()

	resources, err := a.identifiedResources.List(a.labelSelector)
	if err != nil {
		return nil, err
	}

	return resources, nil
}

type AllAndMatchingOpts struct {
	SkipResourceOwnershipCheck      bool
	BlacklistedResourcesByLabelKeys []string
}

// AllAndMatching returns set of all labeled resources
// plus resources that match newResources.
// Returns errors if non-labeled resources were labeled
// with a different value.
func (a *LabeledResources) AllAndMatching(newResources []Resource, opts AllAndMatchingOpts) ([]Resource, error) {
	defer a.logger.DebugFunc("AllAndMatching").Finish()

	resources, err := a.All()
	if err != nil {
		return nil, err
	}

	nonLabeledResources, err := a.findNonLabeledResources(resources, newResources)
	if err != nil {
		return nil, err
	}

	if !opts.SkipResourceOwnershipCheck && len(nonLabeledResources) > 0 {
		err := a.checkResourceOwnership(nonLabeledResources)
		if err != nil {
			return nil, err
		}
	}

	resources = append(resources, nonLabeledResources...)

	err = a.checkBlacklistedLabels(resources, opts.BlacklistedResourcesByLabelKeys)
	if err != nil {
		return nil, err
	}

	return resources, nil
}

func (a *LabeledResources) checkResourceOwnership(resources []Resource) error {
	expectedLabelKey, expectedLabelVal, err := NewSimpleLabel(a.labelSelector).KV()
	if err != nil {
		return err
	}

	var errs []error

	for _, res := range resources {
		if val, found := res.Labels()[expectedLabelKey]; found {
			if val != expectedLabelVal {
				errMsg := "Resource '%s' is associated with a different label value '%s=%s'"
				errs = append(errs, fmt.Errorf(errMsg, res.Description(), expectedLabelKey, val))
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

func (a *LabeledResources) checkBlacklistedLabels(resources []Resource, blacklistedLblKeys []string) error {
	var errs []error

	for _, res := range resources {
		labels := res.Labels()
		for _, blacklistedLblKey := range blacklistedLblKeys {
			if _, found := labels[blacklistedLblKey]; found {
				errMsg := "Resource '%s' has a blacklisted label '%s'"
				errs = append(errs, fmt.Errorf(errMsg, res.Description(), blacklistedLblKey))
			}
		}
	}

	if len(errs) > 0 {
		var msgs []string
		for _, err := range errs {
			msgs = append(msgs, "- "+err.Error())
		}
		return fmt.Errorf("Blacklist errors:\n%s", strings.Join(msgs, "\n"))
	}

	return nil
}

func (a *LabeledResources) findNonLabeledResources(labeledResources, newResources []Resource) ([]Resource, error) {
	defer a.logger.DebugFunc("findNonLabeledResources").Finish()

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
