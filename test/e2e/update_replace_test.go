package e2e

import (
	"strings"
	"testing"
)

func TestUpdateFallbackOnReplace(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}
	kubectl := Kubectl{t, env.Namespace, logger}

	yaml1 := `
---
apiVersion: v1
kind: Service
metadata:
  name: redis-master
spec:
  ports:
  - port: 6380
    targetPort: 6380
  selector:
    app: redis
    tier: backend
    role: master
`

	yaml2 := `
---
apiVersion: v1
kind: Service
metadata:
  name: redis-master
  annotations:
    kapp.k14s.io/update-strategy: fallback-on-replace
spec:
  clusterIP: None
  ports:
  - port: 6380
    targetPort: 6380
  selector:
    app: redis
    tier: backend
    role: master
`

	name := "test-update-fallback-on-replace"
	objKind := "service"
	objName := "redis-master"
	cleanUp := func() {
		kapp.RunWithOpts([]string{"delete", "-a", name}, RunOpts{AllowError: true})
	}

	cleanUp()
	defer cleanUp()

	logger.Section("deploy basic service", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})
	})

	logger.Section("deploy update to service that changes immutable field spec.clusterIP", func() {
		prev := NewPresentClusterResource(objKind, objName, env.Namespace, kubectl)

		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml2)})
		curr := NewPresentClusterResource(objKind, objName, env.Namespace, kubectl)

		if prev.UID() == curr.UID() {
			t.Fatalf("Expected object to be replaced, but found same UID")
		}
	})

	logger.Section("deploy update to service that does not set spec.clusterIP", func() {
		prev := NewPresentClusterResource(objKind, objName, env.Namespace, kubectl)

		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})
		curr := NewPresentClusterResource(objKind, objName, env.Namespace, kubectl)

		if prev.UID() != curr.UID() {
			t.Fatalf("Expected object to be rebased, but found different UID")
		}
	})
}

func TestUpdateAlwaysReplace(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}
	kubectl := Kubectl{t, env.Namespace, logger}

	yaml1 := `
---
apiVersion: v1
kind: Service
metadata:
  name: redis-master
spec:
  ports:
  - port: 6380
    targetPort: 6380
  selector:
    app: redis
    tier: backend
    role: master
`

	yaml2 := `
---
apiVersion: v1
kind: Service
metadata:
  name: redis-master
  annotations:
    kapp.k14s.io/update-strategy: always-replace
spec:
  clusterIP: None
  ports:
  - port: 6380
    targetPort: 6380
  selector:
    app: redis
    tier: backend
    role: master
`

	name := "test-update-always-replace"
	objKind := "service"
	objName := "redis-master"
	cleanUp := func() {
		kapp.RunWithOpts([]string{"delete", "-a", name}, RunOpts{AllowError: true})
	}

	cleanUp()
	defer cleanUp()

	logger.Section("deploy basic service", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})
	})

	logger.Section("deploy update to service that changes immutable field spec.clusterIP", func() {
		prev := NewPresentClusterResource(objKind, objName, env.Namespace, kubectl)

		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml2)})
		curr := NewPresentClusterResource(objKind, objName, env.Namespace, kubectl)

		if prev.UID() == curr.UID() {
			t.Fatalf("Expected object to be replaced, but found same UID")
		}
	})

	logger.Section("deploy update to service that does not set spec.clusterIP", func() {
		prev := NewPresentClusterResource(objKind, objName, env.Namespace, kubectl)

		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})
		curr := NewPresentClusterResource(objKind, objName, env.Namespace, kubectl)

		if prev.UID() != curr.UID() {
			t.Fatalf("Expected object to be rebased, but found different UID")
		}
	})
}
