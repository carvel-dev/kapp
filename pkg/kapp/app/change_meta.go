// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"fmt"
	"time"
)

type ChangeMeta struct {
	StartedAt  time.Time `json:"startedAt"`
	FinishedAt time.Time `json:"finishedAt,omitempty"`

	Successful  *bool  `json:"successful,omitempty"`
	Description string `json:"description,omitempty"`

	Namespaces []string `json:"namespaces,omitempty"`
}

func NewChangeMetaFromString(data string) ChangeMeta {
	var meta ChangeMeta

	err := json.Unmarshal([]byte(data), &meta)
	if err != nil {
		panic(fmt.Sprintf("Decoding app change meta: %s", err))
	}

	return meta
}

func NewChangeMetaFromData(data map[string]string) ChangeMeta {
	return NewChangeMetaFromString(data["spec"])
}

func (m ChangeMeta) AsString() string {
	bytes, err := json.Marshal(m)
	if err != nil {
		panic(fmt.Sprintf("Encoding app change meta: %s", err))
	}

	return string(bytes)
}

func (m ChangeMeta) AsData() map[string]string {
	return map[string]string{"spec": m.AsString()}
}
