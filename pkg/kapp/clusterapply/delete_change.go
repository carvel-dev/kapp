package clusterapply

import (
	"encoding/json"
	"fmt"
	"strings"

	ctldiff "github.com/k14s/kapp/pkg/kapp/diff"
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
	ctlresm "github.com/k14s/kapp/pkg/kapp/resourcesmisc"
	"k8s.io/apimachinery/pkg/types"
)

const (
	deleteStrategyAnnKey        = "kapp.k14s.io/delete-strategy"
	deleteStrategyDefaultAnnKey = ""
	deleteStrategyOrphanAnnKey  = "orphan"

	appLabelKey      = "kapp.k14s.io/app" // TODO duplicated here
	orphanedLabelKey = "kapp.k14s.io/orphaned"
)

var (
	jsonPointerEncoder = strings.NewReplacer("~", "~0", "/", "~1")
)

type DeleteChange struct {
	change              ctldiff.Change
	identifiedResources ctlres.IdentifiedResources
}

func (c DeleteChange) Apply() error {
	res := c.change.ExistingResource()
	strategy := res.Annotations()[deleteStrategyAnnKey]

	switch strategy {
	case deleteStrategyOrphanAnnKey:
		mergePatch := []interface{}{
			// TODO currently we do not account for when '-a label:foo=bar' used
			map[string]interface{}{
				"op":   "remove",
				"path": "/metadata/labels/" + jsonPointerEncoder.Replace(appLabelKey),
			},
			map[string]interface{}{
				"op":    "add",
				"path":  "/metadata/labels/" + jsonPointerEncoder.Replace(orphanedLabelKey),
				"value": "",
			},
		}

		patchJSON, err := json.Marshal(mergePatch)
		if err != nil {
			return err
		}

		_, err = c.identifiedResources.Patch(res, types.JSONPatchType, patchJSON)
		return err

	case deleteStrategyDefaultAnnKey:
		// TODO should we be configuring default garbage collection policy to background?
		// https://kubernetes.io/docs/concepts/workloads/controllers/garbage-collection/
		return c.identifiedResources.Delete(res)

	default:
		return fmt.Errorf("Unknown delete strategy: %s", strategy)
	}

	return nil
}

func (c DeleteChange) IsDoneApplying() (ctlresm.DoneApplyState, []string, error) {
	res := c.change.ExistingResource()

	switch res.Annotations()[deleteStrategyAnnKey] {
	case deleteStrategyOrphanAnnKey:
		return ctlresm.DoneApplyState{Done: true, Successful: true, Message: "Resource orphaned"}, nil, nil
	}

	// it should not matter if change is ignored or not
	// because it should be deleted eventually anyway (thru GC)
	exists, err := c.identifiedResources.Exists(res)
	if err != nil {
		return ctlresm.DoneApplyState{}, nil, err
	}

	return ctlresm.DoneApplyState{Done: !exists, Successful: true}, nil, nil
}
