# Rebase rules for Pods on Calico

Kapp will [merge](https://github.com/vmware-tanzu/carvel-kapp/blob/develop/docs/merge-method.md) resources
with what's in the cluster to determine when to apply changes. The merge can be customized
with [rebaseRules](https://github.com/vmware-tanzu/carvel-kapp/blob/develop/docs/config.md).

In the case of Calico pods get an annotation with their IP address in it. By default this
annotation will be removed by kapp during a deployment, which will fail as an invalid pod
change. Running kapp with the `-c` flag will produces an error similar to this one.

```bash
@@ update pod/bitwarden-http-test (v1) namespace: bitwarden @@
  ...
  3,  3     annotations:
  4     -     cni.projectcalico.org/podIP: 10.1.108.198/32
  5     -     cni.projectcalico.org/podIPs: 10.0.108.198/32
```

The above error was produced deploying the example pod in this folder. The first deploy will
pass, the second will fail.

Adding the [rebase-rule.yml](rebase-rule.yml) will tell kapp to copy existing annotations
during deployments for all v1 Pods, which matches this example pod.

This is only necessary when directly deploying Pods on Calico, when using a Deployment kapp
will compare changes to the Deployment resource which in turn creates the Pod resource(s) where the annotation is added.
