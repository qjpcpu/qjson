package qjson

import (
	"errors"
)

// NodeType describe a json node type
type NodeType int

const (
	// Null json node means null
	Null NodeType = iota
	// String json node means "any text"
	String
	// Bool json node means true/false
	Bool
	// Integer json node means integer
	Integer
	// Float json node means float number
	Float
	// Object json node means k-v object
	Object
	// Array json node means array json nodes
	Array
)

// Node represent json node
type Node struct {
	Type         NodeType
	Value        string
	ObjectValues []ObjectElem
	ArrayValues  []Node
	Color        Color
}

// ObjectElem represent an object
type ObjectElem struct {
	Key   Node
	Value Node
}

/* node methods */

// IsNull tell node is null or not
func (n *Node) IsNull() bool {
	return n.Type == Null
}

// JSONTree represent full json
type JSONTree struct {
	Root Node
}

/* tree methods */

// IsNull tell node is null or not
func (t *JSONTree) IsNull() bool {
	return t.Root.IsNull()
}

// MarshalJSON json marshaller
func (t *JSONTree) MarshalJSON() ([]byte, error) {
	return t.Root.MarshalJSON()
}

// UnmarshalJSON json unmarshaller
func (t *JSONTree) UnmarshalJSON(data []byte) (err error) {
	var tree *JSONTree
	if tree, err = Decode(data); err != nil {
		return err
	}
	*t = *tree
	return
}

// ColorfulMarshal print json with color
func (t *JSONTree) ColorfulMarshal() []byte {
	return t.Root.unsafeMarshal()
}

/* tree generator */
func makeNewTree() *JSONTree {
	return &JSONTree{}
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
	case Integer, Float:
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

// MarshalJSON node is json marshaller too
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
	case Integer, Float:
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

// MarshalJSON object node is json marshaller
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
