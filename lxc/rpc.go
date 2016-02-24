package main

import (
	"encoding/json"
	"errors"
	"os/exec"

	"github.com/zyxar/argo/rpc"
	"github.com/zyxar/xunlei/protocol"
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
	lxhead, err := newOpt(protocol.M.Gid, filename)
	if err != nil {
		return "", err
	}
	return rpcc.AddUri(uri, lxhead)
}

func RPCStatus() ([]byte, error) {
	m, err := rpcc.GetGlobalStat()
	if err != nil {
		return nil, err
	}
	b, _ := json.MarshalIndent(m, "", "  ")
	return b, nil
}

func RPCShutdown(force bool) (string, error) {
	if !force {
		return rpcc.Shutdown()
	}
	return rpcc.ForceShutdown()
}

func launchAria2cDaemon() ([]byte, error) {
	if m, err := rpcc.GetVersion(); err == nil {
		b, _ := json.MarshalIndent(m, "", "  ")
		return b, nil
	}
	cmd := exec.Command("aria2c", "--enable-rpc", "--rpc-listen-all")
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	cmd.Process.Release()
	return []byte("Aria2c daemon launched"), nil
}
