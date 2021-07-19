package qjson

import (
	"strconv"
	"strings"
)

const (
	sharpSym               = "#"
	arrayElemEq            = "=="
	arrayElemContains      = "="
	arrayElemNotEq         = "!=="
	arrayElemNotContains   = "!="
	arrayElemGreaterThan   = ">"
	arrayElemGreaterEqThan = ">="
	arrayElemLessThan      = "<"
	arrayElemLessEqThan    = "<="
)

type stPath struct {
	Name     string
	Selector string
	Op       string
	Val      string
}

func (sp stPath) isArrayElemSelector() bool {
	return sp.Name == "#"
}

func (sp stPath) isInteger() bool {
	_, err := strconv.ParseUint(sp.Name, 10, 64)
	return err == nil
}

func (sp stPath) asInteger() int {
	i, _ := strconv.ParseUint(sp.Name, 10, 64)
	return int(i)
}

func findNode(node *Node, paths []stPath) *Node {
	if len(paths) == 0 {
		return node
	}
	if node == nil {
		return nil
	}
	p := paths[0]
	switch node.Type {
	case Null, String, Bool, Integer, Float:
		/* should never come here */
		return nil
	case Object:
		for _, n := range node.ObjectValues {
			if n.Key.AsString() == p.Name {
				return findNode(n.Value, paths[1:])
			}
		}
	case Array:
		if p.isArrayElemSelector() {
			var list []*Node
			fromList := node.ArrayValues
			if p.Op != "" {
				fromList = filterArrayNodeBySelector(node, p)
			}
			for _, n := range fromList {
				if out := findNode(n, paths[1:]); out != nil {
					list = append(list, out)
				}
			}
			n := CreateArrayNode()
			n.ArrayValues = list
			return n
		} else if p.isInteger() && p.asInteger() < len(node.ArrayValues) {
			return findNode(node.ArrayValues[p.asInteger()], paths[1:])
		}
	}
	return nil
}

func filterArrayNodeBySelector(node *Node, path stPath) []*Node {
	paths, ok := makeStPath(path.Selector)
	if !ok {
		return nil
	}
	val := strings.TrimSuffix(strings.TrimPrefix(path.Val, `"`), `"`)
	var list []*Node
	for _, n := range node.ArrayValues {
		if out := findNode(n, paths); out != nil {
			if isElemMatched(out, path.Op, val) {
				list = append(list, n)
			}
		}
	}
	return list
}

func isElemMatched(n *Node, op string, val string) bool {
	switch n.Type {
	case Array:
		for _, n := range n.ArrayValues {
			if isElemMatched(n, op, val) {
				return true
			}
		}
	case Object:
	default:
		v := n.AsString()
		switch op {
		case arrayElemContains:
			return strings.Contains(v, val)
		case arrayElemEq:
			return v == val
		case arrayElemNotEq:
			return v != val
		case arrayElemNotContains:
			return !strings.Contains(v, val)
		case arrayElemGreaterThan:
			if n.IsNumber() {
				if strings.Contains(val, ".") || strings.Contains(n.AsString(), ".") {
					if num, err := strconv.ParseFloat(val, 64); err == nil {
						return n.AsFloat() > num
					}
				} else {
					if num, err := strconv.ParseInt(val, 10, 64); err == nil {
						return n.AsInt() > num
					}
				}
			} else {
				return n.AsString() > val
			}
		case arrayElemGreaterEqThan:
			if n.IsNumber() {
				if strings.Contains(val, ".") || strings.Contains(n.AsString(), ".") {
					if num, err := strconv.ParseFloat(val, 64); err == nil {
						return n.AsFloat() >= num
					}
				} else {
					if num, err := strconv.ParseInt(val, 10, 64); err == nil {
						return n.AsInt() >= num
					}
				}
			} else {
				return n.AsString() >= val
			}
		case arrayElemLessEqThan:
			if n.IsNumber() {
				if strings.Contains(val, ".") || strings.Contains(n.AsString(), ".") {
					if num, err := strconv.ParseFloat(val, 64); err == nil {
						return n.AsFloat() <= num
					}
				} else {
					if num, err := strconv.ParseInt(val, 10, 64); err == nil {
						return n.AsInt() <= num
					}
				}
			} else {
				return n.AsString() <= val
			}
		case arrayElemLessThan:
			if n.IsNumber() {
				if strings.Contains(val, ".") || strings.Contains(n.AsString(), ".") {
					if num, err := strconv.ParseFloat(val, 64); err == nil {
						return n.AsFloat() < num
					}
				} else {
					if num, err := strconv.ParseInt(val, 10, 64); err == nil {
						return n.AsInt() < num
					}
				}
			} else {
				return n.AsString() < val
			}
		}
	}
	return false
}

