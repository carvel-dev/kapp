// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestOrder(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}

	name := "test-order"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
	}

	cleanUp()
	defer cleanUp()

	logger.Section("deploy ordered", func() {
		yaml1 := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
  annotations: {}
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: import-etcd-into-db
  annotations:
    kapp.k14s.io/change-group: "apps.big.co/import-etcd-into-db"
    kapp.k14s.io/change-rule: "upsert before deleting apps.big.co/etcd" # ref to removed object
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: migrations
  annotations:
    kapp.k14s.io/change-group: "apps.big.co/db-migrations"
    kapp.k14s.io/change-rule: "upsert after upserting apps.big.co/import-etcd-into-db"
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-ing
  annotations:
    kapp.k14s.io/change-group: "apps.big.co/deployment"
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-svc
  annotations:
    kapp.k14s.io/change-group: "apps.big.co/deployment"
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: app
  annotations:
    kapp.k14s.io/change-group: "apps.big.co/deployment"
    kapp.k14s.io/change-rule: "upsert after upserting apps.big.co/db-migrations"
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-health-check
  annotations:
    kapp.k14s.io/change-rule: "upsert after upserting apps.big.co/deployment"
`

		out, _ := kapp.RunWithOpts([]string{"deploy", "--tty", "-f", "-", "-a", name}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})

		out = sortActionsInResult(strings.TrimSpace(replaceSpaces(replaceTarget(replaceTs(out)))))
		expectedOutput := strings.TrimSpace(replaceSpaces(`Changes

Namespace  Name                 Kind       Age  Op      Op st.  Wait to    Rs  Ri  $
kapp-test  app                  ConfigMap  -    create  -       reconcile  -   -  $
^          app-config           ConfigMap  -    create  -       reconcile  -   -  $
^          app-health-check     ConfigMap  -    create  -       reconcile  -   -  $
^          app-ing              ConfigMap  -    create  -       reconcile  -   -  $
^          app-svc              ConfigMap  -    create  -       reconcile  -   -  $
^          import-etcd-into-db  ConfigMap  -    create  -       reconcile  -   -  $
^          migrations           ConfigMap  -    create  -       reconcile  -   -  $

Op:      7 create, 0 delete, 0 update, 0 noop, 0 exists
Wait to: 7 reconcile, 0 delete, 0 noop

<replaced>: ---- applying 4 changes [0/7 done] ----
<replaced>: create configmap/app-config (v1) namespace: kapp-test
<replaced>: create configmap/app-ing (v1) namespace: kapp-test
<replaced>: create configmap/app-svc (v1) namespace: kapp-test
<replaced>: create configmap/import-etcd-into-db (v1) namespace: kapp-test
<replaced>: ---- waiting on 4 changes [0/7 done] ----
<replaced>: ok: reconcile configmap/app-config (v1) namespace: kapp-test
<replaced>: ok: reconcile configmap/app-ing (v1) namespace: kapp-test
<replaced>: ok: reconcile configmap/app-svc (v1) namespace: kapp-test
<replaced>: ok: reconcile configmap/import-etcd-into-db (v1) namespace: kapp-test
<replaced>: ---- applying 1 changes [4/7 done] ----
<replaced>: create configmap/migrations (v1) namespace: kapp-test
<replaced>: ---- waiting on 1 changes [4/7 done] ----
<replaced>: ok: reconcile configmap/migrations (v1) namespace: kapp-test
<replaced>: ---- applying 1 changes [5/7 done] ----
<replaced>: create configmap/app (v1) namespace: kapp-test
<replaced>: ---- waiting on 1 changes [5/7 done] ----
<replaced>: ok: reconcile configmap/app (v1) namespace: kapp-test
<replaced>: ---- applying 1 changes [6/7 done] ----
<replaced>: create configmap/app-health-check (v1) namespace: kapp-test
<replaced>: ---- waiting on 1 changes [6/7 done] ----
<replaced>: ok: reconcile configmap/app-health-check (v1) namespace: kapp-test
<replaced>: ---- applying complete [7/7 done] ----
<replaced>: ---- waiting complete [7/7 done] ----

Succeeded
`))

		require.Equal(t, expectedOutput, out)
	})

	logger.Section("deploy with upsert before delete", func() {
		yaml := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
  annotations: {}
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config2
  annotations:
    kapp.k14s.io/change-rule: "upsert before deleting apps.big.co/import-etcd-into-db" # ref to removed object
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: migrations
  annotations:
    kapp.k14s.io/change-group: "apps.big.co/db-migrations"
    kapp.k14s.io/change-rule: "upsert after upserting apps.big.co/import-etcd-into-db"
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-ing
  annotations:
    kapp.k14s.io/change-group: "apps.big.co/deployment"
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-svc
  annotations:
    kapp.k14s.io/change-group: "apps.big.co/deployment"
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: app
  annotations:
    kapp.k14s.io/change-group: "apps.big.co/deployment"
    kapp.k14s.io/change-rule: "upsert after upserting apps.big.co/db-migrations"
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-health-check
  annotations:
    kapp.k14s.io/change-rule: "upsert after upserting apps.big.co/deployment"
`

		out, _ := kapp.RunWithOpts([]string{"deploy", "--tty", "-f", "-", "-a", name}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml)})

		out = sortActionsInResult(strings.TrimSpace(replaceSpaces(replaceAgeStr(replaceTarget(replaceTs(out))))))
		expectedOutput := strings.TrimSpace(replaceSpaces(`Changes

Namespace  Name                 Kind       Age  Op      Op st.  Wait to    Rs  Ri  $
kapp-test  app-config2          ConfigMap  -    create  -       reconcile  -   -  $
^          import-etcd-into-db  ConfigMap  <replaced>  delete  -       delete     ok  -  $

Op:      1 create, 1 delete, 0 update, 0 noop, 0 exists
Wait to: 1 reconcile, 1 delete, 0 noop

<replaced>: ---- applying 1 changes [0/2 done] ----
<replaced>: create configmap/app-config2 (v1) namespace: kapp-test
<replaced>: ---- waiting on 1 changes [0/2 done] ----
<replaced>: ok: reconcile configmap/app-config2 (v1) namespace: kapp-test
<replaced>: ---- applying 1 changes [1/2 done] ----
<replaced>: delete configmap/import-etcd-into-db (v1) namespace: kapp-test
<replaced>: ---- waiting on 1 changes [1/2 done] ----
<replaced>: ok: delete configmap/import-etcd-into-db (v1) namespace: kapp-test
<replaced>: ---- applying complete [2/2 done] ----
<replaced>: ---- waiting complete [2/2 done] ----

Succeeded
`))

		require.Equal(t, expectedOutput, out)
	})
}

