// Copyright 2026 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package clusterapply_test

import (
	"fmt"
	"testing"

	ctlapply "carvel.dev/kapp/pkg/kapp/clusterapply"
	ctldiff "carvel.dev/kapp/pkg/kapp/diff"
	ctldgraph "carvel.dev/kapp/pkg/kapp/diffgraph"
	"carvel.dev/kapp/pkg/kapp/logger"
	ctlres "carvel.dev/kapp/pkg/kapp/resources"
	"github.com/stretchr/testify/require"
)

func TestClusterChangeSetUnchangedDependency(t *testing.T) {
	for _, order := range []string{"before", "after"} {
		for _, condition := range []string{"", "Complete", "Failed"} {
			t.Run(fmt.Sprintf("%s/condition=%s", order, condition), func(t *testing.T) {
				prerequisiteRule, dependentRule := "", ""
				if order == "before" {
					prerequisiteRule = "upsert before upserting dependent"
				} else {
					dependentRule = "upsert after upserting prerequisite"
				}
				config := fmt.Sprintf(`
apiVersion: batch/v1
kind: Job
metadata:
  name: prerequisite
  annotations:
    kapp.k14s.io/change-group: prerequisite
    %s
status:
  conditions:
  - type: %q
    status: "True"
---
apiVersion: batch/v1
kind: Job
metadata:
  name: dependent
  annotations:
    kapp.k14s.io/change-group: dependent
    %s
`, changeRuleAnnotation(prerequisiteRule), condition, changeRuleAnnotation(dependentRule))

				changes, graph := calculateChangeSet(t, config, []string{"prerequisite"}, ctlapply.ClusterChangeOpts{Wait: true})
				blocked := ctldgraph.NewBlockedChanges(graph)
				if condition == "Complete" {
					require.Len(t, changes, 1)
					require.Equal(t, "dependent", changes[0].Resource().Name())
					require.Len(t, blocked.Unblocked(), 1, "a completed prerequisite must not leave a dangling dependency")
					return
				}

				require.Len(t, changes, 2)
				require.Equal(t, ctlapply.ClusterChangeApplyOpNoop, changes[0].ApplyOp())
				require.Equal(t, ctlapply.ClusterChangeWaitOpOK, changes[0].WaitOp())
				unblocked := blocked.Unblocked()
				require.Len(t, unblocked, 1, "the dependent must wait for the unchanged prerequisite")
				require.Equal(t, "prerequisite", unblocked[0].Change.Resource().Name())
				require.Len(t, blocked.Blocked(), 1)
				if condition == "" {
					blocked.Unblock(unblocked[0])
					require.Empty(t, blocked.Blocked(), "the dependent can proceed once the prerequisite is ready")
				}
			})
		}
	}
}

func TestClusterChangeSetUnchangedDependencyChain(t *testing.T) {
	for _, tc := range []struct {
		name       string
		annotation string
		waitOp     ctlapply.ClusterChangeWaitOp
	}{
		{name: "waiting enabled", waitOp: ctlapply.ClusterChangeWaitOpOK},
		{name: "intermediate waiting disabled", annotation: `kapp.k14s.io/disable-wait: ""`, waitOp: ctlapply.ClusterChangeWaitOpNoop},
	} {
		t.Run(tc.name, func(t *testing.T) {
			config := fmt.Sprintf(`
apiVersion: batch/v1
kind: Job
metadata:
  name: prerequisite
  annotations:
    kapp.k14s.io/change-group: prerequisite
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: intermediate
  annotations:
    %s
    kapp.k14s.io/change-group: intermediate
    kapp.k14s.io/change-rule: upsert after upserting prerequisite
---
apiVersion: batch/v1
kind: Job
metadata:
  name: dependent
  annotations:
    kapp.k14s.io/change-rule: upsert after upserting intermediate
`, tc.annotation)
			changes, graph := calculateChangeSet(t, config, []string{"prerequisite", "intermediate"}, ctlapply.ClusterChangeOpts{Wait: true})
			require.Len(t, changes, 3, "the unchanged intermediate resource must preserve the dependency chain")
			require.Equal(t, ctlapply.ClusterChangeApplyOpNoop, changes[1].ApplyOp())
			require.Equal(t, tc.waitOp, changes[1].WaitOp())
			blocked := ctldgraph.NewBlockedChanges(graph)
			for i, name := range []string{"prerequisite", "intermediate", "dependent"} {
				change := graph.All()[i]
				require.Equal(t, name, change.Change.Resource().Name())
				require.Contains(t, blocked.Unblocked(), change)
				require.Len(t, blocked.Blocked(), 2-i)
				blocked.Unblock(change)
			}
		})
	}
}

func TestClusterChangeSetUnchangedDependencyWaitingDisabled(t *testing.T) {
	for _, tc := range []struct {
		name       string
		annotation string
		wait       bool
	}{
		{name: "wait flag disabled"},
		{name: "disable-wait annotation", annotation: `kapp.k14s.io/disable-wait: ""`, wait: true},
		{name: "noop annotation", annotation: `kapp.k14s.io/noop: ""`, wait: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			config := fmt.Sprintf(`
apiVersion: batch/v1
kind: Job
metadata:
  name: prerequisite
  annotations:
    kapp.k14s.io/change-group: prerequisite
    %s
---
apiVersion: batch/v1
kind: Job
metadata:
  name: dependent
  annotations:
    kapp.k14s.io/change-rule: upsert after upserting prerequisite
`, tc.annotation)
			changes, graph := calculateChangeSet(t, config, []string{"prerequisite"}, ctlapply.ClusterChangeOpts{Wait: tc.wait})
			require.Len(t, changes, 1)
			require.Equal(t, "dependent", changes[0].Resource().Name())
			require.Empty(t, ctldgraph.NewBlockedChanges(graph).Blocked())
		})
	}
}

