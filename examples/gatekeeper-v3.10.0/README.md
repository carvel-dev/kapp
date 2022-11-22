Downloaded via

```bash
ytt \
  -f https://raw.githubusercontent.com/open-policy-agent/gatekeeper/release-3.10/deploy/gatekeeper.yaml \
  -f https://raw.githubusercontent.com/open-policy-agent/gatekeeper/v3.10.0/example/templates/k8srequiredlabels_template.yaml \
  -f https://raw.githubusercontent.com/open-policy-agent/gatekeeper/v3.10.0/example/constraints/all_pod_must_have_gatekeeper_namespaceselector.yaml \
  -f examples/gatekeeper-v3.10.0/exists.yml \
  -f examples/gatekeeper-v3.10.0/overlay.yml > examples/gatekeeper-v3.10.0/config.yml
```

- `exists.yml` is used to wait for CRD created by gatekeeper controller
- `overlay.yml` is used with `ytt` to add ordering
