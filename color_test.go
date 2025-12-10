package qjson

import (
	"strings"
	"testing"
)

func TestColorfulMarshalANSI(t *testing.T) {
	tree, err := Decode([]byte(`{"a":"x"}`))
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}
	out := string(tree.ColorfulMarshal())
	// root key at depth=1 -> bold yellow (1;33)
	if !strings.Contains(out, "\x1b[1;33m\"a\"\x1b[0m") {
		t.Fatalf("missing yellow key color in output: %q", out)
	}
	// value at depth=2 -> bold cyan (1;36)
	if !strings.Contains(out, "\x1b[1;36m\"x\"\x1b[0m") {
		t.Fatalf("missing cyan value color in output: %q", out)
	}
}

func TestColorfulMarshalDepthRotation(t *testing.T) {
	tree, err := Decode([]byte(`{"k":{"kk":{"kkk":"v"}}}`))
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}
	out := string(tree.ColorfulMarshalWithIndent())
	// keys/values colored by depth: 1->yellow, 2->cyan, 3->green, 4->magenta
	expected := []string{
		"\x1b[1;33m\"k\"\x1b[0m",
		"\x1b[1;36m\"kk\"\x1b[0m",
		"\x1b[1;32m\"kkk\"\x1b[0m",
		"\x1b[1;35m\"v\"\x1b[0m",
	}
	for _, seg := range expected {
		if !strings.Contains(out, seg) {
			t.Fatalf("missing color segment %q in output: %q", seg, out)
		}
	}
}

func TestColorResetPresent(t *testing.T) {
	tree, err := Decode([]byte(`{"a":"x","b":1,"c":true,"d":null}`))
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}
	out := string(tree.ColorfulMarshal())
	// ensure ANSI sequences exist and reset appears multiple times
	if !strings.Contains(out, "\x1b[") {
		t.Fatalf("no ANSI sequences found in output: %q", out)
	}
	if cnt := strings.Count(out, "\x1b[0m"); cnt < 5 {
		t.Fatalf("expected at least 5 resets, got %d, output: %q", cnt, out)
	}
}
