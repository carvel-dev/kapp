package clusterapply

import (
	"encoding/json"
	"fmt"

	ctldiff "github.com/k14s/kapp/pkg/kapp/diff"
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
	ctlresm "github.com/k14s/kapp/pkg/kapp/resourcesmisc"
	"k8s.io/apimachinery/pkg/types"
)

const (
	deleteStrategyAnnKey        = "kapp.k14s.io/delete-strategy"
	deleteStrategyDefaultAnnKey = ""
	deleteStrategyOrphanAnnKey  = "orphan"

	orphanedAnnKey = "kapp.k14s.io/orphaned"
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
		mergePatch := map[string]interface{}{
			"metadata": map[string]interface{}{
				"annotations": map[string]string{
					orphanedAnnKey: "",
					// TODO remove current app label?
				},
			},
		}

		patchJSON, err := json.Marshal(mergePatch)
		if err != nil {
			return err
		}

		_, err = c.identifiedResources.Patch(res, types.MergePatchType, patchJSON)
		if err != nil {
			return err
		}

	case deleteStrategyDefaultAnnKey:
		// TODO should we be configuring default garbage collection policy to background?
		// https://kubernetes.io/docs/concepts/workloads/controllers/garbage-collection/
		err := c.identifiedResources.Delete(res)
		if err != nil {
			return err
		}

	default:
		return fmt.Errorf("Unknown delete strategy: %s", strategy)
	}

	return nil
}

func (c DeleteChange) IsDoneApplying() (ctlresm.DoneApplyState, error) {
	res := c.change.ExistingResource()

	switch res.Annotations()[deleteStrategyAnnKey] {
	case deleteStrategyOrphanAnnKey:
		return ctlresm.DoneApplyState{Done: true, Successful: true, Message: "Resource orphaned"}, nil
	}

	// it should not matter if change is ignored or not
	// because it should be deleted eventually anyway (thru GC)
	exists, err := c.identifiedResources.Exists(res)
	if err != nil {
		return ctlresm.DoneApplyState{}, err
	}

	return ctlresm.DoneApplyState{Done: !exists, Successful: true}, nil
}
