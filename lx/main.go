package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/apex/log"
	"github.com/zyxar/argo/rpc"
	"github.com/zyxar/xunlei/protocol"
)

type Term interface {
	ReadLine() (string, error)
	Restore()
}

var (
	errInsufficientArg = errors.New("insufficient arguments")
	errNoSuchTasks     = errors.New("no such tasks")
)

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
		return nil, errors.New("Empty find query.")
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
	var err error
	var line string
	var args []string
	clearscr()
	rpcc.SetNotifier(&rpc.DummyNotifier{})
LOOP:
	for {
		line, err = term.ReadLine()
		if err != nil {
			break
		}
		args = strings.Fields(line)
		if len(args) == 0 {
			continue
		}
		switch args[0] {
		case "version":
			printVersion()
		case "quit", "exit":
			break LOOP
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
