package qjson

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestJSONIndentMarshalWithPanic(t *testing.T) {
	obj := map[string]interface{}{"a": 1, "b": map[string]interface{}{"c": "x"}}
	got := JSONIndentMarshalWithPanic(obj)
	want, _ := json.MarshalIndent(obj, "", "  ")
	// JSONIndentMarshalWithPanic trims trailing newline; MarshalIndent doesn't add one
	if string(got) != string(want) {
		t.Fatalf("indent marshal mismatch\nwant: %s\n got: %s", string(want), string(got))
	}
}

func TestPrettyMarshal(t *testing.T) {
	obj := map[string]interface{}{"k": "v", "x": 1}
	out := string(PrettyMarshal(obj))
	if !strings.Contains(out, "\x1b[") {
		t.Fatalf("expected ANSI sequences in PrettyMarshal output: %q", out)
	}
}

func TestPrettyMarshalWithIndent(t *testing.T) {
	obj := map[string]interface{}{"k": map[string]interface{}{"kk": "v"}, "arr": []interface{}{"x", 1}}
	out := string(PrettyMarshalWithIndent(obj))
	if !strings.Contains(out, "\n") || !strings.Contains(out, "\x1b[") {
		t.Fatalf("expected indent and ANSI sequences in output: %q", out)
	}
	// ensure brackets and commas are present for arrays/objects
	if !strings.Contains(out, "[") || !strings.Contains(out, "]") || !strings.Contains(out, ",") {
		t.Fatalf("expected pretty output to contain structural characters: %q", out)
	}
}

func TestTreeIsNullJSONStringIndentString(t *testing.T) {
	// IsNull true when root nil or Null
	tree := &JSONTree{}
	if !tree.IsNull() {
		t.Fatalf("empty tree should be null")
	}
	// IsNull false when root non-null type
	tr, err := Decode([]byte(`{"a":1}`))
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if tr.IsNull() {
		t.Fatalf("decoded tree should not be null")
	}
	// JSONString
	s1 := tr.JSONString()
	s2 := string(JSONMarshalWithPanic(tr))
	if s1 != s2 {
		t.Fatalf("JSONString mismatch: %q vs %q", s1, s2)
	}
	// JSONIndentString
	is1 := tr.JSONIndentString()
	is2 := string(JSONIndentMarshalWithPanic(tr))
	if is1 != is2 {
		t.Fatalf("JSONIndentString mismatch: %q vs %q", is1, is2)
	}
}

func TestFindAndRemovePaths(t *testing.T) {
	tr, err := Decode([]byte(`{"obj":{"a":1},"arr":[{"x":1},{"x":2}],"s":"t","n":null}`))
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}
	// Find returns nil for invalid selector
	if tr.Find("obj.a.b") != nil {
		t.Fatalf("expected nil for too-long path")
	}
	// Remove object key
	tr.Remove("obj.a")
	if v := tr.Find("obj.a"); v != nil {
		t.Fatalf("expected removed key obj.a")
	}
	// Remove array index
	tr.Remove("arr.0")
	if v := tr.Find("arr.0.x"); v == nil || v.AsString() != "2" {
		t.Fatalf("expected first element now x=2, got: %v", v)
	}
	// Remove all array by selector
	tr.Remove("arr.#")
	v := tr.Find("arr.#")
	if v == nil || len(v.ArrayValues) != 0 {
		t.Fatalf("expected empty array after remove, got: %v", v)
	}
	// also exercise filterArrayNodeBySelector/isElemMatched via find
	tr, _ = Decode([]byte(`{"friends":[{"first":"A","age":44},{"first":"B","age":68}]}`))
	ages := tr.Find("friends.#(age>=47).first")
	if ages.AsJSON() != "[\"B\"]" {
		t.Fatalf("filter by selector mismatch: %s", ages.AsJSON())
	}
	// string comparisons
	names := tr.Find("friends.#(first<\"Z\").first")
	if names.AsJSON() != "[\"A\",\"B\"]" {
		t.Fatalf("string < compare mismatch: %s", names.AsJSON())
	}
	names = tr.Find("friends.#(first<=\"B\").first")
	if names.AsJSON() != "[\"A\",\"B\"]" {
		t.Fatalf("string <= compare mismatch: %s", names.AsJSON())
	}
	names = tr.Find("friends.#(first>=\"B\").first")
	if names.AsJSON() != "[\"B\"]" {
		t.Fatalf("string >= compare mismatch: %s", names.AsJSON())
	}
	names = tr.Find("friends.#(first>\"Z\").first")
	if names.AsJSON() != "[]" {
		t.Fatalf("string > compare mismatch: %s", names.AsJSON())
	}
}

