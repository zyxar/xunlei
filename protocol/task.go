package protocol

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/apex/log"
	"github.com/zyxar/taipei"
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
	UserInfo userInfo `json:"userinfo"`
	// GlobalNew interface{} `json:"global_new"`
	// Time      interface{} `json:"time"` // v1: float64, v-current: string
}

type taskInfo struct {
	Tasks    []Task      `json:"tasks"`
	User     userAccount `json:"user"`
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

func (t *Task) FillBtList() (*btList, error) {
	if !t.IsBt() {
		return nil, errors.New("Not BT task.")
	}
	return FillBtList(t.Id, t.Cid)
}

func (t *Task) Remove() error {
	return t.remove(0)
}

func (t *Task) Purge() error {
	if t.deleted() {
		return t.remove(1)
	}
	err := t.remove(0)
	if err != nil {
		return err
	}
	return t.remove(1)
}

func (t *Task) remove(flag byte) error {
	var delType = t.status()
	if delType == flagInvalid {
		return errors.New("Invalid flag in task.")
	} else if delType == flagPurged {
		return errors.New("Task already purged.")
	} else if flag == 0 && delType == flagDeleted {
		return errors.New("Task already deleted.")
	}
	ct := currentTimestamp()
	uri := fmt.Sprintf(taskdeleteURI, ct, delType, ct)
	data := url.Values{}
	data.Add("taskids", t.Id+",")
	data.Add("databases", "0,")
	data.Add("interfrom", "task")
	r, err := defaultSession.post(uri, data.Encode())
	if err != nil {
		return err
	}
	if ok, _ := regexp.Match(`\{"result":1,"type":`, r); ok {
		log.Debugf("%s\n", r)
		if t.status() == flagDeleted {
			t.Flag = "2"
		} else {
			t.Flag = "1"
		}
		t.Progress = 0
		return nil
	}
	return errUnexpected
}

func (t *Task) Rename(name string) error {
	return defaultSession.renameTask(t.Id, name, t.TaskType)
}

func (t *Task) Pause() error {
	tids := t.Id + ","
	uri := fmt.Sprintf(taskpauseURI, tids, M.Uid, currentTimestamp())
	r, err := defaultSession.get(uri)
	if err != nil {
		return err
	}
	if bytes.Compare(r, []byte("pause_task_resp()")) != 0 {
		return errInvalidResponse
	}
	return nil
}

func (t *Task) Readd() error {
	if t.normal() {
		return errors.New("Task already in progress.")
	}
	if t.purged() {
		return defaultSession.addSimpleTask(t.URL)
	}
	return defaultSession.addSimpleTask(t.URL, t.Id)
}

func (t *Task) Resume() error {
	if t.expired() {
		return errTaskNoRedownCap
	}
	status := t.DownloadStatus
	if status != "5" && status != "3" {
		return errTaskNoRedownCap // only valid for `pending` and `failed` tasks
	}
	form := make([]string, 0, 3)
	v := url.Values{}
	v.Add("id[]", t.Id)
	v.Add("url[]", t.URL)
	v.Add("cid[]", t.Cid)
	v.Add("download_status[]", status)
	v.Add("taskname[]", t.TaskName)
	form = append(form, v.Encode())
	form = append(form, "type=1")
	form = append(form, "interfrom=task")
	uri := fmt.Sprintf(redownloadURI, currentTimestamp())
	r, err := defaultSession.post(uri, strings.Join(form, "&"))
	if err != nil {
		return err
	}
	log.Debugf("%s\n", r)
	return nil
}

func (t *Task) Delay() error {
	return DelayTask(t.Id)
}

func (t Task) GetVodURL() (lurl, hurl string, err error) {
	sid := defaultSession.getCookie("sessionid")
	v := url.Values{}
	v.Add("url", t.URL)
	v.Add("video_name", t.TaskName)
	v.Add("platform", "0")
	v.Add("userid", M.Uid)
	v.Add("vip", "1")
	v.Add("sessionid", sid)
	v.Add("gcid", t.GCid)
	v.Add("cid", t.Cid)
	v.Add("filesize", t.YsFileSize)
	v.Add("cache", strconv.FormatInt(currentTimestamp(), 10))
	v.Add("from", "lxweb")
	v.Add("jsonp", "XL_CLOUD_FX_INSTANCEqueryBack")
	uri := reqGetMethodVodURI + v.Encode()
	r, err := defaultSession.get(uri)
	if err != nil {
		return
	}
	exp := regexp.MustCompile(`XL_CLOUD_FX_INSTANCEqueryBack\((.*)\)`)
	var res struct {
		Resp vodResponse `json:"resp"`
	}
	s := exp.FindSubmatch(r)
	if s == nil {
		err = errInvalidResponse
		return
	}
	json.Unmarshal(s[1], &res)
	fmt.Printf("%+v\n", res.Resp)
	if res.Resp.Status == 0 { // TODO: also check `TransWait`
		for i := range res.Resp.VodList {
			if res.Resp.VodList[i].Spec == 225536 {
				lurl = res.Resp.VodList[i].URL
			} else if res.Resp.VodList[i].Spec == 282880 {
				hurl = res.Resp.VodList[i].URL
			}
		}
	} else {
		err = errors.New(res.Resp.ErrMsg)
	}
	return
}

func (t Task) Verify(path string) bool {
	if t.IsBt() {
		fmt.Println("Verifying [BT]", path)
		var b []byte
		var err error
		if b, err = GetTorrentByHash(t.Cid); err != nil {
			fmt.Println(err)
			return false
		}
		var m *taipei.MetaInfo
		if m, err = taipei.DecodeMetaInfo(b); err != nil {
			fmt.Println(err)
			return false
		}
		taipei.Iconv(m)
		taipei.SetEcho(true)
		g, err := taipei.VerifyContent(m, path)
		taipei.SetEcho(false)
		if err != nil {
			fmt.Println(err)
		}
		return g
	} else if strings.HasPrefix(t.URL, "ed2k://") {
		fmt.Println("Verifying [ED2K]", path)
		h, err := getEd2kHash(path)
		if err != nil {
			fmt.Println(err)
			return false
		}
		if !strings.EqualFold(h, getEd2kHashFromURL(t.URL)) {
			return false
		}
	}
	return true
}
