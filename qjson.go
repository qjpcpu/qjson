package qjson

import (
	"bytes"
)

// NodeType describe a json node type
type NodeType byte

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
	color        Color
	Value        string
	ObjectValues []*ObjectElem
	ArrayValues  []*Node
}

// ObjectElem represent an object
type ObjectElem struct {
	Key   *Node
	Value *Node
}

/* package functions */

// Decode raw json bytes into JSONTree
func Decode(jsonBytes []byte) (*JSONTree, error) {
	tree := makeNewTree()
	offset := searchFirstValidChar(jsonBytes, 0)
	if offset == -1 {
		return tree, nil
	}
	if _, err := decodeAny(jsonBytes, offset, tree); err != nil {
		return nil, err
	}
	return tree, nil
}

/* node methods */

// IsNull tell node is null or not
func (n *Node) IsNull() bool {
	return n.Type == Null
}

// AsTree create sub json tree
func (n *Node) AsTree() *JSONTree {
	tree := makeNewTree()
	tree.Root = *n
	return tree
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
	return t.Root.marshalWithColor()
}

/* tree generator */
func makeNewTree() *JSONTree {
	return &JSONTree{}
}

/* marshalers */
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
	return stringToBytes(buf.String())
}

// MarshalJSON node is json marshaller too
func (n *Node) MarshalJSON() ([]byte, error) {
	buf := bytesPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bytesPool.Put(buf)
	var err error
	switch n.Type {
	case Null:
		buf.WriteString(nullVal)
	case String:
		buf.WriteString(n.Value)
	case Bool:
		if n.Value == trueVal {
			buf.WriteString(trueVal)
		} else {
			buf.WriteString(falseVal)
		}
	case Integer, Float:
		buf.WriteString(n.Value)
	case Object:
		buf.WriteByte(objectStart)
		for i, elem := range n.ObjectValues {
			v, err := elem.MarshalJSON()
			if err != nil {
				return nil, err
			}
			buf.Write(v)
			if i < len(n.ObjectValues)-1 {
				buf.WriteByte(commaChar)
			}
		}
		buf.WriteByte(objectEnd)
	case Array:
		buf.WriteByte(arrayStart)
		for i, elem := range n.ArrayValues {
			v, err := elem.MarshalJSON()
			if err != nil {
				return nil, err
			}
			buf.Write(v)
			if i < len(n.ArrayValues)-1 {
				buf.WriteByte(commaChar)
			}
		}
		buf.WriteByte(arrayEnd)
	}
	return stringToBytes(buf.String()), err
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
	return stringToBytes(buf.String())
}

// MarshalJSON object node is json marshaller
func (e ObjectElem) MarshalJSON() ([]byte, error) {
	buf := bytesPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bytesPool.Put(buf)
	key, err := e.Key.MarshalJSON()
	if err != nil {
		return nil, err
	}
	val, err := e.Value.MarshalJSON()
	if err != nil {
		return nil, err
	}
	buf.Write(key)
	buf.WriteByte(colonChar)
	buf.Write(val)
	return stringToBytes(buf.String()), nil
}
