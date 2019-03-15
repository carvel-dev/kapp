## Using kapp with GitOps workflow

kapp provides a set of commands to make GitOps workflow very easy. Assuming that you have a CI environment or some other place where `kapp` can run based on a trigger (e.g. for every Git repo change) or continiously (e.g. every 5 mins), you can use following command:

```bash
$ ls my-repo
.    ..    app1/    app2/    app3/

$ kapp app-group deploy -g my-env --directory my-repo
```

Above command will deploy an application for each subdirectory in `my-repo` directory (in this case `app1`, `app2` and `app3`). It will also remove old applications if subdirectories are deleted.
