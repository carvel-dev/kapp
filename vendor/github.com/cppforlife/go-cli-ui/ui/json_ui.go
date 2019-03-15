package ui

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	. "github.com/cppforlife/go-cli-ui/ui/table"
)

type JSONUI struct {
	parent UI
	uiResp JSONUIResp

	logTag string
	logger ExternalLogger
}

type JSONUIResp struct {
	Tables []JSONUITableResp
	Blocks []string
	Lines  []string
}

type JSONUITableResp struct {
	Content string
	Header  map[string]string
	Rows    []map[string]string
	Notes   []string
}

func NewJSONUI(parent UI, logger ExternalLogger) *JSONUI {
	return &JSONUI{parent: parent, logTag: "JSONUI", logger: logger}
}

func (ui *JSONUI) ErrorLinef(pattern string, args ...interface{}) {
	ui.addLine(pattern, args)
}

func (ui *JSONUI) PrintLinef(pattern string, args ...interface{}) {
	ui.addLine(pattern, args)
}

func (ui *JSONUI) BeginLinef(pattern string, args ...interface{}) {
	ui.addLine(pattern, args)
}

func (ui *JSONUI) EndLinef(pattern string, args ...interface{}) {
	ui.addLine(pattern, args)
}

func (ui *JSONUI) PrintBlock(block []byte) {
	ui.uiResp.Blocks = append(ui.uiResp.Blocks, string(block))
}

func (ui *JSONUI) PrintErrorBlock(block string) {
	ui.uiResp.Blocks = append(ui.uiResp.Blocks, block)
}

func (ui *JSONUI) PrintTable(table Table) {
	table.FillFirstColumn = true

	header := map[string]string{}

	if len(table.Header) > 0 {
		for i, val := range table.Header {
			if val.Hidden {
				continue
			}

			if val.Key == string(UNKNOWN_HEADER_MAPPING) {
				table.Header[i].Key = strconv.Itoa(i)
			}

			header[table.Header[i].Key] = val.Title
		}
	} else if len(table.AsRows()) > 0 {
		var rawHeaders []Header
		for i, _ := range table.AsRows()[0] {
			val := Header{
				Key:    fmt.Sprintf("col_%d", i),
				Hidden: false,
			}
			header[val.Key] = val.Title
			rawHeaders = append(rawHeaders, val)
		}
		table.Header = rawHeaders
	}

	resp := JSONUITableResp{
		Content: table.Content,
		Header:  header,
		Rows:    ui.stringRows(table.Header, table.AsRows()),
		Notes:   table.Notes,
	}

	ui.uiResp.Tables = append(ui.uiResp.Tables, resp)
}

func (ui *JSONUI) AskForText(_ string) (string, error) {
	panic("Cannot ask for input in JSON UI")
}

func (ui *JSONUI) AskForChoice(_ string, _ []string) (int, error) {
	panic("Cannot ask for a choice in JSON UI")
}

func (ui *JSONUI) AskForPassword(_ string) (string, error) {
	panic("Cannot ask for password in JSON UI")
}

func (ui *JSONUI) AskForConfirmation() error {
	panic("Cannot ask for confirmation in JSON UI")
}

func (ui *JSONUI) IsInteractive() bool {
	return ui.parent.IsInteractive()
}

func (ui *JSONUI) Flush() {
	defer ui.parent.Flush()

	if !reflect.DeepEqual(ui.uiResp, JSONUIResp{}) {
		bytes, err := json.MarshalIndent(ui.uiResp, "", "    ")
		if err != nil {
			ui.logger.Error(ui.logTag, "Failed to marshal UI response")
			return
		}

		ui.parent.PrintBlock(bytes)
	}
}

func (ui *JSONUI) stringRows(header []Header, rows [][]Value) []map[string]string {
	result := []map[string]string{}

	for _, row := range rows {
		data := map[string]string{}

		for i, col := range row {
			if header[i].Hidden {
				continue
			}

			data[header[i].Key] = col.String()
		}

		result = append(result, data)
	}

	return result
}

func (ui *JSONUI) addLine(pattern string, args []interface{}) {
	msg := fmt.Sprintf(pattern, args...)
	ui.uiResp.Lines = append(ui.uiResp.Lines, msg)
	ui.logger.Debug(ui.logTag, msg)
}
