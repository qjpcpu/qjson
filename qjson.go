package qjson

import (
	"bytes"
	"encoding/json"
	"strconv"
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

// Color type
type Color byte

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

// ConvertToJSONTree any object to json tree
func ConvertToJSONTree(obj interface{}) (*JSONTree, error) {
	data, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	return Decode(data)
}

/* node methods */

// IsNull tell node is null or not
func (n *Node) IsNull() bool {
	return n.Type == Null
}

// IsString tell node is string or not
func (n *Node) IsString() bool {
	return n.Type == String
}

// IsBool tell node is boolean or not
func (n *Node) IsBool() bool {
	return n.Type == Bool
}

// IsInteger tell node is num or not
func (n *Node) IsInteger() bool {
	return n.Type == Integer
}

// IsFloat tell node is float or not
func (n *Node) IsFloat() bool {
	return n.Type == Float
}

// IsNumber tell node is number or not
func (n *Node) IsNumber() bool {
	return n.IsFloat() || n.IsInteger()
}

// AsTree create sub json tree
func (n *Node) AsTree() *JSONTree {
	tree := makeNewTree()
	tree.Root = n
	return tree
}

// FindNodeByKey find object value by key
func (n *Node) FindNodeByKey(key string) *Node {
	if elem := n.FindObjectElemByKey(key); elem != nil {
		return elem.Value
	}
	return nil
}

// FindObjectElemByKey find object value by key
func (n *Node) FindObjectElemByKey(key string) *ObjectElem {
	if n.Type != Null && n.Type != Object {
		panic("node type should be object")
	}
	for i, kv := range n.ObjectValues {
		if kv.Key.AsString() == key {
			return n.ObjectValues[i]
		}
	}
	return nil
}

// SetObjectStringElem set kv pair
func (n *Node) SetObjectStringElem(key, value string) {
	for _, elem := range n.ObjectValues {
		if elem.Key.AsString() == key {
			elem.Value.Type = String
			elem.Value.Value = bytesToString(MarshalString([]byte(value)))
			return
		}
	}
	elem := CreateObjectElem()
	elem.Key = CreateStringNode()
	elem.Key.Value = bytesToString(MarshalString([]byte(key)))
	elem.Value = CreateStringNode()
	elem.Value.Value = bytesToString(MarshalString([]byte(value)))
	n.ObjectValues = append(n.ObjectValues, elem)
}

// SetObjectIntElem set kv pair
func (n *Node) SetObjectIntElem(key string, value int64) {
	for _, elem := range n.ObjectValues {
		if elem.Key.AsString() == key {
			elem.Value.Type = Integer
			elem.Value.Value = strconv.FormatInt(value, 10)
			return
		}
	}
	elem := CreateObjectElem()
	elem.Key = CreateStringNode()
	elem.Key.Value = bytesToString(MarshalString([]byte(key)))
	elem.Value = CreateIntegerNode()
	elem.Value.Value = strconv.FormatInt(value, 10)
	n.ObjectValues = append(n.ObjectValues, elem)
}

// SetObjectUintElem set kv pair
func (n *Node) SetObjectUintElem(key string, value uint64) {
	for _, elem := range n.ObjectValues {
		if elem.Key.AsString() == key {
			elem.Value.Type = Integer
			elem.Value.Value = strconv.FormatUint(value, 10)
			return
		}
	}
	elem := CreateObjectElem()
	elem.Key = CreateStringNode()
	elem.Key.Value = bytesToString(MarshalString([]byte(key)))
	elem.Value = CreateIntegerNode()
	elem.Value.Value = strconv.FormatUint(value, 10)
	n.ObjectValues = append(n.ObjectValues, elem)
}

// SetObjectBoolElem set kv pair
func (n *Node) SetObjectBoolElem(key string, value bool) {
	val := falseVal
	if value {
		val = trueVal
	}
	for _, elem := range n.ObjectValues {
		if elem.Key.AsString() == key {
			elem.Value.Type = Bool
			elem.Value.Value = val
			return
		}
	}
	elem := CreateObjectElem()
	elem.Key = CreateStringNode()
	elem.Key.Value = bytesToString(MarshalString([]byte(key)))
	elem.Value = CreateBoolNode()
	elem.Value.Value = val
	n.ObjectValues = append(n.ObjectValues, elem)
}

// SetObjectNodeElem set kv pair
func (n *Node) SetObjectNodeElem(key string, value *Node) {
	for _, elem := range n.ObjectValues {
		if elem.Key.AsString() == key {
			elem.Value = value
			return
		}
	}
	elem := CreateObjectElem()
	elem.Key = CreateStringNode()
	elem.Key.Value = bytesToString(MarshalString([]byte(key)))
	elem.Value = value
	n.ObjectValues = append(n.ObjectValues, elem)
}

// AsMap create map for chilren
func (n *Node) AsMap() map[string]*Node {
	if n.Type != Null && n.Type != Object {
		panic("node type should be object")
	}
	m := make(map[string]*Node)
	for i, kv := range n.ObjectValues {
		m[kv.Key.AsString()] = n.ObjectValues[i].Value
	}
	return m
}

// SetRawValue to node
func (n *Node) SetRawValue(str string) *Node {
	n.Value = str
	return n
}

// SetString to string node
func (n *Node) SetString(str string) *Node {
	n.Value = bytesToString(MarshalString([]byte(str)))
	return n
}

// SetBool to node
func (n *Node) SetBool(b bool) *Node {
	if b {
		n.Value = trueVal
	} else {
		n.Value = falseVal
	}
	return n
}

// SetInt to node
func (n *Node) SetInt(num int64) *Node {
	n.Value = strconv.FormatInt(num, 10)
	return n
}

// SetUnt to node
func (n *Node) SetUint(num uint64) *Node {
	n.Value = strconv.FormatUint(num, 10)
	return n
}

// AsString as string
func (n *Node) AsString() string {
	switch n.Type {
	case String:
		s, err := UnmarshalString([]byte(n.Value))
		if err != nil {
			panic(err)
		}
		return string(s)
	case Bool:
		if n.Value == trueVal {
			return trueVal
		}
		return falseVal
	case Integer, Float:
		return n.Value
	}
	panic("node type should be simple value")
}

// AsBool as boolean
func (n *Node) AsBool() bool {
	if n.Type != Bool {
		panic("node type should be bool value")
	}
	return n.Value == trueVal
}

// AsInt as integer
func (n *Node) AsInt() int64 {
	if n.Type != Integer {
		panic("node type should be integer value")
	}
	i, err := strconv.ParseInt(n.Value, 10, 64)
	if err != nil {
		panic(err)
	}
	return i
}

// AsUint as unsigned integer
func (n *Node) AsUint() uint64 {
	if n.Type != Integer {
		panic("node type should be unsigned integer value")
	}
	i, err := strconv.ParseUint(n.Value, 10, 64)
	if err != nil {
		panic(err)
	}
	return i
}

// AsFloat as float64
func (n *Node) AsFloat() float64 {
	if n.Type != Integer {
		panic("node type should be float value")
	}
	i, err := strconv.ParseFloat(n.Value, 64)
	if err != nil {
		panic(err)
	}
	return i
}

// JSONTree represent full json
type JSONTree struct {
	Root *Node
}

/* tree methods */

// IsNull tell node is null or not
func (t *JSONTree) IsNull() bool {
	return t.Root == nil || t.Root.IsNull()
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

/* tree generator */
func makeNewTree() *JSONTree {
	return &JSONTree{Root: CreateNode()}
}

/* marshalers */

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
