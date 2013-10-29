package main

import (
	"fmt"
	"testing"

	. "github.com/matzoe/xunlei/api"
)

func TestAddUri(t *testing.T) {
	ts, _ := _find("status=normal&group=completed&type=nbt")
	var v *Task
	for i, _ := range ts {
		v = ts[i]
		break
	}
	if v != nil {
		gid, err := RPCAddTask(v.LixianURL, v.TaskName)
		if err != nil {
			t.Error(err)
		}
		fmt.Printf("%s\n", gid)
	}
}
