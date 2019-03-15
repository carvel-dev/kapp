package app

import (
	"fmt"
	"strings"

	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
	ctlresm "github.com/k14s/kapp/pkg/kapp/resourcesmisc"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

type Preparation struct {
	coreClient    kubernetes.Interface
	dynamicClient dynamic.Interface
}

func NewPreparation(coreClient kubernetes.Interface, dynamicClient dynamic.Interface) Preparation {
	return Preparation{coreClient, dynamicClient}
}

func (a Preparation) PrepareResources(resources []ctlres.Resource, opts PrepareResourcesOpts) ([]ctlres.Resource, error) {
	resources, err := ctlres.NewUniqueResources(resources).Resources()
	if err != nil {
		return nil, err
	}

	resources, err = a.placeIntoNamespace(resources, opts)
	if err != nil {
		return nil, err
	}

	return resources, nil
}

func (a Preparation) placeIntoNamespace(resources []ctlres.Resource, opts PrepareResourcesOpts) ([]ctlres.Resource, error) {
	nsMap := map[string]string{}
	for _, nsKV := range opts.MapNamespaces {
		pieces := strings.SplitN(nsKV, "=", 2)
		if len(pieces) != 2 {
			return nil, fmt.Errorf("Expected map namespace '%s' to be in 'src-ns=dst-ns' format", nsKV)
		}
		nsMap[pieces[0]] = pieces[1]
	}

	resTypes := ctlresm.NewResourceTypes(resources, a.coreClient, a.dynamicClient)

	for i, res := range resources {
		isNsed, err := resTypes.IsNamespaced(res)
		if err != nil {
			return nil, err
		}

		if isNsed {
			if len(res.Namespace()) == 0 {
				if len(opts.DefaultNamespace) > 0 {
					res.SetNamespace(opts.DefaultNamespace)
				}
			}

			if len(opts.IntoNamespace) > 0 {
				res.SetNamespace(opts.IntoNamespace)
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
			if len(res.Namespace()) > 0 {
				return nil, fmt.Errorf("Expected resource '%s' to not specify namespace as its kind is not namespaced", res.Description())
			}
		}
	}

	return resources, nil
}

func (a Preparation) ValidateResources(resources []ctlres.Resource, opts PrepareResourcesOpts) error {
	return a.validateAllows(resources, opts)
}

func (a Preparation) validateAllows(resources []ctlres.Resource, opts PrepareResourcesOpts) error {
	if !opts.AllowCheck {
		return nil
	}

	var errs []error

	for _, res := range resources {
		if res.Namespace() == "" {
			if !opts.AllowCluster {
				errs = append(errs, fmt.Errorf("Cluster level resource '%s' is not allowed", res.Description()))
			}
		} else {
			if !opts.InAllowedNamespaces(res.Namespace()) {
				errs = append(errs, fmt.Errorf("Resource '%s' is outside of allowed namespaces", res.Description()))
			}
		}
	}

	if len(errs) > 0 {
		var msgs []string
		for _, err := range errs {
			msgs = append(msgs, "- "+err.Error())
		}
		return fmt.Errorf("Validation errors:\n%s", strings.Join(msgs, "\n"))
	}

	return nil
}

type PrepareResourcesOpts struct {
	AllowCheck         bool
	AllowedNamespaces  []string
	AllowAllNamespaces bool
	AllowCluster       bool

	IntoNamespace    string   // this ns is allowed automatically
	MapNamespaces    []string // this ns is allowed automatically
	DefaultNamespace string   // this ns is allowed automatically
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
