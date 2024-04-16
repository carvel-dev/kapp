// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package preflight

import (
	"context"

	ctldgraph "carvel.dev/kapp/pkg/kapp/diffgraph"
)

type CheckFunc func(context.Context, *ctldgraph.ChangeGraph, CheckConfig) error
type CheckConfig map[string]any
type ConfigFunc func(CheckConfig) error

type Check interface {
	Enabled() bool
	SetEnabled(bool)
	SetConfig(CheckConfig) error
	Run(context.Context, *ctldgraph.ChangeGraph) error
}

type checkImpl struct {
	enabled   bool
	checkFunc CheckFunc

	config     CheckConfig
	configFunc ConfigFunc
}

func NewCheck(cf CheckFunc, sf ConfigFunc, enabled bool) Check {
	return &checkImpl{
		enabled:    enabled,
		checkFunc:  cf,
		configFunc: sf,
	}
}

func (cf *checkImpl) Enabled() bool {
	return cf.enabled
}

func (cf *checkImpl) SetEnabled(enabled bool) {
	cf.enabled = enabled
}

func (cf *checkImpl) SetConfig(config CheckConfig) error {
	cf.config = config
	if cf.configFunc != nil {
		return cf.configFunc(config)
	}
	return nil
}

func (cf *checkImpl) Run(ctx context.Context, changeGraph *ctldgraph.ChangeGraph) error {
	return cf.checkFunc(ctx, changeGraph, cf.config)
}
