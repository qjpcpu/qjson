package qjson

import (
	"errors"
	"strings"
)

type NodeType int

const (
	Null NodeType = iota
	String
	Bool
	Number
	Object
	Array
)

type Node struct {
	Type         NodeType
	Value        string
	ObjectValues []ObjectElem
	ArrayValues  []Node
	Color        Color
}

type ObjectElem struct {
	Key   Node
	Value Node
}

/* node methods */
func (n *Node) IsNull() bool {
	return n.Type == Null
}

func (n *Node) IsInteger() bool {
	return n.Type == Number && !strings.Contains(n.Value, string([]byte{dotChar}))
}

func (n *Node) IsFloat() bool {
	return n.Type == Number && strings.Contains(n.Value, string([]byte{dotChar}))
}

type JSONTree struct {
	Root Node
}

/* tree methods */
func (t *JSONTree) IsNull() bool {
	return t.Root.IsNull()
}

func (t *JSONTree) MarshalJSON() ([]byte, error) {
	return t.Root.MarshalJSON()
}

func (t *JSONTree) UnmarshalJSON(data []byte) error {
	if tree, err := Decode(data); err != nil {
		return err
	} else {
		*t = *tree
	}
	return nil
}

func (t *JSONTree) ColorfulMarshal() []byte {
	return t.Root.unsafeMarshal()
}

/* tree generator */
func makeNewTree() *JSONTree {
	return &JSONTree{}
}

/* node generator */
func makeNullNode() Node {
	return Node{Type: Null}
}

func makeStringNode(val string) Node {
	return Node{Type: String, Value: val}
}

func makeBoolNode(trueFalse string) Node {
	return Node{Type: Bool, Value: trueFalse}
}

func makeNumberNode(v string) Node {
	return Node{Type: Number, Value: v}
}

/* errors */

var (
	errProccessFinished = errors.New("proccess done")
)

/* marshalers */
func (n *Node) unsafeMarshal() []byte {
	fn := n.getColorFunc()
	var buf []byte
	switch n.Type {
	case Null:
		buf = []byte(fn(nullVal))
	case String:
		buf = append(buf, quote)
		buf = append(buf, []byte(fn(n.Value[1:len(n.Value)-1]))...)
		buf = append(buf, quote)
	case Bool:
		if n.Value == trueVal {
			buf = []byte(fn(trueVal))
		} else {
			buf = []byte(fn(falseVal))
		}
	case Number:
		buf = []byte(fn(n.Value))
	case Object:
		buf = append(buf, objectStart)
		for _, elem := range n.ObjectValues {
			v := elem.unsafeMarshal()
			buf = append(buf, v...)
			buf = append(buf, commaChar)
		}
		if size := len(n.ObjectValues); size > 0 {
			buf = buf[:len(buf)-1]
		}
		buf = append(buf, objectEnd)
	case Array:
		buf = append(buf, arrayStart)
		for _, elem := range n.ArrayValues {
			v := elem.unsafeMarshal()
			buf = append(buf, v...)
			buf = append(buf, commaChar)
		}
		if size := len(n.ArrayValues); size > 0 {
			buf = buf[:len(buf)-1]
		}
		buf = append(buf, arrayEnd)
	}
	return buf
}

func (n *Node) MarshalJSON() ([]byte, error) {
	var buf []byte
	var err error
	switch n.Type {
	case Null:
		buf = []byte(nullVal)
	case String:
		buf = []byte(n.Value)
	case Bool:
		if n.Value == trueVal {
			buf = []byte(trueVal)
		} else {
			buf = []byte(falseVal)
		}
	case Number:
		buf = []byte(n.Value)
	case Object:
		buf = append(buf, objectStart)
		for _, elem := range n.ObjectValues {
			v, err := elem.MarshalJSON()
			if err != nil {
				return nil, err
			}
			buf = append(buf, v...)
			buf = append(buf, commaChar)
		}
		if size := len(n.ObjectValues); size > 0 {
			buf = buf[:len(buf)-1]
		}
		buf = append(buf, objectEnd)
	case Array:
		buf = append(buf, arrayStart)
		for _, elem := range n.ArrayValues {
			v, err := elem.MarshalJSON()
			if err != nil {
				return nil, err
			}
			buf = append(buf, v...)
			buf = append(buf, commaChar)
		}
		if size := len(n.ArrayValues); size > 0 {
			buf = buf[:len(buf)-1]
		}
		buf = append(buf, arrayEnd)
	}
	return buf, err
}

func (e ObjectElem) unsafeMarshal() []byte {
	var buf []byte
	key := e.Key.unsafeMarshal()
	val := e.Value.unsafeMarshal()

	buf = append(buf, key...)
	buf = append(buf, colonChar)
	buf = append(buf, val...)
	return buf
}

func (e ObjectElem) MarshalJSON() ([]byte, error) {
	var buf []byte
	key, err := e.Key.MarshalJSON()
	if err != nil {
		return nil, err
	}
	val, err := e.Value.MarshalJSON()
	if err != nil {
		return nil, err
	}
	buf = append(buf, key...)
	buf = append(buf, colonChar)
	buf = append(buf, val...)
	return buf, nil
}
