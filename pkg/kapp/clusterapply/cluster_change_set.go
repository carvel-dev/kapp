package clusterapply

import (
	"fmt"
	"sync"
	"time"

	ctldiff "github.com/k14s/kapp/pkg/kapp/diff"
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
	ctlresm "github.com/k14s/kapp/pkg/kapp/resourcesmisc"
)

const (
	uiSepPrefix = "--- "
)

type ClusterChangeSetOpts struct {
	WaitTimeout       time.Duration
	WaitCheckInterval time.Duration
}

type ClusterChangeSet struct {
	changes              []ctldiff.Change
	opts                 ClusterChangeSetOpts
	clusterChangeFactory ClusterChangeFactory
	ui                   UI
}

func NewClusterChangeSet(changes []ctldiff.Change, opts ClusterChangeSetOpts, clusterChangeFactory ClusterChangeFactory, ui UI) ClusterChangeSet {
	return ClusterChangeSet{changes, opts, clusterChangeFactory, ui}
}

func (c ClusterChangeSet) Apply() error {
	c.ui.Notify(uiSepPrefix + "applying changes")

	// TODO split by crds, etc
	// TODO wait for apply/delete

	isCRD := func(change ctldiff.Change) bool {
		return ctlresm.NewCRDvX(change.NewOrExistingResource()) != nil
	}

	splitChanges := SplitChanges{c.changes, []func(ctldiff.Change) bool{
		// Create CRDs first as new resources may be of their type
		func(change ctldiff.Change) bool { return isCRD(change) && change.Op() != ctldiff.ChangeOpDelete },

		// Create namespaces next as new resources might be inside of them
		func(change ctldiff.Change) bool {
			isNs := ctlres.APIGroupKindMatcher{Kind: "Namespace"}.Matches(change.NewOrExistingResource())
			return isNs && change.Op() != ctldiff.ChangeOpDelete
		},

		SplitChangesRestFunc,

		// Delete CRDs last as they may be used by other resources
		func(change ctldiff.Change) bool { return isCRD(change) && change.Op() == ctldiff.ChangeOpDelete },
	}}

	for _, cs := range splitChanges.ChangesByFunc() {
		var clusterChanges []ClusterChange
		var wg sync.WaitGroup
		applyErrCh := make(chan error, len(cs))

		for _, change := range cs {
			clusterChange := c.clusterChangeFactory.NewClusterChange(change)

			desc := clusterChange.ApplyDescription()
			if len(desc) > 0 {
				c.ui.Notify("%s", desc)
			}

			wg.Add(1)

			go func() {
				defer func() { wg.Done() }()

				err := clusterChange.ApplyOp().Execute()
				applyErrCh <- err
			}()

			clusterChanges = append(clusterChanges, clusterChange)
		}

		wg.Wait()
		close(applyErrCh)

		for range cs {
			err := <-applyErrCh
			if err != nil {
				return err
			}
		}

		err := c.waitForClusterChanges(clusterChanges)
		if err != nil {
			return err
		}
	}

	// TODO apply nonce?

	c.ui.Notify(uiSepPrefix + "changes applied")

	return nil
}

func (c ClusterChangeSet) waitForClusterChanges(changes []ClusterChange) error {
	startTime := time.Now()
	inProgressChanges := changes

	for {
		var newInProgressChanges []ClusterChange

		for _, change := range inProgressChanges {
			desc := change.WaitDescription()
			if len(desc) > 0 {
				c.ui.Notify("waiting on %s", desc)
			}

			state, err := change.IsDoneApplyingOp().Execute()
			if err != nil {
				return fmt.Errorf("waiting on %s errored: %s", desc, err)
			}

			switch {
			case !state.Done:
				newInProgressChanges = append(newInProgressChanges, change)
			case state.Done && !state.Successful:
				msg := ""
				if len(state.Message) > 0 {
					msg += " (" + state.Message + ")"
				}
				return fmt.Errorf("waiting on %s: finished unsuccessfully%s", desc, msg)
			}
		}

		if len(newInProgressChanges) == 0 {
			return nil
		}

		inProgressChanges = newInProgressChanges

		if time.Now().Sub(startTime) > c.opts.WaitTimeout {
			return fmt.Errorf("timed out waiting after %s", c.opts.WaitTimeout)
		}

		time.Sleep(c.opts.WaitCheckInterval)
		c.ui.Notify("")
		c.ui.Notify(uiSepPrefix+" waiting on %d changes", len(inProgressChanges))
	}
}
