package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"reflect"
	"strconv"

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
	go GetTasks()
	web.Get("/gettasks/(.*)", func(ctx *web.Context, val string) {
		flusher, _ := ctx.ResponseWriter.(http.Flusher)
		flusher.Flush()
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
			ctx.NotFound("Invalid Task Group")
			return
		}
		if err != nil {
			ctx.Abort(503, err.Error())
			return
		}
		r, err := json.Marshal(v)
		if err != nil {
			ctx.Abort(503, err.Error())
			return
		}
		ctx.Write(r)
		flusher.Flush()
	})
	web.Get("/self", func(ctx *web.Context) {
		flusher, _ := ctx.ResponseWriter.(http.Flusher)
		flusher.Flush()
		if M.Account == nil {
			ctx.NotFound("Account information not retrieved")
			return
		}
		r, err := json.Marshal(M.Account)
		if err != nil {
			ctx.Abort(503, err.Error())
			return
		}
		ctx.Write(r)
		flusher.Flush()
	})
	web.Get("/raw/tasklist/(.*)", func(ctx *web.Context, val string) {
		flusher, _ := ctx.ResponseWriter.(http.Flusher)
		flusher.Flush()
		page, err := strconv.Atoi(val)
		if err != nil {
			ctx.Abort(503, "Invalid page number")
			return
		}
		b, err := RawTaskList(4, page)
		if err != nil {
			ctx.Abort(503, err.Error())
			return
		}
		ctx.Write(b)
		flusher.Flush()
	})
	web.Get("/raw/task/(.*)", func(ctx *web.Context, val string) {
		flusher, _ := ctx.ResponseWriter.(http.Flusher)
		flusher.Flush()
		if t, ok := M.Tasks[val]; ok {
			if t.IsBt() {
				m, err := RawFillBtList(t.Id, t.Cid, 1)
				if err != nil {
					ctx.Abort(503, err.Error())
					flusher.Flush()
					return
				}
				ctx.Write(m)
				flusher.Flush()
				return
			}
			ctx.WriteString(t.Repr())
			flusher.Flush()
			return
		}
		ctx.NotFound("Task not found")
	})
	web.Run("127.0.0.1:8808")
}
