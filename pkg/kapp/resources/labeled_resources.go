// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	"fmt"
	"strings"
	"sync"

	"github.com/k14s/kapp/pkg/kapp/logger"
	"github.com/k14s/kapp/pkg/kapp/util"
	"k8s.io/apimachinery/pkg/labels"
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

// Modifies passed resources for labels and ownership
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

		for _, t := range lsmFunc(map[string]string{labelKey: labelVal}) {
			err := t.Apply(res)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (a *LabeledResources) GetAssociated(resource Resource, resRefs []ResourceRef) ([]Resource, error) {
	defer a.logger.DebugFunc("GetAssociated").Finish()
	return a.identifiedResources.List(NewAssociationLabel(resource).AsSelector(), resRefs, IdentifiedResourcesListOpts{})
}

func (a *LabeledResources) All(listOpts IdentifiedResourcesListOpts) ([]Resource, error) {
	defer a.logger.DebugFunc("All").Finish()

	resources, err := a.identifiedResources.List(a.labelSelector, nil, listOpts)
	if err != nil {
		return nil, err
	}

	return resources, nil
}

type AllAndMatchingOpts struct {
	ExistingNonLabeledResourcesCheck            bool
	ExistingNonLabeledResourcesCheckConcurrency int
	SkipResourceOwnershipCheck                  bool

	DisallowedResourcesByLabelKeys []string
	LabelErrorResolutionFunc       func(string, string) string

	IdentifiedResourcesListOpts IdentifiedResourcesListOpts
}

// AllAndMatching returns set of all labeled resources
// plus resources that match newResources.
// Returns errors if non-labeled resources were labeled
// with a different value.
func (a *LabeledResources) AllAndMatching(newResources []Resource, opts AllAndMatchingOpts) ([]Resource, error) {
	defer a.logger.DebugFunc("AllAndMatching").Finish()

	resources, err := a.All(opts.IdentifiedResourcesListOpts)
	if err != nil {
		return nil, err
	}

	var nonLabeledResources []Resource

	if opts.ExistingNonLabeledResourcesCheck {
		var err error
		nonLabeledResources, err = a.findNonLabeledResources(
			resources, newResources, opts.ExistingNonLabeledResourcesCheckConcurrency)
		if err != nil {
			return nil, err
		}
	}

	if !opts.SkipResourceOwnershipCheck && len(nonLabeledResources) > 0 {
		err := a.checkResourceOwnership(nonLabeledResources, opts)
		if err != nil {
			return nil, err
		}
	}

	resources = append(resources, nonLabeledResources...)

	err = a.checkDisallowedLabels(resources, opts.DisallowedResourcesByLabelKeys)
	if err != nil {
		return nil, err
	}

	return resources, nil
}

func (a *LabeledResources) checkResourceOwnership(resources []Resource, opts AllAndMatchingOpts) error {
	expectedLabelKey, expectedLabelVal, err := NewSimpleLabel(a.labelSelector).KV()
	if err != nil {
		return err
	}

	var errs []error

	for _, res := range resources {
		if val, found := res.Labels()[expectedLabelKey]; found {
			if val != expectedLabelVal {
				ownerMsg := fmt.Sprintf("different label '%s=%s'", expectedLabelKey, val)
				if opts.LabelErrorResolutionFunc != nil {
					ownerMsgSuggested := opts.LabelErrorResolutionFunc(expectedLabelKey, val)
					if len(ownerMsgSuggested) > 0 {
						ownerMsg = ownerMsgSuggested
					}
				}
				errMsg := "Resource '%s' is already associated with a %s"
				errs = append(errs, fmt.Errorf(errMsg, res.Description(), ownerMsg))
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

func (a *LabeledResources) checkDisallowedLabels(resources []Resource, disallowedLblKeys []string) error {
	var errs []error

	for _, res := range resources {
		labels := res.Labels()
		for _, disallowedLblKey := range disallowedLblKeys {
			if _, found := labels[disallowedLblKey]; found {
				errMsg := "Resource '%s' has a disallowed label '%s'"
				errs = append(errs, fmt.Errorf(errMsg, res.Description(), disallowedLblKey))
			}
		}
	}

	if len(errs) > 0 {
		var msgs []string
		for _, err := range errs {
			msgs = append(msgs, "- "+err.Error())
		}
		return fmt.Errorf("Disallowed labels errors:\n%s", strings.Join(msgs, "\n"))
	}

	return nil
}

func (a *LabeledResources) findNonLabeledResources(labeledResources, newResources []Resource, concurrency int) ([]Resource, error) {
	defer a.logger.DebugFunc("findNonLabeledResources").Finish()

	var foundResources []Resource
	rsMap := map[string]struct{}{}

	for _, res := range labeledResources {
		rsMap[NewUniqueResourceKey(res).String()] = struct{}{}
	}

	var wg sync.WaitGroup
	throttle := util.NewThrottle(concurrency)

	errCh := make(chan error, len(newResources))
	resCh := make(chan Resource, len(newResources))

	for _, res := range newResources {
		res := res // copy

		if _, found := rsMap[NewUniqueResourceKey(res).String()]; !found {
			wg.Add(1)
			go func() {
				throttle.Take()
				defer throttle.Done()

				defer func() { wg.Done() }()

				clusterRes, exists, err := a.identifiedResources.Exists(res, ExistsOpts{})
				if err != nil {
					errCh <- err
					return
				}

				if exists {
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
