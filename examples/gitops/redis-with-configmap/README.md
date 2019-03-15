# Redis With Configmap

Based on [Configuring Redis using a ConfigMap](https://kubernetes.io/docs/tutorials/configuration/configure-redis-using-configmap/#real-world-example-configuring-redis-using-a-configmap).

Configuration has been augmented with templating of a ConfigMap so that when it changes redis Pod is recreated (note in this example, redis is not configured to persistent data between Pod recreates).

```bash
kubectl exec -it redis redis-cli
127.0.0.1:6379> CONFIG GET maxmemory
1) "maxmemory"
2) "2097152"
```
