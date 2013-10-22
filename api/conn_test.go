package api

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var conf struct {
	Id   string `json:"account"`
	Pass string `json:"password"`
}

func init() {
	home := os.Getenv("HOME")
	f, _ := ioutil.ReadFile(filepath.Join(home, ".xltask/config.json"))
	json.Unmarshal(f, &conf)
}

func TestConn(t *testing.T) {
	err := Login(conf.Id, conf.Pass)
	if err != nil {
		t.Fatal(err)
	}
	SaveSession("test_cookie.js")
	err = ResumeSession("test_cookie.js")
	if err != nil {
		t.Fatal(err)
	}
	os.Remove("test_cookie.js")
}

func TestTaskNoFresh(t *testing.T) {
	_, err := tasklist_nofresh(4, 1)
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		for {
			select {
			case <-time.After(time.Second):
				log.Println("tasks in cache:", len(M.Tasks))
			}
		}
	}()
}

func TestProcessTask(t *testing.T) {
	ch := make(chan byte)
	ProcessTaskDaemon(ch, func(t *Task) {
		log.Printf("%s %sB/s %.2f%%\n", t.Id, t.Speed, t.Progress)
	})
	go func() {
		for {
			select {
			case <-time.After(2 * time.Second):
				ch <- 1
			}
		}
	}()
	select {
	case <-time.After(2 * time.Second):
		break
	}
}

func TestGetGid(t *testing.T) {
	gid, err := GetGdriveId()
	if err != nil {
		t.Fatal(err)
	}
	if len(gid) == 0 || gid != M.Gid {
		t.Fatal("Invalid gdriveid")
	}
}

func TestGetCompletedTasks(t *testing.T) {
	ts, err := GetCompletedTasks()
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range ts {
		if v.DownloadStatus != "2" {
			t.Error("Invalid download status")
		}
	}
}

func TestGetIncompletedTasks(t *testing.T) {
	ts, err := GetIncompletedTasks()
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range ts {
		// downloading || failed || pending
		if !v.downloading() && !v.failed() && !v.pending() {
			t.Error("Invalid download status")
		}
	}
}

func TestGetExpiredTasks(t *testing.T) {
	b, err := GetExpiredTasks()
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range b {
		if !v.completed() && v.status() != _FLAG_expired { // expired must be completed
			t.Error("Invalid download status")
		}
	}
}

func TestGetDeletedTasks(t *testing.T) {
	b, err := GetDeletedTasks()
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range b {
		if v.status() != _FLAG_deleted {
			t.Error("Invalid status")
		}
	}
}

func TestAddTask(t *testing.T) {
	err := AddTask("26B092652A3D8263BABDE5D32BEB0F01F6D208F7")
	if err != nil && err != btTaskAlreadyErr {
		t.Error(err)
	}
}

func TestTorrent(t *testing.T) {
	err := GetTorrentFileByHash("26B092652A3D8263BABDE5D32BEB0F01F6D208F7", "test.torrent")
	if err != nil {
		t.Error(err)
	}
	err = addTorrentTask("test.torrent")
	if err != nil && err != btTaskAlreadyErr {
		t.Error(err)
	}
	os.Remove("test.torrent")
}

func TestFillBtListAsync(t *testing.T) {
	for i, _ := range M.Tasks {
		t := M.Tasks[i]
		if t.status() == _FLAG_normal && t.IsBt() {
			FillBtListAsync(t.Id, t.Cid, nil)
		}
	}
	select {
	case <-time.After(5 * time.Second):
		return
	}
}

func TestDispatch(t *testing.T) {
	m, err := DispatchTasks("group=completed")
	if err != nil {
		t.Error(err)
	}
	log.Println("----- group=completed -----")
	for i, _ := range m {
		log.Printf("%v\n", m[i])
	}
	m, err = DispatchTasks("status=deleted")
	if err != nil {
		t.Error(err)
	}
	log.Println("----- status=deleted -----")
	for i, _ := range m {
		log.Printf("%v\n", m[i])
	}
	m, err = DispatchTasks("type=bt")
	if err != nil {
		t.Error(err)
	}
	log.Println("----- type=bt -----")
	for i, _ := range m {
		log.Printf("%v\n", m[i])
	}
	m, err = DispatchTasks("name=monsters")
	if err != nil {
		t.Error(err)
	}
	log.Println("----- name=monsters -----")
	for i, _ := range m {
		log.Printf("%v\n", m[i])
	}
	m, err = DispatchTasks("name=monsters&group=completed&status=deleted&type=bt")
	if err != nil {
		t.Error(err)
	}
	log.Println("----- name=monsters&group=completed&status=deleted&type=bt -----")
	for i, _ := range m {
		log.Printf("%v\n", m[i])
	}
}