func makeStPath(p string) ([]stPath, bool) {
	var paths []stPath
	proj := map[byte]byte{
		'(': ')',
		'"': '"',
	}
	data := []byte(strings.TrimPrefix(p, "."))
	var start int
	for i := 0; i < len(data); {
		if data[i] == '\\' {
			if i+1 < len(data) && data[i+1] == '.' {
				data[i] = 0
			}
			i += 2
			continue
		} else if data[i] == '#' && i+1 < len(data) && data[i+1] == '(' {
			if closeIdx := findCloseSym(data, i+2, len(data), '(', proj); closeIdx == -1 {
				return nil, false
			} else {
				i = closeIdx + 1
				if closeIdx == len(data)-1 {
					paths = append(paths, stPath{Name: removeByte(string(data[start:]), 0)})
				}
				continue
			}
		} else if data[i] == '.' && i > start {
			paths = append(paths, stPath{Name: removeByte(string(data[start:i]), 0)})
			start = i + 1
		} else if i == len(data)-1 {
			paths = append(paths, stPath{Name: removeByte(string(data[start:]), 0)})
			start = i + 1
		}
		i++
	}
	for i, path := range paths {
		paths[i] = reformatStPath(path)
	}
	return paths, true
}

func reformatStPath(p stPath) stPath {
	if len(p.Name) > 3 && p.Name[0] == '#' && p.Name[1] == '(' && p.Name[len(p.Name)-1] == ')' {
		step := p.Name[2 : len(p.Name)-1]
		p.Name = "#"
		p.Selector, p.Op, p.Val = reformatStStep(step)
	}
	return p
}

func reformatStStep(step string) (selector string, op string, val string) {
	idx := strings.Index(step, "#(")
	if idx >= 0 && idx < len(step) {
		s, o, v := reformatStStep(step[idx+2 : len(step)-1])
		if s != "" {
			step = step[:idx] + "#." + s
		} else {
			step = step[:idx] + "#"
		}
		return step, o, v
	} else {
		if idx := strings.Index(step, arrayElemNotEq); idx >= 0 {
			selector = strings.TrimSpace(step[:idx])
			op = arrayElemNotEq
			val = strings.TrimSpace(step[idx+len(op):])
		} else if idx = strings.Index(step, arrayElemNotContains); idx >= 0 {
			selector = strings.TrimSpace(step[:idx])
			op = arrayElemNotContains
			val = strings.TrimSpace(step[idx+len(op):])
		} else if idx = strings.Index(step, arrayElemEq); idx >= 0 {
			selector = strings.TrimSpace(step[:idx])
			op = arrayElemEq
			val = strings.TrimSpace(step[idx+len(op):])
		} else if idx = strings.Index(step, arrayElemGreaterEqThan); idx >= 0 {
			selector = strings.TrimSpace(step[:idx])
			op = arrayElemGreaterEqThan
			val = strings.TrimSpace(step[idx+len(op):])
		} else if idx = strings.Index(step, arrayElemLessEqThan); idx >= 0 {
			selector = strings.TrimSpace(step[:idx])
			op = arrayElemLessEqThan
			val = strings.TrimSpace(step[idx+len(op):])
		} else if idx = strings.Index(step, arrayElemContains); idx >= 0 {
			op = arrayElemContains
			selector = strings.TrimSpace(step[:idx])
			val = strings.TrimSpace(step[idx+len(op):])
		} else if idx = strings.Index(step, arrayElemLessThan); idx >= 0 {
			op = arrayElemLessThan
			selector = strings.TrimSpace(step[:idx])
			val = strings.TrimSpace(step[idx+len(op):])
		} else if idx = strings.Index(step, arrayElemGreaterThan); idx >= 0 {
			op = arrayElemGreaterThan
			selector = strings.TrimSpace(step[:idx])
			val = strings.TrimSpace(step[idx+len(op):])
		}
		return
	}
}

func removeByte(s string, b byte) string {
	var cnt int
	data := []byte(s)
	for i := 0; i < len(data); i++ {
		if data[i] == b {
			cnt++
		} else if cnt > 0 {
			data[i-cnt] = data[i]
		}
	}
	return string(data[:len(data)-cnt])
}

func findCloseSym(data []byte, from, to int, openSym byte, proj map[byte]byte) int {
	closeSym, ok := proj[openSym]
	if !ok {
		return -1
	}
	for i := from; i < to && i < len(data); {
		if data[i] == '\\' {
			i += 2
			continue
		}
		if data[i] == closeSym {
			return i
		}
		if _, ok := proj[data[i]]; ok {
			if nextCloseIndex := findCloseSym(data, i+1, to, data[i], proj); nextCloseIndex == -1 {
				return -1
			} else {
				i = nextCloseIndex + 1
				continue
			}
		}
		i++
	}
	return -1
}