func TestClusterChangeSetPrunesReadyUnchangedDependents(t *testing.T) {
	for _, resource := range []struct {
		apiVersion string
		kind       string
		readyState string
	}{
		{apiVersion: "v1", kind: "Service", readyState: "spec:\n  clusterIP: 10.0.0.1"},
		{apiVersion: "demo.com/v1beta1", kind: "FooStore"},
		{apiVersion: "batch/v1", kind: "Job", readyState: "status:\n  conditions:\n  - type: Complete\n    status: \"True\""},
	} {
		t.Run(resource.kind, func(t *testing.T) {
			config := fmt.Sprintf(`
apiVersion: batch/v1
kind: Job
metadata:
  name: prerequisite
  annotations:
    kapp.k14s.io/change-group: prerequisite
---
apiVersion: %s
kind: %s
metadata:
  name: ready
  annotations:
    kapp.k14s.io/change-group: ready
    kapp.k14s.io/change-rule: upsert after upserting prerequisite
%s
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: ready-leaf
  annotations:
    kapp.k14s.io/change-rule: upsert after upserting ready
`, resource.apiVersion, resource.kind, resource.readyState)

			changes, graph := calculateChangeSet(t, config, []string{"ready", "ready-leaf"}, ctlapply.ClusterChangeOpts{Wait: true})
			require.Len(t, changes, 1, "ready unchanged dependents must not acquire unnecessary reconcile operations")
			require.Equal(t, "prerequisite", changes[0].Resource().Name())
			require.Empty(t, ctldgraph.NewBlockedChanges(graph).Blocked())
		})
	}
}

func TestClusterChangeSetUnchangedDependencyBranches(t *testing.T) {
	config := `
apiVersion: batch/v1
kind: Job
metadata:
  name: prerequisite-a
  annotations:
    kapp.k14s.io/change-group: prerequisite-a
---
apiVersion: batch/v1
kind: Job
metadata:
  name: prerequisite-b
  annotations:
    kapp.k14s.io/change-group: prerequisite-b
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: intermediate-a
  annotations:
    kapp.k14s.io/change-group: intermediate
    kapp.k14s.io/change-rule: upsert after upserting prerequisite-a
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: intermediate-b
  annotations:
    kapp.k14s.io/change-group: intermediate
    kapp.k14s.io/change-rule: upsert after upserting prerequisite-b
---
apiVersion: batch/v1
kind: Job
metadata:
  name: dependent
  annotations:
    kapp.k14s.io/change-rule: upsert after upserting intermediate
`
	changes, graph := calculateChangeSet(t, config, []string{"prerequisite-a", "prerequisite-b", "intermediate-a", "intermediate-b"}, ctlapply.ClusterChangeOpts{Wait: true})
	require.Len(t, changes, 5, "both unchanged branches must preserve their dependencies on pending prerequisites")
	graphChanges := graph.All()
	blocked := ctldgraph.NewBlockedChanges(graph)
	require.Equal(t, graphChanges[:2], blocked.Unblocked())
	blocked.Unblock(graphChanges[0])
	require.Equal(t, graphChanges[3:], blocked.Blocked())
	blocked.Unblock(graphChanges[2])
	require.Equal(t, graphChanges[3:], blocked.Blocked(), "the dependent must wait for both branches")
	blocked.Unblock(graphChanges[1])
	require.Equal(t, graphChanges[4:], blocked.Blocked())
	blocked.Unblock(graphChanges[3])
	require.Empty(t, blocked.Blocked())
}

func changeRuleAnnotation(rule string) string {
	if rule == "" {
		return ""
	}
	return "kapp.k14s.io/change-rule: " + rule
}

func calculateChangeSet(t *testing.T, config string, existingNames []string, opts ctlapply.ClusterChangeOpts) ([]*ctlapply.ClusterChange, *ctldgraph.ChangeGraph) {
	t.Helper()
	resources, err := ctlres.NewFileResource(ctlres.NewBytesSource([]byte(config))).Resources()
	require.NoError(t, err)
	var changes []ctldiff.Change
	for _, resource := range resources {
		var existing ctlres.Resource
		for _, name := range existingNames {
			if resource.Name() == name {
				existing = resource
			}
		}
		changes = append(changes, ctldiff.NewChange(existing, resource, resource, existing, ctldiff.ChangeOpts{}))
	}
	factory := ctlapply.NewClusterChangeFactory(opts, ctlres.IdentifiedResources{}, ctldiff.ChangeFactory{}, ctldiff.ChangeSetFactory{},
		ctlapply.ConvergedResourceFactory{}, nil, nil)
	changeSet := ctlapply.NewClusterChangeSet(changes, ctlapply.ClusterChangeSetOpts{}, factory, nil, nil, nil, logger.NewTODOLogger())
	clusterChanges, graph, err := changeSet.Calculate()
	require.NoError(t, err)
	return clusterChanges, graph
}
