package qjson

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"sync"
)

var (
	nosenseChars = mapToSlice(map[byte]bool{
		'\n': true,
		' ':  true,
		'\t': true,
		'\r': true,
	})
)

const (
	objectStart  byte = '{'
	objectEnd    byte = '}'
	arrayStart   byte = '['
	arrayEnd     byte = ']'
	quote        byte = '"'
	escapeChar   byte = '\\'
	dotChar      byte = '.'
	negativeChar byte = '-'
	colonChar    byte = ':'
	commaChar    byte = ','
)

const (
	nullVal     = "null"
	trueVal     = "true"
	falseVal    = "false"
	dotString   = "."
	nullLen     = len(nullVal)
	trueValLen  = len(trueVal)
	falseValLen = len(falseVal)
	emptyVal    = ""
)

/* decode json entry function */
func decodeAny(jsonBytes []byte, offset int, tree *JSONTree) (int, error) {
	if offset >= len(jsonBytes) {
		return 0, errors.New("bad json")
	}

	var err error
	if nextValueIsObject(jsonBytes, offset) {
		offset, err = fillObjectNode(jsonBytes, offset, tree.Root)
	} else if nextValueIsArray(jsonBytes, offset) {
		offset, err = fillArrayNode(jsonBytes, offset, tree.Root)
	} else if nextValueIsNull(jsonBytes, offset) {
		offset = fillNullNode(jsonBytes, offset, tree.Root)
	} else if nextValueIsBool(jsonBytes, offset) {
		offset = fillBoolNode(jsonBytes, offset, tree.Root)
	} else if nextValueIsNumber(jsonBytes, offset) {
		offset = fillNumberNode(jsonBytes, offset, tree.Root)
	} else if nextValueIsString(jsonBytes, offset) {
		offset = fillStringNode(jsonBytes, offset, tree.Root)
	} else {
		err = fmt.Errorf("unknown json char %s", jsonBytes[offset:])
	}
	if err == nil {
		if searchFirstValidChar(jsonBytes, offset) >= 0 {
			err = fmt.Errorf("should nothing any more after %s", jsonBytes[:offset])
		}
	}
	return offset, err
}

/* node populate functions */

func fillObjectNode(jsonBytes []byte, offset int, node *Node) (int, error) {
	node.Type = Object
	offset++
	for {
		if noffset := searchFirstValidChar(jsonBytes, offset); noffset == -1 {
			return 0, fmt.Errorf("unexpected json char %s", jsonBytes[offset:])
		} else if jsonBytes[noffset] == objectEnd {
			offset = noffset
			break
		} else if jsonBytes[noffset] == commaChar {
			offset = noffset + 1
			continue
		} else {
			offset = noffset
		}
		var err error

		elem := CreateObject()
		elem.Key = CreateNode()
		elem.Value = CreateNode()

		if nextValueIsBool(jsonBytes, offset) {
			offset = fillBoolNode(jsonBytes, offset, elem.Key)
		} else if nextValueIsNumber(jsonBytes, offset) {
			offset = fillNumberNode(jsonBytes, offset, elem.Key)
		} else if nextValueIsString(jsonBytes, offset) {
			offset = fillStringNode(jsonBytes, offset, elem.Key)
		} else {
			return 0, fmt.Errorf("unexpected json char %s", jsonBytes[offset:])
		}
		/* should find : */
		if noffset := nextValueShouldBe(jsonBytes, offset, colonChar); noffset == -1 {
			return 0, fmt.Errorf("expect :, unexpected json char %s", jsonBytes[offset:])
		} else if offset = searchFirstValidChar(jsonBytes, noffset+1); offset == -1 {
			return 0, fmt.Errorf("expect object value, unexpected json char %s", jsonBytes[offset:])
		}

		if nextValueIsObject(jsonBytes, offset) {
			if offset, err = fillObjectNode(jsonBytes, offset, elem.Value); err != nil {
				return 0, err
			}
		} else if nextValueIsArray(jsonBytes, offset) {
			if offset, err = fillArrayNode(jsonBytes, offset, elem.Value); err != nil {
				return 0, err
			}
		} else if nextValueIsNull(jsonBytes, offset) {
			offset = fillNullNode(jsonBytes, offset, elem.Value)
		} else if nextValueIsBool(jsonBytes, offset) {
			offset = fillBoolNode(jsonBytes, offset, elem.Value)
		} else if nextValueIsNumber(jsonBytes, offset) {
			offset = fillNumberNode(jsonBytes, offset, elem.Value)
		} else if nextValueIsString(jsonBytes, offset) {
			offset = fillStringNode(jsonBytes, offset, elem.Value)
		} else {
			return 0, fmt.Errorf("unknown json char %s", jsonBytes[offset:])
		}
		node.ObjectValues = append(node.ObjectValues, elem)
	}
	return offset + 1, nil
}

