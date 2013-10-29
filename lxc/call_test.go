package main

import (
	"fmt"
	"os"
	"testing"

	. "github.com/matzoe/xunlei/api"
)

func init() {
	initConf()
	if err := ResumeSession(cookie_file); err != nil {
		fmt.Println(err)
		if err = Login(conf.Id, conf.Pass); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if err = SaveSession(cookie_file); err != nil {
			fmt.Println(err)
		}
	}
	GetGdriveId()
}

func TestCall(t *testing.T) {
	r, err := Call("GetTasks", nil)
	if err != nil {
		t.Error(err)
	}
	if v, ok := r[0].Interface().([]*Task); ok {
		fmt.Println(len(v), "task(s).")
	}
}
