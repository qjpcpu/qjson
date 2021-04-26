package qjson

import (
	"strconv"
	"strings"
)

type stPath struct {
	Name string
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
			for _, n := range node.ArrayValues {
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

func makeStPath(p string) []stPath {
	var paths []stPath
	data := []byte(strings.TrimPrefix(p, "."))
	var start int
	for i := 0; i < len(data); {
		if data[i] == '\\' {
			if i+1 < len(data) && data[i+1] == '.' {
				data[i] = 0
			}
			i += 2
			continue
		} else if data[i] == '.' && i > start {
			paths = append(paths, stPath{Name: removeByte(string(data[start:i]), 0)})
			start = i + 1
		} else if i == len(data)-1 {
			paths = append(paths, stPath{Name: removeByte(string(data[start:]), 0)})
			start = i + 1
		}
		i += 1
	}
	return paths
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
