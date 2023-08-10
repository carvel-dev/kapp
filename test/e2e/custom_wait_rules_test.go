// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCustomWaitRules(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}
	kubectl := Kubectl{t, env.Namespace, logger}

	config := `
apiVersion: kapp.k14s.io/v1alpha1
kind: Config

waitRules:
  - ytt:
      funcContractV1:
        resource.star: |
          def is_done(resource):
              state = resource.status.currentState
              if state == "Failed":
                return {"done": True, "successful": False, "message": "Current state as Failed"}
              elif state == "Running":
                return {"done": True, "successful": True, "message": "Current state as Running"}
              else:
                return {"done": True, "successful": False, "message": "Not in Failed or Running state"}
              end
          end
    resourceMatchers:
      - apiVersionKindMatcher: {apiVersion: stable.example.com/v1, kind: CronTab}
`

	crdYaml := `
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: crontabs.stable.example.com
spec:
  group: stable.example.com
  versions:
    - name: v1
      served: true
      storage: true
      schema:
        openAPIV3Schema:
          type: object
          properties:
            spec:
              type: object
              properties:
                cronSpec:
                  type: string
                image:
                  type: string
                replicas:
                  type: integer
            status:
              type: object
              properties:
                currentState:
                  type: string
  scope: Namespaced
  names:
    plural: crontabs
    singular: crontab
    kind: CronTab
---
`
	crYaml := `
apiVersion: "stable.example.com/v1"
kind: CronTab
metadata:
  name: my-new-cron-object-1
spec:
  cronSpec: "* * * * */5"
  image: my-awesome-cron-image
status:
  currentState: %s
---
`

	name := "test-custom-wait-rule-contract-v1"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
	}

	cleanUp()
	defer cleanUp()

	logger.Section("deploy resource with current state as running", func() {
		res, err := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name}, RunOpts{
			StdinReader: strings.NewReader(crdYaml + fmt.Sprintf(crYaml, "Running") + config)})
		if err != nil {
			require.Errorf(t, err, "Expected CronTab to be deployed")
		}
		require.Contains(t, res, "Current state as Running")
		NewPresentClusterResource("CronTab", "my-new-cron-object-1", env.Namespace, kubectl)
	})

	cleanUp()

	logger.Section("deploy resource with current state as failed", func() {
		res, err := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name}, RunOpts{
			StdinReader: strings.NewReader(crdYaml + fmt.Sprintf(crYaml, "Failed") + config),
			AllowError:  true,
		})

		require.Contains(t, res, "Current state as Failed")

		require.Contains(t, err.Error(), "kapp: Error: waiting on reconcile crontab/my-new-cron-object-1")
	})
}

func TestYttWaitRules_WithUnblockChanges(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}
	kubectl := Kubectl{t, env.Namespace, logger}

	crd := `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: crontabs.stable.example.com
spec:
  group: stable.example.com
  versions:
    - name: v1
      served: true
      storage: true
      schema:
        openAPIV3Schema:
          type: object
          properties:
            spec:
              type: object
              properties:
                cronSpec:
                  type: string
                image:
                  type: string
                replicas:
                  type: integer
            status:
              type: object
              properties:
                currentState:
                  type: string
  scope: Namespaced
  names:
    plural: crontabs
    singular: crontab
    kind: CronTab`

	yaml := `
apiVersion: kapp.k14s.io/v1alpha1
kind: Config

waitRules:
  - ytt:
      funcContractV1:
        resource.star: |
          def is_done(resource):
              state = resource.status.currentState
              if state == "Progressing":
                return {"done": False, "unblockChanges": True, "message": "Unblock blocked changes"}
              elif state == "Running":
                return {"done": True, "successful": True, "message": "Current state as Running"}
              else:
                return {"done": True, "successful": False, "message": "Not in Failed or Running state"}
              end
          end
    resourceMatchers:
      - apiVersionKindMatcher: {apiVersion: stable.example.com/v1, kind: CronTab}
---
apiVersion: "stable.example.com/v1"
kind: CronTab
metadata:
  name: my-new-cron-object-1
  annotations:
    kapp.k14s.io/change-group: "cr"
spec:
  cronSpec: "* * * * */5"
  image: my-awesome-cron-image
status:
  currentState: Progressing
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-cm
  annotations:
    kapp.k14s.io/change-rule: "upsert after upserting cr"`

	name := "test-custom-wait-rule-contract-v1"
	crdApp := "test-custom-wait-rule-contract-v1-crd"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
		kapp.Run([]string{"delete", "-a", crdApp})
	}

	cleanUp()
	defer cleanUp()

	go func() {
		for {
			time.Sleep(1 * time.Second)
			exists, err := ClusterResourceExists("ConfigMap", "test-cm", kubectl)
			require.NoError(t, err, "Expected error to not have occurred")
			if exists {
				break
			}
		}
		patch := `[{ "op": "replace", "path": "/status/currentState", "value": "Running"}]`
		PatchClusterResource("CronTab", "my-new-cron-object-1", env.Namespace, patch, kubectl)
	}()

	logger.Section("deploy resource with current state as progressing", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", crdApp}, RunOpts{StdinReader: strings.NewReader(crd)})

		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name}, RunOpts{
			StdinReader: strings.NewReader(yaml)})

		out = strings.TrimSpace(replaceSpaces(replaceTarget(replaceTs(out))))

		expectedOutput := strings.TrimSpace(replaceSpaces(`Changes

Namespace  Name                  Kind       Age  Op      Op st.  Wait to    Rs  Ri  $
kapp-test  my-new-cron-object-1  CronTab    -    create  -       reconcile  -   -  $
^          test-cm               ConfigMap  -    create  -       reconcile  -   -  $

Op:      2 create, 0 delete, 0 update, 0 noop, 0 exists
Wait to: 2 reconcile, 0 delete, 0 noop

<replaced>: ---- applying 1 changes [0/2 done] ----
<replaced>: create crontab/my-new-cron-object-1 (stable.example.com/v1) namespace: kapp-test
<replaced>: ---- waiting on 1 changes [0/2 done] ----
<replaced>: ongoing: reconcile crontab/my-new-cron-object-1 (stable.example.com/v1) namespace: kapp-test
<replaced>:  ^ Allowing blocked changes to proceed: Unblock blocked changes
<replaced>: ---- applying 1 changes [1/2 done] ----
<replaced>: create configmap/test-cm (v1) namespace: kapp-test
<replaced>: ---- waiting on 2 changes [0/2 done] ----
<replaced>: ok: reconcile configmap/test-cm (v1) namespace: kapp-test
<replaced>: ---- waiting on 1 changes [1/2 done] ----
<replaced>: ok: reconcile crontab/my-new-cron-object-1 (stable.example.com/v1) namespace: kapp-test
<replaced>:  ^ Current state as Running
<replaced>: ---- applying complete [2/2 done] ----
<replaced>: ---- waiting complete [2/2 done] ----

Succeeded`))
		require.Equal(t, expectedOutput, out)
	})
}
