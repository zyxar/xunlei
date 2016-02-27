package protocol

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/apex/log"
	"github.com/apex/log/handlers/text"
)

var (
	errNoSuchTask       = errors.New("No such TaskId in list")
	errInvalidResponse  = errors.New("Invalid response")
	errUnexpected       = errors.New("Unexpected error")
	errTaskNotCompleted = errors.New("Task not completed")
	errInvalidLogin     = errors.New("Invalid login account")
	errLoginFailed      = errors.New("Login failed")
	errReuseSession     = errors.New("Previous session exipred")
	errBtTaskAlready    = errors.New("Bt task already exists")
	errTaskNoRedownCap  = errors.New("Task not capable for restart")
	errTimeout          = errors.New("Request time out")
	defaultConn         struct {
		*http.Client
		sync.Mutex
		timeout time.Duration
	}
	urlXunleiCom                *url.URL
	urlDynamicCloudVipXunleiCom *url.URL
	// urlLoginXunleiCom           *url.URL
	// urlVipXunleiCom             *url.URL
)

func init() {
	jar, _ := cookiejar.New(nil)
	urlXunleiCom, _ = url.Parse("http://xunlei.com")
	urlDynamicCloudVipXunleiCom, _ = url.Parse("http://dynamic.cloud.vip.xunlei.com")
	// urlLoginXunleiCom, _ = url.Parse("http://login.xunlei.com")
	// urlVipXunleiCom, _ = url.Parse("http://vip.xunlei.com")
	defaultConn.timeout = 3000 * time.Millisecond
	defaultConn.Client = &http.Client{
		Jar: jar,
		Transport: &http.Transport{
			Dial:  (&net.Dialer{Timeout: defaultConn.timeout}).Dial,
			Proxy: http.ProxyFromEnvironment,
		},
	}
	log.SetHandler(text.New(os.Stderr))
	log.SetLevel(log.InfoLevel)
}

func routine(req *http.Request) (*http.Response, error) {
	timeout := false
retry:
	defaultConn.Lock()
	timer := time.AfterFunc(defaultConn.timeout, func() {
		defaultConn.Client.Transport.(*http.Transport).CancelRequest(req)
		timeout = true
	})
	resp, err := defaultConn.Do(req)
	if timer != nil {
		timer.Stop()
	}
	defaultConn.Unlock()
	if err == io.EOF && !timeout {
		goto retry
	}
	if timeout {
		err = errTimeout
	}
	return resp, err
}

