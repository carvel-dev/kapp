package resourcesmisc_test

import (
	"strings"
	"testing"

	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
	ctlresm "github.com/k14s/kapp/pkg/kapp/resourcesmisc"
)

func TestAppsV1StatefulSet(t *testing.T) {
	//reconcile state
	configYAML := `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: web
  generation: 1
spec:
  selector:
    matchLabels:
      app: nginx # has to match .spec.template.metadata.labels
  serviceName: "nginx"
  replicas: 3 # by default is 1
  template:
    metadata:
      labels:
        app: nginx # has to match .spec.selector.matchLabels
    spec:
      containers:
      - name: nginx
        image: k8s.gcr.io/nginx-slim:0.8
        ports:
        - containerPort: 80
          name: web
status:
  observedGeneration: 1
  replicas: 3
  readyReplicas: 2
  updateRevision: "1"
  currentRevision: "1"
  updatedReplicas: 3
  currentReplicas: 3
`

	state := buildStatefulSet(configYAML, t).IsDoneApplying()
	expectedState := ctlresm.DoneApplyState{
		Done:       false,
		Successful: false,
		Message:    "Waiting for 1 replicas to be updated and ready (currently 3 updated and 2 ready)",
	}
	if state != expectedState {
		t.Fatalf("Found incorrect state: %#v", state)
	}

	configYAML = strings.Replace(configYAML, "readyReplicas: 2", "readyReplicas: 3", -1)

	state = buildStatefulSet(configYAML, t).IsDoneApplying()
	if state != (ctlresm.DoneApplyState{Done: true, Successful: true, Message: ""}) {
		t.Fatalf("Found incorrect state: %#v", state)
	}
}

func TestAppsV1StatefulSetCreation(t *testing.T) {
	stsData := `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: web
  generation: 1
spec:
  selector:
    matchLabels:
      app: nginx 
  serviceName: "nginx"
  replicas: 3 
  template:
    metadata:
      labels:
        app: nginx 
    spec:
      containers:
      - name: nginx
        image: k8s.gcr.io/nginx-slim:0.8
        ports:
        - containerPort: 80
          name: web
status:
  replicas: 0
`

	state := buildStatefulSet(stsData, t).IsDoneApplying()
	expectedState := ctlresm.DoneApplyState{
		Done:       false,
		Successful: false,
		Message:    "Waiting for 3 replicas to be updated and ready (currently 0 updated and 0 ready)",
	}
	if state != expectedState {
		t.Fatalf("Found incorrect state: %#v", state)
	}

	//configYAML = strings.Replace(configYAML, "readyReplicas: 2", "readyReplicas: 3", -1)
	//
	//state = buildStatefulSet(configYAML, t).IsDoneApplying()
	//if state != (ctlresm.DoneApplyState{Done: true, Successful: true, Message: ""}) {
	//	t.Fatalf("Found incorrect state: %#v", state)
	//}
}

func TestAppsV1StatefulSetUpdate(t *testing.T) {
	configYAML := `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: web
  generation: 1
spec:
  selector:
    matchLabels:
      app: nginx # has to match .spec.template.metadata.labels
  serviceName: "nginx"
  replicas: 3 # by default is 1
  template:
    metadata:
      labels:
        app: nginx # has to match .spec.selector.matchLabels
    spec:
      containers:
      - name: nginx
        image: k8s.gcr.io/nginx-slim:0.8
        ports:
        - containerPort: 80
          name: web
  updateStrategy:
    type: OnDelete
status:
  observedGeneration: 1
  replicas: 3
  readyReplicas: 2
  updateRevision: "1"
  currentRevision: "1"
  updatedReplicas: 3
  currentReplicas: 3
`

	state := buildStatefulSet(configYAML, t).IsDoneApplying()
	expectedState := ctlresm.DoneApplyState{
		Done:       true,
		Successful: true,
		Message:    "",
	}
	if state != expectedState {
		t.Fatalf("Found incorrect state: %#v", state)
	}

}
func TestAppsV1StatefulSetUpdatePartition(t *testing.T) {
	configYAML := `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: web
  generation: 1
spec:
  selector:
    matchLabels:
      app: nginx # has to match .spec.template.metadata.labels
  serviceName: "nginx"
  replicas: 3 # by default is 1
  template:
    metadata:
      labels:
        app: nginx # has to match .spec.selector.matchLabels
    spec:
      containers:
      - name: nginx
        image: k8s.gcr.io/nginx-slim:0.8
        ports:
        - containerPort: 80
          name: web
  updateStrategy:
    type: OnDelete
status:
  observedGeneration: 1
  replicas: 3
  readyReplicas: 2
  updateRevision: "1"
  currentRevision: "1"
  updatedReplicas: 3
  currentReplicas: 3
`

	state := buildStatefulSet(configYAML, t).IsDoneApplying()
	expectedState := ctlresm.DoneApplyState{
		Done:       true,
		Successful: true,
		Message:    "",
	}
	if state != expectedState {
		t.Fatalf("Found incorrect state: %#v", state)
	}

}

func buildStatefulSet(resourcesBs string, t *testing.T) *ctlresm.AppsV1StatefulSet {
	newResources, err := ctlres.NewFileResource(ctlres.NewBytesSource([]byte(resourcesBs))).Resources()
	if err != nil {
		t.Fatalf("Expected resources to parse")
	}

	return ctlresm.NewAppsV1StatefulSet(newResources[0], nil)
}
