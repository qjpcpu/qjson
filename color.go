package qjson

import (
	"bytes"
	ejson "encoding/json"
	"fmt"
	"reflect"

	"github.com/fatih/color"
)

// PrettyMarshal marshal json with color
func PrettyMarshal(v interface{}) []byte {
	tree, err := Decode(JSONMarshalWithPanic(v))
	if err != nil {
		panic(err)
	}
	tree.ColoredByLevel()
	return tree.ColorfulMarshal()
}

const (
	// Yellow color
	Yellow Color = iota + 1
	// Cyan color
	Cyan
	// Green color
	Green
	// Magenta color
	Magenta
	// Blue color
	Blue
	// Red color
	Red
	// White color
	White
	// Black color
	Black
)

var (
	// MaxColorLevel max render level
	MaxColorLevel = 3
	colorFuncs    = []func(a ...interface{}) string{
		func(e ...interface{}) string {
			var s string
			for _, v := range e {
				s += fmt.Sprint(v)
			}
			return s
		},
		color.New(color.FgYellow).SprintFunc(),
		color.New(color.FgCyan).SprintFunc(),
		color.New(color.FgGreen).SprintFunc(),
		color.New(color.FgMagenta).SprintFunc(),
		color.New(color.FgBlue).SprintFunc(),
		color.New(color.FgRed).SprintFunc(),
		color.New(color.FgWhite, color.BgBlack).SprintFunc(),
		color.New(color.FgBlack, color.BgWhite).SprintFunc(),
	}
)

// ColorfulMarshal print json with color
func (t *JSONTree) ColorfulMarshal() []byte {
	return t.Root.marshalWithColor()
}

// SetSelfColor set current node color
func (n *Node) SetSelfColor(c Color) {
	n.setColor(c, false)
}

// SetColor set color recursive
func (n *Node) SetColor(c Color) {
	n.setColor(c, true)
}

func (n *Node) setLeveledColor(idx int) {
	if idx > MaxColorLevel {
		return
	}
	c := Color(idx % len(colorFuncs))
	switch n.Type {
	case Null:
		n.color = c
	case String:
		n.color = c
	case Bool:
		n.color = c
	case Integer, Float:
		n.color = c
	case Object:
		for i := range n.ObjectValues {
			n.ObjectValues[i].Key.setLeveledColor(idx)
			n.ObjectValues[i].Value.setLeveledColor(idx + 1)
		}
	case Array:
		for i := range n.ArrayValues {
			n.ArrayValues[i].setLeveledColor(idx)
		}
	}
}

func (n *Node) setColor(c Color, recursive bool) {
	switch n.Type {
	case Null:
		n.color = c
	case String:
		n.color = c
	case Bool:
		n.color = c
	case Integer, Float:
		n.color = c
	case Object:
		if recursive {
			for i := range n.ObjectValues {
				n.ObjectValues[i].Key.SetColor(c)
				n.ObjectValues[i].Value.SetColor(c)
			}
		}
	case Array:
		if recursive {
			for i := range n.ArrayValues {
				n.ArrayValues[i].SetColor(c)
			}
		}
	}
}

func (n *Node) getColorFunc() func(...interface{}) string {
	idx := int(n.color) % len(colorFuncs)
	return colorFuncs[idx]
}

func (n *Node) marshalWithColor() []byte {
	fn := n.getColorFunc()
	buf := bytesPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bytesPool.Put(buf)
	switch n.Type {
	case Null:
		buf.WriteString(fn(nullVal))
	case String:
		buf.WriteByte(quote)
		buf.WriteString(fn(n.Value[1 : len(n.Value)-1]))
		buf.WriteByte(quote)
	case Bool:
		if n.Value == trueVal {
			buf.WriteString(fn(trueVal))
		} else {
			buf.WriteString(fn(falseVal))
		}
	case Integer, Float:
		buf.WriteString(fn(n.Value))
	case Object:
		buf.WriteByte(objectStart)
		for i, elem := range n.ObjectValues {
			v := elem.marshalWithColor()
			buf.Write(v)
			if i < len(n.ObjectValues)-1 {
				buf.WriteByte(commaChar)
			}
		}
		buf.WriteByte(objectEnd)
	case Array:
		buf.WriteByte(arrayStart)
		for i, elem := range n.ArrayValues {
			v := elem.marshalWithColor()
			buf.Write(v)
			if i < len(n.ArrayValues)-1 {
				buf.WriteByte(commaChar)
			}
		}
		buf.WriteByte(arrayEnd)
	}
	return copyBytes(buf.Bytes())
}

func (e ObjectElem) marshalWithColor() []byte {
	buf := bytesPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bytesPool.Put(buf)
	key := e.Key.marshalWithColor()
	val := e.Value.marshalWithColor()

	buf.Write(key)
	buf.WriteByte(colonChar)
	buf.Write(val)
	return copyBytes(buf.Bytes())
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

// ColoredByLevel set leveled color
func (t *JSONTree) ColoredByLevel() {
	t.Root.setLeveledColor(1)
}
