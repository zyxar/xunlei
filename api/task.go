package api

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strings"
)

func (this Task) expired() bool {
	// this.Flag == "4"
	return this.status() == _FLAG_expired
}

func (this Task) deleted() bool {
	return this.status() == _FLAG_deleted
}

func (this Task) status() byte {
	switch len(this.Flag) {
	case 0:
		return _FLAG_normal
	case 1:
		t := this.Flag[0] - '0'
		if t < 5 {
			return t
		}
	}
	return _FLAG_invalid
}

func (this *Task) update(t *_ptask_record) {
	if this.Id != t.Id {
		return
	}
	this.Speed = t.Speed
	this.Progress = t.Progress
	this.DownloadStatus = t.DownloadStatus
	this.LixianURL = t.LixianURL
}

func (this *Task) Remove() error {
	tids := this.Id + ","
	var del_type byte = this.status()
	if del_type == _FLAG_invalid || del_type == _FLAG_purged {
		return errors.New("Invalid flag in task.")
	}
	uri := fmt.Sprintf(TASKDELETE_URL, current_timestamp(), del_type, current_timestamp())
	data := url.Values{}
	data.Add("taskids", tids)
	data.Add("databases", "0,")
	data.Add("interfrom", "task")
	r, err := post(uri, data.Encode())
	if err != nil {
		return err
	}
	if ok, _ := regexp.Match(`\{"result":1,"type":`, r); ok {
		log.Printf("%s\n", r)
		if this.status() == _FLAG_deleted {
			this.Flag = "2"
		} else {
			this.Flag = "1"
		}
		return nil
	}
	return unexpectedErr
}

func (this *Task) Rename(name string) error {
	return rename_task(this.Id, name, this.TaskType)
}

func (this *Task) Pause() error {
	tids := this.Id + ","
	uri := fmt.Sprintf(TASKPAUSE_URL, tids, M.Uid, current_timestamp())
	r, err := get(uri)
	if err != nil {
		return err
	}
	if bytes.Compare(r, []byte("pause_task_resp()")) != 0 {
		return invalidResponseErr
	}
	return nil
}

func (this *Task) Readd() error {
	return nil
}

func (this *Task) Restart() error {
	if this.expired() {
		return taskNoRedownCapErr
	}
	status := this.DownloadStatus
	if status != "5" && status != "3" {
		return taskNoRedownCapErr // only valid for `pending` and `failed` tasks
	}
	form := make([]string, 0, 3)
	v := url.Values{}
	v.Add("id[]", this.Id)
	v.Add("url[]", this.URL)
	v.Add("cid[]", this.Cid)
	v.Add("download_status[]", status)
	v.Add("taskname[]", this.TaskName)
	form = append(form, v.Encode())
	form = append(form, "type=1")
	form = append(form, "interfrom=task")
	uri := fmt.Sprintf(REDOWNLOAD_URL, current_timestamp())
	r, err := post(uri, strings.Join(form, "&"))
	if err != nil {
		return err
	}
	log.Printf("%s\n", r)
	return nil
}

func (this *Task) Delay() error {
	return DelayTask(this.Id)
}

func (this *Task) Purge() error {
	err := this.Remove()
	if err != nil {
		return err
	}
	return this.Remove()
}
