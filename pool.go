package qjson

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

var (
	nodePool   = &sync.Pool{New: func() interface{} { return new(Node) }}
	objectPool = &sync.Pool{New: func() interface{} { return new(ObjectElem) }}
	strsPool   = &sync.Pool{New: func() interface{} { return new(strSlice) }}
)

// Release json tree for objects reuse
func (tree *JSONTree) Release() {
	if node := tree.Root; node != nil {
		if atomic.CompareAndSwapPointer((*unsafe.Pointer)((unsafe.Pointer)(&tree.Root)), unsafe.Pointer(node), unsafe.Pointer(nil)) {
			nodePool.Put(node)
		}
	}
}

// CreateNode by pool
func CreateNode() *Node {
	node := nodePool.Get().(*Node)
	if node.ObjectValues != nil {
		for i := range node.ObjectValues {
			objectPool.Put(node.ObjectValues[i])
		}
		node.ObjectValues = nil
	}
	if node.ArrayValues != nil {
		for i := range node.ArrayValues {
			nodePool.Put(node.ArrayValues[i])
		}
		node.ArrayValues = nil
	}
	node.Value = emptyVal
	node.Type = Null
	node.hashId = 0
	return node
}

// CreateObjectElem by pool
func CreateObjectElem() *ObjectElem {
	object := objectPool.Get().(*ObjectElem)
	if object.Key != nil {
		nodePool.Put(object.Key)
		object.Key = nil
	}
	if object.Value != nil {
		nodePool.Put(object.Value)
		object.Value = nil
	}
	return object
}

type strSlice struct {
	Str []string
}

func getStrSlice() *strSlice {
	ss := strsPool.Get().(*strSlice)
	if ss.Str != nil {
		ss.Str = ss.Str[:0]
	}
	return ss
}

func putStrSlice(ss *strSlice) {
	strsPool.Put(ss)
}
