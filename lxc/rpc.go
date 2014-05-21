package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"

	"github.com/matzoe/argo/rpc"
	. "github.com/matzoe/xunlei/protocol"
)

var rpcc *rpc.Client

func init() {
	rpcc = rpc.New("http://localhost:6800/jsonrpc")
}

// {'header':['Cookie: XXXX']}
// --file-allocation=none", "-x5", "-c", "--summary-interval=0", "--follow-torrent=false"
type opt map[string]interface{}

func newOpt(gdriveid, filename string) (opt, error) {
	if len(gdriveid) == 0 {
		return nil, errors.New("Cannot retrieve gdriveid.")
	}
	r := make(map[string]interface{})
	r["header"] = []string{"Cookie: gdriveid=" + gdriveid}
	r["out"] = filename
	r["file-allocation"] = "none"
	r["summary-interval"] = "0"
	r["max-connection-per-server"] = "5"
	r["follow-torrent"] = "false"
	r["continue"] = "true"
	return r, nil
}

func RPCAddTask(uri, filename string) (gid string, err error) {
	lxhead, err := newOpt(M.Gid, filename)
	if err != nil {
		return "", err
	}
	return rpcc.AddUri(uri, lxhead)
}

func RPCStatus() error {
	if m, err := rpcc.GetGlobalStat(); err != nil {
		return err
	} else {
		b, _ := json.MarshalIndent(m, "", "  ")
		fmt.Printf("%s\n", b)
	}
	return nil
}

func launchAria2cDaemon() error {
	if m, err := rpcc.GetVersion(); err == nil {
		fmt.Printf("aria2c version: %v\nenabled features: %v\n", m["version"], m["enabledFeatures"])
		return nil
	}
	cmd := exec.Command("aria2c", "--enable-rpc", "--rpc-listen-all")
	if err := cmd.Start(); err != nil {
		return err
	}
	cmd.Process.Release()
	fmt.Println("Aria2c daemon launched.")
	return nil
}
