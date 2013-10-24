package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/matzoe/xunlei/api"
)

type Term interface {
	ReadLine() (string, error)
	Restore()
}

func _find(req string) (map[string]*api.Task, error) {
	if t, ok := api.M.Tasks[req]; ok {
		return map[string]*api.Task{req: t}, nil
	}
	if ok, _ := regexp.MatchString(`(.+=.+)+`, req); ok {
		return api.FindTasks(req)
	}
	return api.FindTasks("name=" + req)
}

func find(req []string) (map[string]*api.Task, error) {
	if len(req) == 0 {
		return nil, errors.New("Empty find query.")
	} else if len(req) == 1 {
		return _find(req[0])
	}
	return _find("name=" + strings.Join(req, "|"))
}

func fixedLengthName(name string, size int) string {
	l := utf8.RuneCountInString(name)
	var b bytes.Buffer
	var i int = 0
	for i < l && i < size {
		r, s := utf8.DecodeRuneInString(name)
		b.WriteRune(r)
		name = name[s:]
		if s > 1 {
			i += 2
		} else {
			i++
		}
	}
	for i < size {
		b.WriteByte(' ')
		i++
	}
	return b.String()
}

func main() {
	initConf()
	f, _ := ioutil.ReadFile(conf_file)
	json.Unmarshal(f, &conf)
	flag.StringVar(&conf.Id, "login", conf.Id, "login account")
	flag.StringVar(&conf.Pass, "pass", conf.Pass, "password/passhash")
	flag.BoolVar(&printVer, "version", false, "print version")
	flag.Parse()
	if printVer {
		printVersion()
		return
	}
	if err := api.ResumeSession(cookie_file); err != nil {
		log.Println(err)
		if err = api.Login(conf.Id, conf.Pass); err != nil {
			log.Println(err)
			os.Exit(1)
		}
		if err = api.SaveSession(cookie_file); err != nil {
			log.Println(err)
		}
	}
	api.GetGdriveId()

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
					if ts, err = find(cmds[1:]); err == nil {
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
					pay := make(map[string]*struct {
						t *api.Task
						s string
					})
					del := false
					check := conf.checkHash
					for i, _ := range cmds[1:] {
						if strings.HasPrefix(cmds[1:][i], "--") {
							switch cmds[1:][i][2:] {
							case "delete":
								del = true
							case "check":
								check = true
							case "no-check", "nocheck":
								check = false
							}
						} else {
							p := strings.Split(cmds[1:][i], "/")
							m, err := _find(p[0])
							if err == nil {
								for i, _ := range m {
									var filter string
									if len(p) == 1 {
										filter = `.*`
									} else {
										filter = p[1]
									}
									pay[m[i].Id] = &struct {
										t *api.Task
										s string
									}{m[i], filter}
								}
							}
						}
					}
					for i, _ := range pay {
						if err = download(pay[i].t, pay[i].s, true, check); err != nil {
							fmt.Println(err)
						} else if del {
							if err = pay[i].t.Remove(); err != nil {
								fmt.Println(err)
							}
						}
					}
					err = nil
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
					}
					err = nil
				}
			case "rm", "delete":
				if len(cmds) < 2 {
					err = insufficientArgErr
				} else {
					var ts map[string]*api.Task
					if ts, err = find(cmds[1:]); err == nil {
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
					if ts, err = find(cmds[1:]); err == nil {
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
				if len(cmds) > 1 {
					var ts map[string]*api.Task
					if ts, err = find(cmds[1:]); err == nil {
						api.ReAddTasks(ts)
					}
				} else {
					err = insufficientArgErr
				}
			case "pause":
				if len(cmds) > 1 {
					var ts map[string]*api.Task
					if ts, err = find(cmds[1:]); err == nil {
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
			case "resume":
				if len(cmds) > 1 {
					var ts map[string]*api.Task
					if ts, err = find(cmds[1:]); err == nil {
						for i, _ := range ts {
							if err = ts[i].Resume(); err != nil {
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
					if ts, err = find(cmds[1:]); err == nil {
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
					if ts, err = find(cmds[1:]); err == nil {
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
			case "find":
				if len(cmds) == 2 {
					var ts map[string]*api.Task
					if ts, err = find(cmds[1:]); err == nil {
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
					log.Printf("%s %s %sB/s %.2f%%\n", t.Id, fixedLengthName(t.TaskName, 32), t.Speed, t.Progress)
				})
			case "quit", "exit":
				break LOOP
			case "help":
				// TODO
			default:
				err = fmt.Errorf("Unrecognised command: %s", cmds[0])
			}
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}
