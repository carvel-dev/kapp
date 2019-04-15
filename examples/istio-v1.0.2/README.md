# Istio

Downloaded from: https://github.com/knative/serving/releases/tag/v0.2.1

- `config.yml` can be used to get rid cluster changes
- `patches.yml` can be used with `ytt` to get rid of helm specific and/or template inaccuracies
- `cluster-ip-ingressgw.yml` can be used with `ytt` to switch to ClusterIP cluster-ip-ingressgw
  - Example: `ytt t -f examples/istio-v1.0.2/istio.yml --file-mark istio.yml:type=yaml-plain -f examples/istio-v1.0.2/cluster-ip-ingressgw.yml`
