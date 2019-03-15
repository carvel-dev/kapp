# Knative

Downloaded from: https://github.com/knative/serving/releases/tag/v0.2.1

Before installing Knative you must install [Istio](../istio-v1.0.2).

## Small footprint Knative

```
$ kapp deploy -a kn -f examples/knative-*/release-no-mon.yml
```

## Full Knative installation with monitoring

```
$ kapp deploy -a kn -f examples/knative-*/release.yml
```
