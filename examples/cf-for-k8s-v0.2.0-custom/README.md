Downloaded via

```bash
git clone https://github.com/cloudfoundry/cf-for-k8s
# was 86f5a6ed49ff6f132104c52c4a983b202c304211 at the time

cd cf-for-k8s
./hack/generate-values.sh -d cf.cppforlife.io > /tmp/cf-vals.yml
ytt -f config/ -f /tmp/cf-vals.yml > config.yml
```

^ secrets included in config.yml are throwaway
