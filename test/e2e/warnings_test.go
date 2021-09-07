// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	yellowColor = "\u001b[33;1m"
	resetColor  = "\u001b[0m"
)

func TestWarningsFlag(t *testing.T) {
	minorVersion, err := getServerMinorVersion()
	if err != nil {
		t.Fatalf("Error getting k8s server minor version, %v", err)
	}
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
	outputWarning := fmt.Sprintf("\n%sWarning:%s %s", yellowColor, resetColor, customWarning)

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

	expectedOutputTemplate := `
Target cluster 'https://127.0.0.1:60817' (nodes: minikube)

Changes

Namespace  Name  Kind     Conds.  Age  Op      Op st.  Wait to    Rs  Ri  
kapp-test  <cr-name>  CronTab  -       -    create  -       reconcile  -   -  

Op:      1 create, 0 delete, 0 update, 0 noop
Wait to: 1 reconcile, 0 delete, 0 noop

<replaced>: ---- applying 1 changes [0/1 done] ----<output-warning>
<replaced>: create crontab/<cr-name> (stable.example.com/v1alpha1) namespace: kapp-test
<replaced>: ---- waiting on 1 changes [0/1 done] ----
<replaced>: ok: reconcile crontab/<cr-name> (stable.example.com/v1alpha1) namespace: kapp-test
<replaced>: ---- applying complete [1/1 done] ----
<replaced>: ---- waiting complete [1/1 done] ----

Succeeded`

	logger.Section("deploying crd with deprecated version", func() {
		yaml := strings.Replace(crdYaml, "<customWarning>", customWarning, 1)
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", crdName},
			RunOpts{StdinReader: strings.NewReader(yaml)})
	})

	logger.Section("deploying without --warnings flag", func() {
		yaml := strings.Replace(crYaml, "<cr-name>", "cr-1", 1)

		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", crName1},
			RunOpts{StdinReader: strings.NewReader(yaml)})

		expectedOutput := strings.Replace(expectedOutputTemplate, "<cr-name>", "cr-1", 3)

		out = strings.TrimSpace(replaceTarget(replaceSpaces(replaceTs(out))))
		expectedOutput = strings.Replace(expectedOutput, "<output-warning>", outputWarning, 1)
		expectedOutput = strings.TrimSpace(replaceTarget(replaceSpaces(expectedOutput)))

		if expectedOutput != out {
			t.Fatalf("Expected output with warning >>%s<<, but got >>%s<<\n", expectedOutput, out)
		}
	})

	logger.Section("deploying with --warnings flag", func() {
		yaml := strings.Replace(crYaml, "<cr-name>", "cr-2", 1)

		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", crName2, "--warnings=false"},
			RunOpts{StdinReader: strings.NewReader(yaml)})

		expectedOutput := strings.Replace(expectedOutputTemplate, "<cr-name>", "cr-2", 3)

		out = strings.TrimSpace(replaceTarget(replaceSpaces(replaceTs(out))))
		expectedOutput = strings.Replace(expectedOutput, "<output-warning>", "", 1)
		expectedOutput = strings.TrimSpace(replaceTarget(replaceSpaces(expectedOutput)))
		if expectedOutput != out {
			t.Fatalf("Expected output without warning >>%s<< but got >>%s<<\n", expectedOutput, out)
		}
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
