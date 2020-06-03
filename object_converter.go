package qjson

import (
	"reflect"
)

const (
	jsonTagName = "json"
	omitTag     = "-"
	omitEmpty   = "omitempty"
)

type converter int

func (cvt converter) Convert(obj interface{}) (node *Node) {
	if obj == nil {
		return
	}
	tp := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)
	return cvt.ConvertAny(tp, v)
}

func (cvt converter) ConvertAny(tp reflect.Type, v reflect.Value) (node *Node) {
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
			node = cvt.ConvertAny(v.Elem().Type(), v.Elem())
		}
	case reflect.Bool:
		node = CreateBoolNode().SetBool(v.Bool())
	case reflect.Int:
		node = CreateIntegerNode().SetInt(v.Int())
	case reflect.Uint:
		node = CreateIntegerNode().SetUint(v.Uint())
	case reflect.String:
		node = CreateStringNode().SetStringBytes(stringToBytes(v.String()))
	case reflect.Array, reflect.Slice:
		node = CreateArrayNode()
		for i := 0; i < v.Len(); i++ {
			elemTp := tp.Elem()
			if n := cvt.ConvertAny(elemTp, v.Index(i)); n != nil {
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
				if n := cvt.ConvertAny(elemTp, v.MapIndex(key)); n != nil {
					elem := CreateObjectElem()
					elem.Key = cvt.ConvertAny(keyTp, key)
					elem.Value = n
					node.ObjectValues = append(node.ObjectValues, elem)
				}
			}
		}
	case reflect.Struct:
		node = cvt.ConvertObject(tp, v)
	}
	return
}

func (cvt converter) ConvertObject(tp reflect.Type, v reflect.Value) (node *Node) {
	node = CreateObjectNode()
	for i := 0; i < tp.NumField(); i++ {
		fieldType := tp.Field(i)
		fieldVal := v.Field(i)
		name, omitempty, skip := getJSONKey(fieldType)
		if skip || (omitempty && fieldType.Type.Kind() == reflect.Ptr && fieldVal.IsNil()) {
			continue
		}
		val := cvt.ConvertAny(fieldType.Type, fieldVal)
		if val == nil {
			continue
		}
		if fieldType.Anonymous {
			node.ObjectValues = append(node.ObjectValues, val.ObjectValues...)
		} else {
			elem := CreateObjectElem()
			elem.Key = CreateStringNode().SetStringBytes(stringToBytes(name))
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
