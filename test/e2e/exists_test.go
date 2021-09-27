// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func TestExistsOpWithPlaceholderAnn(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}
	clientset, err := getKubernetesClientset()
	if err != nil {
		t.Fatalf("Error creating client, %v", err)
	}
	app := `
apiVersion: v1
kind: Namespace
metadata:
  name: kapp-ns
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: external
  namespace: kapp-ns
  annotations:
    kapp.k14s.io/placeholder: ""
`

	name := "app"
	externalResourceName := "external"
	externalResourceNamespace := "kapp-ns"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
		deleteExternalResource(clientset, externalResourceName, externalResourceNamespace)
	}
	cleanUp()
	defer cleanUp()

	logger.Section("deploying external resource", func() {
		go deployExternalResource(clientset, externalResourceName, externalResourceNamespace)
	})

	logger.Section("deploying app with placeholder annotation", func() {
		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name},
			RunOpts{StdinReader: strings.NewReader(app)})
		expectedOutput := `
Changes

Namespace  Name      Kind       Conds.  Age  Op      Op st.  Wait to    Rs  Ri  $
(cluster)  kapp-ns   Namespace  -       -    create  -       reconcile  -   -  $
kapp-ns    external  ConfigMap  -       -    exists  -       reconcile  -   -  $

Op:      1 create, 0 delete, 0 update, 0 noop
Wait to: 2 reconcile, 0 delete, 0 noop

<replaced>: ---- applying 1 changes [0/2 done] ----
<replaced>: create namespace/kapp-ns (v1) cluster
<replaced>: ---- waiting on 1 changes [0/2 done] ----
<replaced>: ok: reconcile namespace/kapp-ns (v1) cluster
<replaced>: ---- applying 1 changes [1/2 done] ----
<replaced>: exists configmap/external (v1) namespace: kapp-ns
<replaced>:  ^ Retryable error: Placeholder resource doesn't exists
<replaced>: exists configmap/external (v1) namespace: kapp-ns
<replaced>: ---- waiting on 1 changes [1/2 done] ----
<replaced>: ok: reconcile configmap/external (v1) namespace: kapp-ns
<replaced>: ---- applying complete [2/2 done] ----
<replaced>: ---- waiting complete [2/2 done] ----

Succeeded`

		out = strings.TrimSpace(replaceTarget(replaceSpaces(replaceTs(out))))
		expectedOutput = strings.TrimSpace(replaceSpaces(expectedOutput))
		if expectedOutput != out {
			t.Fatalf("Expected output to be >>%s<<, but got >>%s<<\n", expectedOutput, out)
		}
	})

	logger.Section("inspecting app", func() {
		out, _ := kapp.RunWithOpts([]string{"inspect", "-a", name},
			RunOpts{StdinReader: strings.NewReader(app)})
		out = strings.TrimSpace(replaceTarget(replaceSpaces(replaceTs(out))))

		expectedOutput := `
Resources in app 'app'

Namespace  Name     Kind       Owner  Conds.  Rs  Ri  Age  $
(cluster)  kapp-ns  Namespace  kapp   -       ok  -   2s  $

Rs: Reconcile state
Ri: Reconcile information

1 resources

Succeeded`

		expectedOutput = strings.TrimSpace(replaceSpaces(expectedOutput))
		if expectedOutput != out {
			t.Fatalf("Expected output to be >>%s<<, but got >>%s<<\n", expectedOutput, out)
		}
	})
}

func deployExternalResource(clientset *kubernetes.Clientset, name string, namespace string) {
	time.Sleep(2 * time.Second)
	metaData := metav1.ObjectMeta{Name: name}
	data := map[string]string{}
	data["key1"] = "value1"
	configMap := v1.ConfigMap{
		ObjectMeta: metaData,
		Data:       data,
	}
	_, err := clientset.CoreV1().ConfigMaps(namespace).Create(context.TODO(), &configMap, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}
}

func deleteExternalResource(clientset *kubernetes.Clientset, name string, namespace string) {
	err := clientset.CoreV1().ConfigMaps(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		if err.Error() != fmt.Sprintf(`configmaps "%s" not found`, name) {
			panic(err)
		}
	}
}
