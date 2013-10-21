package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/matzoe/xunlei/api"
)

type Term interface {
	ReadLine() (string, error)
	Restore()
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
		// noTasksMatchesErr := errors.New("No task matches.")
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
			case "cls":
				fallthrough
			case "clear":
				clearscr()
			case "show":
				fallthrough
			case "ls":
				ts, err := api.GetTasks()
				if err == nil {
					for i, _ := range ts {
						fmt.Printf("%v\n", ts[i])
					}
				}
			case "ld":
				ts, err := api.GetDeletedTasks()
				if err == nil {
					for i, _ := range ts {
						fmt.Printf("%v\n", ts[i])
					}
				}
			case "le":
				ts, err := api.GetExpiredTasks()
				if err == nil {
					for i, _ := range ts {
						fmt.Printf("%v\n", ts[i])
					}
				}
			case "lc":
				ts, err := api.GetCompletedTasks()
				if err == nil {
					for i, _ := range ts {
						fmt.Printf("%v\n", ts[i])
					}
				}
			case "ll":
				ts, err := api.GetTasks()
				if err == nil {
					for i, _ := range ts {
						fmt.Printf("%v\n", ts[i].Repr())
					}
				}
			case "info":
				if len(cmds) < 2 {
					err = insufficientArgErr
				} else {

				}
			case "dl":
				fallthrough
			case "download":
				if len(cmds) < 2 {
					err = insufficientArgErr
				} else {

				}
			case "add":
				if len(cmds) >= 2 {

				} else {
					err = insufficientArgErr
				}
			case "rm":
				fallthrough
			case "delete":
				if len(cmds) == 2 {

				} else if len(cmds) > 2 {

				} else {
					err = insufficientArgErr
				}
			case "purge":
				if len(cmds) < 2 {
					err = insufficientArgErr
				} else {

				}
			case "readd":
				// re-add tasks from deleted or expired
			case "pause":
				if len(cmds) > 1 {

				} else {
					err = insufficientArgErr
				}
			case "restart":
				if len(cmds) > 1 {

				} else {
					err = insufficientArgErr
				}
			case "rename":
				fallthrough
			case "mv":
				if len(cmds) == 3 {

				} else {
					err = insufficientArgErr
				}
			case "delay":
				if len(cmds) == 2 {

				} else {
					err = insufficientArgErr
				}
			case "link":
				// get lixian_URL of a task
			case "dispatch":
				if len(cmds) == 2 {

				} else {
					err = insufficientArgErr
				}
			case "version":
				printVersion()
			case "update":
				err = api.ProcessTask(func(t *api.Task) {
					log.Printf("%s %sB/s %.2f%%\n", t.Id, t.Speed, t.Progress)
				})
			case "quit":
				fallthrough
			case "exit":
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
