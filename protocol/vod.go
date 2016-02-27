package protocol

import (
	"encoding/json"
	"fmt"
)

// http://vod.xunlei.com/js/list.js

func GetHistoryPlayList() ([]VodHistTask, error) {
	uri := fmt.Sprintf(historyPlayURI, 30, 0, "all", "create", currentTimestamp()) //TODO: eliminate hard-code
	b, err := get(uri)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Data histResponse `json:"resp"`
	}
	err = json.Unmarshal(b, &resp)
	return resp.Data.List, nil
}

func SubmitBt(infohash string, num int) *subbtResponse {
	return nil
}

func QueryProgress() *progressResponse {
	return nil
}

func GetLxtaskList() ([]VodLXTask, error) {
	var payload struct {
		UserInfo struct {
			Uid   string `json:"userid"`
			Name  string `json:"name"`
			NewNo string `json:"newno"`
			VIP   string `json:"vip"`
			IP    string `json:"ip"`
			Sid   string `json:"sessionid"`
			From  int    `json:"from"`
		} `json:"user_info"`
		Offset int   `json:"offset"`
		Num    int   `json:"req_num"`
		Type   int   `json:"req_type"`
		Attr   int   `json:"fileattribute"`
		Time   int64 `json:"t"`
	}
	payload.UserInfo.Uid = getCookie("userid")
	payload.UserInfo.Name = getCookie("usrname")
	payload.UserInfo.NewNo = getCookie("usernewno")
	payload.UserInfo.VIP = getCookie("isvip")
	payload.UserInfo.Sid = getCookie("sessionid")
	payload.Offset = 0
	payload.Num = 30
	payload.Type = 2
	payload.Attr = 1
	payload.Time = currentTimestamp()
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	v, err := post(lxtaskListURI, string(b))
	if err != nil {
		return nil, err
	}
	var resp lxTaskResponse
	err = json.Unmarshal(v, &resp)
	return resp.List, err
}
