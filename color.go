package qjson

import (
	"fmt"
	"strconv"
	"strings"
)

var (
	// ANSI color codes
	ansiReset  = "\x1b[0m"
	colorFuncs = []func(a ...interface{}) string{
		// identity (no color)
		func(e ...interface{}) string {
			var s string
			for _, v := range e {
				s += fmt.Sprint(v)
			}
			return s
		},
		makeAnsiSprint(1, 33),     // bold yellow
		makeAnsiSprint(1, 36),     // bold cyan
		makeAnsiSprint(1, 32),     // bold green
		makeAnsiSprint(1, 35),     // bold magenta
		makeAnsiSprint(1, 34),     // bold blue
		makeAnsiSprint(1, 31),     // bold red
		makeAnsiSprint(1, 37, 40), // bold white on black
		makeAnsiSprint(1, 30, 47), // bold black on white
	}
)

func makeAnsiSprint(codes ...int) func(a ...interface{}) string {
	// build CSI prefix with given codes
	var b strings.Builder
	b.WriteString("\x1b[")
	for i, c := range codes {
		if i > 0 {
			b.WriteByte(';')
		}
		b.WriteString(strconv.Itoa(c))
	}
	b.WriteByte('m')
	prefix := b.String()
	return func(a ...interface{}) string {
		var s string
		for _, v := range a {
			s += fmt.Sprint(v)
		}
		return prefix + s + ansiReset
	}
}

// Formatter json with indent
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
