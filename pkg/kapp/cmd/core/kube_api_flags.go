// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"github.com/spf13/cobra"
)

type KubeAPIFlags struct {
	QPS   float32
	Burst int
}

func (f *KubeAPIFlags) Set(cmd *cobra.Command, _ FlagsFactory) {
	// Similar names are used by kubelet and other controllers
	cmd.PersistentFlags().Float32Var(&f.QPS, "kube-api-qps", 1000, "Set Kubernetes API client QPS limit")
	cmd.PersistentFlags().IntVar(&f.Burst, "kube-api-burst", 1000, "Set Kubernetes API client burst limit")
}

func (f *KubeAPIFlags) Configure(config ConfigFactory) {
	config.ConfigureClient(f.QPS, f.Burst)
}
