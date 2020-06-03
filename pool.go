package qjson

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

var (
	nodePool   = &sync.Pool{New: func() interface{} { return &Node{} }}
	objectPool = &sync.Pool{New: func() interface{} { return &ObjectElem{} }}
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
	if node.Type == Object {
		if len(node.ObjectValues) > 0 {
			for i := range node.ObjectValues {
				objectPool.Put(node.ObjectValues[i])
			}
			node.ObjectValues = node.ObjectValues[:0]
		}
	} else if node.Type == Array {
		if len(node.ArrayValues) > 0 {
			for i := range node.ArrayValues {
				nodePool.Put(node.ArrayValues[i])
			}
			node.ArrayValues = node.ArrayValues[:0]
		}
	}
	node.color = Color(0)
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
