package qjson

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

var (
	nodePool   = &sync.Pool{New: func() interface{} { return new(Node) }}
	objectPool = &sync.Pool{New: func() interface{} { return new(ObjectElem) }}
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