func TestMakeStPathAndHelpers(t *testing.T) {
	paths, ok := makeStPath(`friends.#(age>=47).first`)
	if !ok || len(paths) != 3 {
		t.Fatalf("unexpected path parse: ok=%v len=%d", ok, len(paths))
	}
	// ensure reformat works
	if paths[1].Name != "#" || paths[1].Op == "" || paths[1].Selector == "" {
		t.Fatalf("unexpected selector reformat: %#v", paths[1])
	}
	// isInteger/asInteger
	p2, ok := makeStPath(`arr.10`)
	if !ok || !p2[1].isInteger() || p2[1].asInteger() != 10 {
		t.Fatalf("integer step parse failed: %#v", p2)
	}
	// nested selector parsing
	p3, ok := makeStPath(`friends.#(nets.#(=="fb")).first`)
	if !ok || len(p3) != 3 || p3[1].Name != "#" || p3[1].Op != arrayElemEq {
		t.Fatalf("nested selector parse failed: %#v", p3)
	}
}

func TestDiffAPICoverage(t *testing.T) {
	t1, _ := Decode([]byte(`{"a":1,"b":[1,2],"o":{"x":"y"}}`))
	t2, _ := Decode([]byte(`{"a":"1","b":[1,2,3],"o":{"x":"z","m":1}}`))
	items := Diff(t1, t2)
	if !items.Exist() {
		t.Fatalf("expected diffs to exist")
	}
	s := items.String()
	if !strings.Contains(s, "Total") || !strings.Contains(s, "DiffOf") {
		t.Fatalf("unexpected diff string: %s", s)
	}
	if !strings.Contains(items[0].String(), "DiffOf") {
		t.Fatalf("unexpected item string: %s", items[0].String())
	}
}

func TestRemoveByteFindCloseSym(t *testing.T) {
	// removeByte
	s := removeByte("a.b.c", '.')
	if s != "abc" {
		t.Fatalf("removeByte failed: %q", s)
	}
	// findCloseSym nested and escaped
	data := []byte(`#(a#(b\)c))`)
	proj := map[byte]byte{'(': ')', '"': '"'}
	idx := findCloseSym(data, 2, len(data), '(', proj)
	if idx <= 0 {
		t.Fatalf("findCloseSym failed, got %d", idx)
	}
}
func TestDecodeTrailingGarbage(t *testing.T) {
	// After a valid value, extra garbage should cause error
	if _, err := Decode([]byte("1 1")); err == nil {
		t.Fatalf("expected error for trailing garbage")
	}
}
func TestDecodeObjectMissingColon(t *testing.T) {
	if _, err := Decode([]byte("{\"a\" 1}")); err == nil {
		t.Fatalf("expected error for missing colon in object")
	}
}

func TestStringEncoder_getu4(t *testing.T) {
	if r := getu4([]byte("\\u4f60xx")); r != 0x4f60 {
		t.Fatalf("getu4 expected 0x4f60, got %v", r)
	}
	if r := getu4([]byte("abc")); r != -1 {
		t.Fatalf("getu4 expected -1 for invalid prefix, got %v", r)
	}
}

