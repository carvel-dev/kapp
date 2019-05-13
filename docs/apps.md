## Apps (Applications)

kapp considers a set of resources with the same label as an application. These resources could span any number of namespaces or could be cluster-wide (e.g. CRDs).

kapp has two methods of finding resources:

1. via unique-to-Namespace application name (via `-a my-name` flag), or
2. via user provided label (via `-a label:my-label=val` flag)

First approach is most common as kapp generates a unique label for each tracked application and associates that with an application name.

### List

Applications can be listed via `ls` command:

```bash
$ kapp ls
```

### Deploy

To create or update an application use `deploy` command:

```bash
$ kapp deploy -a my-name -f my-app-config/
```

Deploy command consists of two stages: [resource "diff" stage](diff.md), and [resource "apply" stage](apply.md).

### Delete

To delete an application use `delete` command:

```bash
$ kapp delete -a my-name
```
