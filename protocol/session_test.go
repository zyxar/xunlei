package protocol

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var conf struct {
	Id   string `json:"account"`
	Pass string `json:"password"`
}

var testSession = newSession(3000 * time.Millisecond)

func init() {
	home := os.Getenv("HOME")
	f, _ := ioutil.ReadFile(filepath.Join(home, ".xltask/config.json"))
	json.Unmarshal(f, &conf)
}

func TestConn(t *testing.T) {
	err := testSession.Login(conf.Id, conf.Pass)
	if err != nil {
		t.Fatal(err)
	}
	testSession.SaveSession("test_cookie.js")
	err = testSession.ResumeSession("test_cookie.js")
	if err != nil {
		t.Fatal(err)
	}
	os.Remove("test_cookie.js")
}

func TestTaskNoFresh(t *testing.T) {
	_, err := testSession.tasklistNofresh(4, 1)
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		for {
			select {
			case <-time.After(time.Second):
				fmt.Println("tasks in cache:", len(testSession.cache.Tasks))
			}
		}
	}()
}

func TestProcessTask(t *testing.T) {
	ch := make(chan byte)
	testSession.ProcessTaskDaemon(ch, func(t *Task) error {
		fmt.Printf("%s %sB/s %.2f%%\n", t.Id, t.Speed, t.Progress)
		return nil
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
	gid, err := testSession.GetGdriveId()
	if err != nil {
		t.Fatal(err)
	}
	if len(gid) == 0 || gid != testSession.gid {
		t.Fatal("Invalid gdriveid")
	}
}

func TestGetCompletedTasks(t *testing.T) {
	ts, err := testSession.GetCompletedTasks()
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
	ts, err := testSession.GetIncompletedTasks()
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
	b, err := testSession.GetExpiredTasks()
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range b {
		if !v.completed() && v.status() != flagExpired { // expired must be completed
			t.Error("Invalid download status")
		}
	}
}

func TestGetDeletedTasks(t *testing.T) {
	b, err := testSession.GetDeletedTasks()
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range b {
		if v.status() != flagDeleted {
			t.Error("Invalid status")
		}
	}
}

func TestFind(t *testing.T) {
	m, err := testSession.FindTasks("group=completed")
	if err != nil {
		t.Error(err)
	}
	fmt.Println("----- group=completed -----")
	for i := range m {
		fmt.Printf("%v\n", m[i])
	}
	m, err = testSession.FindTasks("status=deleted")
	if err != nil {
		t.Error(err)
	}
	fmt.Println("----- status=deleted -----")
	for i := range m {
		fmt.Printf("%v\n", m[i])
	}
	m, err = testSession.FindTasks("type=bt")
	if err != nil {
		t.Error(err)
	}
	fmt.Println("----- type=bt -----")
	for i := range m {
		fmt.Printf("%v\n", m[i])
	}
	m, err = testSession.FindTasks("name=monsters")
	if err != nil {
		t.Error(err)
	}
	fmt.Println("----- name=monsters -----")
	for i := range m {
		fmt.Printf("%v\n", m[i])
	}
	m, err = testSession.FindTasks("name=monsters&group=completed&status=deleted&type=bt")
	if err != nil {
		t.Error(err)
	}
	fmt.Println("----- name=monsters&group=completed&status=deleted&type=bt -----")
	for i := range m {
		fmt.Printf("%v\n", m[i])
	}
}
