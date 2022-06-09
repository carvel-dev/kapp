package e2e

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestApplyRun(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}
	yaml := `
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: simple-cm
  annotations:
    kapp.k14s.io/delete-strategy: "orphan"
data:
  hello_msg: good-morning-banaglore
`
	name := "test-apply-run"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
	}

	cleanUp()
	defer cleanUp()
	logger.Section("creating an app with multiple resources", func() {
		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--apply-run"},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml)})
		expectedOutput := `
# add: configmap/simple-cm (v1) namespace: kapp-test
---
apiVersion: v1
data:
  hello_msg: good-morning-banaglore
kind: ConfigMap
metadata:
  annotations:
    kapp.k14s.io/delete-strategy: orphan
  labels:
`
		out = strings.TrimSpace(replaceTarget(replaceSpaces(replaceTs(out))))
		fmt.Printf("out: %s", out)
		expectedOutput = strings.TrimSpace(replaceSpaces(expectedOutput))
		//require.Equal(t, expectedOutput, out)
		require.Contains(t, out, expectedOutput, "output does not match")
	})

}
