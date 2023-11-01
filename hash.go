package qjson

import (
	"hash/fnv"
	"sort"
	"strconv"
	"strings"
)

func (n *Node) Hash() uint64 {
	if n.hashId == 0 {
		return n.hash()
	}
	return n.hashId
}

func (n *Node) Rehash() uint64 {
	return n.hash()
}

func (n *Node) hash() uint64 {
	if n == nil {
		return 0
	}
	switch n.Type {
	case Null:
		n.hashId = 0
	case String, Bool, Integer, Float:
		h := fnv.New64()
		h.Write(stringToBytes(n.Value))
		n.hashId = h.Sum64()
	case Object:
		list := make([]string, len(n.ObjectValues))
		for i, item := range n.ObjectValues {
			list[i] = strconv.FormatUint(item.Key.hash(), 10) + ":" +
				strconv.FormatUint(item.Value.hash(), 10)
		}
		sort.Strings(list)
		h := fnv.New64()
		h.Write(stringToBytes(strings.Join(list, ",")))
		n.hashId = h.Sum64()
	case Array:
		list := make([]string, len(n.ArrayValues))
		for i, item := range n.ArrayValues {
			list[i] = strconv.FormatUint(item.hash(), 10)
		}
		h := fnv.New64()
		h.Write(stringToBytes(strings.Join(list, ",")))
		n.hashId = h.Sum64()
	}

	return n.hashId
}
