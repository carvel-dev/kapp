## Cheatsheet

- `kapp deploy -a app1 -f config/ -c`
  - Deploy app named `app1` with configuration from `config/`

- `ytt -f config/ | kapp deploy -a app1 -f- -c -y`
  - Deploy app named `app1` with configuration piped in

- `kapp deploy -a app1 -f- <(ytt -f config/ )`
  - Deploy app named `app1` with configuration generated inline and with confirmation dialog

- `kapp deploy -a app1 -f config/ -c --diff-context=10`
  - Show more diff context when reviewing changes during deploy

- `kapp deploy -a app1 -f config/ --diff-run`
  - Show diff and exit successfully (without applying any changes)

- `kapp deploy -a app1 -f config/ --logs-all`
  - Show logs from all app `Pods` throughout deploy

- `kapp deploy -a app1 -f config/ --into-ns app1-ns`
  - Rewrite all resources to specify `app1-ns` namespace

- `kapp inspect -a 'label:'`
  - Show all resources in the cluster

- `kapp inspect -a 'label:tier=web'`
  - Show all resources labeled `tier=web` in the cluster

- `kapp inspect -a 'label:!kapp.k14s.io/app' --filter-kind Deployment`
  - Show all `Deployment` resources in the cluster **not** managed by kapp

- `kapp deploy -a label:kapp.k14s.io/is-app-change= --filter-age 500h+ --dangerous-allow-empty-list-of-resources --apply-ignored`
  - Delete all app changes older than 500h (v0.12.0+)
