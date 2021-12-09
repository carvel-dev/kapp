package e2e

import (
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestAppListing(t *testing.T) {
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
`
	yaml2 := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: redis-config
  labels:
    x: "z"
data:
  key: value
`

	name := "test-app-list"
	name2 := "test-app-list-2"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
		kapp.Run([]string{"delete", "-a", name2})
	}

	cleanUp()
	defer cleanUp()
	logger.Section("App listing and filter label", func() {
		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name,
			"--labels", "x=y"}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})
		expectedOutput1 := `
Namespace  Name           Kind     Conds.  Age  Op      Op st.  Wait to    Rs  Ri  
kapp-test  redis-primary  Service  -       -    create  -       reconcile  -   -  

Op:      1 create, 0 delete, 0 update, 0 noop, 0 exists
Wait to: 1 reconcile, 0 delete, 0 noop
`
		require.Contains(t, out, expectedOutput1, "Did not find expected diff output")

		out2, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name2,
			"--labels", "a=b"}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml2)})
		expectedOutput2 := `
Namespace  Name          Kind       Conds.  Age  Op      Op st.  Wait to    Rs  Ri  
kapp-test  redis-config  ConfigMap  -       -    create  -       reconcile  -   -  

Op:      1 create, 0 delete, 0 update, 0 noop, 0 exists
Wait to: 1 reconcile, 0 delete, 0 noop
`
		require.Contains(t, out2, expectedOutput2, "Did not find expected diff output")

		listedApps, _ := kapp.RunWithOpts([]string{"ls"}, RunOpts{Interactive: true})

		expectedAppsList := `
Apps in namespace 'kapp-test'

Name             Namespaces  Lcs   Lca  
test-app-list    kapp-test   true  0s  
test-app-list-2  kapp-test   true  0s  

Lcs: Last Change Successful
Lca: Last Change Age

2 apps
`
		require.Contains(t, listedApps, expectedAppsList, "Did not find expected diff output")

		filteredApps, _ := kapp.RunWithOpts([]string{"ls", "--filter-labels", "x=y"}, RunOpts{Interactive: true})

		expectedFilteredApps := `
Apps in namespace 'kapp-test'

Name           Namespaces  Lcs   Lca  
test-app-list  kapp-test   true  0s  

Lcs: Last Change Successful
Lca: Last Change Age

1 apps

Succeeded
`
		require.Contains(t, filteredApps, expectedFilteredApps, "Did not find expected diff output")
	})
}
