package e2e

import (
	"regexp"
	"sort"
	"strings"
	"testing"
)

func TestOrderBasic(t *testing.T) {
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

		out = sortActionsInResult(strings.TrimSpace(replaceSpaces(replaceTarget(replaceTs(out)))))
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

		out = sortActionsInResult(strings.TrimSpace(replaceSpaces(replaceAgeStr(replaceTarget(replaceTs(out))))))
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

func TestOrderWithNoops(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}

	name := "test-order-noops"
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
  name: first
  annotations:
    kapp.k14s.io/change-group: "first"
data:
  key: val
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: second
  annotations:
    kapp.k14s.io/change-group: "second"
    kapp.k14s.io/change-rule: "upsert after upserting first"
data:
  key: val
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: third
  annotations:
    kapp.k14s.io/change-rule: "upsert after upserting second"
data:
  key: val
`

		out, _ := kapp.RunWithOpts([]string{"deploy", "--tty", "-f", "-", "-a", name}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})

		out = sortActionsInResult(strings.TrimSpace(replaceSpaces(replaceTarget(replaceTs(out)))))
		expectedOutput := strings.TrimSpace(replaceSpaces(`Changes

Namespace  Name    Kind       Conds.  Age  Op      Wait to    Rs  Ri  $
kapp-test  first   ConfigMap  -       -    create  reconcile  -   -  $
^          second  ConfigMap  -       -    create  reconcile  -   -  $
^          third   ConfigMap  -       -    create  reconcile  -   -  $

Op:      3 create, 0 delete, 0 update, 0 noop
Wait to: 3 reconcile, 0 delete, 0 noop

<replaced>: ---- applying 1 changes [0/3 done] ----
<replaced>: create configmap/first (v1) namespace: kapp-test
<replaced>: ---- waiting on 1 changes [0/3 done] ----
<replaced>: ok: reconcile configmap/first (v1) namespace: kapp-test
<replaced>: ---- applying 1 changes [1/3 done] ----
<replaced>: create configmap/second (v1) namespace: kapp-test
<replaced>: ---- waiting on 1 changes [1/3 done] ----
<replaced>: ok: reconcile configmap/second (v1) namespace: kapp-test
<replaced>: ---- applying 1 changes [2/3 done] ----
<replaced>: create configmap/third (v1) namespace: kapp-test
<replaced>: ---- waiting on 1 changes [2/3 done] ----
<replaced>: ok: reconcile configmap/third (v1) namespace: kapp-test
<replaced>: ---- applying complete [3/3 done] ----
<replaced>: ---- waiting complete [3/3 done] ----

Succeeded
`))

		if out != expectedOutput {
			t.Fatalf("Expected output to equal (%d) >>>%s<<< but was (%d) >>>%s<<<",
				len(expectedOutput), expectedOutput, len(out), out)
		}
	})

	logger.Section("deploy update that waits on noop, that waits on update", func() {
		yaml := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: first
  annotations:
    kapp.k14s.io/change-group: "first"
data:
  key: val1
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: second
  annotations:
    kapp.k14s.io/change-group: "second"
    kapp.k14s.io/change-rule: "upsert after upserting first"
data:
  key: val
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: third
  annotations:
    kapp.k14s.io/change-rule: "upsert after upserting second"
data:
  key: val1
`

		out, _ := kapp.RunWithOpts([]string{"deploy", "--tty", "-f", "-", "-a", name}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml)})

		out = sortActionsInResult(strings.TrimSpace(replaceSpaces(replaceAgeStr(replaceTarget(replaceTs(out))))))
		expectedOutput := strings.TrimSpace(replaceSpaces(`Changes

Namespace  Name    Kind       Conds.  Age  Op      Wait to    Rs  Ri  $
kapp-test  first   ConfigMap  -       <replaced>  update  reconcile  ok  -  $
^          second  ConfigMap  -       <replaced>  -       reconcile  ok  -  $
^          third   ConfigMap  -       <replaced>  update  reconcile  ok  -  $

Op:      0 create, 0 delete, 2 update, 1 noop
Wait to: 3 reconcile, 0 delete, 0 noop

<replaced>: ---- applying 1 changes [0/3 done] ----
<replaced>: update configmap/first (v1) namespace: kapp-test
<replaced>: ---- waiting on 1 changes [0/3 done] ----
<replaced>: ok: reconcile configmap/first (v1) namespace: kapp-test
<replaced>: ---- applying 1 changes [1/3 done] ----
<replaced>: noop configmap/second (v1) namespace: kapp-test
<replaced>: ---- waiting on 1 changes [1/3 done] ----
<replaced>: ok: reconcile configmap/second (v1) namespace: kapp-test
<replaced>: ---- applying 1 changes [2/3 done] ----
<replaced>: update configmap/third (v1) namespace: kapp-test
<replaced>: ---- waiting on 1 changes [2/3 done] ----
<replaced>: ok: reconcile configmap/third (v1) namespace: kapp-test
<replaced>: ---- applying complete [3/3 done] ----
<replaced>: ---- waiting complete [3/3 done] ----

Succeeded
`))

		if out != expectedOutput {
			t.Fatalf("Expected output to equal (%d) >>>%s<<< but was (%d) >>>%s<<<",
				len(expectedOutput), expectedOutput, len(out), out)
		}
	})

	logger.Section("deploy update that blocks other noops", func() {
		yaml := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: first
  annotations:
    kapp.k14s.io/change-group: "first"
data:
  key: val2
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: second
  annotations:
    kapp.k14s.io/change-group: "second"
    kapp.k14s.io/change-rule: "upsert after upserting first"
data:
  key: val
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: third
  annotations:
    kapp.k14s.io/change-rule: "upsert after upserting second"
data:
  key: val1
`

		out, _ := kapp.RunWithOpts([]string{"deploy", "--tty", "-f", "-", "-a", name}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml)})

		out = sortActionsInResult(strings.TrimSpace(replaceSpaces(replaceAgeStr(replaceTarget(replaceTs(out))))))
		expectedOutput := strings.TrimSpace(replaceSpaces(`Changes

Namespace  Name    Kind       Conds.  Age  Op      Wait to    Rs  Ri  $
kapp-test  first   ConfigMap  -       <replaced>  update  reconcile  ok  -  $
^          second  ConfigMap  -       <replaced>  -       reconcile  ok  -  $
^          third   ConfigMap  -       <replaced>  -       reconcile  ok  -  $

Op:      0 create, 0 delete, 1 update, 2 noop
Wait to: 3 reconcile, 0 delete, 0 noop

<replaced>: ---- applying 1 changes [0/3 done] ----
<replaced>: update configmap/first (v1) namespace: kapp-test
<replaced>: ---- waiting on 1 changes [0/3 done] ----
<replaced>: ok: reconcile configmap/first (v1) namespace: kapp-test
<replaced>: ---- applying 1 changes [1/3 done] ----
<replaced>: noop configmap/second (v1) namespace: kapp-test
<replaced>: ---- waiting on 1 changes [1/3 done] ----
<replaced>: ok: reconcile configmap/second (v1) namespace: kapp-test
<replaced>: ---- applying 1 changes [2/3 done] ----
<replaced>: noop configmap/third (v1) namespace: kapp-test
<replaced>: ---- waiting on 1 changes [2/3 done] ----
<replaced>: ok: reconcile configmap/third (v1) namespace: kapp-test
<replaced>: ---- applying complete [3/3 done] ----
<replaced>: ---- waiting complete [3/3 done] ----

Succeeded
`))

		if out != expectedOutput {
			t.Fatalf("Expected output to equal (%d) >>>%s<<< but was (%d) >>>%s<<<",
				len(expectedOutput), expectedOutput, len(out), out)
		}
	})

	logger.Section("deploy update waits on noop, hence it's not present", func() {
		yaml := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: first
  annotations:
    kapp.k14s.io/change-group: "first"
data:
  key: val2
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: second
  annotations:
    kapp.k14s.io/change-group: "second"
    kapp.k14s.io/change-rule: "upsert after upserting first"
data:
  key: val3
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: third
  annotations:
    kapp.k14s.io/change-rule: "upsert after upserting second"
data:
  key: val1
`

		out, _ := kapp.RunWithOpts([]string{"deploy", "--tty", "-f", "-", "-a", name}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml)})

		out = sortActionsInResult(strings.TrimSpace(replaceSpaces(replaceAgeStr(replaceTarget(replaceTs(out))))))
		expectedOutput := strings.TrimSpace(replaceSpaces(`Changes

Namespace  Name    Kind       Conds.  Age  Op      Wait to    Rs  Ri  $
kapp-test  second  ConfigMap  -       <replaced>  update  reconcile  ok  -  $
^          third   ConfigMap  -       <replaced>  -       reconcile  ok  -  $

Op:      0 create, 0 delete, 1 update, 1 noop
Wait to: 2 reconcile, 0 delete, 0 noop

<replaced>: ---- applying 1 changes [0/2 done] ----
<replaced>: update configmap/second (v1) namespace: kapp-test
<replaced>: ---- waiting on 1 changes [0/2 done] ----
<replaced>: ok: reconcile configmap/second (v1) namespace: kapp-test
<replaced>: ---- applying 1 changes [1/2 done] ----
<replaced>: noop configmap/third (v1) namespace: kapp-test
<replaced>: ---- waiting on 1 changes [1/2 done] ----
<replaced>: ok: reconcile configmap/third (v1) namespace: kapp-test
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
