package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/matzoe/xunlei/api"
)

type Term interface {
	ReadLine() (string, error)
	Restore()
}

func _dispatch(req string) (map[string]*api.Task, error) {
	if t, ok := api.M.Tasks[req]; ok {
		return map[string]*api.Task{req: t}, nil
	}
	if ok, _ := regexp.MatchString(`(.+=.+)+`, req); ok {
		return api.DispatchTasks(req)
	}
	return api.DispatchTasks("name=" + req)
}

func dispatch(req []string) (map[string]*api.Task, error) {
	if len(req) == 0 {
		return nil, errors.New("Empty dispatch query.")
	} else if len(req) == 1 {
		return _dispatch(req[0])
	}
	return _dispatch("name=" + strings.Join(req, "|"))
}

func main() {
	initConf()
	if printVer {
		printVersion()
		return
	}
	if err := api.ResumeSession(cookie_file); err != nil {
		log.Println(err)
		f, _ := ioutil.ReadFile(conf_file)
		json.Unmarshal(f, &conf)
		err := api.Login(conf.Id, conf.Pass)
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
		if err = api.SaveSession(cookie_file); err != nil {
			log.Println(err)
		}
	}
	api.GetGdriveId()
	api.GetTasks()
	// ch := make(chan byte)
	// api.ProcessTask(ch, func(t *api.Task) {
	// 	log.Printf("%s %sB/s %.2f%%\n", t.Id, t.Speed, t.Progress)
	// })

	term := newTerm()
	defer term.Restore()
	{
		var err error
		insufficientArgErr := errors.New("Insufficient arguments.")
		noTasksMatchesErr := errors.New("No task matches.")
		var line string
		var cmds []string
		clearscr()
	LOOP:
		for {
			line, err = term.ReadLine()
			if err != nil {
				break
			}
			cmds = strings.Fields(line)
			if len(cmds) == 0 {
				continue
			}
			switch cmds[0] {
			case "cls", "clear":
				clearscr()
			case "ls":
				ts, err := api.GetTasks()
				if err == nil {
					k := 0
					for i, _ := range ts {
						fmt.Printf("#%d %v\n", k, ts[i].Coloring())
						k++
					}
				}
			case "ld":
				ts, err := api.GetDeletedTasks()
				if err == nil {
					k := 0
					for i, _ := range ts {
						fmt.Printf("#%d %v\n", k, ts[i].Coloring())
						k++
					}
				}
			case "le":
				ts, err := api.GetExpiredTasks()
				if err == nil {
					k := 0
					for i, _ := range ts {
						fmt.Printf("#%d %v\n", k, ts[i].Coloring())
						k++
					}
				}
			case "lc":
				ts, err := api.GetCompletedTasks()
				if err == nil {
					k := 0
					for i, _ := range ts {
						fmt.Printf("#%d %v\n", k, ts[i].Coloring())
						k++
					}
				}
			case "ll":
				ts, err := api.GetTasks()
				if err == nil {
					k := 0
					for i, _ := range ts {
						fmt.Printf("#%d %v\n", k, ts[i].Repr())
						k++
					}
				}
			case "info":
				if len(cmds) < 2 {
					err = insufficientArgErr
				} else {
					var ts map[string]*api.Task
					if ts, err = dispatch(cmds[1:]); err == nil {
						j := 0
						for i, _ := range ts {
							if ts[i].IsBt() {
								m, err := ts[i].FillBtList()
								fmt.Printf("#%d %v\n", j, ts[i].Repr())
								if err == nil {
									fmt.Printf("%v\n", m)
								}
							} else {
								fmt.Printf("#%d %v\n", j, ts[i].Repr())
							}
							j++
						}
					}
				}
			case "dl", "download":
				if len(cmds) < 2 {
					err = insufficientArgErr
				} else {

				}
			case "add":
				if len(cmds) < 2 {
					err = insufficientArgErr
				} else {
					req := cmds[1:]
					for j, _ := range req {
						if err = api.AddTask(req[j]); err != nil {
							fmt.Println(err)
						}
						err = nil
					}
				}
			case "rm", "delete":
				if len(cmds) < 2 {
					err = insufficientArgErr
				} else {
					var ts map[string]*api.Task
					if ts, err = dispatch(cmds[1:]); err == nil {
						for i, _ := range ts {
							if err = ts[i].Remove(); err != nil {
								fmt.Println(err)
							}
						}
						err = nil
					}
				}
			case "purge":
				if len(cmds) < 2 {
					err = insufficientArgErr
				} else {
					var ts map[string]*api.Task
					if ts, err = dispatch(cmds[1:]); err == nil {
						for i, _ := range ts {
							if err = ts[i].Purge(); err != nil {
								fmt.Println(err)
							}
						}
						err = nil
					}
				}
			case "readd":
				// re-add tasks from deleted or expired
			case "pause":
				if len(cmds) > 1 {
					var ts map[string]*api.Task
					if ts, err = dispatch(cmds[1:]); err == nil {
						for i, _ := range ts {
							if err = ts[i].Pause(); err != nil {
								fmt.Println(err)
							}
						}
						err = nil
					}
				} else {
					err = insufficientArgErr
				}
			case "restart":
				if len(cmds) > 1 {
					var ts map[string]*api.Task
					if ts, err = dispatch(cmds[1:]); err == nil {
						for i, _ := range ts {
							if err = ts[i].Restart(); err != nil {
								fmt.Println(err)
							}
						}
						err = nil
					}
				} else {
					err = insufficientArgErr
				}
			case "rename", "mv":
				if len(cmds) == 3 {
					// must be task id here
					if t, ok := api.M.Tasks[cmds[1]]; ok {
						t.Rename(cmds[2])
					} else {
						err = noTasksMatchesErr
					}
				} else {
					err = insufficientArgErr
				}
			case "delay":
				if len(cmds) < 2 {
					err = insufficientArgErr
				} else {
					var ts map[string]*api.Task
					if ts, err = dispatch(cmds[1:]); err == nil {
						for i, _ := range ts {
							if err = ts[i].Delay(); err != nil {
								fmt.Println(err)
							}
						}
						err = nil
					}
				}
			case "link":
				if len(cmds) == 2 {
					var ts map[string]*api.Task
					if ts, err = dispatch(cmds[1:]); err == nil {
						k := 0
						for i, _ := range ts {
							if !ts[i].IsBt() {
								fmt.Printf("#%d %s: %v\n", k, ts[i].Id, ts[i].LixianURL)
							} else {
								m, err := ts[i].FillBtList()
								if err == nil {
									fmt.Printf("#%d %s:\n", k, ts[i].Id)
									for j, _ := range m.Record {
										fmt.Printf("  #%d %s\n", m.Record[j].Id, m.Record[j].DownURL)
									}
								}
							}
							k++
						}
					}
				} else {
					err = insufficientArgErr
				}
			case "dispatch":
				if len(cmds) == 2 {
					var ts map[string]*api.Task
					if ts, err = dispatch(cmds[1:]); err == nil {
						k := 0
						for i, _ := range ts {
							fmt.Printf("#%d %v\n", k, ts[i].Coloring())
							k++
						}
					}
				} else {
					err = insufficientArgErr
				}
			case "version":
				printVersion()
			case "update":
				err = api.ProcessTask(func(t *api.Task) {
					log.Printf("%s %sB/s %.2f%%\n", t.Id, t.Speed, t.Progress)
				})
			case "quit", "exit":
				break LOOP
			default:
				err = fmt.Errorf("Unrecognised command: %s", cmds[0])
			}
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}
