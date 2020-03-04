## Cheatsheet

### List

- `kapp ls -A`
  - List all app in the cluster (across all namespaces)

### Deploy

- `kapp deploy -a app1 -f config/ -c`
  - Deploy app named `app1` with configuration from `config/`

- `ytt -f config/ | kapp deploy -a app1 -f- -c -y`
  - Deploy app named `app1` with configuration piped in

- `kapp deploy -a app1 -f <(ytt -f config/ )`
  - Deploy app named `app1` with configuration generated inline and with confirmation dialog

- `kapp deploy -a app1 -f config/ -c --diff-context=10`
  - Show more diff context when reviewing changes during deploy

- `kapp deploy -a app1 -f config/ --diff-run`
  - Show diff and exit successfully (without applying any changes)

- `kapp deploy -a app1 -f config/ --logs-all`
  - Show logs from all app `Pods` throughout deploy

- `kapp deploy -a app1 -f config/ --into-ns app1-ns`
  - Rewrite all resources to specify `app1-ns` namespace

### Inspect

- `kapp inspect -a app1`
  - Show summary of all resources in app `app1`

- `kapp inspect -a app1 --tree`
  - Show summary organized as a tree of all resources in app `app1`

- `kapp inspect -a app1 --status`
  - Show status subresources for each resource in app `app1`

- `kapp inspect -a 'label:'`
  - Show all resources in the cluster

- `kapp inspect -a 'label:' --filter-ns some-ns`
  - Show all resources in particular namespace (note that it currently does namespace filtering client-side)

- `kapp inspect -a 'label:tier=web'`
  - Show all resources labeled `tier=web` in the cluster

- `kapp inspect -a 'label:!kapp.k14s.io/app' --filter-kind Deployment`
  - Show all `Deployment` resources in the cluster **not** managed by kapp

- `kapp tools list-labels`
  - See which labels are used in your cluster (add `--values` to see label values)
  
- `kapp tools list-labels --values --tty=false | grep kapp.k14s.io/app`
  - Shows app labels that are still present in the cluster (could be combined with delete command below)

### Delete

- `kapp delete -a 'label:kapp.k14s.io/app=1578599579922603000'`
  - Delete resources under particular label (in this example deleting resources associated with some app)

### Misc

- `kapp deploy -a label:kapp.k14s.io/is-app-change= --filter-age 500h+ --dangerous-allow-empty-list-of-resources --apply-ignored`
  - Delete all app changes older than 500h (v0.12.0+)
