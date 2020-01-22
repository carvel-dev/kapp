package app

import (
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

type AppMeta struct {
	LabelKey   string `json:"labelKey"`
	LabelValue string `json:"labelValue"`

	LastChangeName string     `json:"lastChangeName,omitempty"`
	LastChange     ChangeMeta `json:"lastChange,omitempty"`

	UsedGVs []schema.GroupVersion `json:"usedGVs,omitempty"`
}

func NewAppMetaFromData(data map[string]string) (AppMeta, error) {
	var meta AppMeta

	err := json.Unmarshal([]byte(data["spec"]), &meta)
	if err != nil {
		return AppMeta{}, fmt.Errorf("Parsing app metadata: %s", err)
	}

	return meta, nil
}

func (m AppMeta) AsString() string {
	bytes, err := json.Marshal(m)
	if err != nil {
		panic(fmt.Sprintf("Encoding app meta: %s", err))
	}

	return string(bytes)
}

func (m AppMeta) AsData() map[string]string {
	return map[string]string{"spec": m.AsString()}
}

func (m AppMeta) Labels() map[string]string {
	return map[string]string{m.LabelKey: m.LabelValue}
}