func TestMakeStPathEscapedDot(t *testing.T) {
	// escaped dot path
	p4, ok := makeStPath(`fav\.movie`)
	if !ok || len(p4) != 1 || p4[0].Name != "fav.movie" {
		t.Fatalf("escaped dot parsing failed: %#v", p4)
	}
}

func TestUnquoteBytesValidAndInvalid(t *testing.T) {
	// valid escapes and unicode
	in := []byte("\"hello\\n\\u4f60\"")
	out, ok := unquoteBytes(in)
	if !ok {
		t.Fatalf("unquoteBytes expected ok")
	}
	if !strings.Contains(string(out), "hello\n") || !strings.Contains(string(out), "你") {
		t.Fatalf("unexpected unquote result: %q", string(out))
	}
	// invalid missing closing quote
	_, ok = unquoteBytes([]byte("\"bad"))
	if ok {
		t.Fatalf("unquoteBytes should fail for invalid input")
	}
}

func TestQuoteBytesRoundtrip(t *testing.T) {
	raw := []byte("a\n\\\"b\t")
	q := quoteBytes(raw)
	got, err := stdUnmarshalString(q)
	if err != nil {
		t.Fatalf("roundtrip unmarshal error: %v", err)
	}
	if string(got) != string(raw) {
		t.Fatalf("roundtrip mismatch: %q vs %q", string(raw), string(got))
	}
}

func TestRemoveArrayClear(t *testing.T) {
	tr, err := Decode([]byte(`{"arr":[1,2,3]}`))
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}
	tr.Remove("arr.#")
	n := tr.Find("arr.#")
	if n == nil || len(n.ArrayValues) != 0 {
		t.Fatalf("expected cleared array, got: %v", n)
	}
}

func TestNewCreatesRoot(t *testing.T) {
	tr := New()
	if tr.Root == nil {
		t.Fatalf("New() should create non-nil Root")
	}
}

func TestJSONMarshalWithPanicNil(t *testing.T) {
	if JSONMarshalWithPanic(nil) != nil {
		t.Fatalf("nil input should marshal to nil")
	}
	var p *int
	if JSONMarshalWithPanic(p) != nil {
		t.Fatalf("nil pointer should marshal to nil")
	}
}

func TestStringEncoderASCIIExhaustive(t *testing.T) {
	// cover branches for control chars and safe ASCII
	raw := make([]byte, 128)
	for i := 0; i < 128; i++ {
		raw[i] = byte(i)
	}
	// ensure quotes at ends to make a valid JSON string payload for unquoteBytes
	q := quoteBytes(raw)
	got, err := stdUnmarshalString(q)
	if err != nil {
		t.Fatalf("stdUnmarshalString failed: %v", err)
	}
	if string(got) != string(raw) {
		t.Fatalf("ASCII roundtrip mismatch")
	}
}

func TestStringEncoderUnicodeSpecials(t *testing.T) {
	// include U+2028 and U+2029 to exercise special handling
	s := string([]rune{'a', '\u2028', 'b', '\u2029', 'c'})
	q := quoteBytes([]byte(s))
	got, err := stdUnmarshalString(q)
	if err != nil {
		t.Fatalf("stdUnmarshalString failed: %v", err)
	}
	if string(got) != s {
		t.Fatalf("unicode specials roundtrip mismatch: %q vs %q", s, string(got))
	}
}

func TestUnquoteInvalidHex(t *testing.T) {
	in := []byte{'"', '\\', 'u', 'Z', 'Z', 'Z', 'Z', '"'}
	if _, ok := unquoteBytes(in); ok {
		t.Fatalf("expected unquoteBytes to fail for invalid hex")
	}
}

func TestUnquoteInvalidControl(t *testing.T) {
	in := []byte{'"', 0x01, '"'}
	if _, ok := unquoteBytes(in); ok {
		t.Fatalf("expected unquoteBytes to fail for control char in string")
	}
}

