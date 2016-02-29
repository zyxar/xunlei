package protocol

import (
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
	Random string
	Ret    string
}

type subbtResponse struct {
	Uid  string `json:"userid"`
	Ret  byte   `json:"ret"`
	List []struct {
		GCid     string `json:"gcid"`
		UrlHash  string `json:"url_hash"`
		Name     string `json:"name"`
		Index    int    `json:"index"`
		Cid      string `json:"cid"`
		Size     int64  `json:"file_size"`
		Duration int64  `json:"duration"`
	} `json:"subfile_list"`
	UrlHash  string `json:"main_task_url_hash"`
	InfoHash string `json:"info_hash"`
	Num      int    `json:"record_num"`
}

func (t btList) String() string {
	r := fmt.Sprintf("%s %s %s\n", t.Id, t.InfoId, t.BtNum)
	for i := range t.Record {
		r += fmt.Sprintf("#%d %s %s %s\n", t.Record[i].Id, t.Record[i].FileName, t.Record[i].SizeReadable, t.Record[i].Status)
	}
	return r
}