func fillArrayNode(jsonBytes []byte, offset int, node *Node) (int, error) {
	node.Type = Array
	offset++
	for {
		if noffset := searchFirstValidChar(jsonBytes, offset); noffset == -1 {
			return 0, fmt.Errorf("unexpected json char %s", jsonBytes[offset:])
		} else if jsonBytes[noffset] == arrayEnd {
			offset = noffset
			break
		} else if jsonBytes[noffset] == commaChar {
			offset = noffset + 1
			continue
		} else {
			offset = noffset
		}
		var err error
		elem := CreateNode()
		if nextValueIsBool(jsonBytes, offset) {
			offset = fillBoolNode(jsonBytes, offset, elem)
		} else if nextValueIsNumber(jsonBytes, offset) {
			offset = fillNumberNode(jsonBytes, offset, elem)
		} else if nextValueIsString(jsonBytes, offset) {
			offset = fillStringNode(jsonBytes, offset, elem)
		} else if nextValueIsNull(jsonBytes, offset) {
			offset = fillNullNode(jsonBytes, offset, elem)
		} else if nextValueIsArray(jsonBytes, offset) {
			if offset, err = fillArrayNode(jsonBytes, offset, elem); err != nil {
				return 0, err
			}
		} else if nextValueIsObject(jsonBytes, offset) {
			if offset, err = fillObjectNode(jsonBytes, offset, elem); err != nil {
				return 0, err
			}
		} else {
			return 0, fmt.Errorf("unexpected json char %s", jsonBytes[offset:])
		}

		node.ArrayValues = append(node.ArrayValues, elem)
	}
	return offset + 1, nil

}

func fillNullNode(jsonBytes []byte, offset int, node *Node) int {
	node.Type = Null
	node.Value = nullVal
	return offset + nullLen
}

func fillBoolNode(jsonBytes []byte, offset int, node *Node) int {
	node.Type = Bool
	if bytesToString(jsonBytes[offset:offset+trueValLen]) == trueVal {
		node.Value = trueVal
		offset += trueValLen
	} else {
		node.Value = falseVal
		offset += falseValLen
	}
	return offset
}

func fillNumberNode(jsonBytes []byte, offset int, node *Node) int {
	start := offset
	for ; offset < len(jsonBytes); offset++ {
		b := jsonBytes[offset]
		if b == dotChar || b == negativeChar || isIntegerChar(b) {
			continue
		}
		break
	}

	node.Value = bytesToString(jsonBytes[start:offset])
	if strings.Contains(node.Value, dotString) {
		node.Type = Float
	} else {
		node.Type = Integer
	}
	return offset
}

func fillStringNode(jsonBytes []byte, offset int, node *Node) int {
	start := offset
	var skipNextChar bool
	for offset = offset + 1; offset < len(jsonBytes); offset++ {
		if skipNextChar {
			skipNextChar = false
			continue
		}
		b := jsonBytes[offset]
		if b == escapeChar {
			skipNextChar = true
		} else if b == quote {
			offset++
			break
		}
	}
	node.Type = String
	node.Value = bytesToString(jsonBytes[start:offset])
	return offset
}

/* search first valid char from offset, return found index or -1 */
func searchFirstValidChar(jsonBytes []byte, offset int) int {
	for i := offset; i < len(jsonBytes); i++ {
		b := jsonBytes[i]
		if nosenseChars.Contains(b) {
			continue
		}
		return i
	}
	return -1
}

/* next char detect fuctions */

func nextValueIsNumber(jsonBytes []byte, offset int) bool {
	if jsonBytes[offset] == negativeChar {
		if offset+1 >= len(jsonBytes) || !isIntegerChar(jsonBytes[offset+1]) {
			return false
		}
	} else if isIntegerChar(jsonBytes[offset]) {
	} else {
		return false
	}
	start := offset
	for offset = offset + 1; offset < len(jsonBytes); offset++ {
		if !isIntegerChar(jsonBytes[offset]) && jsonBytes[offset] != dotChar {
			break
		}
	}
	word := string(jsonBytes[start:offset])
	if strings.Count(word, dotString) > 1 {
		return false
	}
	return true
}

func nextValueIsString(jsonBytes []byte, offset int) bool {
	if jsonBytes[offset] != quote {
		return false
	}
	for offset = offset + 1; offset < len(jsonBytes); {
		b := jsonBytes[offset]
		if b == escapeChar {
			offset += 2
			continue
		} else if b == quote {
			return true
		}
		offset++
	}
	return false
}

func nextValueIsObject(jsonBytes []byte, offset int) bool {
	return jsonBytes[offset] == objectStart
}

func nextValueIsArray(jsonBytes []byte, offset int) bool {
	return jsonBytes[offset] == arrayStart
}

func nextValueIsNull(jsonBytes []byte, offset int) bool {
	return offset+nullLen <= len(jsonBytes) && bytesToString(jsonBytes[offset:offset+nullLen]) == nullVal
}

func nextValueIsBool(jsonBytes []byte, offset int) bool {
	if offset+trueValLen <= len(jsonBytes) && bytesToString(jsonBytes[offset:offset+trueValLen]) == trueVal {
		return true
	} else if offset+falseValLen <= len(jsonBytes) && bytesToString(jsonBytes[offset:offset+falseValLen]) == falseVal {
		return true
	}
	return false
}

func nextValueShouldBe(jsonBytes []byte, offset int, options ...byte) int {
	idx := searchFirstValidChar(jsonBytes, offset)
	if idx == -1 {
		return -1
	}
	for _, b := range options {
		if b == jsonBytes[idx] {
			return idx
		}
	}
	return -1
}

/* active object pools */
var bytesPool = sync.Pool{New: func() interface{} { return new(bytes.Buffer) }}
