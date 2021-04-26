package qjson

import (
	"encoding/json"
	"reflect"
)

const (
	jsonTagName = "json"
	omitTag     = "-"
	omitEmpty   = "omitempty"
)

var converterInst = converter(0)

type converter int

func (cvt converter) Convert(obj interface{}) (node *Node, err error) {
	if obj == nil {
		return
	}
	tp := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)
	return cvt.ConvertAny(tp, v)
}

func (cvt converter) stdConvertAny(inf interface{}) (node *Node, err error) {
	if data, err := json.Marshal(inf); err != nil {
		return nil, err
	} else if tree, err := Decode(data); err != nil {
		return nil, err
	} else {
		return tree.Root, nil
	}
}

func (cvt converter) ConvertAny(tp reflect.Type, v reflect.Value) (node *Node, err error) {
	if inf := v.Interface(); inf != nil {
		if _, ok := inf.(json.Marshaler); ok {
			return cvt.stdConvertAny(inf)
		}
	}
	if tp.Kind() == reflect.Ptr {
		if v.IsNil() {
			return
		}
		tp = tp.Elem()
		v = v.Elem()
	}
	switch tp.Kind() {
	case reflect.Interface:
		if v.IsNil() {
			node = CreateNode()
		} else {
			node, err = cvt.ConvertAny(v.Elem().Type(), v.Elem())
		}
	case reflect.Bool:
		node = CreateBoolNode().SetBool(v.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		node = CreateIntegerNode().SetInt(v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		node = CreateIntegerNode().SetUint(v.Uint())
	case reflect.String:
		node = CreateStringNode().setStringBytes(stringToBytes(v.String()))
	case reflect.Array, reflect.Slice:
		node = CreateArrayNode()
		for i := 0; i < v.Len(); i++ {
			elemTp := tp.Elem()
			if n, er := cvt.ConvertAny(elemTp, v.Index(i)); er == nil && n != nil {
				node.ArrayValues = append(node.ArrayValues, n)
			}
		}
	case reflect.Map:
		if !v.IsNil() {
			node = CreateObjectNode()
			keys := v.MapKeys()
			keyTp := tp.Key()
			elemTp := tp.Elem()
			for _, key := range keys {
				if n, er := cvt.ConvertAny(elemTp, v.MapIndex(key)); er == nil && n != nil {
					elem := CreateObjectElem()
					if elem.Key, er = cvt.ConvertAny(keyTp, key); er != nil {
						return nil, er
					}
					elem.Value = n
					node.ObjectValues = append(node.ObjectValues, elem)
				}
			}
		}
	case reflect.Struct:
		node, err = cvt.ConvertObject(tp, v)
	}
	return
}

func (cvt converter) ConvertObject(tp reflect.Type, v reflect.Value) (node *Node, err error) {
	node = CreateObjectNode()
	for i := 0; i < tp.NumField(); i++ {
		fieldType := tp.Field(i)
		fieldVal := v.Field(i)
		name, omitempty, skip := getJSONKey(fieldType)
		if skip || (omitempty && fieldType.Type.Kind() == reflect.Ptr && fieldVal.IsNil()) {
			continue
		}
		val, er := cvt.ConvertAny(fieldType.Type, fieldVal)
		if er != nil {
			return nil, er
		}
		if val == nil {
			continue
		}
		if fieldType.Anonymous {
			node.ObjectValues = append(node.ObjectValues, val.ObjectValues...)
		} else {
			elem := CreateObjectElem()
			elem.Key = CreateStringNode().setStringBytes(stringToBytes(name))
			elem.Value = val
			node.ObjectValues = append(node.ObjectValues, elem)
		}
	}
	return
}

func getJSONKey(field reflect.StructField) (name string, omitempty bool, skip bool) {
	tag := field.Tag.Get(jsonTagName)
	if tag == omitTag {
		skip = true
		return
	}
	if len(tag) == 0 {
		return field.Name, false, false
	}
	data := stringToBytes(tag)
	for i, b := range data {
		if b == ',' {
			return bytesToString(data[:i]), i+1 <= len(tag) && bytesToString(data[i+1:]) == omitEmpty, false
		}
	}
	return tag, false, false
}
