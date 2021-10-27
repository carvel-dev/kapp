// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package diffgraph_test

import (
	"io/ioutil"
	"strings"
	"testing"
	"time"

	ctlconf "github.com/k14s/kapp/pkg/kapp/config"
	ctldgraph "github.com/k14s/kapp/pkg/kapp/diffgraph"
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
	"github.com/stretchr/testify/require"
)

func TestChangeGraphCFForK8sUpsert(t *testing.T) {
	configYAML, err := ioutil.ReadFile("assets/cf-for-k8s.yml")
	require.NoErrorf(t, err, "Reading cf-for-k8s asset")

	configRs, err := ctlres.NewFileResource(ctlres.NewBytesSource([]byte(configYAML))).Resources()
	require.NoErrorf(t, err, "Parsing resources")

	rs, conf, err := ctlconf.NewConfFromResourcesWithDefaults(configRs)
	require.NoErrorf(t, err, "Parsing conf defaults")

	opts := buildGraphOpts{
		resources:           rs,
		op:                  ctldgraph.ActualChangeOpUpsert,
		changeGroupBindings: conf.ChangeGroupBindings(),
		changeRuleBindings:  conf.ChangeRuleBindings(),
	}

	t1 := time.Now()

	graph, err := buildChangeGraphWithOpts(opts, t)
	require.NoErrorf(t, err, "Expected graph to build")

	require.Less(t, time.Now().Sub(t1), time.Duration(1*time.Second), "Graph build took too long")

	output := strings.TrimSpace(graph.PrintLinearizedStr())
	expectedOutput := strings.TrimSpace(cfForK8sExpectedOutputUpsert)

	require.Equal(t, expectedOutput, output)
}

func TestChangeGraphCFForK8sDelete(t *testing.T) {
	configYAML, err := ioutil.ReadFile("assets/cf-for-k8s.yml")
	require.NoErrorf(t, err, "Reading cf-for-k8s asset")

	configRs, err := ctlres.NewFileResource(ctlres.NewBytesSource([]byte(configYAML))).Resources()
	require.NoErrorf(t, err, "Parsing resources")

	rs, conf, err := ctlconf.NewConfFromResourcesWithDefaults(configRs)
	require.NoErrorf(t, err, "Parsing conf defaults")

	opts := buildGraphOpts{
		resources:           rs,
		op:                  ctldgraph.ActualChangeOpDelete,
		changeGroupBindings: conf.ChangeGroupBindings(),
		changeRuleBindings:  conf.ChangeRuleBindings(),
	}

	t1 := time.Now()

	graph, err := buildChangeGraphWithOpts(opts, t)
	require.NoErrorf(t, err, "Expected graph to build")

	require.Less(t, time.Now().Sub(t1), time.Duration(1*time.Second), "Graph build took too long")

	output := strings.TrimSpace(graph.PrintLinearizedStr())
	expectedOutput := strings.TrimSpace(cfForK8sExpectedOutputDelete)

	require.Equal(t, expectedOutput, output)
}

