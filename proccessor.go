package qjson

import (
	"strings"
)

var (
	nosenseChars = map[byte]bool{
		'\n': true,
		' ':  true,
		'\t': true,
	}
	stopChars = map[byte]bool{
		',': true,
	}
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
	nullLen     = len(nullVal)
	trueValLen  = len(trueVal)
	falseValLen = len(falseVal)
)

func searchFirstValidChar(jsonBytes []byte, offset int) int {
	for i := offset; i < len(jsonBytes); i++ {
		b := jsonBytes[i]
		if nosenseChars[b] {
			continue
		}
		return i
	}
	return -1
}

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
	if strings.Count(word, ".") > 1 {
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
	return offset+nullLen <= len(jsonBytes) && string(jsonBytes[offset:offset+nullLen]) == nullVal
}

func nextValueIsBool(jsonBytes []byte, offset int) bool {
	if offset+trueValLen <= len(jsonBytes) && string(jsonBytes[offset:offset+trueValLen]) == trueVal {
		return true
	} else if offset+falseValLen <= len(jsonBytes) && string(jsonBytes[offset:offset+falseValLen]) == falseVal {
		return true
	}
	return false
}

func isIntegerChar(b byte) bool {
	return b >= '0' && b <= '9'
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
