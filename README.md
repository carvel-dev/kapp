# kapp

- Website: https://get-kapp.io
- Slack: [#k14s in Kubernetes slack](https://slack.kubernetes.io)

`kapp` CLI encourages Kubernetes users to manage resources in bulk by working with "Kubernetes applications" (sets of resources with the same label). It focuses on resource diffing, labeling, deployment and deletion. Unlike tools like Helm, `kapp` considers YAML templating and management of packages outside of its scope, though it works great with tools that generate Kubernetes configuration.

![](docs/kapp-deploy-screenshot.png)

See [https://get-kapp.io](https://get-kapp.io) for detailed example workflow.

Features:

- Works with standard Kubernetes YAMLs
- Focuses exclusively on deployment workflow, not packaging or templating
  - but plays well with tools (such as ytt) that produce Kubernetes configuration
- Converges application resources (creates, updates and/or deletes resources) in each deploy
  - based on comparison between provided files and live objects in the cluster
- Separates calculation of changes (diff stage) from application of changes (apply stage)
- Works without admin privileges and does not depend on any custom CRDs
  - making it possible to use kapp as a regular user in a single namespace
- Records application deployment history
- Opt-in resource version management
  - for example, to trigger Deployment rollout when ConfigMap changes
- Works with any group of labeled resources (`kapp -a label:tier=web inspect -t`)
- Works without server side components
- GitOps friendly (`kapp app-group deploy -g all-apps --directory .`)

## Docs

- [Docs](docs/README.md) with topics about diff, apply, gitops, config, talks

## Install

Grab prebuilt binaries from the [Releases page](https://github.com/k14s/kapp/releases).

## Development

```bash
./hack/build.sh
export KAPP_E2E_NAMESPACE=kapp-test
./hack/test-all.sh

# include goog analytics in 'kapp website' command for https://get-kapp.io
# (goog analytics is _not_ included in release binaries)
BUILD_VALUES=./hack/build-values-get-kapp-io.yml ./hack/build.sh
```

`build.sh` depends on [ytt](https://github.com/k14s/ytt).
