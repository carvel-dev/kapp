Downloaded via

```bash
ytt --ignore-unknown-comments \
  -f https://github.com/knative/serving/releases/download/knative-v1.8.0/serving-core.yaml \
  -f https://github.com/knative/serving/releases/download/knative-v1.8.0/serving-crds.yaml \
  -f https://github.com/knative/net-istio/releases/download/knative-v1.8.0/istio.yaml \
  -f https://github.com/knative/net-istio/releases/download/knative-v1.8.0/net-istio.yaml \
  -f https://github.com/knative/serving/releases/download/knative-v1.8.0/serving-default-domain.yaml > examples/knative-v1.8.0/config.yml
```
