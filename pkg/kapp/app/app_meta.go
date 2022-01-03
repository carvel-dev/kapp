// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

type Meta struct {
	LabelKey   string `json:"labelKey"`
	LabelValue string `json:"labelValue"`

	LastChangeName string     `json:"lastChangeName,omitempty"`
	LastChange     ChangeMeta `json:"lastChange,omitempty"`

	UsedGVs []schema.GroupVersion `json:"usedGVs,omitempty"`
}

func NewAppMetaFromData(data map[string]string) (Meta, error) {
	var meta Meta

	err := json.Unmarshal([]byte(data["spec"]), &meta)
	if err != nil {
		return Meta{}, fmt.Errorf("Parsing app metadata: %s", err)
	}

	return meta, nil
}

func (m Meta) AsString() string {
	bytes, err := json.Marshal(m)
	if err != nil {
		panic(fmt.Sprintf("Encoding app meta: %s", err))
	}

	return string(bytes)
}

func (m Meta) AsData() map[string]string {
	return map[string]string{"spec": m.AsString()}
}

func (m Meta) Labels() map[string]string {
	return map[string]string{m.LabelKey: m.LabelValue}
}
