// http://vod.xunlei.com/js/list.js

package api

import (
	"encoding/json"
	// "fmt"
)

func GetHistoryPlayList(num int) *hist_resp {
	return nil
}

func SubmitBt(infohash string, num int) *subbt_resp {
	return nil
}

func QueryProgress() *progress_resp {
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
		Offset int `json:"offset"`
		Num    int `json:"req_num"`
		Type   int `json:"req_type"`
		Attr   int `json:"fileattribute"`
		Time   int `json:"t"`
	}
	domain := "http://xunlei.com"
	payload.UserInfo.Uid = getCookie(domain, "userid")
	payload.UserInfo.Name = getCookie(domain, "usrname")
	payload.UserInfo.NewNo = getCookie(domain, "usernewno")
	payload.UserInfo.VIP = getCookie(domain, "isvip")
	payload.UserInfo.Sid = getCookie(domain, "sessionid")
	payload.Offset = 0
	payload.Num = 30
	payload.Type = 2
	payload.Attr = 1
	payload.Time = current_timestamp()
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	v, err := post(LXTASK_LIST_URL, string(b))
	if err != nil {
		return nil, err
	}
	var resp lxtask_resp
	err = json.Unmarshal(v, &resp)
	return resp.List, err
}
