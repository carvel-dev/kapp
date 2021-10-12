module github.com/k14s/kapp

go 1.16

require (
	github.com/cppforlife/cobrautil v0.0.0-20200514214827-bb86e6965d72
	github.com/cppforlife/color v1.9.1-0.20200716202919-6706ac40b835
	github.com/cppforlife/go-cli-ui v0.0.0-20200716203538-1e47f820817f
	github.com/cppforlife/go-patch v0.2.0
	github.com/ghodss/yaml v1.0.0
	github.com/hashicorp/go-version v1.3.0
	github.com/k14s/difflib v0.0.0-20201117154628-0c031775bf57
	github.com/k14s/ytt v0.36.0
	github.com/mitchellh/go-wordwrap v1.0.1
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	github.com/vmware-tanzu/carvel-kapp-controller v0.27.0
	golang.org/x/crypto v0.0.0-20210616213533-5ff15b29337e // indirect
	golang.org/x/net v0.0.0-20210520170846-37e1c6afe023
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.22.1 // kubernetes-1.22.1
	k8s.io/apimachinery v0.22.1 // kubernetes-1.22.1
	k8s.io/client-go v0.22.1 // kubernetes-1.22.1
)

replace github.com/spf13/cobra => github.com/spf13/cobra v1.1.1
