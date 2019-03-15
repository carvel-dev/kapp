package test

import (
	"encoding/json"
	"testing"

	"github.com/cppforlife/go-cli-ui/ui"
)

func JSONUIFromBytes(t *testing.T, bytes []byte) ui.JSONUIResp {
	resp := ui.JSONUIResp{}

	err := json.Unmarshal(bytes, &resp)
	if err != nil {
		t.Fatalf("Expected to successfully unmarshal JSON UI: %s", err)
	}

	return resp
}
