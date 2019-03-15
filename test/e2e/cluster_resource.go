package e2e

import (
	"fmt"
	"strings"
	"testing"

	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
)

type ClusterResource struct {
	res ctlres.Resource
}

func NewPresentClusterResource(kind, name, ns string, kubectl Kubectl) ClusterResource {
	out, _ := kubectl.RunWithOpts([]string{"get", kind, name, "-n", ns, "-o", "yaml"}, RunOpts{})
	return ClusterResource{ctlres.MustNewResourceFromBytes([]byte(out))}
}

func NewMissingClusterResource(t *testing.T, kind, name, ns string, kubectl Kubectl) {
	_, err := kubectl.RunWithOpts([]string{"get", kind, name, "-n", ns, "-o", "yaml"}, RunOpts{AllowError: true})
	if err == nil || !strings.Contains(err.Error(), "Error from server (NotFound)") {
		t.Fatalf("Expected resource to not exist")
	}
}

func (r ClusterResource) UID() string {
	uid := r.res.UID()
	if len(uid) == 0 {
		panic("Expected UID to be non-empty")
	}
	return uid
}

func (r ClusterResource) Raw() map[string]interface{} {
	return r.res.DeepCopyRaw()
}

func (r ClusterResource) RawPath(path ctlres.Path) interface{} {
	var result interface{} = r.Raw()
	for _, part := range path {
		switch {
		case part.MapKey != nil:
			typedResult, ok := result.(map[string]interface{})
			if !ok {
				panic("Expected to find map")
			}
			result, ok = typedResult[*part.MapKey]
			if !ok {
				panic(fmt.Sprintf("Expected to find key %s", *part.MapKey))
			}

		case part.ArrayIndex != nil:
			typedResult, ok := result.([]interface{})
			if !ok {
				panic("Expected to find array")
			}
			switch {
			case part.ArrayIndex.Index != nil:
				result = typedResult[*part.ArrayIndex.Index]
			case part.ArrayIndex.All != nil:
				panic("Unsupported array index all")
			default:
				panic("Unknown array index")
			}

		default:
			panic("Unknown path part")
		}
	}
	return result
}
