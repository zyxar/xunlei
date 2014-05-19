package main

import (
	"errors"

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
