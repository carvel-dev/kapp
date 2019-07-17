## FAQ

### `Error: Asking for confirmation: EOF`

This probably means you have piped configuration into kapp and did not specify `--yes` (`-y`) flag to continue. It's necessary because kapp can no longer ask for confirmation via stdin. Feel free to re-run the command with `--diff-changes` (`-c`) to make sure pending changes are correct.

### Where to store app resources (i.e. in which namespace)?

`kapp ls` shows list of applications saved in a specified namespace via `-n` flag or whatever namespace selected via `~/.kube/config`. `--all-namespaces` flag can be used to list applications from all namespaces.

To be able to show list of deployed applications, kapp creates (and deletes) ConfigMap for each saved application. Each ConfigMap contains generated label used to label all application resources, plus some other misc. metadata.

There are currently two options where to store these ConfigMaps:

- for each application, keep both ConfigMap representing app and app resources themselves in the same namespace. That namespace will have to be created outside of kapp since kapp will first want to create a ConfigMap representing application.

    ```bash
    $ kubectl create ns app1
    $ kapp deploy -n app1 -f config.yml
    $ kapp ls -n app1
    ```

- create a dedicated namespace (i.e. "state" namespace) to store all ConfigMaps representing apps, and have kapp create Namespace resources for applications from their config.

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
    ...
    ```

    With this approach namespace management (creation, and deletion) is tied to a particular app which makes it a bit easier to track them via configuration.

Note: It's currently not possible to have kapp place app ConfigMap resource into Namespace that kapp creates for that application.