func get(dest string) ([]byte, error) {
	log.Debugf("==> %s", dest)
	req, err := http.NewRequest("GET", dest, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", userAgent)
	req.Header.Add("Accept-Encoding", "gzip, deflate")
	resp, err := routine(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	log.Info(resp.Status)
	if resp.StatusCode/100 > 3 {
		return nil, errors.New(resp.Status)
	}
	return readBody(resp)
}

func post(dest string, data string) ([]byte, error) {
	log.Debugf("==> %s", dest)
	req, err := http.NewRequest("POST", dest, strings.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("User-Agent", userAgent)
	req.Header.Add("Accept-Encoding", "gzip, deflate")
	resp, err := routine(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	log.Info(resp.Status)
	if resp.StatusCode/100 > 3 {
		return nil, errors.New(resp.Status)
	}
	return readBody(resp)
}

func Login(id, passhash string) (err error) {
	var vcode string
	if len(id) == 0 {
		err = errInvalidLogin
		return
	}
	loginURL := fmt.Sprintf("http://login.xunlei.com/check?u=%s&cachetime=%d", id, currentTimestamp())
	u, _ := url.Parse("http://xunlei.com/")
loop:
	if _, err = get(loginURL); err != nil {
		return
	}
	cks := defaultConn.Client.Jar.Cookies(u)
	for i := range cks {
		if cks[i].Name == "check_result" {
			if len(cks[i].Value) < 3 {
				goto loop
			}
			vcode = cks[i].Value[2:]
			vcode = strings.ToUpper(vcode)
			log.Infof("verify_code: %s", vcode)
			break
		}
	}
	v := url.Values{}
	v.Set("u", id)
	v.Set("p", hashPass(passhash, vcode))
	v.Set("verifycode", vcode)
	if _, err = post("http://login.xunlei.com/sec2login/", v.Encode()); err != nil {
		return
	}
	M.Uid = getCookie("userid")
	log.Infof("uid: %s\n", M.Uid)
	if len(M.Uid) == 0 {
		err = errLoginFailed
		return
	}
	var r []byte
	if r, err = get(fmt.Sprintf("%slogin?cachetime=%d&from=0", domainLixianURI, currentTimestamp())); err != nil || len(r) < 512 {
		err = errUnexpected
	}
	return
}

func SaveSession(cookieFile string) error {
	session := [][]*http.Cookie{
		defaultConn.Client.Jar.Cookies(urlXunleiCom),
		defaultConn.Client.Jar.Cookies(urlDynamicCloudVipXunleiCom)}
	r, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(cookieFile, r, 0644)
}

func ResumeSession(cookieFile string) (err error) {
	if cookieFile != "" {
		var data []byte
		data, err = ioutil.ReadFile(cookieFile)
		if err != nil {
			return
		}
		var session [][]*http.Cookie
		if err = json.Unmarshal(data, &session); err != nil {
			return
		}
		if len(session) < 2 {
			err = errors.New("invalid session")
			return
		}
		defaultConn.Client.Jar.SetCookies(urlXunleiCom, session[0])
		defaultConn.Client.Jar.SetCookies(urlDynamicCloudVipXunleiCom, session[1])
	}
	if !IsOn() {
		err = errReuseSession
	}
	return
}

func verifyLogin() bool {
	r, err := get(verifyLoginURI)
	if err != nil {
		log.Error(err.Error())
		return false
	}
	exp := regexp.MustCompile(`.*\((\{.*\})\)`)
	s := exp.FindSubmatch(r)
	if s == nil {
		log.Debugf("Response: %s\n", r)
		return false
	}
	var resp loginResponse
	json.Unmarshal(s[1], &resp)
	if resp.Result == 0 {
		log.Debugf("Response: %s\n", s[1])
		return false
	}
	fmt.Printf(`
-------------+-------------
  user id    |  %s
  user name  |  %s
  user newno |  %s
  user type  |  %s
  nickname   |  %s
  user nick	 |  %s
  vip state	 |  %s
-------------+-------------
`, resp.Data.UserId, resp.Data.UserName,
		resp.Data.UserNewno, resp.Data.UserType,
		resp.Data.Nickname, resp.Data.UserNick, resp.Data.VipState)
	return true
}

func IsOn() bool {
	uid := getCookie("userid")
	if len(uid) == 0 {
		return false
	}
	r, err := get(fmt.Sprintf(taskHomeURI, uid))
	if err != nil {
		return false
	}
	if ok, _ := regexp.Match(`top.location='http://cloud.vip.xunlei.com/task.html\?error=`, r); ok {
		return false
	}
	if len(M.Uid) == 0 {
		M.Uid = uid
	}
	return true
}

func getCookie(name string) string {
	cks := defaultConn.Client.Jar.Cookies(urlXunleiCom)
	for i := range cks {
		if cks[i].Name == name {
			return cks[i].Value
		}
	}
	return ""
}

func GetTasks() ([]*Task, error) {
	accumulated := 0
	page := 1
	var ts []*Task
round:
	b, err := tasklistNofresh(statusMixed, page)
	if err != nil {
		return ts, err
	}
	var resp taskResponse
	err = json.Unmarshal(b, &resp)
	if err != nil {
		return ts, err
	}
	total, _ := strconv.Atoi(resp.Info.User.TotalNum)
	if ts == nil {
		if total <= 0 {
			ts = make([]*Task, 0, len(resp.Info.Tasks))
		} else {
			ts = make([]*Task, 0, total)
		}
	}
	for i := range resp.Info.Tasks {
		resp.Info.Tasks[i].TaskName = unescapeName(resp.Info.Tasks[i].TaskName)
		ts = append(ts, &resp.Info.Tasks[i])
	}
	accumulated += len(ts)
	if accumulated < total {
		page++
		goto round
	}
	M.InvalidateGroup(flagNormal)
	M.pushTasks(ts)
	return ts, err
}

func GetCompletedTasks() ([]*Task, error) {
	accumulated := 0
	page := 1
	var ts []*Task
round:
	b, err := tasklistNofresh(statusCompleted, page)
	if err != nil {
		return ts, err
	}
	var resp taskResponse
	err = json.Unmarshal(b, &resp)
	if err != nil {
		return ts, err
	}
	total, _ := strconv.Atoi(resp.Info.User.TotalNum)
	if ts == nil {
		if total <= 0 {
			ts = make([]*Task, 0, len(resp.Info.Tasks))
		} else {
			ts = make([]*Task, 0, total)
		}
	}
	for i := range resp.Info.Tasks {
		resp.Info.Tasks[i].TaskName = unescapeName(resp.Info.Tasks[i].TaskName)
		ts = append(ts, &resp.Info.Tasks[i])
	}
	accumulated += len(ts)
	if accumulated < total {
		page++
		goto round
	}
	M.pushTasks(ts)
	return ts, err
}

func GetIncompletedTasks() ([]*Task, error) {
	accumulated := 0
	page := 1
	var ts []*Task
round:
	b, err := tasklistNofresh(statusDownloading, page)
	if err != nil {
		return nil, err
	}
	var resp taskResponse
	err = json.Unmarshal(b, &resp)
	if err != nil {
		return nil, err
	}
	total, _ := strconv.Atoi(resp.Info.User.TotalNum)
	if ts == nil {
		if total <= 0 {
			ts = make([]*Task, 0, len(resp.Info.Tasks))
		} else {
			ts = make([]*Task, 0, total)
		}
	}
	for i := range resp.Info.Tasks {
		resp.Info.Tasks[i].TaskName = unescapeName(resp.Info.Tasks[i].TaskName)
		ts = append(ts, &resp.Info.Tasks[i])
	}
	accumulated += len(ts)
	if accumulated < total {
		page++
		goto round
	}
	M.pushTasks(ts)
	return ts, err
}

func GetGdriveId() (gid string, err error) {
	if len(M.Gid) == 0 {
		var b []byte
		b, err = tasklistNofresh(statusMixed, 1)
		if err != nil {
			return
		}
		var resp taskResponse
		err = json.Unmarshal(b, &resp)
		if err != nil {
			return
		}
		M.Gid = resp.Info.User.Cookie
		M.Account = &resp.Info.User
		M.AccountInfo = &resp.UserInfo
	}
	gid = M.Gid
	log.Debugf("gdriveid: %s", gid)
	return
}

// bellow three funcs are for RESTful calling;
// we do not parse content here.
func RawTaskList(category, page int) ([]byte, error) {
	return tasklistNofresh(category, page)
}

func RawTaskListExpired() ([]byte, error) {
	return readExpired()
}

func RawTaskListDeleted(page int) ([]byte, error) {
	return readHistory(page)
}

func tasklistNofresh(tid, page int) ([]byte, error) {
	/*
		tid:
		1 downloading
		2 completed
		4 downloading|completed|expired
		11 deleted - not used now?
		13 expired - not used now?
	*/
	if tid != 4 && tid != 1 && tid != 2 {
		tid = 4
	}
	uri := fmt.Sprintf(showtaskUnfreshURI, tid, page, pageSize, page)
	log.Debugf("==> %s", uri)
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", userAgent)
	req.Header.Add("Accept-Encoding", "gzip, deflate")
	req.AddCookie(&http.Cookie{Name: "pagenum", Value: pageSize})
	resp, err := routine(req)
	if err != nil {
		return nil, err
	}
	log.Info(resp.Status)
	r, err := readBody(resp)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}
	exp := regexp.MustCompile(`rebuild\((\{.*\})\)`)
	s := exp.FindSubmatch(r)
	if s == nil {
		return nil, errInvalidResponse
	}
	return s[1], nil
}

func readExpired() ([]byte, error) {
	uri := fmt.Sprintf(expireHomeURI, M.Uid)
	log.Debugf("==> %s", uri)
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", userAgent)
	req.Header.Add("Accept-Encoding", "gzip, deflate")
	req.AddCookie(&http.Cookie{Name: "lx_nf_all", Value: url.QueryEscape(expiredCk)})
	req.AddCookie(&http.Cookie{Name: "pagenum", Value: pageSize})
	resp, err := routine(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	log.Info(resp.Status)
	return readBody(resp)
}

func GetExpiredTasks() ([]*Task, error) {
	r, err := readExpired()
	ts, _ := parseHistory(r, "4")
	M.InvalidateGroup(flagExpired)
	M.pushTasks(ts)
	return ts, err
}

func GetDeletedTasks() ([]*Task, error) {
	j := 0
	next := true
	var err error
	var r []byte
	var ts []*Task
	tss := make([]*Task, 0, 10)
	for next {
		j++
		r, err = readHistory(j)
		ts, next = parseHistory(r, "1")
		tss = append(tss, ts...)
	}
	M.InvalidateGroup(flagDeleted)
	M.InvalidateGroup(flagPurged)
	M.pushTasks(tss)
	return tss, err
}

func readHistory(page int) ([]byte, error) {
	var uri string
	if page > 0 {
		uri = fmt.Sprintf(historyPageURI, M.Uid, page)
	} else {
		uri = fmt.Sprintf(historyHomeURI, M.Uid)
	}

	log.Debugf("==> %s", uri)
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", userAgent)
	req.Header.Add("Accept-Encoding", "gzip, deflate")
	req.AddCookie(&http.Cookie{Name: "lx_nf_all", Value: url.QueryEscape(deletedCk)})
	req.AddCookie(&http.Cookie{Name: "pagenum", Value: pageSize})
	resp, err := routine(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	log.Info(resp.Status)
	return readBody(resp)
}

func parseHistory(in []byte, ty string) ([]*Task, bool) {
	es := `<input id="d_status(\d+)"[^<>]+value="(.*)" />\s+<input id="dflag\d+"[^<>]+value="(.*)" />\s+<input id="dcid\d+"[^<>]+value="(.*)" />\s+<input id="f_url\d+"[^<>]+value="(.*)" />\s+<input id="taskname\d+"[^<>]+value="(.*)" />\s+<input id="d_tasktype\d+"[^<>]+value="(.*)" />`
	exp := regexp.MustCompile(es)
	s := exp.FindAllSubmatch(in, -1)
	ret := make([]*Task, len(s))
	for i := range s {
		b, _ := strconv.Atoi(string(s[i][7]))
		ret[i] = &Task{Id: string(s[i][1]), DownloadStatus: string(s[i][2]), Cid: string(s[i][4]), URL: string(s[i][5]), TaskName: unescapeName(string(s[i][6])), TaskType: byte(b), Flag: ty}
	}
	exp = regexp.MustCompile(`<li class="next"><a href="([^"]+)">[^<>]*</a></li>`)
	return ret, exp.FindSubmatch(in) != nil
}

func DelayTask(taskid string) error {
	uri := fmt.Sprintf(taskdelayURI, taskid+"_1", "task", currentTimestamp())
	r, err := get(uri)
	if err != nil {
		return err
	}
	exp := regexp.MustCompile(`^task_delay_resp\((.*}),\[.*\]\)`)
	s := exp.FindSubmatch(r)
	if s == nil {
		return errInvalidResponse
	}
	var resp struct {
		K struct {
			Llt string `json:"left_live_time"`
		} `json:"0"`
		Result byte `json:"result"`
	}
	json.Unmarshal(s[1], &resp)
	log.Infof("%s: %s\n", taskid, resp.K.Llt)
	return nil
}

func redownload(tasks []*Task) error {
	form := make([]string, 0, len(tasks)+2)
	for i := range tasks {
		if tasks[i].expired() || !tasks[i].failed() || !tasks[i].pending() {
			continue
		}
		v := url.Values{}
		v.Add("id[]", tasks[i].Id)
		v.Add("url[]", tasks[i].URL)
		v.Add("cid[]", tasks[i].Cid)
		v.Add("download_status[]", tasks[i].DownloadStatus)
		v.Add("taskname[]", tasks[i].TaskName)
		form = append(form, v.Encode())
	}
	if len(form) == 0 {
		return errors.New("No tasks need to restart.")
	}
	form = append(form, "type=1")
	form = append(form, "interfrom=task")
	uri := fmt.Sprintf(redownloadURI, currentTimestamp())
	r, err := post(uri, strings.Join(form, "&"))
	if err != nil {
		return err
	}
	log.Infof("%s\n", r)
	return nil
}

func FillBtList(taskid, infohash string) (*btList, error) {
	var pgsize = btPageSize
retry:
	m, err := fillBtList(taskid, infohash, 1, pgsize)
	if err == io.ErrUnexpectedEOF && pgsize == btPageSize {
		pgsize = "100"
		goto retry
	}
	if err != nil {
		return nil, err
	}
	var list = btList{}
	list.BtNum = m.BtNum
	list.Id = m.Id
	list.InfoId = m.InfoId
	if len(m.Record) > 0 {
		list.Record = append(list.Record, m.Record...)
	}
	total, _ := strconv.Atoi(list.BtNum)
	size, _ := strconv.Atoi(pgsize)
	pageNum := total/size + 1
	next := 2
	for next <= pageNum {
		m, err = fillBtList(taskid, infohash, next, pgsize)
		if err == nil {
			if len(m.Record) > 0 {
				list.Record = append(list.Record, m.Record...)
			}
			next++
		} else {
			log.Error("err in fillBtList()")
		}
	}
	return &list, nil
}

func RawFillBtList(taskid, infohash string, page int) ([]byte, error) {
	var pgsize = btPageSize
retry:
	uri := fmt.Sprintf(fillbtlistURI, taskid, infohash, page, M.Uid, "task", currentTimestamp())
	log.Debugf("==> %s", uri)
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", userAgent)
	req.Header.Add("Accept-Encoding", "gzip, deflate")
	req.AddCookie(&http.Cookie{Name: "pagenum", Value: pgsize})
	resp, err := routine(req)
	if err != nil {
		return nil, err
	}
	log.Info(resp.Status)
	defer resp.Body.Close()
	r, err := readBody(resp)
	if err == io.ErrUnexpectedEOF && pgsize == btPageSize {
		pgsize = "100"
		goto retry
	}
	return r, nil
}

func fillBtList(taskid, infohash string, page int, pgsize string) (*_btList, error) {
	uri := fmt.Sprintf(fillbtlistURI, taskid, infohash, page, M.Uid, "task", currentTimestamp())
	log.Debugf("==> %s", uri)
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", userAgent)
	req.Header.Add("Accept-Encoding", "gzip, deflate")
	req.AddCookie(&http.Cookie{Name: "pagenum", Value: pgsize})
	resp, err := routine(req)
	if err != nil {
		return nil, err
	}
	log.Info(resp.Status)
	r, err := readBody(resp)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}
	exp := regexp.MustCompile(`fill_bt_list\({"Result":(.*)}\)`)
	s := exp.FindSubmatch(r)
	if s == nil {
		exp = regexp.MustCompile(`alert\('(.*)'\);.*`)
		s = exp.FindSubmatch(r)
		if s != nil {
			return nil, errors.New(string(s[1]))
		}
		return nil, errInvalidResponse
	}
	var btList _btList
	json.Unmarshal(s[1], &btList)
	exp = regexp.MustCompile(`\\`)
	for i := range btList.Record {
		btList.Record[i].FileName = exp.ReplaceAllLiteralString(btList.Record[i].FileName, `/`)
		btList.Record[i].FileName = unescapeName(btList.Record[i].FileName)
	}
	return &btList, nil
}

// supported uri schemes:
// 'ed2k', 'http', 'https', 'ftp', 'bt', 'magnet', 'thunder', 'Flashget', 'qqdl'
func AddTask(req string) error {
	ttype := taskTypeOrdinary
	if strings.HasPrefix(req, "magnet:") || strings.Contains(req, "get_torrent?userid=") {
		ttype = taskTypeMagnet
	} else if strings.HasPrefix(req, "ed2k://") {
		ttype = taskTypeEd2k
	} else if strings.HasPrefix(req, "bt://") || strings.HasSuffix(req, ".torrent") {
		ttype = taskTypeBt
	} else if ok, _ := regexp.MatchString(`^[a-zA-Z0-9]{40,40}$`, req); ok {
		ttype = taskTypeBt
		req = "bt://" + req
	} else if hasScheme, _ := regexp.Match(`://`, []byte(req)); !hasScheme {
		ttype = taskTypeBt
	}
	switch ttype {
	case taskTypeOrdinary, taskTypeEd2k:
		return addSimpleTask(req)
	case taskTypeBt:
		return addBtTask(req)
	case taskTypeMagnet:
		return addMagnetTask(req)
	case taskTypeInvalid:
		fallthrough
	default:
	}
	return errUnexpected
}

func AddBatchTasks(urls []string, oids ...string) error {
	// TODO: filter urls
	v := url.Values{}
	for i := 0; i < len(urls); i++ {
		v.Add("cid[]", "")
		v.Add("url[]", url.QueryEscape(urls[i]))
	}
	v.Add("class_id", "0")
	if len(oids) > 0 {
		var b bytes.Buffer
		for i := 0; i < len(oids); i++ {
			b.WriteString("0,")
		}
		v.Add("batch_old_taskid", strings.Join(oids, ","))
		v.Add("batch_old_database", b.String())
		v.Add("interfrom", "history")
	} else {
		v.Add("batch_old_taskid", "0,")
		v.Add("batch_old_database", "0,")
		v.Add("interfrom", "task")
	}
	tm := currentTimestamp()
	uri := fmt.Sprintf(batchtaskcommitURI, tm, tm)
	r, err := post(uri, v.Encode())
	fmt.Printf("%s\n", r)
	return err
}

func addSimpleTask(uri string, oid ...string) error {
	var from string
	if len(oid) > 0 {
		from = "history"
	} else {
		from = "task"
	}
	exp := regexp.MustCompile(`%2C|,`)
	uri = exp.ReplaceAllLiteralString(uri, `.`)
	dest := fmt.Sprintf(taskcheckURI, url.QueryEscape(uri), from, currentRandom(), currentTimestamp())
	r, err := get(dest)
	if err == nil {
		taskPre, err := getTaskPre(r)
		if err != nil {
			return err
		}
		var taskType string
		if strings.HasPrefix(uri, "ed2k://") {
			taskType = strconv.Itoa(taskTypeEd2k)
		} else {
			taskType = strconv.Itoa(taskTypeOrdinary)
			// strings.HasPrefix(uri, "http://") || strings.HasPrefix(uri, "ftp://") || strings.HasPrefix(uri, "https://")
			// return errors.New("Invalid protocol scheme.")
		}
		v := url.Values{}
		v.Add("callback", "ret_task")
		v.Add("uid", M.Uid)
		v.Add("cid", taskPre.Cid)
		v.Add("gcid", taskPre.GCid)
		v.Add("size", taskPre.SizeCost)
		v.Add("goldbean", taskPre.Goldbean)
		v.Add("silverbean", taskPre.Silverbean)
		v.Add("t", taskPre.FileName)
		v.Add("url", uri)
		v.Add("type", taskType)
		if len(oid) > 0 {
			v.Add("o_taskid", oid[0])
			v.Add("o_page", "history")
		} else {
			v.Add("o_page", "task")
			v.Add("o_taskid", "0")
		}
		dest = taskcommitURI + v.Encode()
		if r, err = get(dest); err != nil {
			return err
		}
		if ok, _ := regexp.Match(`ret_task\(.*\)`, r); ok {
			return nil
		}
		return errInvalidResponse
	}
	return err
}

func addBtTask(uri string) error {
	if strings.HasPrefix(uri, "bt://") {
		return addMagnetTask(fmt.Sprintf(gettorrentURI, M.Uid, uri[5:]))
	}
	return addTorrentTask(uri)
}

func addMagnetTask(link string, oid ...string) error {
	uri := fmt.Sprintf(urlqueryURI, url.QueryEscape(link), currentRandom())
	r, err := get(uri)
	if err != nil {
		return err
	}
	exp := regexp.MustCompile(`queryUrl\((1,.*)\)`)
	s := exp.FindSubmatch(r)
	if s == nil {
		if ok, _ := regexp.Match(`queryUrl\(-1,'[0-9A-Za-z]{40,40}'.*`, r); ok {
			return errBtTaskAlready
		}
		return errInvalidResponse
	}
	if task := evalParse(s[1]); task != nil {
		v := url.Values{}
		v.Add("uid", M.Uid)
		v.Add("btname", task.Name)
		v.Add("cid", task.InfoId)
		v.Add("tsize", task.Size)
		findex := strings.Join(task.Index, "_")
		size := strings.Join(task.Sizes, "_")
		v.Add("findex", findex)
		v.Add("size", size)
		if len(oid) > 0 {
			v.Add("from", "history")
			v.Add("o_taskid", oid[0])
			v.Add("o_page", "history")
		} else {
			v.Add("from", "task")
		}
		dest := fmt.Sprintf(bttaskcommitURI, currentTimestamp())
		r, err = post(dest, v.Encode())
		exp = regexp.MustCompile(`jsonp.*\(\{"id":"(\d+)","avail_space":"\d+".*\}\)`)
		s = exp.FindSubmatch(r)
		if s == nil {
			return errInvalidResponse
		}
	} else {
		return errInvalidResponse
	}
	return nil
}

func addTorrentTask(filename string) (err error) {
	var file *os.File
	if file, err = os.Open(filename); err != nil {
		return
	}
	defer file.Close()
	// if _, err = taipei.GetMetaInfo(filename); err != nil {
	//  return
	// }
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	var part io.Writer
	if part, err = writer.CreateFormFile("filepath", filename); err != nil {
		return
	}
	io.Copy(part, file)
	writer.WriteField("random", currentRandom())
	writer.WriteField("interfrom", "task")

	dest := torrentuploadURI
	log.Debugf("==> %s", dest)
	req, err := http.NewRequest("POST", dest, bytes.NewReader(body.Bytes()))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Add("User-Agent", userAgent)
	req.Header.Add("Accept-Encoding", "gzip, deflate")
	resp, err := routine(req)
	if err != nil {
		return
	}
	log.Info(resp.Status)
	r, err := readBody(resp)
	resp.Body.Close()
	exp := regexp.MustCompile(`<script>document\.domain="xunlei\.com";var btResult =(\{.+\});(var btRtcode = 0)*</script>`)
	s := exp.FindSubmatch(r)
	if s != nil {
		var result btupResult
		json.Unmarshal(s[1], &result)
		v := url.Values{}
		v.Add("uid", M.Uid)
		v.Add("btname", result.Name) // TODO: filter illegal char
		v.Add("cid", result.InfoId)
		v.Add("tsize", strconv.Itoa(result.Size))
		findex := make([]string, 0, len(result.List))
		size := make([]string, 0, len(result.List))
		for i := 0; i < len(result.List); i++ {
			findex = append(findex, result.List[i].Id)
			size = append(size, result.List[i].Size)
		}
		v.Add("findex", strings.Join(findex, "_"))
		v.Add("size", strings.Join(size, "_"))
		v.Add("from", "0")
		dest = fmt.Sprintf(bttaskcommitURI, currentTimestamp())
		r, err = post(dest, v.Encode())
		exp = regexp.MustCompile(`jsonp.*\(\{"id":"(\d+)","avail_space":"\d+".*\}\)`)
		s = exp.FindSubmatch(r)
		if s == nil {
			fmt.Printf("%s\n", r)
			return errInvalidResponse
		}
		// tasklistNofresh(4, 1)
		// FillBtList(string(s[1]))
		return nil
	}
	exp = regexp.MustCompile(`parent\.edit_bt_list\((\{.*\}),'`)
	s = exp.FindSubmatch(r)
	if s == nil {
		return errors.New("Add bt task failed.")
	}
	// var result _btup_result
	// json.Unmarshal(s[1], &result)
	return errBtTaskAlready
}

func ProcessTaskDaemon(ch chan byte, callback func(*Task)) {
	if len(M.Tasks) == 0 {
		GetIncompletedTasks()
	}

	go func() {
		for {
			select {
			case <-ch:
				err := processTask(M.Tasks, callback)
				if err != nil {
					log.Errorf("error in ProcessTask(): %v", err)
				}
			case <-time.After(60 * time.Second):
				err := processTask(M.Tasks, callback)
				if err != nil {
					log.Errorf("error in ProcessTask(): %v", err)
					time.Sleep(5 * time.Second)
					ch <- 1
				}
			}
		}
	}()
}

func ProcessTask(callback func(*Task)) error {
	return processTask(M.Tasks, callback)
}

func processTask(tasks map[string]*Task, callback func(*Task)) error {
	l := len(tasks)
	if l == 0 {
		return errors.New("No tasks in progress.")
	}
	ct := currentTimestamp()
	uri := fmt.Sprintf(taskprocessURI, ct, ct)
	v := url.Values{}
	list := make([]string, 0, l)
	nmList := make([]string, 0, l)
	btList := make([]string, 0, l)
	for i := range tasks {
		if tasks[i].status() == flagNormal && tasks[i].DownloadStatus == "1" {
			list = append(list, tasks[i].Id)
			if tasks[i].TaskType == 0 {
				btList = append(btList, tasks[i].Id)
			} else {
				nmList = append(nmList, tasks[i].Id)
			}
		}
	}
	v.Add("list", strings.Join(list, ","))
	v.Add("nm_list", strings.Join(nmList, ","))
	v.Add("bt_list", strings.Join(btList, ","))
	v.Add("uid", M.Uid)
	v.Add("interfrom", "task")
	var r []byte
	var err error
	if r, err = post(uri, v.Encode()); err != nil {
		return err
	}
	exp := regexp.MustCompile(`jsonp\d+\(\{"Process":(.*)\}\)`)
	s := exp.FindSubmatch(r)
	if s == nil {
		return errInvalidResponse
	}
	var res ptaskResponse
	err = json.Unmarshal(s[1], &res)
	if err != nil {
		return err
	}
	for i := range res.List {
		task := tasks[res.List[i].Id]
		task.update(&res.List[i])
		if callback != nil {
			callback(task)
		}
	}
	return nil
}

func GetTorrentByHash(hash string) ([]byte, error) {
	uri := fmt.Sprintf(gettorrentURI, M.Uid, strings.ToUpper(hash))
	r, err := get(uri)
	if err != nil {
		return nil, err
	}
	exp := regexp.MustCompile(`alert\('(.*)'\)`)
	s := exp.FindSubmatch(r)
	if s != nil {
		log.Infof("%s\n", s[1])
		return nil, errInvalidResponse
	}
	return r, nil
}

func GetTorrentFileByHash(hash, file string) error {
	if stat, err := os.Stat(file); err == nil || stat != nil {
		return errors.New("Target file already exists.")
	}
	r, err := GetTorrentByHash(hash)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(file, r, 0644)
}

func PauseTasks(ids []string) error {
	tids := strings.Join(ids, ",")
	tids += ","
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

func DelayAllTasks() error {
	r, err := get(delayonceURI)
	if err != nil {
		return err
	}
	log.Infof("%s\n", r)
	return nil
}

func ReAddTasks(ts map[string]*Task) {
	nbt := make([]*Task, 0, len(ts))
	bt := make([]*Task, 0, len(ts))
	for i := range ts {
		if ts[i].expired() || ts[i].deleted() {
			if ts[i].IsBt() {
				bt = append(bt, ts[i])
			} else {
				nbt = append(nbt, ts[i])
			}
		}
	}
	if len(nbt) == 1 {
		if err := nbt[0].Readd(); err != nil {
			log.Error(err.Error())
		}
	} else if len(nbt) > 1 {
		urls, ids := extractTasks(nbt)
		if err := AddBatchTasks(urls, ids...); err != nil {
			log.Error(err.Error())
		}
	}
	for i := range bt {
		if err := addMagnetTask(fmt.Sprintf(gettorrentURI, M.Uid, bt[i].Cid), bt[i].Id); err != nil {
			log.Error(err.Error())
		}
	}
}

func RenameTask(taskid, newname string) error {
	t := M.getTaskbyId(taskid)
	if t == nil {
		return errNoSuchTask
	}
	return renameTask(taskid, newname, t.TaskType)
}

func renameTask(taskid, newname string, tasktype byte) error {
	v := url.Values{}
	v.Add("taskid", taskid)
	if tasktype == 0 {
		v.Add("bt", "1")
	} else {
		v.Add("bt", "0")
	}
	v.Add("filename", newname)
	r, err := get(renameURI + v.Encode())
	if err != nil {
		return err
	}
	var resp struct {
		Result   int    `json:"result"`
		TaskId   int    `json:"taskid"`
		FileName string `json:"filename"`
	}
	json.Unmarshal(r[1:len(r)-1], &resp)
	if resp.Result != 0 {
		return fmt.Errorf("error in rename task: %d", resp.Result)
	}
	log.Infof("%s => %s", resp.TaskId, resp.FileName)
	return nil
}

func DeleteTasks(ids []string) error {
	return nil
}

func DeleteTask(taskid string) error {
	t := M.getTaskbyId(taskid)
	if t == nil {
		return errNoSuchTask
	}
	return t.Remove()
}

func PurgeTask(taskid string) error {
	t := M.getTaskbyId(taskid)
	if t == nil {
		return errNoSuchTask
	}
	return t.Purge()
}

func ResumeTasks(pattern string) error {
	return nil
}
