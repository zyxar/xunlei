package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

type configure struct {
	Id        string `json:"account"`
	Pass      string `json:"password"`
	CheckHash bool   `json:"check_hash"`
}

var (
	home           string
	configFileName string
	cookieFile     string
	isDaemon       bool
	conf           configure
	version        = "master"
	fromPath       = "github.com/zyxar/xunlei/lxc"
)

func (id *configure) save(cf string) (b []byte, err error) {
	b, err = json.MarshalIndent(id, "", "  ")
	if err != nil {
		return
	}
	err = ioutil.WriteFile(cf, b, 0644)
	return
}

func (id *configure) load(cf string) (b []byte, err error) {
	b, err = ioutil.ReadFile(cf)
	if err != nil {
		return
	}
	err = json.Unmarshal(b, id)
	return
}

var printVer bool

func printVersion() {
	fmt.Printf("lxc version: %s [%s]\n", version, fromPath)
}

func initConf() {
	initHome()
	mkConfigDir()
	configFileName = filepath.Join(home, "config.json")
	cookieFile = filepath.Join(home, "cookie.json")
	conf.CheckHash = true
	conf.load(configFileName)
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
