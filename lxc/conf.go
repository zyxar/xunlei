package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

const (
	version = "master"
)

var home string
var conf_file string
var cookie_file string

var conf struct {
	Id   string `json:"account"`
	Pass string `json:"password"`
}

var printVer bool

func printVersion() {
	fmt.Println("lxc version:", version)
}

func initConf() {
	flag.StringVar(&conf.Id, "login", "", "login account")
	flag.StringVar(&conf.Pass, "pass", "", "password/passhash")
	flag.BoolVar(&printVer, "version", false, "print version")
	flag.Parse()
	initHome()
	mkConfigDir()
	conf_file = filepath.Join(home, "config.json")
	cookie_file = filepath.Join(home, "cookie.json")
}

func mkConfigDir() (err error) {
	if home == "" {
		return os.ErrNotExist
	}
	exists, err := isDirExists(home)
	if err != nil {
		return
	}
	if exists {
		return
	}
	return os.Mkdir(home, 0755)
}

func isDirExists(path string) (bool, error) {
	stat, err := os.Stat(path)
	if err == nil {
		if stat.IsDir() {
			return true, nil
		}
		return false, errors.New(path + " exists but is not a directory")
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
