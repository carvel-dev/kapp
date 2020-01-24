package e2e

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"testing"

	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
)

func TestTemplate(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}
	kubectl := Kubectl{t, env.Namespace, logger}

	depYAML := `---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dep
spec:
  selector:
    matchLabels:
      app: dep
  replicas: 1
  template:
    metadata:
      labels:
        app: dep
    spec:
      containers:
      - name: echo
        image: hashicorp/http-echo
        args:
        - -listen=:80
        - -text=hello
        ports:
        - containerPort: 80
        envFrom:
        - configMapRef:
            name: config
      initContainers:
      - name: echo-init
        image: hashicorp/http-echo
        args:
        - -version
        envFrom:
        - configMapRef:
            name: config
      volumes:
      - name: vol1
        secret:
          secretName: secret
`

	yaml1 := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: config
  annotations:
    kapp.k14s.io/versioned: ""
data:
  key1: val1
---
apiVersion: v1
kind: Secret
metadata:
  name: secret
  annotations:
    kapp.k14s.io/versioned: ""
data:
  key1: val1
` + depYAML

	yaml2 := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: config
  annotations:
    kapp.k14s.io/versioned: ""
data:
  key1: val2
---
apiVersion: v1
kind: Secret
metadata:
  name: secret
  annotations:
    kapp.k14s.io/versioned: ""
data:
  key1: val2
` + depYAML

	expectedYAML1Diff := `
--- create configmap/config-ver-1 (v1) namespace: kapp-test
      0 + apiVersion: v1
      1 + data:
      2 +   key1: val1
      3 + kind: ConfigMap
      4 + metadata:
      5 +   annotations:
      6 +     kapp.k14s.io/versioned: ""
      7 +   labels:
      8 +     -replaced-
      9 +     -replaced-
     10 +   name: config-ver-1
     11 +   namespace: kapp-test
     12 + 
--- create secret/secret-ver-1 (v1) namespace: kapp-test
      0 + apiVersion: v1
      1 + data:
      2 +   key1: val1
      3 + kind: Secret
      4 + metadata:
      5 +   annotations:
      6 +     kapp.k14s.io/versioned: ""
      7 +   labels:
      8 +     -replaced-
      9 +     -replaced-
     10 +   name: secret-ver-1
     11 +   namespace: kapp-test
     12 + 
--- create deployment/dep (apps/v1) namespace: kapp-test
      0 + apiVersion: apps/v1
      1 + kind: Deployment
      2 + metadata:
      3 +   labels:
      4 +     -replaced-
      5 +     -replaced-
      6 +   name: dep
      7 +   namespace: kapp-test
      8 + spec:
      9 +   replicas: 1
     10 +   selector:
     11 +     matchLabels:
     12 +       app: dep
     13 +       -replaced-
     14 +   template:
     15 +     metadata:
     16 +       labels:
     17 +         app: dep
     18 +         -replaced-
     19 +         -replaced-
     20 +     spec:
     21 +       containers:
     22 +       - args:
     23 +         - -listen=:80
     24 +         - -text=hello
     25 +         envFrom:
     26 +         - configMapRef:
     27 +             name: config-ver-1
     28 +         image: hashicorp/http-echo
     29 +         name: echo
     30 +         ports:
     31 +         - containerPort: 80
     32 +       initContainers:
     33 +       - args:
     34 +         - -version
     35 +         envFrom:
     36 +         - configMapRef:
     37 +             name: config-ver-1
     38 +         image: hashicorp/http-echo
     39 +         name: echo-init
     40 +       volumes:
     41 +       - name: vol1
     42 +         secret:
     43 +           secretName: secret-ver-1
     44 + 
`

	expectedYAML2Diff := `
--- create configmap/config-ver-2 (v1) namespace: kapp-test
  ...
  1,  1   data:
  2     -   key1: val1
      2 +   key1: val2
  3,  3   kind: ConfigMap
  4,  4   metadata:
--- create secret/secret-ver-2 (v1) namespace: kapp-test
  ...
  1,  1   data:
  2     -   key1: val1
      2 +   key1: val2
  3,  3   kind: Secret
  4,  4   metadata:
--- update deployment/dep (apps/v1) namespace: kapp-test
  ...
 33, 33           - configMapRef:
 34     -             name: config-ver-1
     34 +             name: config-ver-2
 35, 35           image: hashicorp/http-echo
 36, 36           name: echo
  ...
 43, 43           - configMapRef:
 44     -             name: config-ver-1
     44 +             name: config-ver-2
 45, 45           image: hashicorp/http-echo
 46, 46           name: echo-init
  ...
 49, 49           secret:
 50     -           secretName: secret-ver-1
     50 +           secretName: secret-ver-2
 51, 51   status:
 52, 52     availableReplicas: 1
`

	name := "test-template"
	cleanUp := func() {
		kapp.RunWithOpts([]string{"delete", "-a", name}, RunOpts{AllowError: true})
	}

	depPath := []interface{}{"spec", "template", "spec", "containers", 0, "envFrom", 0, "configMapRef", "name"}

	cleanUp()
	defer cleanUp()

	logger.Section("deploy initial", func() {
		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--diff-changes", "--tty", "--diff-mask=false"},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})
		checkChangesOutput(t, out, expectedYAML1Diff)

		dep := NewPresentClusterResource("deployment", "dep", env.Namespace, kubectl)
		NewPresentClusterResource("configmap", "config-ver-1", env.Namespace, kubectl)

		val := dep.RawPath(ctlres.NewPathFromInterfaces(depPath))
		if !reflect.DeepEqual(val, "config-ver-1") {
			t.Fatalf("Expected value to be updated")
		}
	})

	logger.Section("deploy update that changes configmap", func() {
		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--diff-changes", "--tty", "--diff-mask=false"},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml2)})
		checkChangesOutput(t, out, expectedYAML2Diff)

		dep := NewPresentClusterResource("deployment", "dep", env.Namespace, kubectl)
		NewPresentClusterResource("configmap", "config-ver-1", env.Namespace, kubectl)
		NewPresentClusterResource("configmap", "config-ver-2", env.Namespace, kubectl)

		val := dep.RawPath(ctlres.NewPathFromInterfaces(depPath))
		if !reflect.DeepEqual(val, "config-ver-2") {
			t.Fatalf("Expected value to be updated")
		}
	})

	logger.Section("deploy update that has no changes", func() {
		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--diff-changes", "--tty", "--diff-mask=false"},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml2)})
		checkChangesOutput(t, out, "")

		dep := NewPresentClusterResource("deployment", "dep", env.Namespace, kubectl)
		NewPresentClusterResource("configmap", "config-ver-1", env.Namespace, kubectl)
		NewPresentClusterResource("configmap", "config-ver-2", env.Namespace, kubectl)

		val := dep.RawPath(ctlres.NewPathFromInterfaces(depPath))
		if !reflect.DeepEqual(val, "config-ver-2") {
			t.Fatalf("Expected value to be updated")
		}
	})

	// TODO deploy via patch or filter
}

func checkChangesOutput(t *testing.T, actualOutput, expectedOutput string) {
	replaceAnns := regexp.MustCompile("kapp\\.k14s\\.io\\/(app|association): .+")
	actualOutput = replaceAnns.ReplaceAllString(actualOutput, "-replaced-")

	actualOutput = strings.TrimSpace(strings.Split(replaceTarget(actualOutput), "Changes")[0])
	expectedOutput = strings.TrimSpace(expectedOutput)

	// Useful for debugging:
	// printLines("actual", actualOutput)
	// printLines("expected", expectedOutput)

	if actualOutput != expectedOutput {
		t.Fatalf("Expected output to match:  %d >>>%s<<< vs %d >>>%s<<<",
			len(actualOutput), actualOutput, len(expectedOutput), expectedOutput)
	}
}

func printLines(heading, str string) {
	fmt.Printf("%s:\n", heading)
	for _, line := range strings.Split(str, "\n") {
		fmt.Printf(">>>%s<<<\n", line)
	}
}
