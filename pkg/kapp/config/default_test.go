// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package config_test

import (
	"strings"
	"testing"

	"carvel.dev/kapp/pkg/kapp/config"
	ctldiff "carvel.dev/kapp/pkg/kapp/diff"
	ctlres "carvel.dev/kapp/pkg/kapp/resources"
	"github.com/stretchr/testify/require"
)

func TestDefaultTemplateRules(t *testing.T) {
	_, defaultConfig, err := config.NewConfFromResourcesWithDefaults([]ctlres.Resource{})
	require.NoError(t, err)
	changeFactory := ctldiff.NewChangeFactory(defaultConfig.RebaseMods(), defaultConfig.DiffAgainstLastAppliedFieldExclusionMods(), defaultConfig.DiffAgainstExistingFieldExclusionMods(), ctldiff.ChangeOpts{false})

	testCases := []struct {
		description  string
		newYAML      []byte
		expectedDiff string
	}{
		{
			description: `kappctrl.k14s.io/v1alpha1/App`,
			newYAML: []byte(`
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-configmap
  annotations:
    kapp.k14s.io/versioned: ""
---
apiVersion: v1
kind: Secret
metadata:
  name: test-secret
  annotations:
    kapp.k14s.io/versioned: ""
---
apiVersion: kappctrl.k14s.io/v1alpha1
kind: App
metadata:
  name: test
spec:
  cluster:
    kubeconfigSecretRef:
      name: test-secret
  fetch:
    - inline:
        pathsFrom:
          - secretRef:
              name: test-secret
          - configMapRef:
              name: test-configmap
    - imgpkgBundle:
        secretRef:
          name: test-secret
    - http:
        secretRef:
          name: test-secret
    - git:
        secretRef:
          name: test-secret
    - helmChart:
        repository:
          secretRef:
            name: test-secret
  template:
    - ytt:
        inline:
          pathsFrom:
            - secretRef:
                name: test-secret
            - configMapRef:
                name: test-configmap
        valuesFrom:
          - secretRef:
              name: test-secret
          - configMapRef:
              name: test-configmap
    - helmTemplate:
        valuesFrom:
          - secretRef:
              name: test-secret
          - configMapRef:
              name: test-configmap
    - cue:
        valuesFrom:
          - secretRef:
              name: test-secret
          - configMapRef:
              name: test-configmap
    - sops:
        pgp:
          privateKeySecretRef:
            name: test-secret
`),
			expectedDiff: strings.TrimLeft(`
  0,  0 + apiVersion: v1
  0,  1 + kind: ConfigMap
  0,  2 + metadata:
  0,  3 +   annotations:
  0,  4 +     kapp.k14s.io/versioned: ""
  0,  5 +   name: test-configmap-ver-1
  0,  6 + 
  0,  0 + apiVersion: v1
  0,  1 + kind: Secret
  0,  2 + metadata:
  0,  3 +   annotations:
  0,  4 +     kapp.k14s.io/versioned: ""
  0,  5 +   name: test-secret-ver-1
  0,  6 + 
  0,  0 + apiVersion: kappctrl.k14s.io/v1alpha1
  0,  1 + kind: App
  0,  2 + metadata:
  0,  3 +   name: test
  0,  4 + spec:
  0,  5 +   cluster:
  0,  6 +     kubeconfigSecretRef:
  0,  7 +       name: test-secret-ver-1
  0,  8 +   fetch:
  0,  9 +   - inline:
  0, 10 +       pathsFrom:
  0, 11 +       - secretRef:
  0, 12 +           name: test-secret-ver-1
  0, 13 +       - configMapRef:
  0, 14 +           name: test-configmap-ver-1
  0, 15 +   - imgpkgBundle:
  0, 16 +       secretRef:
  0, 17 +         name: test-secret-ver-1
  0, 18 +   - http:
  0, 19 +       secretRef:
  0, 20 +         name: test-secret-ver-1
  0, 21 +   - git:
  0, 22 +       secretRef:
  0, 23 +         name: test-secret-ver-1
  0, 24 +   - helmChart:
  0, 25 +       repository:
  0, 26 +         secretRef:
  0, 27 +           name: test-secret-ver-1
  0, 28 +   template:
  0, 29 +   - ytt:
  0, 30 +       inline:
  0, 31 +         pathsFrom:
  0, 32 +         - secretRef:
  0, 33 +             name: test-secret-ver-1
  0, 34 +         - configMapRef:
  0, 35 +             name: test-configmap-ver-1
  0, 36 +       valuesFrom:
  0, 37 +       - secretRef:
  0, 38 +           name: test-secret-ver-1
  0, 39 +       - configMapRef:
  0, 40 +           name: test-configmap-ver-1
  0, 41 +   - helmTemplate:
  0, 42 +       valuesFrom:
  0, 43 +       - secretRef:
  0, 44 +           name: test-secret-ver-1
  0, 45 +       - configMapRef:
  0, 46 +           name: test-configmap-ver-1
  0, 47 +   - cue:
  0, 48 +       valuesFrom:
  0, 49 +       - secretRef:
  0, 50 +           name: test-secret-ver-1
  0, 51 +       - configMapRef:
  0, 52 +           name: test-configmap-ver-1
  0, 53 +   - sops:
  0, 54 +       pgp:
  0, 55 +         privateKeySecretRef:
  0, 56 +           name: test-secret-ver-1
  0, 57 + 
`, "\n"),
		},
		{
			description: `packaging.carvel.dev/v1alpha1/PackageRepository`,
			newYAML: []byte(`
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-configmap
  annotations:
    kapp.k14s.io/versioned: ""
---
apiVersion: v1
kind: Secret
metadata:
  name: test-secret
  annotations:
    kapp.k14s.io/versioned: ""
---
apiVersion: packaging.carvel.dev/v1alpha1
kind: PackageRepository
metadata:
  name: test
spec:
  fetch:
    inline:
      pathsFrom:
      - configMapRef:
          name: test-configmap
      - secretRef:
          name: test-secret
`),
			expectedDiff: strings.TrimLeft(`
  0,  0 + apiVersion: v1
  0,  1 + kind: ConfigMap
  0,  2 + metadata:
  0,  3 +   annotations:
  0,  4 +     kapp.k14s.io/versioned: ""
  0,  5 +   name: test-configmap-ver-1
  0,  6 + 
  0,  0 + apiVersion: v1
  0,  1 + kind: Secret
  0,  2 + metadata:
  0,  3 +   annotations:
  0,  4 +     kapp.k14s.io/versioned: ""
  0,  5 +   name: test-secret-ver-1
  0,  6 + 
  0,  0 + apiVersion: packaging.carvel.dev/v1alpha1
  0,  1 + kind: PackageRepository
  0,  2 + metadata:
  0,  3 +   name: test
  0,  4 + spec:
  0,  5 +   fetch:
  0,  6 +     inline:
  0,  7 +       pathsFrom:
  0,  8 +       - configMapRef:
  0,  9 +           name: test-configmap-ver-1
  0, 10 +       - secretRef:
  0, 11 +           name: test-secret-ver-1
  0, 12 + 
`, "\n"),
		},
		{
			description: `packaging.carvel.dev/v1alpha1/PackageInstall`,
			newYAML: []byte(`
---
apiVersion: v1
kind: Secret
metadata:
  name: test-secret
  annotations:
    kapp.k14s.io/versioned: ""
---
apiVersion: packaging.carvel.dev/v1alpha1
kind: PackageInstall
metadata:
  name: test
spec:
  values:
  - secretRef:
      name: test-secret
  cluster:
    kubeconfigSecretRef:
      name: test-secret
`),
			expectedDiff: strings.TrimLeft(`
  0,  0 + apiVersion: v1
  0,  1 + kind: Secret
  0,  2 + metadata:
  0,  3 +   annotations:
  0,  4 +     kapp.k14s.io/versioned: ""
  0,  5 +   name: test-secret-ver-1
  0,  6 + 
  0,  0 + apiVersion: packaging.carvel.dev/v1alpha1
  0,  1 + kind: PackageInstall
  0,  2 + metadata:
  0,  3 +   name: test
  0,  4 + spec:
  0,  5 +   cluster:
  0,  6 +     kubeconfigSecretRef:
  0,  7 +       name: test-secret-ver-1
  0,  8 +   values:
  0,  9 +   - secretRef:
  0, 10 +       name: test-secret-ver-1
  0, 11 + 
`, "\n"),
		},
	}

	for _, testCase := range testCases {
		// Deserialize YAML into resources
		docs, err := ctlres.NewYAMLFile(ctlres.NewBytesSource(testCase.newYAML)).Docs()
		require.NoError(t, err)
		var newResources []ctlres.Resource
		for _, doc := range docs {
			bytes, err := ctlres.NewResourcesFromBytes(doc)
			require.NoError(t, err)
			newResources = append(newResources, bytes...)
		}

		// Calculate changes with default template rules
		changes, err := ctldiff.NewChangeSetWithVersionedRs([]ctlres.Resource{}, newResources, defaultConfig.TemplateRules(), ctldiff.ChangeSetOpts{}, changeFactory).Calculate()
		require.NoError(t, err)

		// Compare against expected diff
		var diff strings.Builder
		for _, change := range changes {
			diff.WriteString(change.ConfigurableTextDiff().Full().FullString())
		}
		require.Equal(t, testCase.expectedDiff, diff.String())
	}
}
