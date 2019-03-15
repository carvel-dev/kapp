package app

import (
	"encoding/json"
	"fmt"
)

type AppMeta struct {
	LabelKey   string `json:"labelKey"`
	LabelValue string `json:"labelValue"`

	LastChangeName string     `json:"lastChangeName,omitempty"`
	LastChange     ChangeMeta `json:"lastChange,omitempty"`
}

func NewAppMetaFromString(data string) AppMeta {
	var meta AppMeta

	err := json.Unmarshal([]byte(data), &meta)
	if err != nil {
		panic(fmt.Sprintf("Decoding app meta: %s", err))
	}

	return meta
}

func NewAppMetaFromData(data map[string]string) AppMeta {
	return NewAppMetaFromString(data["spec"])
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
