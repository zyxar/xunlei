package main

import (
	"bytes"
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

type payload struct {
	Signature string
	Action    string
	Data      interface{}
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

	// GET - ls, info
	// ctx.SetHeader("Access-Control-Allow-Origin", "*", true)
	web.Get("/tasks/([0-4]|l[cdeis])", func(ctx *web.Context, val string) {
		flusher, _ := ctx.ResponseWriter.(http.Flusher)
		defer flusher.Flush()
		var v []*Task
		var err error
		switch val {
		case "0", "ls":
			v, err = GetTasks()
		case "1", "li":
			v, err = GetIncompletedTasks()
		case "2", "lc":
			v, err = GetCompletedTasks()
		case "3", "ld":
			v, err = GetDeletedTasks()
		case "4", "le":
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
		ctx.SetHeader("Content-Type", "application/json", true)
		ctx.Write(r)
	})
	web.Get("/session", func(ctx *web.Context) {
		flusher, _ := ctx.ResponseWriter.(http.Flusher)
		defer flusher.Flush()
		if M.Account == nil {
			ctx.NotFound("Account information not retrieved")
			return
		}
		r, err := json.Marshal(M.Account)
		if err != nil {
			ctx.Abort(503, err.Error())
			return
		}
		ctx.SetHeader("Content-Type", "application/json", true)
		ctx.Write(r)
	})
	web.Get("/tasks/raw/([0-9]+)", func(ctx *web.Context, val string) {
		flusher, _ := ctx.ResponseWriter.(http.Flusher)
		defer flusher.Flush()
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
		ctx.SetHeader("Content-Type", "application/json", true)
		ctx.Write(b)
	})
	web.Get("/task/bt/([0-9]+)/(.*)", func(ctx *web.Context, taskId string, taskHash string) {
		flusher, _ := ctx.ResponseWriter.(http.Flusher)
		defer flusher.Flush()
		m, err := RawFillBtList(taskId, taskHash, 1)
		if err != nil {
			ctx.Abort(503, err.Error())
			return
		}
		if bytes.HasPrefix(m, []byte("fill_bt_list")) {
			ctx.SetHeader("Content-Type", "application/javascript", true)
		} else {
			ctx.SetHeader("Content-Type", "text/plain", true)
		}
		ctx.Write(m)
	})
	// POST - relogin, saveconf, loadconf, savesession
	web.Post("/session", func(ctx *web.Context) {})
	// POST - add, readd
	web.Post("/task", func(ctx *web.Context) {

	})
	// PUT - delay(All), pause, resume, rename, dl, dt, ti
	web.Put("/task/(.*)", func(ctx *web.Context, val string) {})
	// DELETE - rm, purge, GOODBYE
	web.Delete("/task/(.*)", func(ctx *web.Context, val string) {})
	web.Delete("/session/(.*)", func(ctx *web.Context, val string) {})
	web.Run("127.0.0.1:8808")
}
