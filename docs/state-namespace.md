## State namespace

To show list of deployed applications (via `kapp ls`), kapp manages metadata `ConfigMap` for each saved application. Each metadata `ConfigMap` contains generated label used to label all application resources. Additionally kapp creates `ConfigMap` per each deploy to record deployment history (seen via `kapp app-change list -a app1`).

`-n` (`--namespace`) flag allows to control which namespace is used for finding and storing metadata `ConfigMaps`. If namespace is not explicitly specified your current namespace is selected from kube config (typically `~/.kube/config`).

There are currently two approaches to deciding which namespace to use for storing metadata `ConfigMaps`:

- for each application, keep metadata `ConfigMap` and app resources themselves in the same namespace. That namespace will have to be created before running `kapp deploy` since kapp will first want to create a `ConfigMap` representing application.

    ```bash
    $ kubectl create ns app1
    $ kapp deploy -n app1 -f config.yml
    $ kapp ls -n app1
    ```

- create a dedicated namespace to store metadata `ConfigMaps` representing apps, and have kapp create `Namespace` resources for applications from their config. With this approach namespace management (creation and deletion) is tied to a particular app configuration which makes it a bit easier to track `Namespaces` via configuration.

   ```bash
    $ kubectl create ns apps
    $ kapp deploy -n apps -f app1/config.yml
    $ kapp deploy -n apps -f app2/config.yml
    $ kapp ls -n apps
    ```

    for example, `app1/config.yml` may look like this:

    ```yaml
    apiVersion: v1
    kind: Namespace
    metadata:
      name: app1
    ---
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: dep
      namespace: app1
    ...
    ```

Note: It's currently not possible to have kapp place app `ConfigMap` resource into `Namespace` that kapp creates for that application.

### App Changes

As mentioned above, app changes (stored as `ConfigMap`) are stored in state namespace. App changes do not store any information necessary for kapp to operate, but rather act as informational records. There is currently no cap on how many app changes are kept per app.

To remove older app changes, use `kapp app-change gc -a app1` which by default will keep 200 most recent changes (as of v0.12.0). Alternatively use `--app-changes-max-to-keep` flag on the `deploy` command to control number of changes kept at the time of deploy.
