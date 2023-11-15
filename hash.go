package qjson

import (
	"bytes"
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
		n.hashId = bytesHash(stringToBytes(n.Value))
	case Object:
		list := getStrSlice()
		defer putStrSlice(list)
		for _, item := range n.ObjectValues {
			list.Str = append(list.Str, strconv.FormatUint(item.Key.hash(), 10)+":"+
				strconv.FormatUint(item.Value.hash(), 10))
		}
		sort.Strings(list.Str)
		n.hashId = bytesHash(stringToBytes(strings.Join(list.Str, ",")))
	case Array:
		buf := bytesPool.Get().(*bytes.Buffer)
		buf.Reset()
		defer bytesPool.Put(buf)
		for _, item := range n.ArrayValues {
			buf.Write(stringToBytes(strconv.FormatUint(item.hash(), 10)))
			buf.WriteByte(',')
		}
		n.hashId = bytesHash(buf.Bytes())
	}

	return n.hashId
}

func bytesHash(bs []byte) uint64 {
	h := fnv.New64()
	h.Write(bs)
	return h.Sum64()
}
