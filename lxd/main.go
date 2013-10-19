package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"

	"github.com/hoisie/web"
	"github.com/matzoe/xunlei/api"
)

var conf struct {
	Id   string `json:"account"`
	Pass string `json:"password"`
}

var register = map[string]interface{}{
	"AddBatchTasks":        api.AddBatchTasks,
	"AddTask":              api.AddTask,
	"DelayTask":            api.DelayTask,
	"DeleteTask":           api.DeleteTask,
	"DeleteTasks":          api.DeleteTasks,
	"FillBtList":           api.FillBtList,
	"GetCompletedTasks":    api.GetCompletedTasks,
	"GetDeletedTasks":      api.GetDeletedTasks,
	"GetExpiredTasks":      api.GetExpiredTasks,
	"GetGdriveId":          api.GetGdriveId,
	"GetIncompletedTasks":  api.GetIncompletedTasks,
	"GetTasks":             api.GetTasks,
	"GetTorrentByHash":     api.GetTorrentByHash,
	"GetTorrentFileByHash": api.GetTorrentFileByHash,
	"PauseTasks":           api.PauseTasks,
	"ProcessTask":          api.ProcessTask,
	"PurgeTask":            api.PurgeTask,
	"ReAddAllExpiredTasks": api.ReAddAllExpiredTasks,
	"RenameTask":           api.RenameTask,
	"RestartTasks":         api.RestartTasks,
}

func Call(name string, params []string) (result []reflect.Value, err error) {
	fn := reflect.ValueOf(register[name])
	if len(params) != fn.Type().NumIn() {
		err = errors.New("Incompatible parameters")
		return
	}
	args := make([]reflect.Value, len(params))
	for k, param := range params {
		args[k] = reflect.ValueOf(param)
	}
	result = fn.Call(args)
	return
}

func init() {
	home := os.Getenv("HOME")
	cf := filepath.Join(home, ".xltask/cookie.json")
	if err := api.ResumeSession(cf); err != nil {
		log.Println(err)
		f, _ := ioutil.ReadFile(filepath.Join(home, ".xltask/config.json"))
		json.Unmarshal(f, &conf)
		err := api.Login(conf.Id, conf.Pass)
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
		if err = api.SaveSession(cf); err != nil {
			log.Println(err)
		}
	}
	api.GetGdriveId()
	api.GetTasks()
	ch := make(chan byte)
	api.ProcessTask(ch, func(t *api.Task) {
		log.Printf("%s %sB/s %.2f%%\n", t.Id, t.Speed, t.Progress)
	})
}

func main() {
	web.Get("/gettasks/(.*)", func(ctx *web.Context, val string) string {
		var v []*api.Task
		var err error
		switch val {
		case "0":
			v, err = api.GetTasks()
		case "1":
			v, err = api.GetIncompletedTasks()
		case "2":
			v, err = api.GetCompletedTasks()
		case "3":
			v, err = api.GetDeletedTasks()
		case "4":
			v, err = api.GetExpiredTasks()
		default:
			return "Invalid Task Group"
		}
		if err != nil {
			return err.Error()
		}
		r, err := json.Marshal(v)
		if err != nil {
			return err.Error()
		}
		return string(r)
	})
	web.Run("127.0.0.1:8808")
}
