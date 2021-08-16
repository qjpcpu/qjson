package qjson

import (
	"bytes"
	ejson "encoding/json"
	"reflect"
)

/* package functions */

// New json tree
func New() *JSONTree {
	return makeNewTree()
}

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
func ConvertToJSONTree(obj interface{}) (tree *JSONTree, err error) {
	tree = makeNewTree()
	if tree.Root, err = converterInst.Convert(obj); err != nil {
		return tree, err
	}
	return tree, nil
}

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

// JSONMarshalWithPanic json marshal with panic
func JSONMarshalWithPanic(t interface{}) []byte {
	return jsonMarshalWithPanic(t, false)
}

// JSONIndentMarshalWithPanic json marshal with panic
func JSONIndentMarshalWithPanic(t interface{}) []byte {
	return jsonMarshalWithPanic(t, true)
}

func jsonMarshalWithPanic(t interface{}, indent bool) []byte {
	if t == nil {
		return nil
	}
	if v := reflect.ValueOf(t); v.Kind() == reflect.Ptr && v.IsNil() {
		return nil
	}
	buffer := &bytes.Buffer{}
	encoder := ejson.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	if indent {
		encoder.SetIndent("", "  ")
	}
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
