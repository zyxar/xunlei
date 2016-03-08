package protocol

import (
	"os"
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

func TestGetVerifyURL(t *testing.T) {
	for i := 0; i < 100; i++ {
		url1 := verifyBaseURLs[i%len(verifyBaseURLs)]
		url2 := getVerifyURL()
		if url1 != url2 {
			t.Errorf("URL not match: expected %s, got %s", url1, url2)
		}
	}
}

func TestGetVerifyImage(t *testing.T) {
	for i := 0; i < 10; i++ {
		w, err := defaultSession.(*session).getVerifyImage()
		if err != nil {
			t.Error(err)
		}
		w.WriteTo(os.Stdout)
	}
}
