package main

import (
	"encoding/json"
	"errors"
	"log"
	"reflect"

	"github.com/hoisie/web"
	. "github.com/matzoe/xunlei/api"
)

var register = map[string]interface{}{
	"AddBatchTasks":        AddBatchTasks,
	"AddTask":              AddTask,
	"DelayTask":            DelayTask,
	"DeleteTask":           DeleteTask,
	"DeleteTasks":          DeleteTasks,
	"FillBtList":           FillBtList,
	"GetCompletedTasks":    GetCompletedTasks,
	"GetDeletedTasks":      GetDeletedTasks,
	"GetExpiredTasks":      GetExpiredTasks,
	"GetGdriveId":          GetGdriveId,
	"GetIncompletedTasks":  GetIncompletedTasks,
	"GetTasks":             GetTasks,
	"GetTorrentByHash":     GetTorrentByHash,
	"GetTorrentFileByHash": GetTorrentFileByHash,
	"PauseTasks":           PauseTasks,
	"ProcessTask":          ProcessTask,
	"PurgeTask":            PurgeTask,
	"DelayAllTasks":        DelayAllTasks,
	"RenameTask":           RenameTask,
	"ResumeTasks":          ResumeTasks,
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
}

func daemonLoop() {
	ch := make(chan byte)
	ProcessTaskDaemon(ch, func(t *Task) {
		log.Printf("%s %s %sB/s %.2f%%\n", t.Id, fixedLengthName(t.TaskName, 32), t.Speed, t.Progress)
	})
	web.Get("/gettasks/(.*)", func(ctx *web.Context, val string) string {
		var v []*Task
		var err error
		switch val {
		case "0":
			v, err = GetTasks()
		case "1":
			v, err = GetIncompletedTasks()
		case "2":
			v, err = GetCompletedTasks()
		case "3":
			v, err = GetDeletedTasks()
		case "4":
			v, err = GetExpiredTasks()
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
	web.Get("/self", func(ctx *web.Context) string {
		if M.Account == nil {
			return "{}"
		}
		v, err := json.Marshal(M.Account)
		if err != nil {
			return err.Error()
		}
		return string(v)
	})
	web.Run("127.0.0.1:8808")
}
