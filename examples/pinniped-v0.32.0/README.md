Downloaded via

```bash
ytt \
  -f https://get.pinniped.dev/v0.32.0/install-pinniped-concierge.yaml \
  -f https://get.pinniped.dev/v0.32.0/install-local-user-authenticator.yaml \
  -f https://get.pinniped.dev/v0.32.0/install-pinniped-supervisor.yaml > config.yml
```
