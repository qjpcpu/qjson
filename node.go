package qjson

import (
	"bytes"
	"strconv"
	"strings"
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
	Value        string
	ObjectValues []*ObjectElem
	ArrayValues  []*Node
}

// ObjectElem represent an object
type ObjectElem struct {
	Key   *Node
	Value *Node
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

// FindObjectElemByKeyRecursive find object value by key
func (n *Node) FindObjectElemByKeyRecursive(keyWithDot string) *ObjectElem {
	keys := strings.Split(keyWithDot, ".")
	current, cnt := n, 0
	var elem *ObjectElem
	for _, key := range keys {
		cnt++
		if elem = current.FindObjectElemByKey(key); elem == nil {
			break
		}
		if current = elem.Value; current == nil {
			break
		}
	}
	if cnt == len(keys) {
		return elem
	}
	return nil
}

// RemoveObjectElemByKey remove object element
func (n *Node) RemoveObjectElemByKey(key string) bool {
	if n.Type != Null && n.Type != Object {
		panic("node type should be object")
	}
	size := len(n.ObjectValues)
	var delCnt int
	for i := 0; i < size; i++ {
		if n.ObjectValues[i].Key.AsString() == key {
			delCnt++
		} else if delCnt > 0 {
			n.ObjectValues[i-delCnt] = n.ObjectValues[i]
		}
	}
	if delCnt > 0 {
		for i := 0; i < delCnt; i++ {
			objectPool.Put(n.ObjectValues[size-i-1])
		}
		n.ObjectValues = n.ObjectValues[:size-delCnt]
	}
	return delCnt > 0
}

// SetObjectStringElem set kv pair
func (n *Node) SetObjectStringElem(key, value string) *Node {
	for _, elem := range n.ObjectValues {
		if elem.Key.AsString() == key {
			elem.Value.Type = String
			elem.Value.Value = bytesToString(stdMarshalString([]byte(value)))
			return n
		}
	}
	elem := CreateObjectElem()
	elem.Key = CreateStringNode()
	elem.Key.Value = bytesToString(stdMarshalString([]byte(key)))
	elem.Value = CreateStringNode()
	elem.Value.Value = bytesToString(stdMarshalString([]byte(value)))
	n.ObjectValues = append(n.ObjectValues, elem)
	return n
}

// SetObjectIntElem set kv pair
func (n *Node) SetObjectIntElem(key string, value int64) *Node {
	for _, elem := range n.ObjectValues {
		if elem.Key.AsString() == key {
			elem.Value.Type = Integer
			elem.Value.Value = strconv.FormatInt(value, 10)
			return n
		}
	}
	elem := CreateObjectElem()
	elem.Key = CreateStringNode()
	elem.Key.Value = bytesToString(stdMarshalString([]byte(key)))
	elem.Value = CreateIntegerNode()
	elem.Value.Value = strconv.FormatInt(value, 10)
	n.ObjectValues = append(n.ObjectValues, elem)
	return n
}

// SetObjectUintElem set kv pair
func (n *Node) SetObjectUintElem(key string, value uint64) *Node {
	for _, elem := range n.ObjectValues {
		if elem.Key.AsString() == key {
			elem.Value.Type = Integer
			elem.Value.Value = strconv.FormatUint(value, 10)
			return n
		}
	}
	elem := CreateObjectElem()
	elem.Key = CreateStringNode()
	elem.Key.Value = bytesToString(stdMarshalString([]byte(key)))
	elem.Value = CreateIntegerNode()
	elem.Value.Value = strconv.FormatUint(value, 10)
	n.ObjectValues = append(n.ObjectValues, elem)
	return n
}

// SetObjectBoolElem set kv pair
func (n *Node) SetObjectBoolElem(key string, value bool) *Node {
	val := falseVal
	if value {
		val = trueVal
	}
	for _, elem := range n.ObjectValues {
		if elem.Key.AsString() == key {
			elem.Value.Type = Bool
			elem.Value.Value = val
			return n
		}
	}
	elem := CreateObjectElem()
	elem.Key = CreateStringNode()
	elem.Key.Value = bytesToString(stdMarshalString([]byte(key)))
	elem.Value = CreateBoolNode()
	elem.Value.Value = val
	n.ObjectValues = append(n.ObjectValues, elem)
	return n
}

// SetObjectNodeElem set kv pair
func (n *Node) SetObjectNodeElem(key string, value *Node) *Node {
	for _, elem := range n.ObjectValues {
		if elem.Key.AsString() == key {
			elem.Value = value
			return n
		}
	}
	elem := CreateObjectElem()
	elem.Key = CreateStringNode()
	elem.Key.Value = bytesToString(stdMarshalString([]byte(key)))
	elem.Value = value
	n.ObjectValues = append(n.ObjectValues, elem)
	return n
}

// AddObjectElem to node
func (n *Node) AddObjectElem(elem *ObjectElem) *Node {
	n.ObjectValues = append(n.ObjectValues, elem)
	return n
}

// AddArrayElem to node
func (n *Node) AddArrayElem(elem *Node) *Node {
	n.ArrayValues = append(n.ArrayValues, elem)
	return n
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
	n.Value = bytesToString(stdMarshalString(stringToBytes(str)))
	return n
}

// setStringBytes to string node
func (n *Node) setStringBytes(bts []byte) *Node {
	n.Value = bytesToString(stdMarshalString(bts))
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

// SetFloat to node
func (n *Node) SetFloat(f float64, prec int) *Node {
	n.Value = strconv.FormatFloat(f, 'f', prec, 64)
	return n
}

// SetInt to node
func (n *Node) SetInt(num int64) *Node {
	n.Value = strconv.FormatInt(num, 10)
	return n
}

// SetUint to node
func (n *Node) SetUint(num uint64) *Node {
	n.Value = strconv.FormatUint(num, 10)
	return n
}

// AsString as string
func (n *Node) AsString() string {
	switch n.Type {
	case String:
		s, err := stdUnmarshalString([]byte(n.Value))
		if err != nil {
			panic(err)
		}
		return bytesToString(s)
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
	return copyBytes(buf.Bytes()), err
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
	return copyBytes(buf.Bytes()), nil
}
