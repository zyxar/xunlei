package protocol

import (
	"testing"
)

func TestUnescapeName(t *testing.T) {
	var a = "&amp;amp;amp;amp;amp;ttt"
	v := unescapeName(a)
	if v != "&ttt" {
		t.Error(v)
	}
	var b = "&lt;amp;"
	v = unescapeName(b)
	if v != "<amp;" {
		t.Error(v)
	}
}
