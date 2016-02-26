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
	conf           configure
	isDaemon       bool
	home           string
	configFileName string
	cookieFile     string
	hashes         string
)

const (
	version    = "master"
	binaryName = "github.com/zyxar/xunlei/lx"
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
	if len(hashes) > 0 {
		fmt.Printf("[%s]\nversions:\n", binaryName)
		var deps map[string]string
		if json.Unmarshal([]byte(hashes), &deps) == nil {
			for k, v := range deps {
				fmt.Printf("     %s:\r\t\t\t\t%s\n", k, v)
			}
		}
		fmt.Println()
	} else {
		fmt.Printf("[%s] version: %s\n", binaryName, version)
	}
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
