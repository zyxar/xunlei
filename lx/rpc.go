package main

import (
	"encoding/json"
	"errors"
	"os/exec"
	"time"

	"github.com/zyxar/argo/rpc"
	"github.com/zyxar/xunlei/protocol"
)

var rpcc rpc.Protocol

func init() {
	rpcc = rpc.New("http://localhost:6800/jsonrpc")
}

// {'header':['Cookie: XXXX']}
// --file-allocation=none", "-x5", "-c", "--summary-interval=0", "--follow-torrent=false"
func newOpt(gdriveid, filename string) (rpc.Option, error) {
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

func rpcAddTask(uri, filename string) (gid string, err error) {
	lxhead, err := newOpt(protocol.M.Gid, filename)
	if err != nil {
		return "", err
	}
	return rpcc.AddURI(uri, lxhead)
}

func rpcStatus() ([]byte, error) {
	m, err := rpcc.GetGlobalStat()
	if err != nil {
		return nil, err
	}
	b, _ := json.MarshalIndent(m, "", "  ")
	return b, nil
}

func rpcShutdown(force bool) (string, error) {
	if !force {
		return rpcc.Shutdown()
	}
	return rpcc.ForceShutdown()
}

func rpcGetVersion() (b []byte, err error) {
	m, err := rpcc.GetVersion()
	if err == nil {
		b, _ = json.MarshalIndent(m, "", "  ")
	}
	return
}

func launchAria2cDaemon() (b []byte, err error) {
	if b, err = rpcGetVersion(); err == nil {
		return
	}
	cmd := exec.Command("aria2c", "--enable-rpc", "--rpc-listen-all")
	if err = cmd.Start(); err != nil {
		return
	}
	cmd.Process.Release()
	time.Sleep(time.Second)
	b, err = rpcGetVersion()
	return
}
