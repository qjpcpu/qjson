package qjson

// JSONTree represent full json
type JSONTree struct {
	Root *Node
}

/* tree methods */

// IsNull tell node is null or not
func (tree *JSONTree) IsNull() bool {
	return tree.Root == nil || tree.Root.IsNull()
}

// MarshalJSON json marshaller
func (tree *JSONTree) MarshalJSON() ([]byte, error) {
	return tree.Root.MarshalJSON()
}

// UnmarshalJSON json unmarshaller
func (tree *JSONTree) UnmarshalJSON(data []byte) (err error) {
	var tree1 *JSONTree
	if tree1, err = Decode(data); err != nil {
		return err
	}
	*tree = *tree1
	return
}

// JSONString tree to string
func (tree *JSONTree) JSONString() string {
	return string(JSONMarshalWithPanic(tree))
}

// ColorfulMarshal print json with color
func (tree *JSONTree) ColorfulMarshal() []byte {
	return new(Formatter).Format(tree)
}

// ColorfulMarshalWithIndent print json with indent
func (tree *JSONTree) ColorfulMarshalWithIndent() []byte {
	return NewFormatter().Format(tree)
}

/* tree generator */
func makeNewTree() *JSONTree {
	return &JSONTree{Root: CreateNode()}
}

// Find json node/nodes by path selector
func (tree *JSONTree) Find(path string) *Node {
	return findNode(tree.Root, makeStPath(path))
}
