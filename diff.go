package qjson

import (
	"fmt"
	"strconv"
	"strings"
)

type DiffType int

const undefined = "undefined"

const (
	DiffOfType DiffType = iota
	DiffOfValue
)

func (t DiffType) String() string {
	switch t {
	case DiffOfType:
		return "Type"
	case DiffOfValue:
		return "Value"
	default:
		return ""
	}
}

func Diff(t1, t2 *JSONTree) DiffItems {
	d := &differ{}
	d.diffNode(t1.Root, t2.Root, "")
	return d.diffList
}

type DiffItem struct {
	Type        DiffType
	Path        string
	Left, Right string
}

type DiffItems []DiffItem

func (items DiffItems) Exist() bool { return len(items) > 0 }

func (items DiffItems) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Total %v diff\n", len(items)))
	for _, item := range items {
		sb.WriteString(item.String() + "\n")
	}
	return sb.String()
}

func (item DiffItem) String() string {
	return fmt.Sprintf("DiffOf%v Path: %v Left:%v Right:%v", item.Type.String(), item.Path, item.Left, item.Right)
}

type differ struct {
	diffList []DiffItem
}

func (d *differ) diffNode(n1, n2 *Node, prefix string) {
	if n1.Type != n2.Type {
		d.addDiff(DiffOfType, prefix, n1, n2)
		return
	}
	switch n1.Type {
	case String, Bool, Integer, Float:
		if n1.Value != n2.Value {
			d.addDiff(DiffOfValue, prefix, n1, n2)
		}
	case Object:
		d.diffObject(n1, n2, prefix)
	case Array:
		d.diffArray(n1, n2, prefix)
	}
}

func (d *differ) diffArray(n1, n2 *Node, prefix string) {
	if len(n1.ArrayValues) != len(n2.ArrayValues) {
		d.addDiff(DiffOfValue, prefix, n1, n2)
		return
	}
	for i, item := range n1.ArrayValues {
		d.diffNode(item, n2.ArrayValues[i], d.appendPath(prefix, strconv.Itoa(i)))
	}
}

func (d *differ) diffObject(n1, n2 *Node, prefix string) {
	left, right := n1.AsMap(), n2.AsMap()
	for k, v := range left {
		v2, ok := right[k]
		if !ok {
			d.addDiff(DiffOfValue, d.appendPath(prefix, k), v, nil)
			continue
		}
		d.diffNode(v, v2, d.appendPath(prefix, k))
		delete(right, k)
	}
	for k, v := range right {
		v2, ok := left[k]
		if !ok {
			d.addDiff(DiffOfValue, d.appendPath(prefix, k), nil, v)
			continue
		}
		d.diffNode(v, v2, d.appendPath(prefix, k))
	}
}

func (d *differ) appendPath(prefix string, suffix ...string) string {
	if len(suffix) > 0 {
		return prefix + "." + strings.Join(suffix, ".")
	}
	return prefix
}

func (d *differ) addDiff(t DiffType, prefix string, ln, rn *Node) {
	lv, rv := undefined, undefined
	if ln != nil {
		lv = ln.AsJSON()
	}
	if rn != nil {
		rv = rn.AsJSON()
	}
	d.diffList = append(d.diffList, DiffItem{Type: t, Path: strings.TrimPrefix(prefix, "."), Left: lv, Right: rv})
}
