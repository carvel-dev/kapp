// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"fmt"
	"strings"
	"time"

	ctlres "carvel.dev/kapp/pkg/kapp/resources"
	ctlresm "carvel.dev/kapp/pkg/kapp/resourcesmisc"
)

const (
	nonceAnnKey = "kapp.k14s.io/nonce"
)

type Preparation struct {
	resourceTypes ctlres.ResourceTypes
	opts          PrepareResourcesOpts
}

type PrepareResourcesOpts struct {
	BeforeModificationFunc func([]ctlres.Resource) []ctlres.Resource

	AllowCheck         bool
	AllowedNamespaces  []string
	AllowAllNamespaces bool
	AllowCluster       bool

	IntoNamespace    string   // this ns is allowed automatically
	MapNamespaces    []string // this ns is allowed automatically
	DefaultNamespace string   // this ns is allowed automatically
}

func NewPreparation(resourceTypes ctlres.ResourceTypes, opts PrepareResourcesOpts) Preparation {
	return Preparation{resourceTypes, opts}
}

func (a Preparation) PrepareResources(resources []ctlres.Resource) ([]ctlres.Resource, error) {
	err := a.validateBasicInfo(resources)
	if err != nil {
		return nil, err
	}

	resources, err = ctlres.NewUniqueResources(resources).Resources()
	if err != nil {
		return nil, err
	}

	resources = a.opts.BeforeModificationFunc(resources)

	resources, err = a.placeIntoNamespace(resources)
	if err != nil {
		return nil, err
	}

	resources, err = a.addNonce(resources)
	if err != nil {
		return nil, err
	}

	return resources, nil
}

func (a Preparation) placeIntoNamespace(resources []ctlres.Resource) ([]ctlres.Resource, error) {
	nsMap := map[string]string{}
	for _, nsKV := range a.opts.MapNamespaces {
		pieces := strings.SplitN(nsKV, "=", 2)
		if len(pieces) != 2 {
			return nil, fmt.Errorf("Expected map namespace '%s' to be in 'src-ns=dst-ns' format", nsKV)
		}
		nsMap[pieces[0]] = pieces[1]
	}

	resTypes := ctlresm.NewResourceTypes(resources, a.resourceTypes)

	for i, res := range resources {
		isNsed, err := resTypes.IsNamespaced(res)
		if err != nil {
			return nil, err
		}

		if isNsed {
			if len(res.Namespace()) == 0 {
				if len(a.opts.DefaultNamespace) > 0 {
					res.SetNamespace(a.opts.DefaultNamespace)
				}
			}

			if len(a.opts.IntoNamespace) > 0 {
				res.SetNamespace(a.opts.IntoNamespace)
			}

			if len(nsMap) > 0 {
				if dstNs, found := nsMap[res.Namespace()]; found {
					res.SetNamespace(dstNs)
				} else {
					return nil, fmt.Errorf("Expected to find mapped namespace for '%s'", res.Namespace())
				}
			}

			resources[i] = res
		} else {
			res.RemoveNamespace()
		}
	}

	return resources, nil
}

func (a Preparation) addNonce(resources []ctlres.Resource) ([]ctlres.Resource, error) {
	addNonceMod := ctlres.StringMapAppendMod{
		ResourceMatcher: ctlres.AllMatcher{},
		Path:            ctlres.NewPathFromStrings([]string{"metadata", "annotations"}),
		KVs: map[string]string{
			nonceAnnKey: fmt.Sprintf("%d", time.Now().UTC().UnixNano()),
		},
	}

	for _, res := range resources {
		if val, found := res.Annotations()[nonceAnnKey]; found {
			if val != "" {
				return nil, fmt.Errorf("Expected annotation '%s' on resource '%s' to have value ''",
					nonceAnnKey, res.Description())
			}

			err := addNonceMod.Apply(res)
			if err != nil {
				return nil, err
			}
		}
	}
	return resources, nil
}

func (a Preparation) validateBasicInfo(resources []ctlres.Resource) error {
	var errs []error

	for _, res := range resources {
		if res.Kind() == "" {
			errs = append(errs, fmt.Errorf("Expected 'kind' on resource '%s' to be non-empty (%s)", res.Description(), res.Origin()))
		}
		if res.APIVersion() == "" {
			errs = append(errs, fmt.Errorf("Expected 'apiVersion' on resource '%s' to be non-empty (%s)", res.Description(), res.Origin()))
		}
		if res.Name() == "" {
			errs = append(errs, fmt.Errorf("Expected 'metadata.name' on resource '%s' to be non-empty (%s)", res.Description(), res.Origin()))
		}
	}

	return a.combinedErr(errs)
}

func (a Preparation) ValidateResources(resources []ctlres.Resource) error {
	return a.validateAllows(resources)
}

func (a Preparation) validateAllows(resources []ctlres.Resource) error {
	if !a.opts.AllowCheck {
		return nil
	}

	var errs []error

	for _, res := range resources {
		if res.Namespace() == "" {
			if !a.opts.AllowCluster {
				errs = append(errs, fmt.Errorf("Cluster level resource '%s' is not allowed (%s)", res.Description(), res.Origin()))
			}
		} else {
			if !a.opts.InAllowedNamespaces(res.Namespace()) {
				errs = append(errs, fmt.Errorf("Resource '%s' is outside of allowed namespaces (%s)", res.Description(), res.Origin()))
			}
		}
	}

	return a.combinedErr(errs)
}

func (a Preparation) combinedErr(errs []error) error {
	if len(errs) > 0 {
		var msgs []string
		for _, err := range errs {
			msgs = append(msgs, "- "+err.Error())
		}
		return fmt.Errorf("Validation errors:\n%s", strings.Join(msgs, "\n"))
	}

	return nil
}

func (o PrepareResourcesOpts) InAllowedNamespaces(ns string) bool {
	if len(o.AllowedNamespaces) == 0 && o.AllowAllNamespaces {
		return true
	}

	for _, n := range o.AllowedNamespaces {
		if ns == n {
			return true
		}
	}

	if len(o.IntoNamespace) > 0 && ns == o.IntoNamespace {
		return true
	}
	if len(o.DefaultNamespace) > 0 && ns == o.DefaultNamespace {
		return true
	}
	if len(o.MapNamespaces) > 0 {
		// TODO consolidate parsing
		for _, kv := range o.MapNamespaces {
			pieces := strings.SplitN(kv, "=", 2)
			if len(pieces) == 2 && ns == pieces[1] {
				return true
			}
		}
	}

	return false
}
