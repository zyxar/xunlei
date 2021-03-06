package main

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/zyxar/xunlei/fetch"
	"github.com/zyxar/xunlei/protocol"
)

var worker = fetch.DefaultFetcher

type taskSink func(uri, filename string, echo bool) error

func dl(uri, filename string, echo bool) error { //TODO: check file existence
	gid, err := protocol.GetGdriveId()
	if err != nil {
		return err
	}
	if len(gid) == 0 {
		return errors.New("gdriveid missing")
	}
	return worker.Fetch(uri, gid, filename, echo)
}

func download(t *protocol.Task, filter string, echo, verify bool, sink taskSink) error {
	if t.IsBt() {
		m, err := protocol.FillBtList(t)
		if err != nil {
			return err
		}
		var fullpath string
		for j := range m.Record {
			if m.Record[j].Status == "2" {
				if ok, _ := regexp.MatchString(`(?i)`+filter, m.Record[j].FileName); ok {
					fmt.Println("Downloading", m.Record[j].FileName, "...")
					if len(m.Record) == 1 { // choose not to use torrent info, to reduce network transportation
						fullpath = m.Record[j].FileName
					} else {
						fullpath = filepath.Join(t.TaskName, m.Record[j].FileName)
					}
					if err = sink(m.Record[j].DownURL, fullpath, echo); err != nil {
						return err
					}
				} else {
					fmt.Printf("Skip unselected task %s\n", m.Record[j].FileName)
				}
			} else {
				fmt.Printf("Skip incompleted task %s\n", m.Record[j].FileName)
			}
		}
	} else {
		if len(t.LixianURL) == 0 {
			return errors.New("Target file not ready for downloading.")
		}
		fmt.Println("Downloading", t.TaskName, "...")
		if err := sink(t.LixianURL, t.TaskName, echo); err != nil {
			return err
		}
	}
	if verify && !protocol.VerifyTask(t, t.TaskName) {
		return errors.New("Verification failed.")
	}
	return nil
}
