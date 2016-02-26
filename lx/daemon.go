package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"time"

	"github.com/conclave/web" // fork from github.com/hoisie/web
	"github.com/zyxar/xunlei/protocol"
)

var register = map[string]interface{}{
	"AddBatchTasks":        protocol.AddBatchTasks,
	"AddTask":              protocol.AddTask,
	"DelayTask":            protocol.DelayTask,
	"DeleteTask":           protocol.DeleteTask,
	"DeleteTasks":          protocol.DeleteTasks,
	"FillBtList":           protocol.FillBtList,
	"GetCompletedTasks":    protocol.GetCompletedTasks,
	"GetDeletedTasks":      protocol.GetDeletedTasks,
	"GetExpiredTasks":      protocol.GetExpiredTasks,
	"GetGdriveId":          protocol.GetGdriveId,
	"GetIncompletedTasks":  protocol.GetIncompletedTasks,
	"GetTasks":             protocol.GetTasks,
	"GetTorrentByHash":     protocol.GetTorrentByHash,
	"GetTorrentFileByHash": protocol.GetTorrentFileByHash,
	"PauseTasks":           protocol.PauseTasks,
	"ProcessTask":          protocol.ProcessTask,
	"PurgeTask":            protocol.PurgeTask,
	"DelayAllTasks":        protocol.DelayAllTasks,
	"RenameTask":           protocol.RenameTask,
	"ResumeTasks":          protocol.ResumeTasks,
}

type payload struct {
	Signature string
	Action    string
	Data      interface{}
}

type response struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
}

func (r *response) json() []byte {
	if b, err := json.MarshalIndent(r, "", "  "); err == nil {
		return b
	}
	return nil
}

func makeResponse(e bool, msg string) []byte {
	return (&response{e, msg}).json()
}

