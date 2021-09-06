// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package yaml

import (
	"encoding"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// jsonNumber is the interface of the encoding/json.Number datatype.
// Repeating the interface here avoids a dependency on encoding/json, and also
// supports other libraries like jsoniter, which use a similar datatype with
// the same interface. Detecting this interface is useful when dealing with
// structures containing json.Number, which is a string under the hood. The
// encoder should prefer the use of Int64(), Float64() and string(), in that
// order, when encoding this type.
type jsonNumber interface {
	Float64() (float64, error)
	Int64() (int64, error)
	String() string
}

type encoder struct {
	emitter yamlEmitterT
	event   yamlEventT
	out     []byte
	flow    bool
	// doneInit holds whether the initial stream_start_event has been
	// emitted.
	doneInit bool
}

func newEncoder() *encoder {
	e := &encoder{}
	yamlEmitterInitialize(&e.emitter)
	yamlEmitterSetOutputString(&e.emitter, &e.out)
	yamlEmitterSetUnicode(&e.emitter, true)
	return e
}

func newEncoderWithWriter(w io.Writer) *encoder {
	e := &encoder{}
	yamlEmitterInitialize(&e.emitter)
	yamlEmitterSetOutputWriter(&e.emitter, w)
	yamlEmitterSetUnicode(&e.emitter, true)
	return e
}

func (e *encoder) init() {
	if e.doneInit {
		return
	}
	yamlStreamStartEventInitialize(&e.event, yamlUtf8Encoding)
	e.emit()
	e.doneInit = true
}

func (e *encoder) finish() {
	e.emitter.openEnded = false
	yamlStreamEndEventInitialize(&e.event)
	e.emit()
}

func (e *encoder) destroy() {
	yamlEmitterDelete(&e.emitter)
}

func (e *encoder) emit() {
	// This will internally delete the e.event value.
	e.must(yamlEmitterEmit(&e.emitter, &e.event))
}

func (e *encoder) must(ok bool) {
	if !ok {
		msg := e.emitter.problem
		if msg == "" {
			msg = "unknown problem generating YAML content"
		}
		failf("%s", msg)
	}
}

func (e *encoder) marshalDoc(tag string, in reflect.Value) {
	e.init()
	yamlDocumentStartEventInitialize(&e.event, nil, nil, true)
	e.emit()
	e.marshal(tag, in)
	yamlDocumentEndEventInitialize(&e.event, true)
	e.emit()
}

func (e *encoder) marshal(tag string, in reflect.Value) {
	if !in.IsValid() || in.Kind() == reflect.Ptr && in.IsNil() {
		e.nilv()
		return
	}
	iface := in.Interface()
	switch m := iface.(type) {
	case jsonNumber:
		integer, err := m.Int64()
		if err == nil {
			// In this case the json.Number is a valid int64
			in = reflect.ValueOf(integer)
			break
		}
		float, err := m.Float64()
		if err == nil {
			// In this case the json.Number is a valid float64
			in = reflect.ValueOf(float)
			break
		}
		// fallback case - no number could be obtained
		in = reflect.ValueOf(m.String())
	case time.Time, *time.Time:
		// Although time.Time implements TextMarshaler,
		// we don't want to treat it as a string for YAML
		// purposes because YAML has special support for
		// timestamps.
	case Marshaler:
		v, err := m.MarshalYAML()
		if err != nil {
			fail(err)
		}
		if v == nil {
			e.nilv()
			return
		}
		in = reflect.ValueOf(v)
	case encoding.TextMarshaler:
		text, err := m.MarshalText()
		if err != nil {
			fail(err)
		}
		in = reflect.ValueOf(string(text))
	case nil:
		e.nilv()
		return
	}
	switch in.Kind() {
	case reflect.Interface:
		e.marshal(tag, in.Elem())
	case reflect.Map:
		e.mapv(tag, in)
	case reflect.Ptr:
		if in.Type() == ptrTimeType {
			e.timev(tag, in.Elem())
		} else {
			e.marshal(tag, in.Elem())
		}
	case reflect.Struct:
		if in.Type() == timeType {
			e.timev(tag, in)
		} else {
			e.structv(tag, in)
		}
	case reflect.Slice, reflect.Array:
		if in.Type().Elem() == mapItemType {
			e.itemsv(tag, in)
		} else {
			e.slicev(tag, in)
		}
	case reflect.String:
		e.stringv(tag, in)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if in.Type() == durationType {
			e.stringv(tag, reflect.ValueOf(iface.(time.Duration).String()))
		} else {
			e.intv(tag, in)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		e.uintv(tag, in)
	case reflect.Float32, reflect.Float64:
		e.floatv(tag, in)
	case reflect.Bool:
		e.boolv(tag, in)
	default:
		panic("cannot marshal type: " + in.Type().String())
	}
}

func (e *encoder) mapv(tag string, in reflect.Value) {
	e.mappingv(tag, func() {
		keys := keyList(in.MapKeys())
		sort.Sort(keys)
		for _, k := range keys {
			e.marshal("", k)
			e.marshal("", in.MapIndex(k))
		}
	})
}

func (e *encoder) itemsv(tag string, in reflect.Value) {
	e.mappingv(tag, func() {
		slice := in.Convert(reflect.TypeOf([]MapItem{})).Interface().([]MapItem)
		for _, item := range slice {
			e.marshal("", reflect.ValueOf(item.Key))
			e.marshal("", reflect.ValueOf(item.Value))
		}
	})
}

func (e *encoder) structv(tag string, in reflect.Value) {
	sinfo, err := getStructInfo(in.Type())
	if err != nil {
		panic(err)
	}
	e.mappingv(tag, func() {
		for _, info := range sinfo.FieldsList {
			var value reflect.Value
			if info.Inline == nil {
				value = in.Field(info.Num)
			} else {
				value = in.FieldByIndex(info.Inline)
			}
			if info.OmitEmpty && isZero(value) {
				continue
			}
			e.marshal("", reflect.ValueOf(info.Key))
			e.flow = info.Flow
			e.marshal("", value)
		}
		if sinfo.InlineMap >= 0 {
			m := in.Field(sinfo.InlineMap)
			if m.Len() > 0 {
				e.flow = false
				keys := keyList(m.MapKeys())
				sort.Sort(keys)
				for _, k := range keys {
					if _, found := sinfo.FieldsMap[k.String()]; found {
						panic(fmt.Sprintf("Can't have key %q in inlined map; conflicts with struct field", k.String()))
					}
					e.marshal("", k)
					e.flow = false
					e.marshal("", m.MapIndex(k))
				}
			}
		}
	})
}

func (e *encoder) mappingv(tag string, f func()) {
	implicit := tag == ""
	style := yamlBlockMappingStyle
	if e.flow {
		e.flow = false
		style = yamlFlowMappingStyle
	}
	yamlMappingStartEventInitialize(&e.event, nil, []byte(tag), implicit, style)
	e.emit()
	f()
	yamlMappingEndEventInitialize(&e.event)
	e.emit()
}

func (e *encoder) slicev(tag string, in reflect.Value) {
	implicit := tag == ""
	style := yamlBlockSequenceStyle
	if e.flow {
		e.flow = false
		style = yamlFlowSequenceStyle
	}
	e.must(yamlSequenceStartEventInitialize(&e.event, nil, []byte(tag), implicit, style))
	e.emit()
	n := in.Len()
	for i := 0; i < n; i++ {
		e.marshal("", in.Index(i))
	}
	e.must(yamlSequenceEndEventInitialize(&e.event))
	e.emit()
}

// isBase60 returns whether s is in base 60 notation as defined in YAML 1.1.
//
// The base 60 float notation in YAML 1.1 is a terrible idea and is unsupported
// in YAML 1.2 and by this package, but these should be marshalled quoted for
// the time being for compatibility with other parsers.
func isBase60Float(s string) (result bool) {
	// Fast path.
	if s == "" {
		return false
	}
	c := s[0]
	if !(c == '+' || c == '-' || c >= '0' && c <= '9') || strings.IndexByte(s, ':') < 0 {
		return false
	}
	// Do the full match.
	return base60float.MatchString(s)
}

// From http://yaml.org/type/float.html, except the regular expression there
// is bogus. In practice parsers do not enforce the "\.[0-9_]*" suffix.
var base60float = regexp.MustCompile(`^[-+]?[0-9][0-9_]*(?::[0-5]?[0-9])+(?:\.[0-9_]*)?$`)

func (e *encoder) stringv(tag string, in reflect.Value) {
	var style yamlScalarStyleT
	s := in.String()
	canUsePlain := true
	switch {
	case !utf8.ValidString(s):
		if tag == yamlBinaryTag {
			failf("explicitly tagged !!binary data must be base64-encoded")
		}
		if tag != "" {
			failf("cannot marshal invalid UTF-8 data as %s", shortTag(tag))
		}
		// It can't be encoded directly as YAML so use a binary tag
		// and encode it as base64.
		tag = yamlBinaryTag
		s = encodeBase64(s)
	case tag == "":
		// Check to see if it would resolve to a specific
		// tag when encoded unquoted. If it doesn't,
		// there's no need to quote it.
		rtag, _ := resolve("", s)
		canUsePlain = rtag == yamlStrTag && !isBase60Float(s)
	}
	// Note: it's possible for user code to emit invalid YAML
	// if they explicitly specify a tag and a string containing
	// text that's incompatible with that tag.
	switch {
	case strings.Contains(s, "\n"):
		style = yamlLiteralScalarStyle
	case canUsePlain:
		style = yamlPlainScalarStyle
	default:
		style = yamlDoubleQuotedScalarStyle
	}
	e.emitScalar(s, "", tag, style)
}

func (e *encoder) boolv(tag string, in reflect.Value) {
	var s string
	if in.Bool() {
		s = "true"
	} else {
		s = "false"
	}
	e.emitScalar(s, "", tag, yamlPlainScalarStyle)
}

func (e *encoder) intv(tag string, in reflect.Value) {
	s := strconv.FormatInt(in.Int(), 10)
	e.emitScalar(s, "", tag, yamlPlainScalarStyle)
}

func (e *encoder) uintv(tag string, in reflect.Value) {
	s := strconv.FormatUint(in.Uint(), 10)
	e.emitScalar(s, "", tag, yamlPlainScalarStyle)
}

func (e *encoder) timev(tag string, in reflect.Value) {
	t := in.Interface().(time.Time)
	s := t.Format(time.RFC3339Nano)
	e.emitScalar(s, "", tag, yamlPlainScalarStyle)
}

func (e *encoder) floatv(tag string, in reflect.Value) {
	// Issue #352: When formatting, use the precision of the underlying value
	precision := 64
	if in.Kind() == reflect.Float32 {
		precision = 32
	}

	s := strconv.FormatFloat(in.Float(), 'g', -1, precision)
	switch s {
	case "+Inf":
		s = ".inf"
	case "-Inf":
		s = "-.inf"
	case "NaN":
		s = ".nan"
	}
	e.emitScalar(s, "", tag, yamlPlainScalarStyle)
}

func (e *encoder) nilv() {
	e.emitScalar("null", "", "", yamlPlainScalarStyle)
}

func (e *encoder) emitScalar(value, anchor, tag string, style yamlScalarStyleT) {
	implicit := tag == ""
	e.must(yamlScalarEventInitialize(&e.event, []byte(anchor), []byte(tag), []byte(value), implicit, implicit, style))
	e.emit()
}
