package protocol

import (
	"fmt"
	"strconv"
)

var (
	stats = []string{
		colorBgYellow + "waiting" + colorReset,
		colorBgMagenta + "downloading" + colorReset,
		colorBgGreen + "completed" + colorReset,
		colorBgRed + "failed" + colorReset,
		colorBgBlue + "pending" + colorReset,
		colorBgCyan + "expired" + colorReset,
	}
	coloring = []string{
		colorFrontYellow,
		colorFrontMagenta,
		colorFrontGreen,
		colorFrontRed,
		colorFrontCyan,
		colorFrontBlue,
		colorReset,
	}
)

type TaskCallback func(*Task) error

type taskResponse struct {
	Rtcode   int      `json:"rtcode"`
	Info     taskInfo `json:"info"`
	UserInfo UserInfo `json:"userinfo"`
	// GlobalNew interface{} `json:"global_new"`
	// Time      interface{} `json:"time"` // v1: float64, v-current: string
}

type taskInfo struct {
	Tasks    []Task      `json:"tasks"`
	User     UserAccount `json:"user"`
	ShowArc  int         `json:"show_arc"`
	TotalNum string      `json:"total_num"`
}

type taskPrepare struct {
	Cid        string
	GCid       string
	SizeCost   string
	FileName   string
	Goldbean   string
	Silverbean string
	// IsFull     string
	// Random     []byte
}

type ptaskRecord struct {
	Id             string  `json:"tid"`
	URL            string  `json:"url"`
	Speed          string  `json:"speed"`
	Progress       float32 `json:"fpercent"` // diff between 'percent' and 'fpercent'?
	LeaveTime      string  `json:"leave_time"`
	Size           string  `json:"fsize"`
	DownloadStatus string  `json:"download_status"`
	LixianURL      string  `json:"lixian_url"`
	LeftLiveTime   string  `json:"left_live_time"`
	TaskType       string  `json:"tasktype"`
	FileSize       string  `json:"filesize"`
}

type ptaskResponse struct {
	List []ptaskRecord `json:"Record"`
	Info struct {
		DownNum string `json:"downloading_num"`
		WaitNum string `json:"waiting_num"`
	} `json:"Task"`
}

type Task struct {
	Id   string `json:"id"`
	Flag string `json:"flag"`
	// Database            string `json:"database"`
	// ClassValue          string `json:"class_value"`
	GlobalId       string  `json:"global_id"`
	ResType        string  `json:"restype"`
	FileSize       string  `json:"filesize"`
	FileType       string  `json:"filetype"`
	Cid            string  `json:"cid"`
	GCid           string  `json:"gcid"`
	TaskName       string  `json:"taskname"`
	DownloadStatus string  `json:"download_status"`
	Speed          string  `json:"speed"`
	Progress       float32 `json:"progress"`
	// UsedTime            string `json:"used_time"`
	LeftLiveTime string `json:"left_live_time"`
	LixianURL    string `json:"lixian_url"`
	URL          string `json:"url"`
	ReferURL     string `json:"refer_url"`
	Cookie       string `json:"cookie"`
	// Vod                 string `json:"vod"`
	Status string `json:"status"`
	// Message             string `json:"message"`
	// DtCommitted         string `json:"dt_committed"`
	// DtDeleted           string `json:"dt_deleted"`
	// ListSum             string `json:"list_sum"`
	// FinishSum           string `json:"finish_sum"`
	// FlagKilledInASecond string `json:"flag_killed_in_a_second"`
	// ResCount            string `json:"res_count"`
	// UsingResCount       string `json:"using_res_count"`
	// VerifyFlag          string `json:"verify_flag"`
	// VerifyTime          string `json:"verify_time"`
	// ProgressText 			 string `json:"progress_text"`
	// ProgressImg         string `json:"progress_img"`
	// ProgressClass       string `json:"progress_class"`
	LeftTime string `json:"left_time"`
	UserId   int    `json:"userid"`
	// OpenFormat string `json:"openformat"`
	// TaskNameShow        string `json:"taskname_show"`
	// Ext                 string `json:"ext"`
	// ExtShow             string `json:"ext_show"`
	TaskType byte `json:"tasktype"`
	// FormatImg           string `json:"format_img"`
	// ResCountDegree      byte   `json:"res_count_degree"`
	YsFileSize string `json:"ysfilesize"`
	// BtMovie             byte   `json:"bt_movie"`
	UserType string `json:"user_type"`
}

func (t Task) Coloring() string {
	j, _ := strconv.Atoi(t.DownloadStatus)
	k, _ := strconv.Atoi(t.Flag)
	if k == 4 {
		j += k
	}
	status := stats[j]
	return fmt.Sprintf("%s%s %s %s %s%s %.1f%% %s%s", coloring[j], t.Id, t.TaskName, status, coloring[j], t.FileSize, t.Progress, trimHTMLFontTag(t.LeftLiveTime), colorReset)
}

func (t Task) String() string {
	return fmt.Sprintf("%s %s [%s] %s %.1f%% %s", t.Id, t.TaskName, t.DownloadStatus, t.FileSize, t.Progress, trimHTMLFontTag(t.LeftLiveTime))
}

func (t Task) Repr() string {
	j, _ := strconv.Atoi(t.DownloadStatus)
	k, _ := strconv.Atoi(t.Flag)
	if k == 4 {
		j += k
	}
	status := stats[j]
	ret := coloring[j] + t.Id + " " + t.TaskName + " " + status + coloring[j] + " " + t.FileSize + " " + trimHTMLFontTag(t.LeftLiveTime) + "\n"
	if t.Cid != "" {
		ret += t.Cid + " "
	}
	if t.GCid != "" {
		ret += t.GCid + "\n"
	}
	ret += t.URL
	if t.LixianURL != "" {
		ret += "\n" + t.LixianURL
	}
	return ret + colorReset
}

func (t Task) expired() bool {
	return t.status() == flagExpired
}

func (t Task) purged() bool {
	return t.status() == flagPurged
}

func (t Task) deleted() bool {
	return t.status() == flagDeleted
}

func (t Task) normal() bool {
	return t.status() == flagNormal
}

func (t Task) IsBt() bool {
	return t.TaskType == 0
}

func (t Task) waiting() bool {
	return t.DownloadStatus == "0"
}

func (t Task) completed() bool {
	return t.DownloadStatus == "2"
}

func (t Task) downloading() bool {
	return t.DownloadStatus == "1"
}

func (t Task) failed() bool {
	return t.DownloadStatus == "3"
}

func (t Task) pending() bool {
	return t.DownloadStatus == "5"
}

func (t Task) status() byte {
	switch len(t.Flag) {
	case 0:
		return flagNormal
	case 1:
		t := t.Flag[0] - '0'
		if t < 5 {
			return t
		}
	}
	return flagInvalid
}

func (t *Task) update(r *ptaskRecord) {
	if t.Id != r.Id {
		return
	}
	t.Speed = r.Speed
	t.Progress = r.Progress
	t.DownloadStatus = r.DownloadStatus
	t.LixianURL = r.LixianURL
}
