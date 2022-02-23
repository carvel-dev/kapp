package app

import (
	"fmt"
	"strings"

	"github.com/cppforlife/go-cli-ui/ui"
)

type InMemoryUI struct {
	ui.UI
	diffBuffer    strings.Builder
	summaryBuffer strings.Builder
}

func (ui *InMemoryUI) BeginLinef(pattern string, args ...interface{}) {
	ui.diffBuffer.WriteString(fmt.Sprintf(pattern, args...))
}

func (ui *InMemoryUI) PrintBlock(bytes []byte) {
	ui.diffBuffer.Write(bytes)
}
