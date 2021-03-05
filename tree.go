package qjson

// JSONTree represent full json
type JSONTree struct {
	Root *Node
}

/* tree methods */

// IsNull tell node is null or not
func (t *JSONTree) IsNull() bool {
	return t.Root == nil || t.Root.IsNull()
}

// MarshalJSON json marshaller
func (t *JSONTree) MarshalJSON() ([]byte, error) {
	return t.Root.MarshalJSON()
}

// UnmarshalJSON json unmarshaller
func (t *JSONTree) UnmarshalJSON(data []byte) (err error) {
	var tree *JSONTree
	if tree, err = Decode(data); err != nil {
		return err
	}
	*t = *tree
	return
}

// JSONString tree to string
func (t *JSONTree) JSONString() string {
	return string(JSONMarshalWithPanic(t))
}

// ColorfulMarshal print json with color
func (t *JSONTree) ColorfulMarshal() []byte {
	return new(Formatter).Format(t)
}

// ColorfulMarshalWithIndent print json with indent
func (t *JSONTree) ColorfulMarshalWithIndent() []byte {
	return NewFormatter().Format(t)
}

/* tree generator */
func makeNewTree() *JSONTree {
	return &JSONTree{Root: CreateNode()}
}