func TestUnquoteInvalidUTF8RuneError(t *testing.T) {
	// invalid utf-8 sequence inside quotes should be coerced to replacement char
	in := []byte{'"', 0xE0, '"'}
	out, ok := unquoteBytes(in)
	if !ok {
		t.Fatalf("expected unquoteBytes to succeed with replacement char")
	}
	if !strings.Contains(string(out), "\ufffd") {
		t.Fatalf("expected replacement char in output, got: %q", string(out))
	}
}

func TestUnquoteSurrogatePairs(t *testing.T) {
	// musical G clef U+1D11E as surrogate pair \uD834\uDD1E
	in := []byte("\"\\uD834\\uDD1E\"")
	out, ok := unquoteBytes(in)
	if !ok {
		t.Fatalf("expected valid surrogate pair")
	}
	if string(out) != "\U0001D11E" {
		t.Fatalf("unexpected surrogate decode: %q", string(out))
	}
	// lone high surrogate -> replacement char
	in = []byte("\"\\uD800\"")
	out, ok = unquoteBytes(in)
	if !ok {
		t.Fatalf("expected ok for lone surrogate (replacement)")
	}
}

func TestUnquoteInvalidEscapeAndShortUnicode(t *testing.T) {
	// invalid escape sequence \x
	if _, ok := unquoteBytes([]byte("\"\\x\"")); ok {
		t.Fatalf("expected invalid escape to fail")
	}
	// short unicode sequence
	if _, ok := unquoteBytes([]byte("\"\\u12\"")); ok {
		t.Fatalf("expected short unicode escape to fail")
	}
	// supported simple escapes (without \/ to avoid Go unknown escape)
	// careful escaping in Go literal: "\b\f\n\r\t\\\"'"
	out, ok := unquoteBytes([]byte("\"\\b\\f\\n\\r\\t\\\\\\\"'\""))
	if !ok {
		t.Fatalf("expected valid simple escapes")
	}
	if string(out) != "\b\f\n\r\t\\\"'" {
		t.Fatalf("unexpected simple escapes decode: %q", string(out))
	}
}

func TestNextValueDetectors(t *testing.T) {
	if !nextValueIsNumber([]byte("-12"), 0) {
		t.Fatalf("-12 should be number")
	}
	if !nextValueIsNumber([]byte("1e+10"), 0) {
		t.Fatalf("1e+10 should be number")
	}
	if !nextValueIsNumber([]byte("1e-10"), 0) {
		t.Fatalf("1e-10 should be number")
	}
	if nextValueIsNumber([]byte("1e"), 0) {
		t.Fatalf("1e should be invalid number")
	}
	if nextValueIsNumber([]byte("1e+"), 0) {
		t.Fatalf("1e+ should be invalid number")
	}
	if !nextValueIsString([]byte("\"a\\\"b\""), 0) {
		t.Fatalf("escaped quote should be valid string")
	}
	if nextValueIsString([]byte("\"a"), 0) {
		t.Fatalf("missing closing quote should be invalid string")
	}
	if !nextValueIsNull([]byte("null"), 0) {
		t.Fatalf("null should be detected")
	}
	if !nextValueIsBool([]byte("true"), 0) || !nextValueIsBool([]byte("false"), 0) {
		t.Fatalf("bool should be detected")
	}
	if nextValueShouldBe([]byte("  :"), 0, ':') == -1 {
		t.Fatalf("should find colon after whitespace")
	}
	if searchFirstValidChar([]byte("  \n\t  a"), 0) != 6 {
		t.Fatalf("searchFirstValidChar index mismatch")
	}
	if searchFirstValidChar([]byte("    \t\n \r"), 0) != -1 {
		t.Fatalf("expected -1 when no valid char")
	}
	if nextValueShouldBe([]byte(" ,"), 0, ':') != -1 {
		t.Fatalf("should not find colon among options")
	}
}

func TestFindFailureUnmatchedSelector(t *testing.T) {
	tr, err := Decode([]byte(`{"a":1}`))
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if tr.Find(`#(age>=`) != nil {
		t.Fatalf("expected nil for unmatched selector")
	}
}

