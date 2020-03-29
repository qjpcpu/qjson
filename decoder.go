package qjson

import (
	"errors"
	"fmt"
	"strings"
)

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

func decodeAny(jsonBytes []byte, offset int, tree *JSONTree) (int, error) {
	if offset >= len(jsonBytes) {
		return 0, errors.New("bad json")
	}

	var err error
	if nextValueIsObject(jsonBytes, offset) {
		offset, err = fillObjectNode(jsonBytes, offset, &tree.Root)
	} else if nextValueIsArray(jsonBytes, offset) {
		offset, err = fillArrayNode(jsonBytes, offset, &tree.Root)
	} else if nextValueIsNull(jsonBytes, offset) {
		offset = fillNullNode(jsonBytes, offset, &tree.Root)
	} else if nextValueIsBool(jsonBytes, offset) {
		offset = fillBoolNode(jsonBytes, offset, &tree.Root)
	} else if nextValueIsNumber(jsonBytes, offset) {
		offset = fillNumberNode(jsonBytes, offset, &tree.Root)
	} else if nextValueIsString(jsonBytes, offset) {
		offset = fillStringNode(jsonBytes, offset, &tree.Root)
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
		elem := ObjectElem{}
		if nextValueIsBool(jsonBytes, offset) {
			offset = fillBoolNode(jsonBytes, offset, &elem.Key)
		} else if nextValueIsNumber(jsonBytes, offset) {
			offset = fillNumberNode(jsonBytes, offset, &elem.Key)
		} else if nextValueIsString(jsonBytes, offset) {
			offset = fillStringNode(jsonBytes, offset, &elem.Key)
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
			if offset, err = fillObjectNode(jsonBytes, offset, &elem.Value); err != nil {
				return 0, err
			}
		} else if nextValueIsArray(jsonBytes, offset) {
			if offset, err = fillArrayNode(jsonBytes, offset, &elem.Value); err != nil {
				return 0, err
			}
		} else if nextValueIsNull(jsonBytes, offset) {
			offset = fillNullNode(jsonBytes, offset, &elem.Value)
		} else if nextValueIsBool(jsonBytes, offset) {
			offset = fillBoolNode(jsonBytes, offset, &elem.Value)
		} else if nextValueIsNumber(jsonBytes, offset) {
			offset = fillNumberNode(jsonBytes, offset, &elem.Value)
		} else if nextValueIsString(jsonBytes, offset) {
			offset = fillStringNode(jsonBytes, offset, &elem.Value)
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
		elem := Node{}
		if nextValueIsBool(jsonBytes, offset) {
			offset = fillBoolNode(jsonBytes, offset, &elem)
		} else if nextValueIsNumber(jsonBytes, offset) {
			offset = fillNumberNode(jsonBytes, offset, &elem)
		} else if nextValueIsString(jsonBytes, offset) {
			offset = fillStringNode(jsonBytes, offset, &elem)
		} else if nextValueIsNull(jsonBytes, offset) {
			offset = fillNullNode(jsonBytes, offset, &elem)
		} else if nextValueIsArray(jsonBytes, offset) {
			if offset, err = fillArrayNode(jsonBytes, offset, &elem); err != nil {
				return 0, err
			}
		} else if nextValueIsObject(jsonBytes, offset) {
			if offset, err = fillObjectNode(jsonBytes, offset, &elem); err != nil {
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
	if string(jsonBytes[offset:offset+trueValLen]) == trueVal {
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

	node.Value = string(jsonBytes[start:offset])
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
	node.Value = string(jsonBytes[start:offset])
	return offset
}
