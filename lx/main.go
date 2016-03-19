package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"unicode/utf8"

	"github.com/apex/log"
	"github.com/zyxar/xunlei/protocol"
)

type Term interface {
	ReadLine() (string, error)
	Restore()
}

func main() {
	initConf()
	printVer := flag.Bool("version", false, "print version")
	debug := flag.Bool("debug", false, "set log level to debug")
	flag.Parse()
	if *printVer {
		printVersion()
		return
	}
	if *debug {
		log.SetLevel(log.DebugLevel)
	}
	var err error
	if err = protocol.ResumeSession(cookieFile); err != nil {
		log.Warn(err.Error())
		if err = protocol.Login(conf.Id, conf.Pass); err != nil {
			log.Warn(err.Error())
			os.Exit(1)
		}
		if err = protocol.SaveSession(cookieFile); err != nil {
			log.Warn(err.Error())
		}
	}

	protocol.GetGdriveId()
	term := newTerm()
	var quit = func(code int) {
		term.Restore()
		os.Exit(code)
	}

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGUSR1)
		for s := range c {
			fmt.Printf("%v caught, exit\n", s)
			quit(1)
			break
		}
	}()

	var line string
	var args []string
	clearscr()
	rpcc.SetNotifier(&notifier{})

	for {
		line, err = term.ReadLine()
		if err == io.EOF {
			quit(0)
		} else if err != nil {
			fmt.Println(err)
			continue
		}
		args = strings.Fields(line)
		if len(args) == 0 {
			continue
		}
		switch args[0] {
		case "version":
			printVersion()
		case "quit", "exit":
			quit(0)
		default:
			if cmd, ok := Cmds[args[0]]; ok {
				err = cmd.fn(args[1:]...)
			} else {
				fmt.Printf("unrecognised command: %s\n", args[0])
				continue
			}
		}
		if err != nil {
			fmt.Println(err)
		}
	}
	term.Restore()
}

func query(req string) (map[string]*protocol.Task, error) {
	if t, ok := protocol.GetTaskById(req); ok {
		return map[string]*protocol.Task{req: t}, nil
	}
	if ok, _ := regexp.MatchString(`(.+=.+)+`, req); ok {
		return protocol.FindTasks(req)
	}
	return protocol.FindTasks("name=" + req)
}

func find(req []string) (map[string]*protocol.Task, error) {
	if len(req) == 0 {
		return nil, errTaskNotFound
	} else if len(req) == 1 {
		return query(req[0])
	}
	return query("name=" + strings.Join(req, "|"))
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
