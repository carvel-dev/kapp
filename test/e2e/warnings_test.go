// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func TestWarningsFlag(t *testing.T) {
	minorVersion, err := getServerMinorVersion()
	require.NoErrorf(t, err, "Error getting k8s server minor version")

	if minorVersion < 19 {
		t.Skip("Skipping test as warnings weren't introduced before v1.19")
	}
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, Logger{}}
	crdName := "test-no-warnings-crd"
	crName1 := "test-no-warnings-cr1"
	crName2 := "test-no-warnings-cr2"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", crName1})
		kapp.Run([]string{"delete", "-a", crName2})
		kapp.Run([]string{"delete", "-a", crdName})
	}

	cleanUp()
	defer cleanUp()
	customWarning := "example.com/v1alpha1 CronTab is deprecated; use example.com/v1 CronTab"

	crdYaml := `
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: crontabs.stable.example.com
spec:
  group: stable.example.com
  versions:
  - name: v1alpha1
    served: true
    storage: false
    deprecated: true
    deprecationWarning: "<customWarning>"
    schema:
      openAPIV3Schema:
        type: object
  - name: v1
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
  scope: Namespaced
  names:
    plural: crontabs
    singular: crontab
    kind: CronTab
`
	crYaml := `
apiVersion: "stable.example.com/v1alpha1"
kind: CronTab
metadata:
  name: <cr-name>
spec:
  cronSpec: "* * * * */5"
  image: my-awesome-cron-image
`

	logger.Section("deploying crd with deprecated version", func() {
		yaml := strings.Replace(crdYaml, "<customWarning>", customWarning, 1)
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", crdName},
			RunOpts{StdinReader: strings.NewReader(yaml)})
	})

	logger.Section("deploying without --warnings flag", func() {
		yaml := strings.Replace(crYaml, "<cr-name>", "cr-1", 1)
		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", crName1},
			RunOpts{StdinReader: strings.NewReader(yaml)})

		expectedOutput := `
Changes

Namespace  Name  Kind     Conds.  Age  Op      Op st.  Wait to    Rs  Ri  
kapp-test  cr-1  CronTab  -       -    create  -       reconcile  -   -  

Op:      1 create, 0 delete, 0 update, 0 noop
Wait to: 1 reconcile, 0 delete, 0 noop

<replaced>: ---- applying 1 changes [0/1 done] ----
Warning: <custom-warning>
<replaced>: create crontab/cr-1 (stable.example.com/v1alpha1) namespace: kapp-test
<replaced>: ---- waiting on 1 changes [0/1 done] ----
<replaced>: ok: reconcile crontab/cr-1 (stable.example.com/v1alpha1) namespace: kapp-test
<replaced>: ---- applying complete [1/1 done] ----
<replaced>: ---- waiting complete [1/1 done] ----

Succeeded`

		out = strings.TrimSpace(replaceTarget(replaceSpaces(replaceTs(out))))
		expectedOutput = strings.Replace(expectedOutput, "<custom-warning>", customWarning, 1)
		expectedOutput = strings.TrimSpace(replaceSpaces(expectedOutput))

		require.Equal(t, expectedOutput, out)
	})

	logger.Section("deploying with --warnings flag", func() {
		yaml := strings.Replace(crYaml, "<cr-name>", "cr-2", 1)

		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", crName2, "--warnings=false"},
			RunOpts{StdinReader: strings.NewReader(yaml)})

		expectedOutput := `
Changes

Namespace  Name  Kind     Conds.  Age  Op      Op st.  Wait to    Rs  Ri  
kapp-test  cr-2  CronTab  -       -    create  -       reconcile  -   -  

Op:      1 create, 0 delete, 0 update, 0 noop
Wait to: 1 reconcile, 0 delete, 0 noop

<replaced>: ---- applying 1 changes [0/1 done] ----
<replaced>: create crontab/cr-2 (stable.example.com/v1alpha1) namespace: kapp-test
<replaced>: ---- waiting on 1 changes [0/1 done] ----
<replaced>: ok: reconcile crontab/cr-2 (stable.example.com/v1alpha1) namespace: kapp-test
<replaced>: ---- applying complete [1/1 done] ----
<replaced>: ---- waiting complete [1/1 done] ----

Succeeded`

		out = strings.TrimSpace(replaceTarget(replaceSpaces(replaceTs(out))))
		expectedOutput = strings.TrimSpace(replaceSpaces(expectedOutput))

		require.Equal(t, expectedOutput, out)
	})
}

func getServerMinorVersion() (minorVersion int, err error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	config, err := kubeConfig.ClientConfig()
	if err != nil {
		return minorVersion, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return minorVersion, err
	}

	sv, err := clientset.Discovery().ServerVersion()
	if err != nil {
		return minorVersion, err
	}
	minorVersion, err = strconv.Atoi(sv.Minor)
	if err != nil {
		return minorVersion, err
	}

	return minorVersion, err
}
