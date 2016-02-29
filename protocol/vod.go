package protocol

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// http://vod.xunlei.com/js/list.js

type VodLXTask struct {
	Vod       int    `json:"vod"`
	GCid      string `json:"gcid"`
	Status    int    `json:"download_status"`
	SecDown   int    `json:"sec_down"`
	FinishNum int    `json:"finish_num"`
	TotalNum  int    `json:"total_num"`
	Cid       string `json:"cid"`
	URL       string `json:"url"`
	FileType  int    `json:"filetype"`
	ResType   int    `json:"restype"`
	Flag      int    `json:"flag"`
	Size      int64  `json:"filesize"`
	Id        string `json:"taskid"`
	Progress  int    `json:"progress"`
	Name      string `json:"taskname"`
	LixianURL string `json:"lixian_url"`
	LeftTime  int    `json:"left_live_time"`
}

type vodLxTaskResponse struct {
	List []VodLXTask `json:"tasklist"`
	Ret  string      `json:"ret"`
	User struct {
		From      int    `json:"from"`
		Name      string `json:"name"`
		IP        string `json:"ip"`
		NewNO     string `json:"newno"`
		Uid       string `json:"userid"`
		VIP       string `json:"vip"`
		Type      string `json:"user_type"`
		Sid       string `json:"sessionid"`
		AvalSpace string `json:"available_space"`
		MaxStore  string `json:"max_store"`
	} `json:"user_info"`
	Count    string `json:"task_total_cnt"`
	ErrMsg   string `json:"error_msg"`
	CountAll string `json:"task_total_cnt_all"`
}

type mbResponse struct {
	// url: http://i.vod.xunlei.com/miaobo/
	// infohash/507625D95D7695F9A1F91EC41A633D250419DF26?jsonp=jsonp1383446499423&t=0.7223064277786762
	InfoHash string `json:"infohash"`
	Result   byte   `json:"result"`
	Index    int    `json:"idx"`
}

type VodInfo struct {
	URLs   []string `json:"vod_urls"`
	Spec   int      `json:"spec_id"`
	URL    string   `json:"vod_url"`
	HasSub byte     `json:"has_subtitle"`
}

type vodResponse struct {
	Status    byte   `json:"status"`
	UrlHash   string `json:"url_hash"`
	TransWait int    `json:"tans_wait"`
	Uid       string `json:"userid"`
	Ret       byte   `json:"ret"`
	SrcInfo   struct {
		Name string `json:"file_name"`
		Cid  string `json:"cid"`
		Size string `json:"file_size"`
		GCid string `json:"gcid"`
	} `json:"src_info"`
	Duration  int64 `json:"duration"`
	VodPermit struct {
		Msg string `json:"msg"`
		Ret byte   `json:"ret"`
	} `json:"vod_permit"`
	ErrMsg  string    `json:"error_msg"`
	VodList []VodInfo `json:"vodinfo_list"`
}

type VodHistTask struct {
	GCid     string `json:"gcid"`
	UrlHash  string `json:"url_hash"`
	Cid      string `json:"cid"`
	URL      string `json:"url"`
	Name     string `json:"file_name"`
	Type     byte   `json:"tasktype"`
	Src      string `json:"src_url"`
	Size     int64  `json:"file_size"`
	Duration int    `json:"duration"`
	Played   string `json:"playtime"`
	Created  string `json:"createtime"`
}

type vodHistResponse struct {
	List   []VodHistTask `json:"history_play_list"`
	MaxNum int           `json:"max_num"`
	Uid    string        `json:"userid"`
	Ret    byte          `json:"ret"`
	Start  string        `json:"start_t"`
	End    string        `json:"end_t"`
	Num    int           `json:"record_num"`
	Type   string        `json:"type"`
}

func (t VodLXTask) String() string {
	name, _ := url.QueryUnescape(t.Name)
	return fmt.Sprintf("%s %s [%d] %dMB %d%% %dDays", t.Id, name, t.Status, t.Size/1024/1204, t.Progress/100, t.LeftTime/3600/24)
}

func (t VodHistTask) String() string {
	name, _ := url.QueryUnescape(t.Name)
	return fmt.Sprintf("%s %dMB %d", name, t.Size/1024/1204, t.Duration)
}

func GetHistoryPlayList() ([]VodHistTask, error) {
	uri := fmt.Sprintf(historyPlayURI, 30, 0, "all", "create", currentTimestamp()) //TODO: eliminate hard-code
	b, err := defaultSession.get(uri)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Data vodHistResponse `json:"resp"`
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
	payload.UserInfo.Uid = defaultSession.getCookie("userid")
	payload.UserInfo.Name = defaultSession.getCookie("usrname")
	payload.UserInfo.NewNo = defaultSession.getCookie("usernewno")
	payload.UserInfo.VIP = defaultSession.getCookie("isvip")
	payload.UserInfo.Sid = defaultSession.getCookie("sessionid")
	payload.Offset = 0
	payload.Num = 30
	payload.Type = 2
	payload.Attr = 1
	payload.Time = currentTimestamp()
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	v, err := defaultSession.post(lxtaskListURI, string(b))
	if err != nil {
		return nil, err
	}
	var resp vodLxTaskResponse
	err = json.Unmarshal(v, &resp)
	return resp.List, err
}
