package qjson

const defaultQueueSize = 256

var (
	nodeQueue   = newQueue(defaultQueueSize)
	objectQueue = newQueue(defaultQueueSize)
)

// Release json tree for objects reuse
func (tree *JSONTree) Release() {
	node := tree.Root
	nodeQueue.Put(node)
	v := &tree.Root
	*v = nil
}

// CreateNode by pool
func CreateNode() *Node {
	if v, ok := nodeQueue.Get(); ok {
		node := v.(*Node)
		if node.Type == Object {
			if len(node.ObjectValues) > 0 {
				for i := range node.ObjectValues {
					objectQueue.Put(node.ObjectValues[i])
				}
				node.ObjectValues = node.ObjectValues[:0]
			}
		} else if node.Type == Array {
			if len(node.ArrayValues) > 0 {
				for i := range node.ArrayValues {
					nodeQueue.Put(node.ArrayValues[i])
				}
				node.ArrayValues = node.ArrayValues[:0]
			}
		}
		node.color = Color(0)
		node.Value = emptyVal
		node.Type = Null
		return node
	}
	return &Node{}
}

// CreateObject by pool
func CreateObject() *ObjectElem {
	if v, ok := objectQueue.Get(); ok {
		object := v.(*ObjectElem)
		if object.Key != nil {
			nodeQueue.Put(object.Key)
			object.Key = nil
		}
		if object.Value != nil {
			nodeQueue.Put(object.Value)
			object.Value = nil
		}
		return object
	}
	return &ObjectElem{}
}