func errorMsg(msg string) []byte {
	return makeResponse(true, msg)
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

func verify(sig string) bool {
	return true
}

func unpack(ctx *web.Context, action func(*payload)) {
	flusher, _ := ctx.ResponseWriter.(http.Flusher)
	defer flusher.Flush()
	var v payload
	if body, err := ioutil.ReadAll(ctx.Request.Body); err == nil {
		defer ctx.Request.Body.Close()
		if err := json.Unmarshal(body, &v); err != nil {
			ctx.WriteHeader(400)
			ctx.Write(errorMsg("INSUFFICIENT PAYLOAD"))
			return
		}
		if !verify(v.Signature) {
			ctx.WriteHeader(403)
			ctx.Write(errorMsg("INVALID SIGNATURE"))
			return
		}
		action(&v)
		return
	} else {
		ctx.WriteHeader(400)
		ctx.Write(errorMsg(err.Error()))
	}
}

func daemonLoop() {
	ch := make(chan byte)
	protocol.ProcessTaskDaemon(ch, func(t *protocol.Task) {
	})
	go protocol.GetTasks()

	// GET - ls, info
	// ctx.SetHeader("Access-Control-Allow-Origin", "*", true)
	web.Get("/task/([0-4]|l[cdeis])", func(ctx *web.Context, val string) {
		flusher, _ := ctx.ResponseWriter.(http.Flusher)
		defer flusher.Flush()
		var v []*protocol.Task
		var err error
		switch val {
		case "0", "ls":
			v, err = protocol.GetTasks()
		case "1", "li":
			v, err = protocol.GetIncompletedTasks()
		case "2", "lc":
			v, err = protocol.GetCompletedTasks()
		case "3", "ld":
			v, err = protocol.GetDeletedTasks()
		case "4", "le":
			v, err = protocol.GetExpiredTasks()
		default:
			ctx.WriteHeader(404)
			ctx.Write(errorMsg("INVALID TASK GROUP"))
			return
		}
		if err != nil {
			ctx.WriteHeader(503)
			ctx.Write(errorMsg(err.Error()))
			return
		}
		r, err := json.Marshal(v)
		if err != nil {
			ctx.WriteHeader(503)
			ctx.Write(errorMsg(err.Error()))
			return
		}
		ctx.SetHeader("Content-Type", "application/json", true)
		ctx.Write(r)
	})
	web.Get("/session", func(ctx *web.Context) {
		flusher, _ := ctx.ResponseWriter.(http.Flusher)
		defer flusher.Flush()
		if protocol.M.Account == nil {
			ctx.WriteHeader(404)
			ctx.Write(errorMsg("ACCOUNT INFORMATION NOT RETRIEVED"))
			return
		}
		r, err := json.Marshal(protocol.M.Account)
		if err != nil {
			ctx.WriteHeader(503)
			ctx.Write(errorMsg(err.Error()))
			return
		}
		ctx.SetHeader("Content-Type", "application/json", true)
		ctx.Write(r)
	})
	web.Get("/task/raw/([0-9]+)", func(ctx *web.Context, val string) {
		flusher, _ := ctx.ResponseWriter.(http.Flusher)
		defer flusher.Flush()
		page, err := strconv.Atoi(val)
		if err != nil {
			ctx.WriteHeader(503)
			ctx.Write(errorMsg("INVALID PAGE NUMBER"))
			return
		}
		b, err := protocol.RawTaskList(4, page)
		if err != nil {
			ctx.WriteHeader(503)
			ctx.Write(errorMsg(err.Error()))
			return
		}
		ctx.SetHeader("Content-Type", "application/json", true)
		ctx.Write(b)
	})
	web.Get("/task/bt/([0-9]+)/(.*)", func(ctx *web.Context, taskId string, taskHash string) {
		flusher, _ := ctx.ResponseWriter.(http.Flusher)
		defer flusher.Flush()
		m, err := protocol.RawFillBtList(taskId, taskHash, 1)
		if err != nil {
			ctx.WriteHeader(503)
			ctx.Write(errorMsg(err.Error()))
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
	web.Post("/session", func(ctx *web.Context) {
		unpack(ctx, func(v *payload) {
			var err error
			switch v.Action {
			case "relogin":
				if !protocol.IsOn() {
					if err = protocol.Login(conf.Id, conf.Pass); err != nil {
						ctx.WriteHeader(400)
						ctx.Write(errorMsg(err.Error()))
					} else if err = protocol.SaveSession(cookieFile); err != nil {
						ctx.Write(errorMsg(err.Error()))
					} else {
						ctx.Write(makeResponse(false, "SESSION STATUS OK"))
					}
				} else {
					ctx.Write(makeResponse(false, "SESSION STATUS OK"))
				}
			case "saveconf", "save_conf":
				conf.Pass = protocol.EncryptPass(conf.Pass)
				if _, err := conf.save(configFileName); err != nil {
					ctx.WriteHeader(400)
					ctx.Write(errorMsg(err.Error()))
					return
				}
				ctx.Write(makeResponse(false, "CONFIGURATION SAVED"))
			case "loadconf", "load_conf":
				if _, err = conf.load(configFileName); err != nil {
					ctx.WriteHeader(400)
					ctx.Write(errorMsg(err.Error()))
					return
				}
				ctx.Write(makeResponse(false, "CONFIGURATION RELOADED"))
			case "savesession", "save_session":
				if err := protocol.SaveSession(cookieFile); err != nil {
					ctx.WriteHeader(400)
					ctx.Write(errorMsg(err.Error()))
					return
				}
				ctx.Write(makeResponse(false, "SESSION COOKIE SAVED"))
			default:
				ctx.WriteHeader(501)
				ctx.Write(errorMsg("NOT IMPLEMENTED"))
			}
		})
	})
	// POST - add, readd
	web.Post("/task", func(ctx *web.Context) {
		unpack(ctx, func(v *payload) {
			var err error
			switch v.Action {
			case "add":
				switch url := v.Data.(type) {
				case string:
					err = protocol.AddTask(url)
				case []string:
					for i := range url {
						if err = protocol.AddTask(url[i]); err != nil {
							break
						}
					}
				default:
					err = fmt.Errorf("INVALID PAYLOAD DATA")
				}
				if err != nil {
					ctx.WriteHeader(400)
					ctx.Write(errorMsg(err.Error()))
					return
				}
				ctx.Write(makeResponse(false, "OK"))
			case "readd":
				ctx.WriteHeader(501)
				ctx.Write(errorMsg("NOT IMPLEMENTED"))
			default:
				ctx.WriteHeader(501)
				ctx.Write(errorMsg("NOT IMPLEMENTED"))
			}
		})
	})
	// PUT - delay(All), pause, resume, rename, dl, dt, ti
	web.Put("/task", func(ctx *web.Context) {
		unpack(ctx, func(v *payload) {
			var err error
			switch v.Action {
			case "delayAll":
				err = protocol.DelayAllTasks()
			case "pause":
				ctx.WriteHeader(501)
				ctx.Write(errorMsg("NOT IMPLEMENTED"))
			case "resume":
				ctx.WriteHeader(501)
				ctx.Write(errorMsg("NOT IMPLEMENTED"))
			case "rename":
				ctx.WriteHeader(501)
				ctx.Write(errorMsg("NOT IMPLEMENTED"))
			case "delay":
				ctx.WriteHeader(501)
				ctx.Write(errorMsg("NOT IMPLEMENTED"))
			case "dl", "download":
				ctx.WriteHeader(501)
				ctx.Write(errorMsg("NOT IMPLEMENTED"))
			case "dt", "download_torent":
				ctx.WriteHeader(501)
				ctx.Write(errorMsg("NOT IMPLEMENTED"))
			case "ti", "torrent_info":
				ctx.WriteHeader(501)
				ctx.Write(errorMsg("NOT IMPLEMENTED"))
			default:
				ctx.WriteHeader(501)
				ctx.Write(errorMsg("NOT IMPLEMENTED"))
			}
			if err != nil {
				ctx.WriteHeader(400)
				ctx.Write(errorMsg(err.Error()))
				return
			}
		})
	})
	// DELETE - rm, purge, GOODBYE
	web.Delete("/task", func(ctx *web.Context) {
		unpack(ctx, func(v *payload) {
			var err error
			switch v.Action {
			case "remove", "delete", "rm":
				ctx.WriteHeader(501)
				ctx.Write(errorMsg("NOT IMPLEMENTED"))
			case "purge":
				ctx.WriteHeader(501)
				ctx.Write(errorMsg("NOT IMPLEMENTED"))
			default:
				ctx.WriteHeader(501)
				ctx.Write(errorMsg("NOT IMPLEMENTED"))
			}
			if err != nil {
				ctx.WriteHeader(400)
				ctx.Write(errorMsg(err.Error()))
				return
			}
		})
	})
	web.Delete("/session", func(ctx *web.Context) {
		unpack(ctx, func(v *payload) {
			if v.Action == "GOODBYE" {
				ctx.Write(makeResponse(false, "GOODBYE!"))
				time.AfterFunc(time.Second, func() {
					os.Exit(0)
				})
				return
			}
			ctx.WriteHeader(405)
			ctx.Write(errorMsg("ACTION NOT ALLOWED"))
		})
	})
	web.Run("127.0.0.1:8808")
}
