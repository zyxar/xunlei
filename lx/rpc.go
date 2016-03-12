package main

import (
	"errors"
	"fmt"

	"github.com/zyxar/argo/rpc"
	"github.com/zyxar/xunlei/protocol"
)

var rpcc, _ = rpc.New("http://localhost:6800/jsonrpc")

// {'header':['Cookie: XXXX']}
// --file-allocation=none", "-x5", "-c", "--summary-interval=0", "--follow-torrent=false"
func newOpt(gdriveid, filename string) (rpc.Option, error) {
	if len(gdriveid) == 0 {
		return nil, errors.New("gdriveid not found")
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

func rpcAddTask(uri, filename string) (string, error) {
	gid, err := protocol.GetGdriveId()
	if err != nil {
		return "", err
	}
	lxhead, err := newOpt(gid, filename)
	if err != nil {
		return "", err
	}
	return rpcc.AddURI(uri, lxhead)
}

type notifier struct {
}

func (*notifier) OnStart(es []rpc.Event) {
	var info []rpc.FileInfo
	var err error
	for e := range es {
		info, err = rpcc.GetFiles(es[e].Gid)
		if err != nil {
			fmt.Println(es[e].Gid, err)
			continue
		}
		fmt.Printf("\r(%s) %s started.\n", es[e].Gid, info[0].Path)
	}
}
func (*notifier) OnPause(es []rpc.Event) {
	var info []rpc.FileInfo
	var err error
	for e := range es {
		info, err = rpcc.GetFiles(es[e].Gid)
		if err != nil {
			fmt.Println(es[e].Gid, err)
			continue
		}
		fmt.Printf("\r(%s) %s paused.\n", es[e].Gid, info[0].Path)
	}
}
func (*notifier) OnStop(es []rpc.Event) {
	var info []rpc.FileInfo
	var err error
	for e := range es {
		info, err = rpcc.GetFiles(es[e].Gid)
		if err != nil {
			fmt.Println(es[e].Gid, err)
			continue
		}
		fmt.Printf("\r(%s) %s stopped.\n", es[e].Gid, info[0].Path)
	}
}
func (*notifier) OnComplete(es []rpc.Event) {
	var info []rpc.FileInfo
	var err error
	for e := range es {
		info, err = rpcc.GetFiles(es[e].Gid)
		if err != nil {
			fmt.Println(es[e].Gid, err)
			continue
		}
		fmt.Printf("\r(%s) %s completed.\n", es[e].Gid, info[0].Path)
	}
}
func (*notifier) OnError(es []rpc.Event) {
	var info []rpc.FileInfo
	var err error
	for e := range es {
		info, err = rpcc.GetFiles(es[e].Gid)
		if err != nil {
			fmt.Println(es[e].Gid, err)
			continue
		}
		fmt.Printf("\r(%s) %s error.\n", es[e].Gid, info[0].Path)
	}
}
func (*notifier) OnBtComplete(es []rpc.Event) {
	var info []rpc.FileInfo
	var err error
	for e := range es {
		info, err = rpcc.GetFiles(es[e].Gid)
		if err != nil {
			fmt.Println(es[e].Gid, err)
			continue
		}
		fmt.Printf("\r(%s) %s completed.\n", es[e].Gid, info[0].Path)
	}
}
