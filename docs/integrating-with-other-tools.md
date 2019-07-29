## Integrating With Other Tools

This is a non-exhaustive list of tools that are commonly used with kapp.

### ytt and kbld

We recommend to use kapp with [ytt](https://get-ytt.io) and [kbld](https://get-kbld.io) to cover your configuration templating and image building needs. Typical workflow may look like this:

```bash
ytt -f config/ | kbld -f - | kapp deploy -a app1 -f- -c -y
```

### Helm

If you want to take advantage of both Helm templating and kapp deployment mechanisms, you can use `helm template` command to build configuration, and have kapp apply to the cluster:

```bash
helm template ... | kapp deploy -a app1 -f- -c -y
```

### PV labeling controller

If you want to have better visibility into which persistent volumes (PVs) are associated with persistent volume claims (PVCs), you can install [https://github.com/k14s/pv-labeling-controller](https://github.com/k14s/pv-labeling-controller) so that it copies several kapp applied labels to associated PVs. Once that's done you will see PVs in `kapp inspect` output.
