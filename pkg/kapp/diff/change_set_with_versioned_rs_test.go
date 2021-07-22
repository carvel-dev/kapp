package diff

import (
	"testing"

	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
)

func TestChangeSet_ExistingVersioned_NewNonVersioned_Resource(t *testing.T) {
	newRs, _ := ctlres.NewResourcesFromBytes([]byte(`
apiVersion: v1
kind: Secret
metadata:
  name: secret
`))

	existingRes, _ := ctlres.NewResourcesFromBytes([]byte(`
apiVersion: v1
kind: Secret
metadata:
  name: secret-ver-1
  annotations:
    kapp.k14s.io/versioned: ""
`))

	changeSetWithVerRes := NewChangeSetWithVersionedRs(existingRes, newRs, nil, ChangeSetOpts{}, ChangeFactory{})

	changes, err := changeSetWithVerRes.Calculate()

	if err != nil {
		t.Fatalf("Expected non-error")
	}

	if changes[0].Op() != ChangeOpDelete {
		t.Fatalf("Expected to get deleted: actual >>>%s<<< vs expected >>>%s<<<", changes[0].Op(), ChangeOpDelete)
	}

	if changes[1].Op() != ChangeOpAdd {
		t.Fatalf("Expected to get added: actual >>>%s<<< vs expected >>>%s<<<", changes[1].Op(), ChangeOpAdd)
	}

	actualDiff1 := changes[0].ConfigurableTextDiff().Full().FullString()

	expectedDiff1 := `  0,  0 - apiVersion: v1
  1,  0 - kind: Secret
  2,  0 - metadata:
  3,  0 -   annotations:
  4,  0 -     kapp.k14s.io/versioned: ""
  5,  0 -   name: secret-ver-1
  6,  0 - 
`
	if actualDiff1 != expectedDiff1 {
		t.Fatalf("Expected diff to match: actual >>>%s<<< vs expected >>>%s<<< %d %d", actualDiff1, expectedDiff1, len(actualDiff1), len(expectedDiff1))
	}

	actualDiff2 := changes[1].ConfigurableTextDiff().Full().FullString()

	expectedDiff2 := `  0,  0 + apiVersion: v1
  0,  1 + kind: Secret
  0,  2 + metadata:
  0,  3 +   name: secret
  0,  4 + 
`
	if actualDiff2 != expectedDiff2 {
		t.Fatalf("Expected diff to match: actual >>>%s<<< vs expected >>>%s<<< %d %d", actualDiff2, expectedDiff2, len(actualDiff2), len(expectedDiff2))
	}
}

func TestChangeSet_ExistingNonVersioned_NewVersioned_Resource(t *testing.T) {
	newRs, _ := ctlres.NewResourcesFromBytes([]byte(`
apiVersion: v1
kind: Secret
metadata:
  name: secret
  annotations:
    kapp.k14s.io/versioned: ""
`))

	existingRes, _ := ctlres.NewResourcesFromBytes([]byte(`
apiVersion: v1
kind: Secret
metadata:
  name: secret
`))

	changeSetWithVerRes := NewChangeSetWithVersionedRs(existingRes, newRs, nil, ChangeSetOpts{}, ChangeFactory{})

	changes, err := changeSetWithVerRes.Calculate()

	if err != nil {
		t.Fatalf("Expected non-error")
	}

	if changes[0].Op() != ChangeOpAdd {
		t.Fatalf("Expected to get added: actual >>>%s<<< vs expected >>>%s<<<", changes[0].Op(), ChangeOpAdd)
	}

	if changes[1].Op() != ChangeOpDelete {
		t.Fatalf("Expected to get deleted: actual >>>%s<<< vs expected >>>%s<<<", changes[1].Op(), ChangeOpDelete)
	}

	actualDiff1 := changes[0].ConfigurableTextDiff().Full().FullString()

	expectedDiff1 := `  0,  0 + apiVersion: v1
  0,  1 + kind: Secret
  0,  2 + metadata:
  0,  3 +   annotations:
  0,  4 +     kapp.k14s.io/versioned: ""
  0,  5 +   name: secret-ver-1
  0,  6 + 
`
	if actualDiff1 != expectedDiff1 {
		t.Fatalf("Expected diff to match: actual >>>%s<<< vs expected >>>%s<<< %d %d", actualDiff1, expectedDiff1, len(actualDiff1), len(expectedDiff1))
	}

	actualDiff2 := changes[1].ConfigurableTextDiff().Full().FullString()

	expectedDiff2 := `  0,  0 - apiVersion: v1
  1,  0 - kind: Secret
  2,  0 - metadata:
  3,  0 -   name: secret
  4,  0 - 
`
	if actualDiff2 != expectedDiff2 {
		t.Fatalf("Expected diff to match: actual >>>%s<<< vs expected >>>%s<<< %d %d", actualDiff2, expectedDiff2, len(actualDiff2), len(expectedDiff2))
	}
}

func TestChangeSet_ExistingNonVersioned_NewVersioneKeepOrg_Resource(t *testing.T) {
	newRs, _ := ctlres.NewResourcesFromBytes([]byte(`
apiVersion: v1
kind: Secret
metadata:
  name: secret
  annotations:
    kapp.k14s.io/versioned: ""
    kapp.k14s.io/versioned-keep-original: ""
`))

	existingRes, _ := ctlres.NewResourcesFromBytes([]byte(`
apiVersion: v1
kind: Secret
metadata:
  name: secret
`))

	changeSetWithVerRes := NewChangeSetWithVersionedRs(existingRes, newRs, nil, ChangeSetOpts{}, ChangeFactory{})

	changes, err := changeSetWithVerRes.Calculate()

	if err != nil {
		t.Fatalf("Expected non-error")
	}

	if changes[0].Op() != ChangeOpAdd {
		t.Fatalf("Expected to get added: actual >>>%s<<< vs expected >>>%s<<<", changes[0].Op(), ChangeOpAdd)
	}

	if changes[1].Op() != ChangeOpUpdate {
		t.Fatalf("Expected to get updated: actual >>>%s<<< vs expected >>>%s<<<", changes[1].Op(), ChangeOpUpdate)
	}

	actualDiff1 := changes[0].ConfigurableTextDiff().Full().FullString()

	expectedDiff1 := `  0,  0 + apiVersion: v1
  0,  1 + kind: Secret
  0,  2 + metadata:
  0,  3 +   annotations:
  0,  4 +     kapp.k14s.io/versioned: ""
  0,  5 +     kapp.k14s.io/versioned-keep-original: ""
  0,  6 +   name: secret-ver-1
  0,  7 + 
`
	if actualDiff1 != expectedDiff1 {
		t.Fatalf("Expected diff to match: actual >>>%s<<< vs expected >>>%s<<< %d %d", actualDiff1, expectedDiff1, len(actualDiff1), len(expectedDiff1))
	}

	actualDiff2 := changes[1].ConfigurableTextDiff().Full().FullString()

	expectedDiff2 := `  0,  0   apiVersion: v1
  1,  1   kind: Secret
  2,  2   metadata:
  3,  3 +   annotations:
  3,  4 +     kapp.k14s.io/versioned: ""
  3,  5 +     kapp.k14s.io/versioned-keep-original: ""
  3,  6     name: secret
  4,  7   
`
	if actualDiff2 != expectedDiff2 {
		t.Fatalf("Expected diff to match: actual >>>%s<<< vs expected >>>%s<<< %d %d", actualDiff2, expectedDiff2, len(actualDiff2), len(expectedDiff2))
	}
}