const (
	cfForK8sExpectedOutputUpsert = `
(upsert) clusterrole/kpack-watcher (rbac.authorization.k8s.io/v1) cluster
(upsert) namespace/cf-workloads-staging (v1) cluster
(upsert) podsecuritypolicy/cf-workloads-app-psp (policy/v1beta1) cluster
(upsert) podsecuritypolicy/cf-workloads-privileged-app-psp (policy/v1beta1) cluster
(upsert) podsecuritypolicy/eirini (policy/v1beta1) cluster
(upsert) namespace/kpack (v1) cluster
(upsert) customresourcedefinition/builds.build.pivotal.io (apiextensions.k8s.io/v1beta1) cluster
(upsert) customresourcedefinition/builders.build.pivotal.io (apiextensions.k8s.io/v1beta1) cluster
(upsert) customresourcedefinition/clusterbuilders.build.pivotal.io (apiextensions.k8s.io/v1beta1) cluster
(upsert) clusterrole/kpack-controller-admin (rbac.authorization.k8s.io/v1) cluster
(upsert) customresourcedefinition/custombuilders.experimental.kpack.pivotal.io (apiextensions.k8s.io/v1beta1) cluster
(upsert) customresourcedefinition/customclusterbuilders.experimental.kpack.pivotal.io (apiextensions.k8s.io/v1beta1) cluster
(upsert) customresourcedefinition/images.build.pivotal.io (apiextensions.k8s.io/v1beta1) cluster
(upsert) customresourcedefinition/sourceresolvers.build.pivotal.io (apiextensions.k8s.io/v1beta1) cluster
(upsert) customresourcedefinition/stacks.experimental.kpack.pivotal.io (apiextensions.k8s.io/v1beta1) cluster
(upsert) customresourcedefinition/stores.experimental.kpack.pivotal.io (apiextensions.k8s.io/v1beta1) cluster
(upsert) mutatingwebhookconfiguration/resource.webhook.kpack.pivotal.io (admissionregistration.k8s.io/v1beta1) cluster
(upsert) clusterrole/kpack-webhook-mutatingwebhookconfiguration-admin (rbac.authorization.k8s.io/v1) cluster
(upsert) namespace/cf-blobstore (v1) cluster
(upsert) customresourcedefinition/routebulksyncs.apps.cloudfoundry.org (apiextensions.k8s.io/v1beta1) cluster
(upsert) clusterrole/istio-reader-istio-system (rbac.authorization.k8s.io/v1) cluster
(upsert) customresourcedefinition/attributemanifests.config.istio.io (apiextensions.k8s.io/v1beta1) cluster
(upsert) customresourcedefinition/clusterrbacconfigs.rbac.istio.io (apiextensions.k8s.io/v1beta1) cluster
(upsert) customresourcedefinition/destinationrules.networking.istio.io (apiextensions.k8s.io/v1beta1) cluster
(upsert) customresourcedefinition/envoyfilters.networking.istio.io (apiextensions.k8s.io/v1beta1) cluster
(upsert) customresourcedefinition/gateways.networking.istio.io (apiextensions.k8s.io/v1beta1) cluster
(upsert) customresourcedefinition/httpapispecbindings.config.istio.io (apiextensions.k8s.io/v1beta1) cluster
(upsert) customresourcedefinition/httpapispecs.config.istio.io (apiextensions.k8s.io/v1beta1) cluster
(upsert) customresourcedefinition/meshpolicies.authentication.istio.io (apiextensions.k8s.io/v1beta1) cluster
(upsert) customresourcedefinition/policies.authentication.istio.io (apiextensions.k8s.io/v1beta1) cluster
(upsert) customresourcedefinition/quotaspecbindings.config.istio.io (apiextensions.k8s.io/v1beta1) cluster
(upsert) customresourcedefinition/quotaspecs.config.istio.io (apiextensions.k8s.io/v1beta1) cluster
(upsert) customresourcedefinition/rbacconfigs.rbac.istio.io (apiextensions.k8s.io/v1beta1) cluster
(upsert) customresourcedefinition/rules.config.istio.io (apiextensions.k8s.io/v1beta1) cluster
(upsert) customresourcedefinition/serviceentries.networking.istio.io (apiextensions.k8s.io/v1beta1) cluster
(upsert) customresourcedefinition/servicerolebindings.rbac.istio.io (apiextensions.k8s.io/v1beta1) cluster
(upsert) customresourcedefinition/serviceroles.rbac.istio.io (apiextensions.k8s.io/v1beta1) cluster
(upsert) customresourcedefinition/virtualservices.networking.istio.io (apiextensions.k8s.io/v1beta1) cluster
(upsert) customresourcedefinition/adapters.config.istio.io (apiextensions.k8s.io/v1beta1) cluster
(upsert) customresourcedefinition/instances.config.istio.io (apiextensions.k8s.io/v1beta1) cluster
(upsert) customresourcedefinition/templates.config.istio.io (apiextensions.k8s.io/v1beta1) cluster
(upsert) customresourcedefinition/handlers.config.istio.io (apiextensions.k8s.io/v1beta1) cluster
(upsert) customresourcedefinition/sidecars.networking.istio.io (apiextensions.k8s.io/v1beta1) cluster
(upsert) customresourcedefinition/authorizationpolicies.security.istio.io (apiextensions.k8s.io/v1beta1) cluster
(upsert) namespace/istio-system (v1) cluster
(upsert) clusterrole/istio-citadel-istio-system (rbac.authorization.k8s.io/v1) cluster
(upsert) clusterrole/istio-galley-istio-system (rbac.authorization.k8s.io/v1) cluster
(upsert) clusterrole/istio-sidecar-injector-istio-system (rbac.authorization.k8s.io/v1) cluster
(upsert) mutatingwebhookconfiguration/istio-sidecar-injector (admissionregistration.k8s.io/v1beta1) cluster
(upsert) clusterrole/istio-pilot-istio-system (rbac.authorization.k8s.io/v1) cluster
(upsert) clusterrole/istio-policy (rbac.authorization.k8s.io/v1) cluster
(upsert) clusterrole/istio-mixer-istio-system (rbac.authorization.k8s.io/v1) cluster
(upsert) namespace/metacontroller (v1) cluster
(upsert) clusterrole/metacontroller (rbac.authorization.k8s.io/v1) cluster
(upsert) clusterrole/aggregate-metacontroller-view (rbac.authorization.k8s.io/v1) cluster
(upsert) clusterrole/aggregate-metacontroller-edit (rbac.authorization.k8s.io/v1) cluster
(upsert) customresourcedefinition/compositecontrollers.metacontroller.k8s.io (apiextensions.k8s.io/v1beta1) cluster
(upsert) customresourcedefinition/decoratorcontrollers.metacontroller.k8s.io (apiextensions.k8s.io/v1beta1) cluster
(upsert) customresourcedefinition/controllerrevisions.metacontroller.k8s.io (apiextensions.k8s.io/v1beta1) cluster
(upsert) namespace/cf-db (v1) cluster
(upsert) namespace/cf-system (v1) cluster
(upsert) namespace/cf-workloads (v1) cluster
---
(upsert) role/kpack-watcher-pod-logs-reader (rbac.authorization.k8s.io/v1) namespace: cf-workloads
(upsert) configmap/cloud-controller-ng-yaml (v1) namespace: cf-system
(upsert) configmap/nginx (v1) namespace: cf-system
(upsert) secret/opi-secrets (v1) namespace: cf-system
(upsert) serviceaccount/cc-api-service-account (v1) namespace: cf-system
(upsert) secret/cc-kpack-registry-auth-secret (v1) namespace: cf-workloads-staging
(upsert) serviceaccount/cc-kpack-registry-service-account (v1) namespace: cf-workloads-staging
(upsert) networkpolicy/deny-app-ingress (networking.k8s.io/v1) namespace: cf-workloads
(upsert) serviceaccount/eirini (v1) namespace: cf-workloads
(upsert) serviceaccount/eirini-privileged (v1) namespace: cf-workloads
(upsert) serviceaccount/opi (v1) namespace: cf-system
(upsert) configmap/eirini (v1) namespace: cf-system
(upsert) role/cf-workloads-app-role (rbac.authorization.k8s.io/v1) namespace: cf-workloads
(upsert) role/cf-workloads-privileged-app-role (rbac.authorization.k8s.io/v1) namespace: cf-workloads
(upsert) role/eirini-role (rbac.authorization.k8s.io/v1) namespace: cf-workloads
(upsert) role/eirini-role (rbac.authorization.k8s.io/v1) namespace: cf-system
(upsert) secret/eirini-internal-tls-certs (v1) namespace: cf-system
(upsert) networkpolicy/allow-app-ingress-from-ingressgateway (networking.k8s.io/v1) namespace: cf-workloads
(upsert) secret/app-registry-credentials (v1) namespace: cf-workloads
(upsert) secret/istio-ingressgateway-certs (v1) namespace: istio-system
(upsert) serviceaccount/controller (v1) namespace: kpack
(upsert) serviceaccount/webhook (v1) namespace: kpack
(upsert) role/kpack-webhook-certs-admin (rbac.authorization.k8s.io/v1) namespace: kpack
(upsert) serviceaccount/fluentd-service-account (v1) namespace: cf-system
(upsert) clusterrole/pod-namespace-read (rbac.authorization.k8s.io/v1) namespace: cf-system
(upsert) configmap/fluentd-config (v1) namespace: cf-system
(upsert) secret/log-cache-ca (v1) namespace: cf-system
(upsert) secret/log-cache (v1) namespace: cf-system
(upsert) secret/log-cache-metrics (v1) namespace: cf-system
(upsert) secret/log-cache-gateway (v1) namespace: cf-system
(upsert) secret/log-cache-syslog (v1) namespace: cf-system
(upsert) serviceaccount/metric-proxy (v1) namespace: cf-system
(upsert) clusterrole/metric-proxy (rbac.authorization.k8s.io/v1) namespace: cf-system
(upsert) secret/metric-proxy-cert (v1) namespace: cf-system
(upsert) secret/metric-proxy-ca (v1) namespace: cf-system
(upsert) secret/cf-blobstore-minio (v1) namespace: cf-blobstore
(upsert) configmap/cf-blobstore-minio (v1) namespace: cf-blobstore
(upsert) persistentvolumeclaim/cf-blobstore-minio (v1) namespace: cf-blobstore
(upsert) serviceaccount/cf-blobstore-minio (v1) namespace: cf-blobstore
(upsert) configmap/cfroutesync-config (v1) namespace: cf-system
(upsert) secret/cfroutesync (v1) namespace: cf-system
(upsert) compositecontroller/cfroutesync (metacontroller.k8s.io/v1alpha1) cluster
(upsert) serviceaccount/istio-reader-service-account (v1) namespace: istio-system
(upsert) poddisruptionbudget/istio-citadel (policy/v1beta1) namespace: istio-system
(upsert) serviceaccount/istio-citadel-service-account (v1) namespace: istio-system
(upsert) configmap/galley-envoy-config (v1) namespace: istio-system
(upsert) configmap/istio-mesh-galley (v1) namespace: istio-system
(upsert) configmap/istio-galley-configuration (v1) namespace: istio-system
(upsert) poddisruptionbudget/istio-galley (policy/v1beta1) namespace: istio-system
(upsert) serviceaccount/istio-galley-service-account (v1) namespace: istio-system
(upsert) poddisruptionbudget/ingressgateway (policy/v1beta1) namespace: istio-system
(upsert) role/istio-ingressgateway-sds (rbac.authorization.k8s.io/v1) namespace: istio-system
(upsert) serviceaccount/istio-ingressgateway-service-account (v1) namespace: istio-system
(upsert) configmap/injector-mesh (v1) namespace: istio-system
(upsert) poddisruptionbudget/istio-sidecar-injector (policy/v1beta1) namespace: istio-system
(upsert) serviceaccount/istio-sidecar-injector-service-account (v1) namespace: istio-system
(upsert) configmap/istio-sidecar-injector (v1) namespace: istio-system
(upsert) configmap/pilot-envoy-config (v1) namespace: istio-system
(upsert) configmap/istio (v1) namespace: istio-system
(upsert) meshpolicy/default (authentication.istio.io/v1alpha1) cluster
(upsert) poddisruptionbudget/istio-pilot (policy/v1beta1) namespace: istio-system
(upsert) serviceaccount/istio-pilot-service-account (v1) namespace: istio-system
(upsert) configmap/policy-envoy-config (v1) namespace: istio-system
(upsert) poddisruptionbudget/istio-policy (policy/v1beta1) namespace: istio-system
(upsert) serviceaccount/istio-policy-service-account (v1) namespace: istio-system
(upsert) configmap/telemetry-envoy-config (v1) namespace: istio-system
(upsert) poddisruptionbudget/istio-telemetry (policy/v1beta1) namespace: istio-system
(upsert) serviceaccount/istio-mixer-service-account (v1) namespace: istio-system
(upsert) networkpolicy/pilot-network-policy (networking.k8s.io/v1) namespace: istio-system
(upsert) networkpolicy/citadel-network-policy (networking.k8s.io/v1) namespace: istio-system
(upsert) networkpolicy/mixer-network-policy (networking.k8s.io/v1) namespace: istio-system
(upsert) networkpolicy/sidecar-injector-network-policy (networking.k8s.io/v1) namespace: istio-system
(upsert) networkpolicy/ingressgateway-network-policy-pilot (networking.k8s.io/v1) namespace: istio-system
(upsert) serviceaccount/metacontroller (v1) namespace: metacontroller
(upsert) secret/cf-db-admin-secret (v1) namespace: cf-db
(upsert) secret/cf-db-credentials (v1) namespace: cf-db
(upsert) configmap/cf-db-postgresql-init-scripts (v1) namespace: cf-db
(upsert) configmap/uaa-config (v1) namespace: cf-system
(upsert) serviceaccount/uaa (v1) namespace: cf-system
(upsert) secret/uaa-certs (v1) namespace: cf-system
---
(upsert) clusterrolebinding/kpack-watcher-binding (rbac.authorization.k8s.io/v1) cluster
(upsert) rolebinding/kpack-watcher-pod-logs-binding (rbac.authorization.k8s.io/v1) namespace: cf-workloads
(upsert) clusterrolebinding/cc-api-service-account-superuser (rbac.authorization.k8s.io/v1) cluster
(upsert) rolebinding/cf-workloads-app-rolebinding (rbac.authorization.k8s.io/v1) namespace: cf-workloads
(upsert) rolebinding/cf-workloads-privileged-app-rolebinding (rbac.authorization.k8s.io/v1) namespace: cf-workloads
(upsert) rolebinding/eirini-rolebinding (rbac.authorization.k8s.io/v1) namespace: cf-workloads
(upsert) rolebinding/eirini-rolebinding (rbac.authorization.k8s.io/v1) namespace: cf-system
(upsert) clusterrolebinding/kpack-controller-admin-binding (rbac.authorization.k8s.io/v1) cluster
(upsert) rolebinding/kpack-webhook-certs-admin-binding (rbac.authorization.k8s.io/v1) namespace: kpack
(upsert) clusterrolebinding/kpack-webhook-certs-mutatingwebhookconfiguration-admin-binding (rbac.authorization.k8s.io/v1) cluster
(upsert) clusterrolebinding/fluentd-service-account-pod-namespace-read (rbac.authorization.k8s.io/v1) cluster
(upsert) clusterrolebinding/metric-proxy (rbac.authorization.k8s.io/v1) cluster
(upsert) clusterrolebinding/istio-reader-istio-system (rbac.authorization.k8s.io/v1) cluster
(upsert) clusterrolebinding/istio-citadel-istio-system (rbac.authorization.k8s.io/v1) cluster
(upsert) clusterrolebinding/istio-galley-admin-role-binding-istio-system (rbac.authorization.k8s.io/v1) cluster
(upsert) rolebinding/istio-ingressgateway-sds (rbac.authorization.k8s.io/v1) namespace: istio-system
(upsert) clusterrolebinding/istio-sidecar-injector-admin-role-binding-istio-system (rbac.authorization.k8s.io/v1) cluster
(upsert) clusterrolebinding/istio-pilot-istio-system (rbac.authorization.k8s.io/v1) cluster
(upsert) clusterrolebinding/istio-policy-admin-role-binding-istio-system (rbac.authorization.k8s.io/v1) cluster
(upsert) clusterrolebinding/istio-mixer-admin-role-binding-istio-system (rbac.authorization.k8s.io/v1) cluster
(upsert) clusterrolebinding/metacontroller (rbac.authorization.k8s.io/v1) cluster
---
(upsert) deployment/capi-api-server (apps/v1) namespace: cf-system
(upsert) deployment/capi-kpack-watcher (apps/v1) namespace: cf-system
(upsert) builder/cf-autodetect-builder (build.pivotal.io/v1alpha1) namespace: cf-workloads-staging
(upsert) deployment/capi-clock (apps/v1) namespace: cf-system
(upsert) deployment/capi-deployment-updater (apps/v1) namespace: cf-system
(upsert) service/capi (v1) namespace: cf-system
(upsert) deployment/capi-worker (apps/v1) namespace: cf-system
(upsert) virtualservice/capi-external-virtual-service (networking.istio.io/v1alpha3) namespace: cf-system
(upsert) service/eirini (v1) namespace: cf-system
(upsert) deployment/eirini (apps/v1) namespace: cf-system
(upsert) gateway/istio-ingressgateway (networking.istio.io/v1alpha3) namespace: cf-system
(upsert) deployment/kpack-controller (apps/v1) namespace: kpack
(upsert) service/kpack-webhook (v1) namespace: kpack
(upsert) deployment/kpack-webhook (apps/v1) namespace: kpack
(upsert) policy/cf-blobstore-allow-plaintext (authentication.istio.io/v1alpha1) namespace: cf-blobstore
(upsert) service/log-cache (v1) namespace: cf-system
(upsert) service/log-cache-syslog (v1) namespace: cf-system
(upsert) virtualservice/log-cache-external-virtual-service (networking.istio.io/v1alpha3) namespace: cf-system
(upsert) daemonset/fluentd (apps/v1) namespace: cf-system
(upsert) deployment/log-cache (apps/v1) namespace: cf-system
(upsert) service/metric-proxy (v1) namespace: cf-system
(upsert) deployment/metric-proxy (apps/v1) namespace: cf-system
(upsert) service/cf-blobstore-minio (v1) namespace: cf-blobstore
(upsert) deployment/cf-blobstore-minio (apps/v1) namespace: cf-blobstore
(upsert) deployment/cfroutesync (apps/v1) namespace: cf-system
(upsert) service/cfroutesync (v1) namespace: cf-system
(upsert) routebulksync/route-bulk-sync (apps.cloudfoundry.org/v1alpha1) namespace: cf-workloads
(upsert) authorizationpolicy/cfroutesync-auth-metacontroller (security.istio.io/v1beta1) namespace: cf-system
(upsert) authorizationpolicy/cfroutesync-auth-prometheus (security.istio.io/v1beta1) namespace: cf-system
(upsert) handler/cf-prometheus (config.istio.io/v1alpha2) namespace: istio-system
(upsert) instance/cfrequestcount (config.istio.io/v1alpha2) namespace: istio-system
(upsert) rule/cf-promhttp (config.istio.io/v1alpha2) namespace: istio-system
(upsert) gateway/istio-ingress (networking.istio.io/v1alpha3) namespace: cf-workloads
(upsert) deployment/istio-citadel (apps/v1) namespace: istio-system
(upsert) service/istio-citadel (v1) namespace: istio-system
(upsert) deployment/istio-galley (apps/v1) namespace: istio-system
(upsert) service/istio-galley (v1) namespace: istio-system
(upsert) daemonset/istio-ingressgateway (apps/v1) namespace: istio-system
(upsert) gateway/ingressgateway (networking.istio.io/v1alpha3) namespace: istio-system
(upsert) service/istio-ingressgateway (v1) namespace: istio-system
(upsert) sidecar/default (networking.istio.io/v1alpha3) namespace: istio-system
(upsert) deployment/istio-sidecar-injector (apps/v1) namespace: istio-system
(upsert) service/istio-sidecar-injector (v1) namespace: istio-system
(upsert) horizontalpodautoscaler/istio-pilot (autoscaling/v2beta1) namespace: istio-system
(upsert) deployment/istio-pilot (apps/v1) namespace: istio-system
(upsert) service/istio-pilot (v1) namespace: istio-system
(upsert) horizontalpodautoscaler/istio-policy (autoscaling/v2beta1) namespace: istio-system
(upsert) destinationrule/istio-policy (networking.istio.io/v1alpha3) namespace: istio-system
(upsert) deployment/istio-policy (apps/v1) namespace: istio-system
(upsert) service/istio-policy (v1) namespace: istio-system
(upsert) horizontalpodautoscaler/istio-telemetry (autoscaling/v2beta1) namespace: istio-system
(upsert) attributemanifest/istioproxy (config.istio.io/v1alpha2) namespace: istio-system
(upsert) attributemanifest/kubernetes (config.istio.io/v1alpha2) namespace: istio-system
(upsert) instance/requestcount (config.istio.io/v1alpha2) namespace: istio-system
(upsert) instance/requestduration (config.istio.io/v1alpha2) namespace: istio-system
(upsert) instance/requestsize (config.istio.io/v1alpha2) namespace: istio-system
(upsert) instance/responsesize (config.istio.io/v1alpha2) namespace: istio-system
(upsert) instance/tcpbytesent (config.istio.io/v1alpha2) namespace: istio-system
(upsert) instance/tcpbytereceived (config.istio.io/v1alpha2) namespace: istio-system
(upsert) instance/tcpconnectionsopened (config.istio.io/v1alpha2) namespace: istio-system
(upsert) instance/tcpconnectionsclosed (config.istio.io/v1alpha2) namespace: istio-system
(upsert) handler/prometheus (config.istio.io/v1alpha2) namespace: istio-system
(upsert) rule/promhttp (config.istio.io/v1alpha2) namespace: istio-system
(upsert) rule/promtcp (config.istio.io/v1alpha2) namespace: istio-system
(upsert) rule/promtcpconnectionopen (config.istio.io/v1alpha2) namespace: istio-system
(upsert) rule/promtcpconnectionclosed (config.istio.io/v1alpha2) namespace: istio-system
(upsert) handler/kubernetesenv (config.istio.io/v1alpha2) namespace: istio-system
(upsert) rule/kubeattrgenrulerule (config.istio.io/v1alpha2) namespace: istio-system
(upsert) rule/tcpkubeattrgenrulerule (config.istio.io/v1alpha2) namespace: istio-system
(upsert) instance/attributes (config.istio.io/v1alpha2) namespace: istio-system
(upsert) destinationrule/istio-telemetry (networking.istio.io/v1alpha3) namespace: istio-system
(upsert) deployment/istio-telemetry (apps/v1) namespace: istio-system
(upsert) service/istio-telemetry (v1) namespace: istio-system
(upsert) sidecar/default (networking.istio.io/v1alpha3) namespace: cf-workloads
(upsert) statefulset/metacontroller (apps/v1) namespace: metacontroller
(upsert) service/cf-db-postgresql-headless (v1) namespace: cf-db
(upsert) service/cf-db-postgresql (v1) namespace: cf-db
(upsert) statefulset/cf-db-postgresql (apps/v1) namespace: cf-db
(upsert) policy/cf-db-allow-plaintext (authentication.istio.io/v1alpha1) namespace: cf-db
(upsert) deployment/uaa (apps/v1) namespace: cf-system
(upsert) service/uaa (v1) namespace: cf-system
(upsert) virtualservice/uaa-external-virtual-service (networking.istio.io/v1alpha3) namespace: cf-system
`

	cfForK8sExpectedOutputDelete = `
(delete) builder/cf-autodetect-builder (build.pivotal.io/v1alpha1) namespace: cf-workloads-staging
(delete) virtualservice/capi-external-virtual-service (networking.istio.io/v1alpha3) namespace: cf-system
(delete) gateway/istio-ingressgateway (networking.istio.io/v1alpha3) namespace: cf-system
(delete) policy/cf-blobstore-allow-plaintext (authentication.istio.io/v1alpha1) namespace: cf-blobstore
(delete) virtualservice/log-cache-external-virtual-service (networking.istio.io/v1alpha3) namespace: cf-system
(delete) compositecontroller/cfroutesync (metacontroller.k8s.io/v1alpha1) cluster
(delete) routebulksync/route-bulk-sync (apps.cloudfoundry.org/v1alpha1) namespace: cf-workloads
(delete) authorizationpolicy/cfroutesync-auth-metacontroller (security.istio.io/v1beta1) namespace: cf-system
(delete) authorizationpolicy/cfroutesync-auth-prometheus (security.istio.io/v1beta1) namespace: cf-system
(delete) handler/cf-prometheus (config.istio.io/v1alpha2) namespace: istio-system
(delete) instance/cfrequestcount (config.istio.io/v1alpha2) namespace: istio-system
(delete) rule/cf-promhttp (config.istio.io/v1alpha2) namespace: istio-system
(delete) gateway/istio-ingress (networking.istio.io/v1alpha3) namespace: cf-workloads
(delete) gateway/ingressgateway (networking.istio.io/v1alpha3) namespace: istio-system
(delete) sidecar/default (networking.istio.io/v1alpha3) namespace: istio-system
(delete) meshpolicy/default (authentication.istio.io/v1alpha1) cluster
(delete) destinationrule/istio-policy (networking.istio.io/v1alpha3) namespace: istio-system
(delete) attributemanifest/istioproxy (config.istio.io/v1alpha2) namespace: istio-system
(delete) attributemanifest/kubernetes (config.istio.io/v1alpha2) namespace: istio-system
(delete) instance/requestcount (config.istio.io/v1alpha2) namespace: istio-system
(delete) instance/requestduration (config.istio.io/v1alpha2) namespace: istio-system
(delete) instance/requestsize (config.istio.io/v1alpha2) namespace: istio-system
(delete) instance/responsesize (config.istio.io/v1alpha2) namespace: istio-system
(delete) instance/tcpbytesent (config.istio.io/v1alpha2) namespace: istio-system
(delete) instance/tcpbytereceived (config.istio.io/v1alpha2) namespace: istio-system
(delete) instance/tcpconnectionsopened (config.istio.io/v1alpha2) namespace: istio-system
(delete) instance/tcpconnectionsclosed (config.istio.io/v1alpha2) namespace: istio-system
(delete) handler/prometheus (config.istio.io/v1alpha2) namespace: istio-system
(delete) rule/promhttp (config.istio.io/v1alpha2) namespace: istio-system
(delete) rule/promtcp (config.istio.io/v1alpha2) namespace: istio-system
(delete) rule/promtcpconnectionopen (config.istio.io/v1alpha2) namespace: istio-system
(delete) rule/promtcpconnectionclosed (config.istio.io/v1alpha2) namespace: istio-system
(delete) handler/kubernetesenv (config.istio.io/v1alpha2) namespace: istio-system
(delete) rule/kubeattrgenrulerule (config.istio.io/v1alpha2) namespace: istio-system
(delete) rule/tcpkubeattrgenrulerule (config.istio.io/v1alpha2) namespace: istio-system
(delete) instance/attributes (config.istio.io/v1alpha2) namespace: istio-system
(delete) destinationrule/istio-telemetry (networking.istio.io/v1alpha3) namespace: istio-system
(delete) sidecar/default (networking.istio.io/v1alpha3) namespace: cf-workloads
(delete) policy/cf-db-allow-plaintext (authentication.istio.io/v1alpha1) namespace: cf-db
(delete) virtualservice/uaa-external-virtual-service (networking.istio.io/v1alpha3) namespace: cf-system
---
(delete) customresourcedefinition/builds.build.pivotal.io (apiextensions.k8s.io/v1beta1) cluster
(delete) customresourcedefinition/builders.build.pivotal.io (apiextensions.k8s.io/v1beta1) cluster
(delete) customresourcedefinition/clusterbuilders.build.pivotal.io (apiextensions.k8s.io/v1beta1) cluster
(delete) customresourcedefinition/custombuilders.experimental.kpack.pivotal.io (apiextensions.k8s.io/v1beta1) cluster
(delete) customresourcedefinition/customclusterbuilders.experimental.kpack.pivotal.io (apiextensions.k8s.io/v1beta1) cluster
(delete) customresourcedefinition/images.build.pivotal.io (apiextensions.k8s.io/v1beta1) cluster
(delete) customresourcedefinition/sourceresolvers.build.pivotal.io (apiextensions.k8s.io/v1beta1) cluster
(delete) customresourcedefinition/stacks.experimental.kpack.pivotal.io (apiextensions.k8s.io/v1beta1) cluster
(delete) customresourcedefinition/stores.experimental.kpack.pivotal.io (apiextensions.k8s.io/v1beta1) cluster
(delete) customresourcedefinition/routebulksyncs.apps.cloudfoundry.org (apiextensions.k8s.io/v1beta1) cluster
(delete) customresourcedefinition/attributemanifests.config.istio.io (apiextensions.k8s.io/v1beta1) cluster
(delete) customresourcedefinition/clusterrbacconfigs.rbac.istio.io (apiextensions.k8s.io/v1beta1) cluster
(delete) customresourcedefinition/destinationrules.networking.istio.io (apiextensions.k8s.io/v1beta1) cluster
(delete) customresourcedefinition/envoyfilters.networking.istio.io (apiextensions.k8s.io/v1beta1) cluster
(delete) customresourcedefinition/gateways.networking.istio.io (apiextensions.k8s.io/v1beta1) cluster
(delete) customresourcedefinition/httpapispecbindings.config.istio.io (apiextensions.k8s.io/v1beta1) cluster
(delete) customresourcedefinition/httpapispecs.config.istio.io (apiextensions.k8s.io/v1beta1) cluster
(delete) customresourcedefinition/meshpolicies.authentication.istio.io (apiextensions.k8s.io/v1beta1) cluster
(delete) customresourcedefinition/policies.authentication.istio.io (apiextensions.k8s.io/v1beta1) cluster
(delete) customresourcedefinition/quotaspecbindings.config.istio.io (apiextensions.k8s.io/v1beta1) cluster
(delete) customresourcedefinition/quotaspecs.config.istio.io (apiextensions.k8s.io/v1beta1) cluster
(delete) customresourcedefinition/rbacconfigs.rbac.istio.io (apiextensions.k8s.io/v1beta1) cluster
(delete) customresourcedefinition/rules.config.istio.io (apiextensions.k8s.io/v1beta1) cluster
(delete) customresourcedefinition/serviceentries.networking.istio.io (apiextensions.k8s.io/v1beta1) cluster
(delete) customresourcedefinition/servicerolebindings.rbac.istio.io (apiextensions.k8s.io/v1beta1) cluster
(delete) customresourcedefinition/serviceroles.rbac.istio.io (apiextensions.k8s.io/v1beta1) cluster
(delete) customresourcedefinition/virtualservices.networking.istio.io (apiextensions.k8s.io/v1beta1) cluster
(delete) customresourcedefinition/adapters.config.istio.io (apiextensions.k8s.io/v1beta1) cluster
(delete) customresourcedefinition/instances.config.istio.io (apiextensions.k8s.io/v1beta1) cluster
(delete) customresourcedefinition/templates.config.istio.io (apiextensions.k8s.io/v1beta1) cluster
(delete) customresourcedefinition/handlers.config.istio.io (apiextensions.k8s.io/v1beta1) cluster
(delete) customresourcedefinition/sidecars.networking.istio.io (apiextensions.k8s.io/v1beta1) cluster
(delete) customresourcedefinition/authorizationpolicies.security.istio.io (apiextensions.k8s.io/v1beta1) cluster
(delete) customresourcedefinition/compositecontrollers.metacontroller.k8s.io (apiextensions.k8s.io/v1beta1) cluster
(delete) customresourcedefinition/decoratorcontrollers.metacontroller.k8s.io (apiextensions.k8s.io/v1beta1) cluster
(delete) customresourcedefinition/controllerrevisions.metacontroller.k8s.io (apiextensions.k8s.io/v1beta1) cluster
---
(delete) deployment/capi-api-server (apps/v1) namespace: cf-system
(delete) deployment/capi-kpack-watcher (apps/v1) namespace: cf-system
(delete) clusterrole/kpack-watcher (rbac.authorization.k8s.io/v1) cluster
(delete) clusterrolebinding/kpack-watcher-binding (rbac.authorization.k8s.io/v1) cluster
(delete) role/kpack-watcher-pod-logs-reader (rbac.authorization.k8s.io/v1) namespace: cf-workloads
(delete) rolebinding/kpack-watcher-pod-logs-binding (rbac.authorization.k8s.io/v1) namespace: cf-workloads
(delete) configmap/cloud-controller-ng-yaml (v1) namespace: cf-system
(delete) namespace/cf-workloads-staging (v1) cluster
(delete) deployment/capi-clock (apps/v1) namespace: cf-system
(delete) deployment/capi-deployment-updater (apps/v1) namespace: cf-system
(delete) configmap/nginx (v1) namespace: cf-system
(delete) secret/opi-secrets (v1) namespace: cf-system
(delete) serviceaccount/cc-api-service-account (v1) namespace: cf-system
(delete) clusterrolebinding/cc-api-service-account-superuser (rbac.authorization.k8s.io/v1) cluster
(delete) secret/cc-kpack-registry-auth-secret (v1) namespace: cf-workloads-staging
(delete) serviceaccount/cc-kpack-registry-service-account (v1) namespace: cf-workloads-staging
(delete) service/capi (v1) namespace: cf-system
(delete) deployment/capi-worker (apps/v1) namespace: cf-system
(delete) networkpolicy/deny-app-ingress (networking.k8s.io/v1) namespace: cf-workloads
(delete) podsecuritypolicy/cf-workloads-app-psp (policy/v1beta1) cluster
(delete) podsecuritypolicy/cf-workloads-privileged-app-psp (policy/v1beta1) cluster
(delete) podsecuritypolicy/eirini (policy/v1beta1) cluster
(delete) serviceaccount/eirini (v1) namespace: cf-workloads
(delete) serviceaccount/eirini-privileged (v1) namespace: cf-workloads
(delete) serviceaccount/opi (v1) namespace: cf-system
(delete) configmap/eirini (v1) namespace: cf-system
(delete) role/cf-workloads-app-role (rbac.authorization.k8s.io/v1) namespace: cf-workloads
(delete) role/cf-workloads-privileged-app-role (rbac.authorization.k8s.io/v1) namespace: cf-workloads
(delete) role/eirini-role (rbac.authorization.k8s.io/v1) namespace: cf-workloads
(delete) role/eirini-role (rbac.authorization.k8s.io/v1) namespace: cf-system
(delete) rolebinding/cf-workloads-app-rolebinding (rbac.authorization.k8s.io/v1) namespace: cf-workloads
(delete) rolebinding/cf-workloads-privileged-app-rolebinding (rbac.authorization.k8s.io/v1) namespace: cf-workloads
(delete) rolebinding/eirini-rolebinding (rbac.authorization.k8s.io/v1) namespace: cf-workloads
(delete) rolebinding/eirini-rolebinding (rbac.authorization.k8s.io/v1) namespace: cf-system
(delete) service/eirini (v1) namespace: cf-system
(delete) deployment/eirini (apps/v1) namespace: cf-system
(delete) secret/eirini-internal-tls-certs (v1) namespace: cf-system
(delete) networkpolicy/allow-app-ingress-from-ingressgateway (networking.k8s.io/v1) namespace: cf-workloads
(delete) secret/app-registry-credentials (v1) namespace: cf-workloads
(delete) secret/istio-ingressgateway-certs (v1) namespace: istio-system
(delete) namespace/kpack (v1) cluster
(delete) deployment/kpack-controller (apps/v1) namespace: kpack
(delete) serviceaccount/controller (v1) namespace: kpack
(delete) clusterrole/kpack-controller-admin (rbac.authorization.k8s.io/v1) cluster
(delete) clusterrolebinding/kpack-controller-admin-binding (rbac.authorization.k8s.io/v1) cluster
(delete) service/kpack-webhook (v1) namespace: kpack
(delete) mutatingwebhookconfiguration/resource.webhook.kpack.pivotal.io (admissionregistration.k8s.io/v1beta1) cluster
(delete) deployment/kpack-webhook (apps/v1) namespace: kpack
(delete) serviceaccount/webhook (v1) namespace: kpack
(delete) role/kpack-webhook-certs-admin (rbac.authorization.k8s.io/v1) namespace: kpack
(delete) rolebinding/kpack-webhook-certs-admin-binding (rbac.authorization.k8s.io/v1) namespace: kpack
(delete) clusterrole/kpack-webhook-mutatingwebhookconfiguration-admin (rbac.authorization.k8s.io/v1) cluster
(delete) clusterrolebinding/kpack-webhook-certs-mutatingwebhookconfiguration-admin-binding (rbac.authorization.k8s.io/v1) cluster
(delete) serviceaccount/fluentd-service-account (v1) namespace: cf-system
(delete) clusterrole/pod-namespace-read (rbac.authorization.k8s.io/v1) namespace: cf-system
(delete) clusterrolebinding/fluentd-service-account-pod-namespace-read (rbac.authorization.k8s.io/v1) cluster
(delete) configmap/fluentd-config (v1) namespace: cf-system
(delete) service/log-cache (v1) namespace: cf-system
(delete) service/log-cache-syslog (v1) namespace: cf-system
(delete) secret/log-cache-ca (v1) namespace: cf-system
(delete) secret/log-cache (v1) namespace: cf-system
(delete) secret/log-cache-metrics (v1) namespace: cf-system
(delete) secret/log-cache-gateway (v1) namespace: cf-system
(delete) secret/log-cache-syslog (v1) namespace: cf-system
(delete) daemonset/fluentd (apps/v1) namespace: cf-system
(delete) deployment/log-cache (apps/v1) namespace: cf-system
(delete) serviceaccount/metric-proxy (v1) namespace: cf-system
(delete) clusterrole/metric-proxy (rbac.authorization.k8s.io/v1) namespace: cf-system
(delete) clusterrolebinding/metric-proxy (rbac.authorization.k8s.io/v1) cluster
(delete) service/metric-proxy (v1) namespace: cf-system
(delete) secret/metric-proxy-cert (v1) namespace: cf-system
(delete) secret/metric-proxy-ca (v1) namespace: cf-system
(delete) deployment/metric-proxy (apps/v1) namespace: cf-system
(delete) namespace/cf-blobstore (v1) cluster
(delete) secret/cf-blobstore-minio (v1) namespace: cf-blobstore
(delete) configmap/cf-blobstore-minio (v1) namespace: cf-blobstore
(delete) persistentvolumeclaim/cf-blobstore-minio (v1) namespace: cf-blobstore
(delete) serviceaccount/cf-blobstore-minio (v1) namespace: cf-blobstore
(delete) service/cf-blobstore-minio (v1) namespace: cf-blobstore
(delete) deployment/cf-blobstore-minio (apps/v1) namespace: cf-blobstore
(delete) configmap/cfroutesync-config (v1) namespace: cf-system
(delete) secret/cfroutesync (v1) namespace: cf-system
(delete) deployment/cfroutesync (apps/v1) namespace: cf-system
(delete) service/cfroutesync (v1) namespace: cf-system
(delete) clusterrole/istio-reader-istio-system (rbac.authorization.k8s.io/v1) cluster
(delete) clusterrolebinding/istio-reader-istio-system (rbac.authorization.k8s.io/v1) cluster
(delete) namespace/istio-system (v1) cluster
(delete) serviceaccount/istio-reader-service-account (v1) namespace: istio-system
(delete) clusterrole/istio-citadel-istio-system (rbac.authorization.k8s.io/v1) cluster
(delete) clusterrolebinding/istio-citadel-istio-system (rbac.authorization.k8s.io/v1) cluster
(delete) deployment/istio-citadel (apps/v1) namespace: istio-system
(delete) poddisruptionbudget/istio-citadel (policy/v1beta1) namespace: istio-system
(delete) service/istio-citadel (v1) namespace: istio-system
(delete) serviceaccount/istio-citadel-service-account (v1) namespace: istio-system
(delete) clusterrole/istio-galley-istio-system (rbac.authorization.k8s.io/v1) cluster
(delete) clusterrolebinding/istio-galley-admin-role-binding-istio-system (rbac.authorization.k8s.io/v1) cluster
(delete) configmap/galley-envoy-config (v1) namespace: istio-system
(delete) configmap/istio-mesh-galley (v1) namespace: istio-system
(delete) configmap/istio-galley-configuration (v1) namespace: istio-system
(delete) deployment/istio-galley (apps/v1) namespace: istio-system
(delete) poddisruptionbudget/istio-galley (policy/v1beta1) namespace: istio-system
(delete) service/istio-galley (v1) namespace: istio-system
(delete) serviceaccount/istio-galley-service-account (v1) namespace: istio-system
(delete) daemonset/istio-ingressgateway (apps/v1) namespace: istio-system
(delete) poddisruptionbudget/ingressgateway (policy/v1beta1) namespace: istio-system
(delete) role/istio-ingressgateway-sds (rbac.authorization.k8s.io/v1) namespace: istio-system
(delete) rolebinding/istio-ingressgateway-sds (rbac.authorization.k8s.io/v1) namespace: istio-system
(delete) service/istio-ingressgateway (v1) namespace: istio-system
(delete) serviceaccount/istio-ingressgateway-service-account (v1) namespace: istio-system
(delete) clusterrole/istio-sidecar-injector-istio-system (rbac.authorization.k8s.io/v1) cluster
(delete) clusterrolebinding/istio-sidecar-injector-admin-role-binding-istio-system (rbac.authorization.k8s.io/v1) cluster
(delete) configmap/injector-mesh (v1) namespace: istio-system
(delete) deployment/istio-sidecar-injector (apps/v1) namespace: istio-system
(delete) mutatingwebhookconfiguration/istio-sidecar-injector (admissionregistration.k8s.io/v1beta1) cluster
(delete) poddisruptionbudget/istio-sidecar-injector (policy/v1beta1) namespace: istio-system
(delete) service/istio-sidecar-injector (v1) namespace: istio-system
(delete) serviceaccount/istio-sidecar-injector-service-account (v1) namespace: istio-system
(delete) configmap/istio-sidecar-injector (v1) namespace: istio-system
(delete) horizontalpodautoscaler/istio-pilot (autoscaling/v2beta1) namespace: istio-system
(delete) clusterrole/istio-pilot-istio-system (rbac.authorization.k8s.io/v1) cluster
(delete) clusterrolebinding/istio-pilot-istio-system (rbac.authorization.k8s.io/v1) cluster
(delete) configmap/pilot-envoy-config (v1) namespace: istio-system
(delete) configmap/istio (v1) namespace: istio-system
(delete) deployment/istio-pilot (apps/v1) namespace: istio-system
(delete) poddisruptionbudget/istio-pilot (policy/v1beta1) namespace: istio-system
(delete) service/istio-pilot (v1) namespace: istio-system
(delete) serviceaccount/istio-pilot-service-account (v1) namespace: istio-system
(delete) horizontalpodautoscaler/istio-policy (autoscaling/v2beta1) namespace: istio-system
(delete) clusterrole/istio-policy (rbac.authorization.k8s.io/v1) cluster
(delete) clusterrolebinding/istio-policy-admin-role-binding-istio-system (rbac.authorization.k8s.io/v1) cluster
(delete) configmap/policy-envoy-config (v1) namespace: istio-system
(delete) deployment/istio-policy (apps/v1) namespace: istio-system
(delete) poddisruptionbudget/istio-policy (policy/v1beta1) namespace: istio-system
(delete) service/istio-policy (v1) namespace: istio-system
(delete) serviceaccount/istio-policy-service-account (v1) namespace: istio-system
(delete) horizontalpodautoscaler/istio-telemetry (autoscaling/v2beta1) namespace: istio-system
(delete) clusterrole/istio-mixer-istio-system (rbac.authorization.k8s.io/v1) cluster
(delete) clusterrolebinding/istio-mixer-admin-role-binding-istio-system (rbac.authorization.k8s.io/v1) cluster
(delete) configmap/telemetry-envoy-config (v1) namespace: istio-system
(delete) deployment/istio-telemetry (apps/v1) namespace: istio-system
(delete) poddisruptionbudget/istio-telemetry (policy/v1beta1) namespace: istio-system
(delete) service/istio-telemetry (v1) namespace: istio-system
(delete) serviceaccount/istio-mixer-service-account (v1) namespace: istio-system
(delete) networkpolicy/pilot-network-policy (networking.k8s.io/v1) namespace: istio-system
(delete) networkpolicy/citadel-network-policy (networking.k8s.io/v1) namespace: istio-system
(delete) networkpolicy/mixer-network-policy (networking.k8s.io/v1) namespace: istio-system
(delete) networkpolicy/sidecar-injector-network-policy (networking.k8s.io/v1) namespace: istio-system
(delete) networkpolicy/ingressgateway-network-policy-pilot (networking.k8s.io/v1) namespace: istio-system
(delete) namespace/metacontroller (v1) cluster
(delete) serviceaccount/metacontroller (v1) namespace: metacontroller
(delete) clusterrole/metacontroller (rbac.authorization.k8s.io/v1) cluster
(delete) clusterrolebinding/metacontroller (rbac.authorization.k8s.io/v1) cluster
(delete) clusterrole/aggregate-metacontroller-view (rbac.authorization.k8s.io/v1) cluster
(delete) clusterrole/aggregate-metacontroller-edit (rbac.authorization.k8s.io/v1) cluster
(delete) statefulset/metacontroller (apps/v1) namespace: metacontroller
(delete) namespace/cf-db (v1) cluster
(delete) secret/cf-db-admin-secret (v1) namespace: cf-db
(delete) secret/cf-db-credentials (v1) namespace: cf-db
(delete) configmap/cf-db-postgresql-init-scripts (v1) namespace: cf-db
(delete) service/cf-db-postgresql-headless (v1) namespace: cf-db
(delete) service/cf-db-postgresql (v1) namespace: cf-db
(delete) statefulset/cf-db-postgresql (apps/v1) namespace: cf-db
(delete) namespace/cf-system (v1) cluster
(delete) configmap/uaa-config (v1) namespace: cf-system
(delete) deployment/uaa (apps/v1) namespace: cf-system
(delete) service/uaa (v1) namespace: cf-system
(delete) serviceaccount/uaa (v1) namespace: cf-system
(delete) secret/uaa-certs (v1) namespace: cf-system
(delete) namespace/cf-workloads (v1) cluster
`
)
