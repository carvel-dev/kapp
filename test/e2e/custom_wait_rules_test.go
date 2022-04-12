// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCustomWaitRules(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}
	//kubectl := Kubectl{t, env.Namespace, logger}

	config := `
apiVersion: kapp.k14s.io/v1alpha1
kind: Config

waitRules:
  - ytt:
      funcContractV1:
        rules.star: |
          def check_status(resource):
              state = resource.status.currentState
              if state == "Failed":
                return {"Done":True, "Successful": False, "Message": ""}
              elif state == "Running":
                return {"Done":True, "Successful": True, "Message": ""}
              else:
                return {"Done":False, "Successful": False, "Message": "Not in Failed or Running state"}
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
    shortNames:
      - ct
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

	logger.Section("deploy when resource current state as running)", func() {
		res, err := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(crdYaml + fmt.Sprintf(crYaml, "Running") + config)})
		if err != nil {
			require.Errorf(t, err, "")
		}
		require.Contains(t, res, "Succeeded")
	})
}
