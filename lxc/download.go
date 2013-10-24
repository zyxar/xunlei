package main

import (
	"errors"
	"log"
	"path/filepath"
	"regexp"

	"github.com/matzoe/xunlei/api"
	"github.com/matzoe/xunlei/fetch"
)

var worker fetch.Fetcher = fetch.DefaultFetcher

func dl(uri, filename string, echo bool) error { //TODO: check file existence
	if len(api.M.Gid) == 0 {
		return errors.New("gdriveid missing.")
	}
	return worker.Fetch(uri, api.M.Gid, filename, echo)
}

func download(t *api.Task, filter string, echo, verify bool) error {
	if t.IsBt() {
		m, err := t.FillBtList()
		if err != nil {
			return err
		}
		for j, _ := range m.Record {
			if m.Record[j].Status == "2" {
				if ok, _ := regexp.MatchString(`(?i)`+filter, m.Record[j].FileName); ok {
					log.Println("Downloading", m.Record[j].FileName, "...")
					if err = dl(m.Record[j].DownURL, filepath.Join(t.TaskName, m.Record[j].FileName), echo); err != nil {
						return err
					}
				} else {
					log.Printf("Skip unselected task %s", m.Record[j].FileName)
				}
			} else {
				log.Printf("Skip incompleted task %s", m.Record[j].FileName)
			}
		}
	} else {
		if len(t.LixianURL) == 0 {
			return errors.New("Target file not ready for downloading.")
		}
		log.Println("Downloading", t.TaskName, "...")
		if err := dl(t.LixianURL, t.TaskName, echo); err != nil {
			return err
		}
	}
	if verify && !t.Verify(t.TaskName) {
		return errors.New("Verification failed.")
	}
	return nil
}
