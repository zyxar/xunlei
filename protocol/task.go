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

func trim(raw string) string {
	exp := regexp.MustCompile(`<font color='([a-z]*)'>(.*)</font>`)
	s := exp.FindStringSubmatch(raw)
	if s == nil {
		return raw
	}
	return s[2]
}

func (t Task) Coloring() string {
	j, _ := strconv.Atoi(t.DownloadStatus)
	k, _ := strconv.Atoi(t.Flag)
	if k == 4 {
		j += k
	}
	status := stats[j]
	return fmt.Sprintf("%s%s %s %s %s%s %.1f%% %s%s", coloring[j], t.Id, t.TaskName, status, coloring[j], t.FileSize, t.Progress, trim(t.LeftLiveTime), colorReset)
}

func (t Task) String() string {
	return fmt.Sprintf("%s %s [%s] %s %.1f%% %s", t.Id, t.TaskName, t.DownloadStatus, t.FileSize, t.Progress, trim(t.LeftLiveTime))
}

func (t Task) Repr() string {
	j, _ := strconv.Atoi(t.DownloadStatus)
	k, _ := strconv.Atoi(t.Flag)
	if k == 4 {
		j += k
	}
	status := stats[j]
	ret := coloring[j] + t.Id + " " + t.TaskName + " " + status + coloring[j] + " " + t.FileSize + " " + trim(t.LeftLiveTime) + "\n"
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
	r, err := post(uri, data.Encode())
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
	return renameTask(t.Id, name, t.TaskType)
}

func (t *Task) Pause() error {
	tids := t.Id + ","
	uri := fmt.Sprintf(taskpauseURI, tids, M.Uid, currentTimestamp())
	r, err := get(uri)
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
		return addSimpleTask(t.URL)
	}
	return addSimpleTask(t.URL, t.Id)
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
	r, err := post(uri, strings.Join(form, "&"))
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
	sid := getCookie("sessionid")
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
	r, err := get(uri)
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

func (t _btList) String() string {
	r := fmt.Sprintf("%s %s %s/%d\n", t.Id, t.InfoId, t.BtNum, t.BtPerNum)
	for i := range t.Record {
		r += fmt.Sprintf("#%d %s %s %s\n", t.Record[i].Id, t.Record[i].FileName, t.Record[i].SizeReadable, t.Record[i].Status)
	}
	return r
}

func (t btList) String() string {
	r := fmt.Sprintf("%s %s %s\n", t.Id, t.InfoId, t.BtNum)
	for i := range t.Record {
		r += fmt.Sprintf("#%d %s %s %s\n", t.Record[i].Id, t.Record[i].FileName, t.Record[i].SizeReadable, t.Record[i].Status)
	}
	return r
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

func (t VodLXTask) String() string {
	name, _ := url.QueryUnescape(t.Name)
	return fmt.Sprintf("%s %s [%d] %dMB %d%% %dDays", t.Id, name, t.Status, t.Size/1024/1204, t.Progress/100, t.LeftTime/3600/24)
}

func (t VodHistTask) String() string {
	name, _ := url.QueryUnescape(t.Name)
	return fmt.Sprintf("%s %dMB %d", name, t.Size/1024/1204, t.Duration)
}
