package ui

import (
	"io/ioutil"
)

func NewNoopUI() *WriterUI {
	return NewWriterUI(ioutil.Discard, ioutil.Discard, NewNoopLogger())
}