func TestRemoveNoPaths(t *testing.T) {
	tr, _ := Decode([]byte(`{"x":1}`))
	// empty or invalid path should not panic
	tr.Remove("")
	tr.Remove("#(")
}

func TestNodeSetObjectUpdatesAndRemoveArrayIndex(t *testing.T) {
	tr, _ := Decode([]byte(`{"o":{"k":"v","n":1},"arr":[1,2]}`))
	tr.Root.Find("o").SetObjectStringElem("k", "vv")
	tr.Root.Find("o").SetObjectUintElem("n", 10)
	tr.Root.Find("o").SetObjectBoolElem("b", true)
	if tr.Root.Find("o.k").AsString() != "vv" {
		t.Fatalf("expected updated string value")
	}
	if tr.Root.Find("o.n").AsInt() != int64(10) {
		t.Fatalf("expected updated uint value")
	}
	if tr.Root.Find("o.b").AsString() != "true" {
		t.Fatalf("expected added bool value")
	}
	// invalid remove index
	if tr.Root.Find("arr").RemoveArrayElemByIndex(-1) {
		t.Fatalf("expected false for invalid index")
	}
	if tr.Root.Find("arr").RemoveArrayElemByIndex(10) {
		t.Fatalf("expected false for out-of-range index")
	}
}

func TestNodeSettersAndGetters(t *testing.T) {
	n := CreateNode()
	// SetString and AsString
	n.Type = String
	n.SetString("hello")
	if n.AsString() != "hello" {
		t.Fatalf("AsString mismatch")
	}
	// SetBool and AsBool/AsString
	n.Type = Bool
	n.SetBool(true)
	if !n.AsBool() || n.AsString() != "true" {
		t.Fatalf("bool setters/getters mismatch")
	}
	n.SetBool(false)
	if n.AsBool() || n.AsString() != "false" {
		t.Fatalf("bool setters/getters mismatch false")
	}
	// SetInt/SetUint/SetFloat and AsInt/AsUint/AsFloat
	n.Type = Integer
	n.SetInt(-12)
	if n.AsInt() != -12 {
		t.Fatalf("AsInt mismatch")
	}
	n.SetUint(42)
	if n.AsUint() != 42 {
		t.Fatalf("AsUint mismatch")
	}
	n.Type = Float
	n.SetFloat(3.14, 2)
	if n.AsFloat() != 3.14 {
		t.Fatalf("AsFloat mismatch")
	}
}

func TestGetObjectElemByKeyAndAsMap(t *testing.T) {
	tr, _ := Decode([]byte(`{"o":{"k":"v","i":1}}`))
	obj := tr.Root.Find("o")
	kv := obj.GetObjectElemByKey("k")
	if kv == nil || kv.Value.AsString() != "v" {
		t.Fatalf("GetObjectElemByKey failed")
	}
	m := obj.AsMap()
	if m["i"].AsInt() != 1 {
		t.Fatalf("AsMap missing or wrong value")
	}
}

func TestPoolReleaseAndCreate(t *testing.T) {
	// Ensure Release returns node to pool and CreateNode resets fields
	tr, _ := Decode([]byte(`{"a":1,"o":{"k":"v"},"arr":[1,2]}`))
	if tr.Root == nil {
		t.Fatalf("root should not be nil")
	}
	tr.Release()
	if tr.Root != nil {
		t.Fatalf("root should be nil after release")
	}
	n := CreateNode()
	if n.Type != Null || n.Value != emptyVal || n.ObjectValues != nil || n.ArrayValues != nil {
		t.Fatalf("CreateNode should reset fields")
	}
	obj := CreateObjectElem()
	if obj.Key != nil || obj.Value != nil {
		t.Fatalf("CreateObjectElem should reset fields")
	}
}

func TestJSONMarshalWithPanicError(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic on unsupported type")
		}
	}()
	type Bad struct{ C chan int }
	_ = JSONMarshalWithPanic(Bad{C: make(chan int)})
}
