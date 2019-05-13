# Kapp on cluster

This example shows how to run `kapp` continiously in cluster, and converge a set of apps available in a repositiory.

To install perform:

```bash
$ ytt t -f examples/kapp-on-cluster/ | kapp --yes deploy -a kapp-on-cluster -f - --diff-changes
```

See `cron.yml` for details on what this script does every 1m.
