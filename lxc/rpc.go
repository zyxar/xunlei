package main

import (
	"github.com/matzoe/argo/rpc"
)

var rpcc *rpc.Client

func init() {
	rpcc = rpc.New("http://localhost:6800/jsonrpc")
}

func RPCAddTask(uri, gdriveid string) (gid string, err error) {
	// {'header':['Cookie: gdriveid=XXXX']}
	return rpcc.AddUri(uri, "COOKIE")
}
