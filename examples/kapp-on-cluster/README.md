# Kapp on cluster

This example shows simple way to run `kapp` continiously in cluster, and converge a set of apps available in a Git repositiory.

To install perform:

```bash
$ ytt -f examples/kapp-on-cluster/ | kapp --yes deploy -a kapp-on-cluster -f - -c
```

See `cron.yml` for details on what this script does every 1m.
