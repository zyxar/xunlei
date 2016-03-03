package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/zyxar/taipei"
	"github.com/zyxar/xunlei/protocol"
)

type Method struct {
	name  string
	fn    func(args ...string) error
	usage func()
}

func init() {
	Cmds["cc"] = Cmds["cache_clean"]
	Cmds["sub"] = Cmds["submit"]
	Cmds["dl"] = Cmds["download"]
	Cmds["delete"] = Cmds["rm"]
	Cmds["mv"] = Cmds["rename"]
}

var Cmds = map[string]*Method{
	"ison": &Method{name: "ison", fn: func(args ...string) (err error) {
		fmt.Println(protocol.IsOn())
		return
	}},
	"me": &Method{name: "me", fn: func(args ...string) (err error) {
		fmt.Printf("%#v\n", *protocol.GetAccount())
		return
	}},
	"relogin": &Method{name: "relogin", fn: func(args ...string) (err error) {
		if !protocol.IsOn() {
			if err = protocol.Login(conf.Id, conf.Pass); err != nil {
				return
			}
			err = protocol.SaveSession(cookieFile)
			return
		}
		fmt.Println("Already logon.")
		return
	}},
	"saveconf": &Method{name: "saveconf", fn: func(args ...string) (err error) {
		conf.Pass = protocol.EncryptPass(conf.Pass)
		b, err := conf.save(configFileName)
		if err == nil {
			fmt.Printf("%s\n", b)
		}
		return
	}},
	"loadconf": &Method{name: "loadconf", fn: func(args ...string) (err error) {
		if _, err = conf.load(configFileName); err == nil {
			fmt.Printf("%+v\n", conf)
		}
		return
	}},
	"savesession": &Method{name: "savesession", fn: func(args ...string) (err error) {
		if err = protocol.SaveSession(cookieFile); err == nil {
			fmt.Println("[done]")
		}
		return
	}},
	"ls": &Method{name: "ls", fn: func(args ...string) (err error) {
		ts, err := protocol.GetTasks()
		if err != nil {
			return
		}
		k := 0
		for i := range ts {
			fmt.Printf("#%d %v\n", k, ts[i].Coloring())
			k++
		}
		return
	}},
	"ld": &Method{name: "ld", fn: func(args ...string) (err error) {
		ts, err := protocol.GetDeletedTasks()
		if err != nil {
			return
		}
		k := 0
		for i := range ts {
			fmt.Printf("#%d %v\n", k, ts[i].Coloring())
			k++
		}
		return
	}},
	"le": &Method{name: "le", fn: func(args ...string) (err error) {
		ts, err := protocol.GetExpiredTasks()
		if err != nil {
			return
		}
		k := 0
		for i := range ts {
			fmt.Printf("#%d %v\n", k, ts[i].Coloring())
			k++
		}
		return
	}},
	"lc": &Method{name: "lc", fn: func(args ...string) (err error) {
		ts, err := protocol.GetCompletedTasks()
		if err != nil {
			return
		}
		k := 0
		for i := range ts {
			fmt.Printf("#%d %v\n", k, ts[i].Coloring())
			k++
		}
		return
	}},
	"ll": &Method{name: "ll", fn: func(args ...string) (err error) {
		ts, err := protocol.GetTasks()
		if err != nil {
			return
		}
		k := 0
		for i := range ts {
			fmt.Printf("#%d %v\n", k, ts[i].Repr())
			k++
		}
		return
	}},
	"head": &Method{name: "head", fn: func(args ...string) (err error) {
		var num = 10
		if len(args) > 0 {
			num, err = strconv.Atoi(args[1])
			if err != nil {
				num = 10
			}
		}
		ts, err := protocol.GetTasks()
		if err != nil {
			return
		}
		if len(ts) == 0 {
			err = errors.New("Empty task list")
			return
		} else if len(ts) < num {
			num = len(ts)
		}
		k := 0
		for i := range ts[:num] {
			fmt.Printf("#%d %v\n", k, ts[i].Coloring())
			k++
		}
		return
	}},
	"cache_clean": &Method{name: "cache_clean", fn: func(args ...string) (err error) {
		if len(args) < 1 {
			err = errInsufficientArg
			return
		}
		switch args[0] {
		case "normal":
			protocol.InvalidateCache(0)
		case "deleted":
			protocol.InvalidateCache(1)
		case "purged":
			protocol.InvalidateCache(2)
		case "invalid":
			protocol.InvalidateCache(3)
		case "expired":
			protocol.InvalidateCache(4)
		case "all":
			protocol.InvalidateCacheAll()
		}
		return
	}},
	"info": &Method{name: "info", fn: func(args ...string) (err error) {
		if len(args) < 1 {
			err = errInsufficientArg
			return
		}
		ts, err := find(args)
		if err != nil {
			return
		}
		j := 0
		for i := range ts {
			if ts[i].IsBt() {
				m, err := protocol.FillBtList(ts[i])
				fmt.Printf("#%d %v\n", j, ts[i].Repr())
				if err == nil {
					fmt.Printf("%v\n", m)
				}
			} else {
				fmt.Printf("#%d %v\n", j, ts[i].Repr())
			}
			j++
		}
		return
	}},
	"launch": &Method{name: "launch", fn: func(args ...string) (err error) {
		info, err := rpcc.LaunchAria2cDaemon()
		if err == nil {
			fmt.Printf("aria2: %+v\n", info)
		}
		return
	}},
	"status": &Method{name: "status", fn: func(args ...string) (err error) {
		stat, err := rpcc.GetGlobalStat()
		if err == nil {
			fmt.Printf("%+v\n", stat)
		}
		return
	}},
	"kill": &Method{name: "kill", fn: func(args ...string) (err error) {
		var s string
		if len(args) >= 1 && (args[0] == "-9" || args[0] == "-f") {
			s, err = rpcc.ForceShutdown()
		} else {
			s, err = rpcc.Shutdown()
		}
		if err == nil {
			fmt.Println(s)
		}
		return
	}},
	"submit": &Method{name: "submit", fn: func(args ...string) (err error) {
		if len(args) < 1 {
			err = errInsufficientArg
			return
		}
		pay := make(map[string]*struct {
			t *protocol.Task
			s string
		})
		for i := range args {
			p := strings.Split(args[i], "/")
			m, err := query(p[0])
			if err == nil {
				for i := range m {
					var filter string
					if len(p) == 1 {
						filter = `.*`
					} else {
						filter = p[1]
					}
					pay[m[i].Id] = &struct {
						t *protocol.Task
						s string
					}{m[i], filter}
				}
			}
		}
		for i := range pay {
			if err = download(pay[i].t, pay[i].s, false, false, func(uri, filename string, echo bool) error {
				_, err := rpcAddTask(uri, filename)
				return err
			}); err != nil {
				fmt.Println(err)
			}
		}
		err = nil
		return
	}},
	"verify": &Method{name: "verify", fn: func(args ...string) (err error) {
		if len(args) < 1 {
			err = errInsufficientArg
			return
		}
		ts, err := find(args)
		if err == nil {
			for i := range ts {
				if _, err = os.Stat(ts[i].TaskName); err != nil {
					fmt.Println(err)
					continue
				}
				fmt.Printf("Task verified? %v\n", protocol.VerifyTask(ts[i], ts[i].TaskName))
			}
			err = nil
		}
		return
	}},
	"download": &Method{name: "download", fn: func(args ...string) (err error) {
		if len(args) < 1 {
			err = errInsufficientArg
			return
		}
		pay := make(map[string]*struct {
			t *protocol.Task
			s string
		})
		del := false
		check := conf.CheckHash
		for i := range args {
			if strings.HasPrefix(args[i], "--") {
				switch args[i][2:] {
				case "delete":
					del = true
				case "check":
					check = true
				case "no-check", "nocheck":
					check = false
				}
			} else {
				p := strings.Split(args[i], "/")
				m, err := query(p[0])
				if err == nil {
					for i := range m {
						var filter string
						if len(p) == 1 {
							filter = `.*`
						} else {
							filter = p[1]
						}
						pay[m[i].Id] = &struct {
							t *protocol.Task
							s string
						}{m[i], filter}
					}
				}
			}
		}
		for i := range pay {
			if err = download(pay[i].t, pay[i].s, true, check, dl); err != nil {
				fmt.Println(err)
			} else if del {
				if err = protocol.DeleteTask(pay[i].t); err != nil {
					fmt.Println(err)
				}
			}
		}
		err = nil
		return
	}},
	"dt": &Method{name: "dt", fn: func(args ...string) (err error) {
		if len(args) > 0 {
			ts, err := find(args)
			if err == nil { // TODO: improve find query
				for i := range ts {
					if ts[i].IsBt() {
						if err = protocol.GetTorrentFileByHash(ts[i].Cid, ts[i].TaskName+".torrent"); err != nil {
							fmt.Println(err)
						}
					}
				}
				err = nil
			}
		} else {
			err = errInsufficientArg
		}
		return
	}},
	"ti": &Method{name: "ti", fn: func(args ...string) (err error) {
		if len(args) > 0 {
			ts, err := find(args)
			if err == nil { // TODO: improve find query
				for i := range ts {
					if ts[i].IsBt() {
						if b, err := protocol.GetTorrentByHash(ts[i].Cid); err != nil {
							fmt.Println(err)
						} else {
							if m, err := taipei.DecodeMetaInfo(b); err != nil {
								fmt.Println(err)
							} else {
								taipei.Iconv(m)
								fmt.Println(m)
							}
						}
					}
				}
				err = nil
			}
		} else {
			err = errInsufficientArg
		}
		return
	}},
	"add": &Method{name: "add", fn: func(args ...string) (err error) {
		if len(args) < 1 {
			err = errInsufficientArg
			return
		}
		for i := range args {
			if err = protocol.AddTask(args[i]); err != nil {
				fmt.Println(err)
			}
		}
		err = nil
		return
	}},
	"rm": &Method{name: "rm", fn: func(args ...string) (err error) {
		if len(args) < 1 {
			err = errInsufficientArg
			return
		}
		ts, err := find(args)
		if err == nil {
			for i := range ts {
				if err = protocol.DeleteTask(ts[i]); err != nil {
					fmt.Println(err)
				}
			}
			err = nil
		}
		return
	}},
	"purge": &Method{name: "purge", fn: func(args ...string) (err error) {
		if len(args) < 1 {
			err = errInsufficientArg
			return
		}
		ts, err := find(args)
		if err == nil {
			for i := range ts {
				if err = protocol.PurgeTask(ts[i]); err != nil {
					fmt.Println(err)
				}
			}
			err = nil
		}
		return
	}},
	"readd": &Method{name: "readd", fn: func(args ...string) (err error) {
		if len(args) > 0 {
			ts, err := find(args)
			if err == nil {
				protocol.ReAddTasks(ts)
			}
		} else {
			err = errInsufficientArg
		}
		return
	}},
	"delayall": &Method{name: "delayall", fn: func(args ...string) (err error) {
		err = protocol.DelayAllTasks()
		return
	}},
	"pause": &Method{name: "pause", fn: func(args ...string) (err error) {
		if len(args) > 0 {
			ts, err := find(args)
			if err == nil {
				for i := range ts {
					if err = protocol.PauseTask(ts[i]); err != nil {
						fmt.Println(err)
					}
				}
				err = nil
			}
		} else {
			err = errInsufficientArg
		}
		return
	}},
	"resume": &Method{name: "resume", fn: func(args ...string) (err error) {
		if len(args) > 0 {
			ts, err := find(args)
			if err == nil {
				for i := range ts {
					if err = protocol.ResumeTask(ts[i]); err != nil {
						fmt.Println(err)
					}
				}
				err = nil
			}
		} else {
			err = errInsufficientArg
		}
		return
	}},
	"rename": &Method{name: "rename", fn: func(args ...string) (err error) {
		if len(args) > 2 {
			// must be task id here
			if t, ok := protocol.GetTaskById(args[0]); ok {
				err = protocol.RenameTask(t, strings.Join(args[1:], " "))
			} else {
				err = errNoSuchTasks
			}
		} else {
			err = errInsufficientArg
		}
		return
	}},
	"delay": &Method{name: "delay", fn: func(args ...string) (err error) {
		if len(args) < 1 {
			err = errInsufficientArg
		} else {
			ts, err := find(args)
			if err == nil {
				for i := range ts {
					if err = protocol.DelayTask(ts[i]); err != nil {
						fmt.Println(err)
					}
				}
				err = nil
			}
		}
		return
	}},
	"link": &Method{name: "link", fn: func(args ...string) (err error) {
		if len(args) > 0 {
			ts, err := find(args)
			if err == nil {
				k := 0
				for i := range ts {
					if !ts[i].IsBt() {
						fmt.Printf("#%d %s: %v\n", k, ts[i].Id, ts[i].LixianURL)
					} else {
						m, err := protocol.FillBtList(ts[i])
						if err == nil {
							fmt.Printf("#%d %s:\n", k, ts[i].Id)
							for j := range m.Record {
								fmt.Printf("  #%d %s\n", j, m.Record[j].DownURL)
							}
						} else {
							fmt.Println(err)
						}
					}
					k++
				}
			}
		} else {
			err = errInsufficientArg
		}
		return
	}},
	"find": &Method{name: "find", fn: func(args ...string) (err error) {
		if len(args) > 0 {
			ts, err := find(args)
			if err == nil {
				k := 0
				for i := range ts {
					fmt.Printf("#%d %v\n", k, ts[i].Coloring())
					k++
				}
			}
		} else {
			fmt.Println(`pattern == "name=abc&group=completed&status=normal&type=bt"`)
			err = errInsufficientArg
		}
		return
	}},
	"update": &Method{name: "update", fn: func(args ...string) (err error) {
		err = protocol.ProcessTask(func(t *protocol.Task) error {
			fmt.Printf("%s %s %sB/s %.2f%%\n", t.Id, fixedLengthName(t.TaskName, 32), t.Speed, t.Progress)
			return nil
		})
		return
	}},
	"help": &Method{name: "help", fn: func(args ...string) (err error) {
		return
	}},
}
