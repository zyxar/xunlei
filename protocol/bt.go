package protocol

import (
	"encoding/json"
	"fmt"
)

type btList struct {
	Id       string     `json:"Tid"`
	InfoId   string     `json:"Infoid"`
	BtNum    string     `json:"btnum"`
	BtPerNum int        `json:"btpernum"`
	NowPage  int        `json:"now_page"`
	Record   []btRecord `json:"Record"`
}

type btRecord struct {
	Id           int    `json:"id"`
	FileName     string `json:"title"`
	Status       string `json:"download_status"`
	Cid          string `json:"cid"`
	SizeReadable string `json:"size"`
	Percent      int    `json:"percent"`
	TaskId       string `json:"taskid"`
	LiveTime     string `json:"livetime"`
	DownURL      string `json:"downurl"`
	FileSize     string `json:"filesize"`
	Verify       string `json:"verify"`
	URL          string `json:"url"`
	DirName      string `json:"dirtitle"`
}

type btFileList struct {
	Id    string `json:"id"`
	Size  string `json:"subsize"`
	Sizef string `json:"subformatsize"`
	Valid byte   `json:"valid"`
	Index string `json:"findex"`
	Name  string `json:"subtitle"`
	Ext   string `json:"ext"`
}

type btUploadResponse struct {
	Ret    int          `json:"ret_value"`
	InfoId string       `json:"infoid"`
	Name   string       `json:"ftitle"`
	Size   int          `json:"btsize"`
	IsFull string       `json:"is_full"`
	List   []btFileList `json:"filelist"`
	Random string       `json:"random"`
}

type btQueryResponse struct {
	InfoId string
	Size   string
	Name   string
	IsFull string
	Files  []string
	Sizesf []string
	Sizes  []string
	Picked []string
	Ext    []string
	Index  []string
	Random json.RawMessage
	Ret    json.RawMessage
}

func (response btQueryResponse) String() string {
	b, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return ""
	}
	return string(b)
}

func (t btList) String() string {
	r := fmt.Sprintf("%s %s %s\n", t.Id, t.InfoId, t.BtNum)
	for i := range t.Record {
		r += fmt.Sprintf("#%d %s %s %s\n", t.Record[i].Id, t.Record[i].FileName, t.Record[i].SizeReadable, t.Record[i].Status)
	}
	return r
}

type ErrorMessage struct {
	Code    json.RawMessage `json:"rtcode,omitempty"`
	Message string          `json:"msg,omitempty"`
}

// jsonp1457358333644({"progress":-12}) => -11/-12: need verify_code
// jsonp1457357084473({"id":"","avail_space":null,"progress":2,"time":0.47122287750244,"rtcode":"75","msg":"\u8be5\u8d44\u6e90\u88ab\u4e3e\u62a5\uff0c\u65e0\u6cd5\u6dfb\u52a0\u5230\u79bb\u7ebf\u7a7a\u95f4[0975]"})
// jsonp1457357441211({"id":"xxx","avail_space":"xxx","time":1.2841351032257,"progress":1}) => Success
type btSumbissionResponse struct {
	Id             string          `json:"id"`
	AvailableSpace json.RawMessage `json:"avail_space"`
	Time           json.RawMessage `json:"time"`
	Progress       int             `json:"progress"`
	*ErrorMessage
}
