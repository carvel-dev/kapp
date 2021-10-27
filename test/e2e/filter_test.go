// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFilter(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}
	yaml1 := `
---
apiVersion: v1
kind: Service
metadata:
  name: redis-primary
  labels:
    x: "y"
spec:
  ports:
  - port: 6380
    targetPort: 6380
  selector:
    app: redis
    tier: backend
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: redis-config
  labels:
    x: "z"
data:
  key: value
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: redis-config2
  labels:
    x: "a"
data:
  key: value
`

	name := "test-filter"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
	}

	cleanUp()
	defer cleanUp()
	logger.Section("multiple filter labels", func() {
		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--diff-run",
			"--filter-labels", "x=y,x=z"},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})
		expectedOutput1 := `
Namespace  Name           Kind       Conds.  Age  Op      Op st.  Wait to    Rs  Ri  
kapp-test  redis-config   ConfigMap  -       -    create  -       reconcile  -   -  
^          redis-primary  Service    -       -    create  -       reconcile  -   -  

Op:      2 create, 0 delete, 0 update, 0 noop
Wait to: 2 reconcile, 0 delete, 0 noop
`
		require.Contains(t, out, expectedOutput1, "Did not find expected diff output")
	})

	logger.Section("not equal filter label", func() {
		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--diff-run",
			"--filter-labels", "x!=y"},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})
		expectedOutput2 := `
Namespace  Name           Kind       Conds.  Age  Op      Op st.  Wait to    Rs  Ri  
kapp-test  redis-config   ConfigMap  -       -    create  -       reconcile  -   -  
^          redis-config2  ConfigMap  -       -    create  -       reconcile  -   -  

Op:      2 create, 0 delete, 0 update, 0 noop
Wait to: 2 reconcile, 0 delete, 0 noop
`
		require.Contains(t, out, expectedOutput2, "Did not find expected diff output")
	})

	logger.Section("test filter flag", func() {
		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--diff-run",
			"--filter", `{"resource":{"kinds":["Service"]}}`},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})
		expectedOutput3 := `
Namespace  Name           Kind     Conds.  Age  Op      Op st.  Wait to    Rs  Ri  
kapp-test  redis-primary  Service  -       -    create  -       reconcile  -   -  

Op:      1 create, 0 delete, 0 update, 0 noop
Wait to: 1 reconcile, 0 delete, 0 noop
`
		require.Contains(t, out, expectedOutput3, "Did not find expected diff output")
	})

	logger.Section("test multiple filter flags together", func() {
		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--diff-run",
			"--filter-kind", "ConfigMap",
			"--filter-labels", "x=a"},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})

		expectedOutput4 := `
Namespace  Name           Kind       Conds.  Age  Op      Op st.  Wait to    Rs  Ri  
kapp-test  redis-config2  ConfigMap  -       -    create  -       reconcile  -   -  

Op:      1 create, 0 delete, 0 update, 0 noop
Wait to: 1 reconcile, 0 delete, 0 noop
`
		require.Contains(t, out, expectedOutput4, "Did not find expected diff output")
	})
}