func TestSupportUnblockingChanges(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}
	kubectl := Kubectl{t, env.Namespace, logger}

	name := "test-support-unblocking-changes"
	crdAppName := "crd-app"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
		kapp.Run([]string{"delete", "-a", crdAppName})
	}

	cleanUp()
	defer cleanUp()

	crd := `
---
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
              default: {}
              properties:
                conditions:
                  type: array
                  default: 
                    - {}
                  items:
                    type: object
                    default: {}
                    properties:
                      status:
                        type: string
                        default: "True"
                      type:
                        type: string
                        default: Progressing
  scope: Namespaced
  names:
    plural: crontabs
    singular: crontab
    kind: CronTab`

	yaml := `apiVersion: kapp.k14s.io/v1alpha1
kind: Config
waitRules:
- conditionMatchers:
  - type: Progressing
    status: "True"
    unblockChanges: true
  - type: Available
    status: "True"
    success: true
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
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-cm
  annotations:
    kapp.k14s.io/change-rule: "upsert after upserting cr"`

	go func() {
		for {
			time.Sleep(1 * time.Second)
			exists, err := ClusterResourceExists("ConfigMap", "test-cm", kubectl)
			require.NoError(t, err, "Expected error to not have occurred")
			if exists {
				break
			}
		}
		patch := `[{ "op": "add", "path": "/status", "value": {"conditions": [{"type": "Available", "status": "True"}]}}]`
		PatchClusterResource("CronTab", "my-new-cron-object-1", env.Namespace, patch, kubectl)
	}()

	logger.Section("deploying with supporting unblocking changes", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", crdAppName}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(crd)})

		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml)})

		out = strings.TrimSpace(replaceSpaces(replaceAgeStr(replaceTarget(replaceTs(out)))))
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
<replaced>:  ^ Allowing blocked changes to proceed: Encountered condition Progressing == True: $
<replaced>: ---- applying 1 changes [1/2 done] ----
<replaced>: create configmap/test-cm (v1) namespace: kapp-test
<replaced>: ---- waiting on 2 changes [0/2 done] ----
<replaced>: ok: reconcile configmap/test-cm (v1) namespace: kapp-test
<replaced>: ---- waiting on 1 changes [1/2 done] ----
<replaced>: ok: reconcile crontab/my-new-cron-object-1 (stable.example.com/v1) namespace: kapp-test
<replaced>:  ^ Encountered successful condition Available == True:  (message: )
<replaced>: ---- applying complete [2/2 done] ----
<replaced>: ---- waiting complete [2/2 done] ----

Succeeded
`))

		require.Contains(t, out, expectedOutput)
	})
}

func replaceTs(result string) string {
	return regexp.MustCompile("\\d{1,2}:\\d{1,2}:\\d{1,2}(AM|PM)").ReplaceAllString(result, "<replaced>")
}

func replaceTarget(result string) string {
	return regexp.MustCompile("Target cluster .+\n").ReplaceAllString(result, "")
}

func replaceAgeStr(result string) string {
	return regexp.MustCompile("\\d+(s|m|h|d)\\s+").ReplaceAllString(result, "<replaced>  ")
}

func replaceSpaces(result string) string {
	// result = strings.Replace(result, " ", "_", -1) // useful for debugging
	result = strings.Replace(result, " \n", " $\n", -1) // explicit endline
	return result
}

// Sort action lines with replaced timestamp prefix
func sortActionsInResult(result string) string {
	var lines []string
	var firstIdx, lastIdx int
	for i, line := range strings.Split(result, "\n") {
		lines = append(lines, line)
		if strings.HasPrefix(line, "<replaced>:") {
			if firstIdx == 0 {
				firstIdx = i
			}
			lastIdx = i
		}
	}
	resultLines := append([]string{}, lines[:firstIdx]...)
	resultLines = append(resultLines, sortActionLines(lines[firstIdx:lastIdx])...)
	resultLines = append(resultLines, lines[lastIdx:]...)
	return strings.Join(resultLines, "\n")
}

// Sort actions within each section alphabetically
func sortActionLines(lines []string) []string {
	type actionBucket struct {
		TitleLine  string
		OtherLines []string
	}

	var buckets []*actionBucket
	var lastBucket *actionBucket

	for _, line := range lines {
		if strings.HasPrefix(line, "<replaced>: ----") {
			newBucket := &actionBucket{TitleLine: line}
			buckets = append(buckets, newBucket)
			lastBucket = newBucket
		} else {
			lastBucket.OtherLines = append(lastBucket.OtherLines, line)
		}
	}

	var resultLines []string
	for _, bucket := range buckets {
		resultLines = append(resultLines, bucket.TitleLine)
		sort.Strings(bucket.OtherLines)
		for _, line := range bucket.OtherLines {
			resultLines = append(resultLines, line)
		}
	}
	return resultLines
}
