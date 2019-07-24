### Apply orderding

By default kapp has few builtin change rules to make sure:

- CRDs and namespaces are created before most resources
- CRDs are deleted last (after CRs)

kapp also allows to customize change order via following resource annotations:

- `kapp.k14s.io/change-group` annotation to group one or more resource changes into arbitrarily named group. Example: `apps.big.co/db-migrations`.
- `kapp.k14s.io/change-rule` annotation to control when resource change should be applied (created, updated, or deleted) relative to other changes. You can specify multiple change rules by suffixing each annotation with a `.x` where `x` is some number (e.g. `kapp.k14s.io/change-rule.1`).

`kapp.k14s.io/change-rule` format is as follows: `(upsert|delete) (after|before) (upserting|deleting) <name>`. For example:

- `kapp.k14s.io/change-rule: "upsert after upserting apps.big.co/db-migrations"`
- `kapp.k14s.io/change-rule: "delete before upserting apps.big.co/service"`

#### Example

Following example shows how to run `job/migrations`, start and wait for `deployment/app`, and finally `job/app-health-check`.

```yaml
kind: ConfigMap
metadata:
  name: app-config
  annotations: {}
#...
---
kind: Job
metadata:
  name: migrations
  annotations:
    kapp.k14s.io/change-group: "apps.big.co/db-migrations"
#...
---
kind: Service
metadata:
  name: app
  annotations:
    kapp.k14s.io/change-group: "apps.big.co/deployment"
#...
---
kind: Ingress
metadata:
  name: app
  annotations:
    kapp.k14s.io/change-group: "apps.big.co/deployment"
#...
---
kind: Deployment
metadata:
  name: app
  annotations:
    kapp.k14s.io/change-group: "apps.big.co/deployment"
    kapp.k14s.io/change-rule: "upsert after upserting apps.big.co/db-migrations"
#...
---
kind: Job
metadata:
  name: app-health-check
  annotations:
    kapp.k14s.io/change-rule: "upsert after upserting apps.big.co/deployment"
#...
```
