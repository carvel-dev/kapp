package e2e

import (
	"regexp"
	"strings"
	"testing"
)

func TestOrder(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}

	name := "test-order"
	cleanUp := func() {
		kapp.RunWithOpts([]string{"delete", "-a", name}, RunOpts{AllowError: true})
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

		out = strings.TrimSpace(replaceSpaces(replaceTarget(replaceTs(out))))
		expectedOutput := strings.TrimSpace(replaceSpaces(`Changes

Namespace  Name                 Kind       Conds.  Age  Op      Wait to    Rs  Ri  $
kapp-test  app                  ConfigMap  -       -    create  reconcile  -   -  $
^          app-config           ConfigMap  -       -    create  reconcile  -   -  $
^          app-health-check     ConfigMap  -       -    create  reconcile  -   -  $
^          app-ing              ConfigMap  -       -    create  reconcile  -   -  $
^          app-svc              ConfigMap  -       -    create  reconcile  -   -  $
^          import-etcd-into-db  ConfigMap  -       -    create  reconcile  -   -  $
^          migrations           ConfigMap  -       -    create  reconcile  -   -  $

Op:      7 create, 0 delete, 0 update, 0 noop
Wait to: 7 reconcile, 0 delete, 0 noop

<replaced>: ---- applying 4 changes [0/7 done] ----
<replaced>: create configmap/app-config (v1) namespace: kapp-test
<replaced>: create configmap/import-etcd-into-db (v1) namespace: kapp-test
<replaced>: create configmap/app-ing (v1) namespace: kapp-test
<replaced>: create configmap/app-svc (v1) namespace: kapp-test
<replaced>: ---- waiting on 4 changes [0/7 done] ----
<replaced>: ok: reconcile configmap/app-config (v1) namespace: kapp-test
<replaced>: ok: reconcile configmap/import-etcd-into-db (v1) namespace: kapp-test
<replaced>: ok: reconcile configmap/app-ing (v1) namespace: kapp-test
<replaced>: ok: reconcile configmap/app-svc (v1) namespace: kapp-test
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

		if out != expectedOutput {
			t.Fatalf("Expected output to equal (%d) >>>%s<<< but was (%d) >>>%s<<<",
				len(expectedOutput), expectedOutput, len(out), out)
		}
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

		out = strings.TrimSpace(replaceSpaces(replaceAgeStr(replaceTarget(replaceTs(out)))))
		expectedOutput := strings.TrimSpace(replaceSpaces(`Changes

Namespace  Name                 Kind       Conds.  Age  Op      Wait to    Rs  Ri  $
kapp-test  app-config2          ConfigMap  -       -    create  reconcile  -   -  $
^          import-etcd-into-db  ConfigMap  -       <replaced>  delete  delete     ok  -  $

Op:      1 create, 1 delete, 0 update, 0 noop
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

		if out != expectedOutput {
			t.Fatalf("Expected output to equal (%d) >>>%s<<< but was (%d) >>>%s<<<",
				len(expectedOutput), expectedOutput, len(out), out)
		}
	})
}

func replaceTs(result string) string {
	return regexp.MustCompile("\\d{1,2}:\\d{1,2}:\\d{1,2}(AM|PM)").ReplaceAllString(result, "<replaced>")
}

func replaceTarget(result string) string {
	return regexp.MustCompile("Target cluster .+\n").ReplaceAllString(result, "")
}

func replaceAgeStr(result string) string {
	return regexp.MustCompile("\\d+(s|m|h)\\s+").ReplaceAllString(result, "<replaced>  ")
}

func replaceSpaces(result string) string {
	// result = strings.Replace(result, " ", "_", -1) // useful for debugging
	result = strings.Replace(result, " \n", " $\n", -1) // explicit endline
	return result
}
