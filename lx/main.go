package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/apex/log"
	"github.com/zyxar/taipei"
	"github.com/zyxar/xunlei/protocol"
)

type Term interface {
	ReadLine() (string, error)
	Restore()
}

func _find(req string) (map[string]*protocol.Task, error) {
	if t, ok := protocol.M.Tasks[req]; ok {
		return map[string]*protocol.Task{req: t}, nil
	}
	if ok, _ := regexp.MatchString(`(.+=.+)+`, req); ok {
		return protocol.FindTasks(req)
	}
	return protocol.FindTasks("name=" + req)
}

func find(req []string) (map[string]*protocol.Task, error) {
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
	var i int
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
	flag.BoolVar(&printVer, "version", false, "print version")
	flag.BoolVar(&isDaemon, "d", false, "run as daemon/server")
	loop := flag.Bool("loop", false, "start daemon loop in background")
	closeFds := flag.Bool("close-fds", false, "close stdout,stderr,stdin")
	debug := flag.Bool("debug", false, "set log level to debug")
	flag.Parse()
	if printVer {
		printVersion()
		return
	}
	if *debug {
		log.SetLevel(log.DebugLevel)
	}
	var login = func() error {
		if err := protocol.ResumeSession(cookieFile); err != nil {
			log.Warn(err.Error())
			if err = protocol.Login(conf.Id, conf.Pass); err != nil {
				return err
			}
			if err = protocol.SaveSession(cookieFile); err != nil {
				return err
			}
		}
		return nil
	}

	if isDaemon {
		cmd := exec.Command(os.Args[0], "-loop", "-close-fds")
		err := cmd.Start()
		if err != nil {
			fmt.Println(err)
		}
		cmd.Process.Release() // FIXME: find a proper way to detect daemon error and call cmd.Process.Kill().
		return
	}

	if *closeFds {
		os.Stdout.Close()
		os.Stderr.Close()
		os.Stdin.Close()
	}

	if *loop {
		go func() {
			if err := login(); err != nil {
				os.Exit(1)
			}
			protocol.GetGdriveId()
		}()
		daemonLoop()
		return
	}

	if err := login(); err != nil {
		os.Exit(1)
	}
	protocol.GetGdriveId()
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
			case "ison":
				fmt.Println(protocol.IsOn())
			case "me":
				fmt.Printf("%#v\n", *protocol.M.Account)
			case "relogin":
				if !protocol.IsOn() {
					if err = protocol.Login(conf.Id, conf.Pass); err != nil {
						fmt.Println(err)
					} else if err = protocol.SaveSession(cookieFile); err != nil {
						fmt.Println(err)
					}
				} else {
					fmt.Println("Already logon.")
				}
			case "saveconf":
				{
					conf.Pass = protocol.EncryptPass(conf.Pass)
					b, err := conf.save(configFileName)
					if err == nil {
						fmt.Printf("%s\n", b)
					}
				}
			case "loadconf":
				{
					if _, err = conf.load(configFileName); err == nil {
						fmt.Printf("%+v\n", conf)
					}
				}
			case "savesession":
				{
					if err := protocol.SaveSession(cookieFile); err != nil {
						fmt.Println(err)
					} else {
						fmt.Println("[done]")
					}
				}
			case "cls", "clear":
				clearscr()
			case "ls":
				ts, err := protocol.GetTasks()
				if err == nil {
					k := 0
					for i := range ts {
						fmt.Printf("#%d %v\n", k, ts[i].Coloring())
						k++
					}
				} else {
					fmt.Println(err)
				}
			case "ld":
				ts, err := protocol.GetDeletedTasks()
				if err == nil {
					k := 0
					for i := range ts {
						fmt.Printf("#%d %v\n", k, ts[i].Coloring())
						k++
					}
				} else {
					fmt.Println(err)
				}
			case "le":
				ts, err := protocol.GetExpiredTasks()
				if err == nil {
					k := 0
					for i := range ts {
						fmt.Printf("#%d %v\n", k, ts[i].Coloring())
						k++
					}
				} else {
					fmt.Println(err)
				}
			case "lc":
				ts, err := protocol.GetCompletedTasks()
				if err == nil {
					k := 0
					for i := range ts {
						fmt.Printf("#%d %v\n", k, ts[i].Coloring())
						k++
					}
				} else {
					fmt.Println(err)
				}
			case "ll":
				var ts []*protocol.Task
				ts, err = protocol.GetTasks()
				if err == nil {
					k := 0
					for i := range ts {
						fmt.Printf("#%d %v\n", k, ts[i].Repr())
						k++
					}
				}
			case "head":
				var ts []*protocol.Task
				var num = 10
				if len(cmds) > 1 {
					num, err = strconv.Atoi(cmds[1])
					if err != nil {
						num = 10
					}
				}
				ts, err = protocol.GetTasks()
				if len(ts) == 0 {
					err = errors.New("Empty task list")
				} else if len(ts) < num {
					num = len(ts)
				}
				if err == nil {
					k := 0
					for i := range ts[:num] {
						fmt.Printf("#%d %v\n", k, ts[i].Coloring())
						k++
					}
				}
			case "cache_clean", "cc":
				if len(cmds) < 2 {
					err = insufficientArgErr
				} else {
					switch cmds[1] {
					case "normal":
						protocol.M.InvalidateGroup(0)
					case "deleted":
						protocol.M.InvalidateGroup(1)
					case "purged":
						protocol.M.InvalidateGroup(2)
					case "invalid":
						protocol.M.InvalidateGroup(3)
					case "expired":
						protocol.M.InvalidateGroup(4)
					case "all":
						protocol.M.InvalidateAll()
					}
				}
			case "info":
				if len(cmds) < 2 {
					err = insufficientArgErr
				} else {
					var ts map[string]*protocol.Task
					if ts, err = find(cmds[1:]); err == nil {
						j := 0
						for i := range ts {
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
			case "launch":
				var b []byte
				b, err = launchAria2cDaemon()
				if err == nil {
					fmt.Printf("%s\n", b)
				}
			case "status":
				var b []byte
				b, err = rpcStatus()
				if err == nil {
					fmt.Printf("%s\n", b)
				}
			case "kill":
				var s string
				force := false
				if len(cmds) >= 2 && (cmds[1] == "-9" || cmds[1] == "-f") {
					force = true
				}
				s, err = rpcShutdown(force)
				if err == nil {
					fmt.Println(s)
				}
			case "submit", "sub":
				if len(cmds) < 2 {
					err = insufficientArgErr
				} else {
					pay := make(map[string]*struct {
						t *protocol.Task
						s string
					})
					for i := range cmds[1:] {
						p := strings.Split(cmds[1:][i], "/")
						m, err := _find(p[0])
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
				}
			case "verify":
				if len(cmds) < 2 {
					err = insufficientArgErr
				} else {
					var ts map[string]*protocol.Task
					if ts, err = find(cmds[1:]); err == nil {
						for i := range ts {
							if _, err = os.Stat(ts[i].TaskName); err != nil {
								fmt.Println(err)
								continue
							}
							fmt.Printf("Task verified? %v\n", ts[i].Verify(ts[i].TaskName))
						}
						err = nil
					}
				}
			case "dl", "download":
				if len(cmds) < 2 {
					err = insufficientArgErr
				} else {
					pay := make(map[string]*struct {
						t *protocol.Task
						s string
					})
					del := false
					check := conf.CheckHash
					for i := range cmds[1:] {
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
							if err = pay[i].t.Remove(); err != nil {
								fmt.Println(err)
							}
						}
					}
					err = nil
				}
			case "dt":
				if len(cmds) > 1 {
					var ts map[string]*protocol.Task
					if ts, err = find(cmds[1:]); err == nil { // TODO: improve find query
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
					err = insufficientArgErr
				}
			case "ti":
				if len(cmds) > 1 {
					var ts map[string]*protocol.Task
					if ts, err = find(cmds[1:]); err == nil { // TODO: improve find query
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
					err = insufficientArgErr
				}
			case "add":
				if len(cmds) < 2 {
					err = insufficientArgErr
				} else {
					req := cmds[1:]
					for j := range req {
						if err = protocol.AddTask(req[j]); err != nil {
							fmt.Println(err)
						}
					}
					err = nil
				}
			case "rm", "delete":
				if len(cmds) < 2 {
					err = insufficientArgErr
				} else {
					var ts map[string]*protocol.Task
					if ts, err = find(cmds[1:]); err == nil {
						for i := range ts {
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
					var ts map[string]*protocol.Task
					if ts, err = find(cmds[1:]); err == nil {
						for i := range ts {
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
					var ts map[string]*protocol.Task
					if ts, err = find(cmds[1:]); err == nil {
						protocol.ReAddTasks(ts)
					}
				} else {
					err = insufficientArgErr
				}
			case "delayall":
				{
					protocol.DelayAllTasks()
				}
			case "pause":
				if len(cmds) > 1 {
					var ts map[string]*protocol.Task
					if ts, err = find(cmds[1:]); err == nil {
						for i := range ts {
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
					var ts map[string]*protocol.Task
					if ts, err = find(cmds[1:]); err == nil {
						for i := range ts {
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
				if len(cmds) > 2 {
					// must be task id here
					if t, ok := protocol.M.Tasks[cmds[1]]; ok {
						t.Rename(strings.Join(cmds[2:], " "))
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
					var ts map[string]*protocol.Task
					if ts, err = find(cmds[1:]); err == nil {
						for i := range ts {
							if err = ts[i].Delay(); err != nil {
								fmt.Println(err)
							}
						}
						err = nil
					}
				}
			case "link":
				if len(cmds) > 1 {
					var ts map[string]*protocol.Task
					if ts, err = find(cmds[1:]); err == nil {
						k := 0
						for i := range ts {
							if !ts[i].IsBt() {
								fmt.Printf("#%d %s: %v\n", k, ts[i].Id, ts[i].LixianURL)
							} else {
								m, err := ts[i].FillBtList()
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
					err = insufficientArgErr
				}
			case "find":
				if len(cmds) > 1 {
					var ts map[string]*protocol.Task
					if ts, err = find(cmds[1:]); err == nil {
						k := 0
						for i := range ts {
							fmt.Printf("#%d %v\n", k, ts[i].Coloring())
							k++
						}
					}
				} else {
					fmt.Println(`pattern == "name=abc&group=completed&status=normal&type=bt"`)
					err = insufficientArgErr
				}
			case "st":
				{
				}
			case "rpc":
				{
				}
			case "play":
				if len(cmds) > 1 {
					var ts map[string]*protocol.Task
					if ts, err = find(cmds[1:]); err == nil {
						for i := range ts {
							low, high, err := ts[i].GetVodURL()
							if err == nil {
								fmt.Printf("%s\n%s\n", low, high)
							}
						}
					}
				} else {
					err = insufficientArgErr
				}
			case "vod":
				if len(cmds) > 1 {
					var list interface{}
					for i := range cmds[1:] {
						switch cmds[1:][i] {
						case "hist":
							list, err = protocol.GetHistoryPlayList()
						case "lx":
							list, err = protocol.GetLxtaskList()
						default:
							err = errors.New("Unkown vod command.")
						}
						if err == nil {
							fmt.Println(list)
						}
					}
				} else {
					err = insufficientArgErr
				}
			case "version":
				printVersion()
			case "update":
				err = protocol.ProcessTask(func(t *protocol.Task) {
					fmt.Printf("%s %s %sB/s %.2f%%\n", t.Id, fixedLengthName(t.TaskName, 32), t.Speed, t.Progress)
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
