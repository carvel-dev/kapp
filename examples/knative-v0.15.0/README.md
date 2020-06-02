Downloaded via

```bash
ytt --ignore-unknown-comments \
  -f https://github.com/knative/serving/releases/download/v0.15.0/serving-storage-version-migration.yaml \
  -f https://github.com/knative/serving/releases/download/v0.15.0/serving-crds.yaml \
  -f https://github.com/knative/serving/releases/download/v0.15.0/serving-core.yaml \
  -f https://github.com/knative/net-kourier/releases/download/v0.15.0/kourier.yaml \
  -f https://github.com/knative/serving/releases/download/v0.15.0/monitoring.yaml \
-f https://github.com/knative/eventing/releases/download/v0.15.0/eventing.yaml > config.yml
```
