package qjson

import (
	"bytes"
	ejson "encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/fatih/color"
)

// PrettyMarshal marshal json with color
func PrettyMarshal(v interface{}) []byte {
	tree, err := Decode(JSONMarshalWithPanic(v))
	if err != nil {
		panic(err)
	}
	return tree.ColorfulMarshal()
}

// PrettyMarshalWithIndent marshal json with indent
func PrettyMarshalWithIndent(v interface{}) []byte {
	tree, err := Decode(JSONMarshalWithPanic(v))
	if err != nil {
		panic(err)
	}
	return tree.ColorfulMarshalWithIndent()
}

var (
	colorFuncs = []func(a ...interface{}) string{
		func(e ...interface{}) string {
			var s string
			for _, v := range e {
				s += fmt.Sprint(v)
			}
			return s
		},
		color.New(color.FgYellow, color.Bold).SprintFunc(),
		color.New(color.FgCyan, color.Bold).SprintFunc(),
		color.New(color.FgGreen, color.Bold).SprintFunc(),
		color.New(color.FgMagenta, color.Bold).SprintFunc(),
		color.New(color.FgBlue, color.Bold).SprintFunc(),
		color.New(color.FgRed, color.Bold).SprintFunc(),
		color.New(color.FgWhite, color.BgBlack, color.Bold).SprintFunc(),
		color.New(color.FgBlack, color.BgWhite, color.Bold).SprintFunc(),
	}
)

// ColorfulMarshal print json with color
func (t *JSONTree) ColorfulMarshal() []byte {
	return new(Formatter).Format(t)
}

// ColorfulMarshalWithIndent print json with indent
func (t *JSONTree) ColorfulMarshalWithIndent() []byte {
	return NewFormatter().Format(t)
}

// JSONMarshalWithPanic json marshal with panic
func JSONMarshalWithPanic(t interface{}) []byte {
	if t == nil {
		return nil
	}
	if v := reflect.ValueOf(t); v.Kind() == reflect.Ptr && v.IsNil() {
		return nil
	}
	buffer := &bytes.Buffer{}
	encoder := ejson.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(t); err != nil {
		panic(err)
	}
	ret := buffer.Bytes()
	// golang's encoder would always append a '\n', so we should drop it
	if len(ret) > 0 {
		ret = ret[:len(ret)-1]
	}
	return ret
}

type Formatter struct {
	Indent int
}

// NewFormatter returns a new formatter with following default values.
func NewFormatter() *Formatter {
	return &Formatter{
		Indent: 2,
	}
}

// Format JSONTree
func (f *Formatter) Format(v *JSONTree) []byte {
	if v == nil || v.Root == nil {
		return nil
	}
	s := f.pretty(v.Root, 1)
	return []byte(s)
}

func (f *Formatter) pretty(node *Node, depth int) string {
	if node == nil {
		return ""
	}
	fn := f.getColorFuncByDepth(depth)
	switch node.Type {
	case String, Bool, Float, Integer:
		return fn(node.Value)
	case Null:
		return fn(nullVal)
	case Object:
		return f.processMap(node, depth)
	case Array:
		return f.processArray(node, depth)
	}

	return ""
}

func (f *Formatter) processMap(m *Node, depth int) string {
	if m == nil {
		return ""
	}
	currentIndent := f.generateIndent(depth - 1)
	nextIndent := f.generateIndent(depth)
	rows := []string{}

	if len(m.ObjectValues) == 0 {
		return "{}"
	}

	fn := f.getColorFuncByDepth(depth)
	for _, elem := range m.ObjectValues {
		k := fn(elem.Key.Value)
		v := f.pretty(elem.Value, depth+1)
		var row string
		if f.isNoIndent() {
			row = fmt.Sprintf("%s:%s", k, v)
		} else {
			row = fmt.Sprintf("%s%s: %s", nextIndent, k, v)
		}

		rows = append(rows, row)
	}
	if f.isNoIndent() {
		return fmt.Sprintf("{%s}", strings.Join(rows, ","))
	}
	return fmt.Sprintf("{\n%s\n%s}", strings.Join(rows, ",\n"), currentIndent)
}

func (f *Formatter) processArray(a *Node, depth int) string {
	if a == nil {
		return ""
	}
	currentIndent := f.generateIndent(depth - 1)
	nextIndent := f.generateIndent(depth)
	rows := []string{}

	if len(a.ArrayValues) == 0 {
		return "[]"
	}

	for _, val := range a.ArrayValues {
		c := f.pretty(val, depth+1)
		var row string
		if f.isNoIndent() {
			row = c
		} else {
			row = nextIndent + c
		}
		rows = append(rows, row)
	}
	if f.isNoIndent() {
		return fmt.Sprintf("[%s]", strings.Join(rows, ","))
	}
	return fmt.Sprintf("[\n%s\n%s]", strings.Join(rows, ",\n"), currentIndent)
}

func (f *Formatter) generateIndent(depth int) string {
	return strings.Join(make([]string, f.Indent*depth+1), " ")
}

func (f *Formatter) isNoIndent() bool {
	return f.Indent == 0
}

func (f *Formatter) getColorFuncByDepth(depth int) func(...interface{}) string {
	idx := depth % len(colorFuncs)
	return colorFuncs[idx]
}
