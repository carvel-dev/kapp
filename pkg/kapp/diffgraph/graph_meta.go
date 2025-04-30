// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package diffgraph

type RenderedNodeData struct {
	Name         string            `json:"name"`
	Namespace    *string           `json:"namespace"`
	ChangeGroups []string          `json:"changeGroups"`
	GroupKind    RenderedGroupKind `json:"groupKind"`
	Op           ActualChangeOp    `json:"op"`
}

type RenderedNode struct {
	ID   string           `json:"id"`
	Data RenderedNodeData `json:"data"`
}

type RenderedGroupKind struct {
	Group string `json:"group"`
	Kind  string `json:"kind"`
}

type RenderedEdge struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

type RenderedGraph struct {
	Nodes []RenderedNode `json:"nodes"`
	Edges []RenderedEdge `json:"edges"`
}
