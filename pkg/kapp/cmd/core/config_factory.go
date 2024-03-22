// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"fmt"
	"net"
	"os"
	"strings"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type ConfigFactory interface {
	ConfigurePathResolver(func() (string, error))
	ConfigureContextResolver(func() (string, error))
	ConfigureYAMLResolver(func() (string, error))
	ConfigureClient(float32, int)
	RESTConfig() (*rest.Config, error)
	DefaultNamespace() (string, error)
}

type ConfigFactoryImpl struct {
	pathResolverFunc    func() (string, error)
	contextResolverFunc func() (string, error)
	yamlResolverFunc    func() (string, error)

	qps   float32
	burst int
}

var _ ConfigFactory = &ConfigFactoryImpl{}

func NewConfigFactoryImpl() *ConfigFactoryImpl {
	return &ConfigFactoryImpl{}
}

func (f *ConfigFactoryImpl) ConfigurePathResolver(resolverFunc func() (string, error)) {
	f.pathResolverFunc = resolverFunc
}

func (f *ConfigFactoryImpl) ConfigureContextResolver(resolverFunc func() (string, error)) {
	f.contextResolverFunc = resolverFunc
}

func (f *ConfigFactoryImpl) ConfigureYAMLResolver(resolverFunc func() (string, error)) {
	f.yamlResolverFunc = resolverFunc
}

func (f *ConfigFactoryImpl) ConfigureClient(qps float32, burst int) {
	f.qps = qps
	f.burst = burst
}

func (f *ConfigFactoryImpl) RESTConfig() (*rest.Config, error) {
	isExplicitYAMLConfig, config, err := f.clientConfig()
	if err != nil {
		return nil, err
	}

	restConfig, err := config.ClientConfig()
	if err != nil {
		prefixMsg := ""
		if isExplicitYAMLConfig {
			prefixMsg = " (explicit config given)"
		}

		hintMsg := ""
		if strings.Contains(err.Error(), "no configuration has been provided") {
			hintMsg = "Ensure cluster name is specified correctly in context configuration"
		}
		if len(hintMsg) > 0 {
			hintMsg = " (hint: " + hintMsg + ")"
		}

		return nil, fmt.Errorf("Building Kubernetes config%s: %w%s", prefixMsg, err, hintMsg)
	}

	if f.qps > 0.0 {
		restConfig.QPS = f.qps
		restConfig.Burst = f.burst
	}

	return restConfig, nil
}

func (f *ConfigFactoryImpl) DefaultNamespace() (string, error) {
	_, config, err := f.clientConfig()
	if err != nil {
		return "", err
	}

	name, _, err := config.Namespace()
	return name, err
}

func (f *ConfigFactoryImpl) clientConfig() (bool, clientcmd.ClientConfig, error) {
	path, err := f.pathResolverFunc()
	if err != nil {
		return false, nil, fmt.Errorf("Resolving config path: %w", err)
	}

	context, err := f.contextResolverFunc()
	if err != nil {
		return false, nil, fmt.Errorf("Resolving config context: %w", err)
	}

	configYAML, err := f.yamlResolverFunc()
	if err != nil {
		return false, nil, fmt.Errorf("Resolving config YAML: %w", err)
	}

	if len(configYAML) > 0 {
		kubernetesHost := os.Getenv("KUBERNETES_SERVICE_HOST")
		kubernetesServicePort := os.Getenv("KUBERNETES_SERVICE_PORT")
		envHostPort := net.JoinHostPort(kubernetesHost, kubernetesServicePort)
		if kubernetesServicePort == "" {
			// client-go will manually add the port based on http/https
			envHostPort = kubernetesHost
		}
		configYAML = strings.ReplaceAll(configYAML, "${KAPP_KUBERNETES_SERVICE_HOST_PORT}", envHostPort)
		config, err := clientcmd.NewClientConfigFromBytes([]byte(configYAML))
		return true, config, err
	}

	// Based on https://github.com/kubernetes/kubernetes/blob/30c7df5cd822067016640aa267714204ac089172/staging/src/k8s.io/cli-runtime/pkg/genericclioptions/config_flags.go#L124
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	overrides := &clientcmd.ConfigOverrides{}

	if len(path) > 0 {
		loadingRules.ExplicitPath = path
	}
	if len(context) > 0 {
		overrides.CurrentContext = context
	}

	return false, clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, overrides), nil
}
