package main

import (
	"fmt"
	"testing"

	. "github.com/matzoe/xunlei/api"
)

func TestCall(t *testing.T) {
	r, err := Call("GetTasks", nil)
	if err != nil {
		t.Error(err)
	}
	if v, ok := r[0].Interface().([]*Task); ok {
		fmt.Println(len(v), "task(s).")
	}
}
