package main

import (
	"errors"
	"path/filepath"
	"regexp"

	"github.com/golang/glog"
	. "github.com/zyxar/xunlei/fetch"
	. "github.com/zyxar/xunlei/protocol"
)

var worker Fetcher = DefaultFetcher

type taskSink func(uri, filename string, echo bool) error

func dl(uri, filename string, echo bool) error { //TODO: check file existence
	if len(M.Gid) == 0 {
		return errors.New("gdriveid missing.")
	}
	return worker.Fetch(uri, M.Gid, filename, echo)
}

func download(t *Task, filter string, echo, verify bool, sink taskSink) error {
	if t.IsBt() {
		m, err := t.FillBtList()
		if err != nil {
			return err
		}
		var fullpath string
		for j := range m.Record {
			if m.Record[j].Status == "2" {
				if ok, _ := regexp.MatchString(`(?i)`+filter, m.Record[j].FileName); ok {
					glog.V(2).Infoln("Downloading", m.Record[j].FileName, "...")
					if len(m.Record) == 1 { // choose not to use torrent info, to reduce network transportation
						fullpath = m.Record[j].FileName
					} else {
						fullpath = filepath.Join(t.TaskName, m.Record[j].FileName)
					}
					if err = sink(m.Record[j].DownURL, fullpath, echo); err != nil {
						return err
					}
				} else {
					glog.V(3).Infof("Skip unselected task %s", m.Record[j].FileName)
				}
			} else {
				glog.V(2).Infof("Skip incompleted task %s", m.Record[j].FileName)
			}
		}
	} else {
		if len(t.LixianURL) == 0 {
			return errors.New("Target file not ready for downloading.")
		}
		glog.V(2).Infoln("Downloading", t.TaskName, "...")
		if err := sink(t.LixianURL, t.TaskName, echo); err != nil {
			return err
		}
	}
	if verify && !t.Verify(t.TaskName) {
		return errors.New("Verification failed.")
	}
	return nil
}
