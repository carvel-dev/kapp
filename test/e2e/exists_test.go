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
	"k8s.io/client-go/tools/clientcmd"
)

func TestExistsOpWithPlaceholderAnn(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	config, err := kubeConfig.ClientConfig()
	if err != nil {
		t.Fatalf("Error getting kube config, %v", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
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
kind: Secret
metadata:
  name: secret
data:
  key1: val1
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: external
  annotations:
    kapp.k14s.io/placeholder: ""
`

	name := "app"
	externalResourceName := "external"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
		deleteExternalResource(env, clientset, externalResourceName)
	}
	cleanUp()
	defer cleanUp()

	logger.Section("deploying external resource", func() {
		go deployExternalResource(env, clientset, externalResourceName)
	})

	logger.Section("deploying app with placeholder annotation", func() {
		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name},
			RunOpts{StdinReader: strings.NewReader(app)})
		expectedOutput := `
Changes

Namespace  Name      Kind       Conds.  Age  Op      Op st.  Wait to    Rs  Ri  
(cluster)  kapp-ns   Namespace  -       -    create  -       reconcile  -   -  
kapp-test  external  ConfigMap  -       -    exists  -       reconcile  -   -  
^          secret    Secret     -       -    create  -       reconcile  -   -  

Op:      2 create, 0 delete, 0 update, 0 noop
Wait to: 3 reconcile, 0 delete, 0 noop

<replaced>: ---- applying 1 changes [0/3 done] ----
<replaced>: create namespace/kapp-ns (v1) cluster
<replaced>: ---- waiting on 1 changes [0/3 done] ----
<replaced>: ok: reconcile namespace/kapp-ns (v1) cluster
<replaced>: ---- applying 2 changes [1/3 done] ----
<replaced>: exists configmap/external (v1) namespace: kapp-test
<replaced>:  ^ Retryable error: Resource doesn't exists
<replaced>: create secret/secret (v1) namespace: kapp-test
<replaced>: ---- waiting on 1 changes [1/3 done] ----
<replaced>: ok: reconcile secret/secret (v1) namespace: kapp-test
<replaced>: ---- applying 1 changes [2/3 done] ----
<replaced>: exists configmap/external (v1) namespace: kapp-test
<replaced>: ---- waiting on 1 changes [2/3 done] ----
<replaced>: ok: reconcile configmap/external (v1) namespace: kapp-test
<replaced>: ---- applying complete [3/3 done] ----
<replaced>: ---- waiting complete [3/3 done] ----

Succeeded`

		out = strings.TrimSpace(replaceTarget(replaceSpaces(replaceTs(out))))
		expectedOutput = strings.TrimSpace(replaceSpaces(expectedOutput))
		if expectedOutput != out {
			t.Fatalf("Expected output to be >>%s<<, but got >>%s<<\n", expectedOutput, out)
		}
	})
}

func deployExternalResource(env Env, clientset *kubernetes.Clientset, name string) {
	time.Sleep(2 * time.Second)
	metaData := metav1.ObjectMeta{Name: name}
	data := map[string]string{}
	data["key1"] = "value1"
	configMap := v1.ConfigMap{
		ObjectMeta: metaData,
		Data:       data,
	}
	_, err := clientset.CoreV1().ConfigMaps(env.Namespace).Create(context.TODO(), &configMap, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}
}

func deleteExternalResource(env Env, clientset *kubernetes.Clientset, name string) {
	err := clientset.CoreV1().ConfigMaps(env.Namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		if err.Error() != fmt.Sprintf(`configmaps "%s" not found`, name) {
			panic(err)
		}
	}
}
