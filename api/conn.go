package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/matzoe/xunlei/cookiejar"
)

var noSuchTaskErr error
var invalidResponseErr error
var unexpectedErr error
var taskNotCompletedErr error
var invalidLoginErr error
var loginFailedErr error
var ReuseSessionErr error
var btTaskAlreadyErr error
var taskNoRedownCapErr error
var defaultConn struct {
	*http.Client
	sync.Mutex
}

func init() {
	jar, _ := cookiejar.New(nil)
	defaultConn.Client = &http.Client{nil, nil, jar}
	defaultConn.Mutex = sync.Mutex{}

	noSuchTaskErr = errors.New("No such TaskId in list.")
	invalidResponseErr = errors.New("Invalid response.")
	unexpectedErr = errors.New("Unexpected error.")
	taskNotCompletedErr = errors.New("Task not completed.")
	invalidLoginErr = errors.New("Invalid login account.")
	loginFailedErr = errors.New("Login failed.")
	ReuseSessionErr = errors.New("Previous session exipred.")
	btTaskAlreadyErr = errors.New("Bt task already exists.")
	taskNoRedownCapErr = errors.New("Task not capable for restart.")
}

func get(dest string) ([]byte, error) {
	log.Println("==>", dest)
	req, err := http.NewRequest("GET", dest, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", user_agent)
	req.Header.Add("Accept-Encoding", "gzip, deflate")
retry:
	defaultConn.Lock()
	resp, err := defaultConn.Do(req)
	defaultConn.Unlock()
	if err == io.EOF {
		goto retry
	}
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	log.Println(resp.Status)
	if resp.StatusCode/100 > 3 {
		return nil, errors.New(resp.Status)
	}
	return readBody(resp)
}

func post(dest string, data string) ([]byte, error) {
	log.Println("==>", dest)
	req, err := http.NewRequest("POST", dest, strings.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("User-Agent", user_agent)
	req.Header.Add("Accept-Encoding", "gzip, deflate")
retry:
	defaultConn.Lock()
	resp, err := defaultConn.Do(req)
	defaultConn.Unlock()
	if err == io.EOF {
		goto retry
	}
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	log.Println(resp.Status)
	if resp.StatusCode/100 > 3 {
		return nil, errors.New(resp.Status)
	}
	return readBody(resp)
}

func Login(id, passhash string) (err error) {
	var vcode string
	if len(id) == 0 {
		err = invalidLoginErr
		return
	}
	loginUrl := fmt.Sprintf("http://login.xunlei.com/check?u=%s&cachetime=%d", id, current_timestamp())
	u, _ := url.Parse("http://xunlei.com/")
loop:
	if _, err = get(loginUrl); err != nil {
		return
	}
	cks := defaultConn.Client.Jar.Cookies(u)
	for i, _ := range cks {
		if cks[i].Name == "check_result" {
			if len(cks[i].Value) < 3 {
				goto loop
			}
			vcode = cks[i].Value[2:]
			vcode = strings.ToUpper(vcode)
			log.Println("verify_code:", vcode)
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
	M.Uid = getCookie("http://xunlei.com", "userid")
	log.Printf("uid: %s\n", M.Uid)
	if len(M.Uid) == 0 {
		err = loginFailedErr
		return
	}
	var r []byte
	if r, err = get(fmt.Sprintf("%slogin?cachetime=%d&from=0", DOMAIN_LIXIAN, current_timestamp())); err != nil || len(r) < 512 {
		err = unexpectedErr
	}
	return
}

func SaveSession(cookieFile string) error {
	return defaultConn.Client.Jar.(*cookiejar.Jar).Save(cookieFile)
}

func ResumeSession(cookieFile string) (err error) {
	if cookieFile != "" {
		if err = defaultConn.Client.Jar.(*cookiejar.Jar).Load(cookieFile); err != nil {
			err = errors.New("Invalid cookie file.")
			return
		}
	}
	if !IsOn() {
		err = ReuseSessionErr
	}
	return
}

func IsOn() bool {
	uid := getCookie("http://xunlei.com", "userid")
	if len(uid) == 0 {
		return false
	}
	r, err := get(fmt.Sprintf(TASK_HOME, uid))
	if err != nil {
		return false
	}
	if ok, _ := regexp.Match(`top.location='http://cloud.vip.xunlei.com/task.html\?error=`, r); ok {
		// log.Println("previous login timeout")
		return false
	}
	if len(M.Uid) == 0 {
		M.Uid = uid
	}
	return true
}

func getCookie(uri, name string) string {
	u, _ := url.Parse(uri)
	cks := defaultConn.Client.Jar.Cookies(u)
	for i, _ := range cks {
		if cks[i].Name == name {
			return cks[i].Value
		}
	}
	return ""
}

func GetTasks() ([]*Task, error) {
	b, err := tasklist_nofresh(_STATUS_mixed, 1)
	if err != nil {
		return nil, err
	}
	var resp _task_resp
	err = json.Unmarshal(b, &resp)
	if err != nil {
		return nil, err
	}
	ts := make([]*Task, 0, len(resp.Info.Tasks))
	for i, _ := range resp.Info.Tasks {
		resp.Info.Tasks[i].TaskName = unescapeName(resp.Info.Tasks[i].TaskName)
		ts = append(ts, &resp.Info.Tasks[i])
	}
	M.invalidateGroup(_FLAG_normal)
	M.pushTasks(ts)
	return ts, err
}

func GetCompletedTasks() ([]*Task, error) {
	b, err := tasklist_nofresh(_STATUS_completed, 1)
	if err != nil {
		return nil, err
	}
	var resp _task_resp
	err = json.Unmarshal(b, &resp)
	if err != nil {
		return nil, err
	}
	ts := make([]*Task, 0, len(resp.Info.Tasks))
	for i, _ := range resp.Info.Tasks {
		resp.Info.Tasks[i].TaskName = unescapeName(resp.Info.Tasks[i].TaskName)
		ts = append(ts, &resp.Info.Tasks[i])
	}
	M.pushTasks(ts)
	return ts, err
}

func GetIncompletedTasks() ([]*Task, error) {
	b, err := tasklist_nofresh(_STATUS_downloading, 1)
	if err != nil {
		return nil, err
	}
	var resp _task_resp
	err = json.Unmarshal(b, &resp)
	if err != nil {
		return nil, err
	}
	ts := make([]*Task, 0, len(resp.Info.Tasks))
	for i, _ := range resp.Info.Tasks {
		resp.Info.Tasks[i].TaskName = unescapeName(resp.Info.Tasks[i].TaskName)
		ts = append(ts, &resp.Info.Tasks[i])
	}
	M.pushTasks(ts)
	return ts, err
}

func GetGdriveId() (gid string, err error) {
	if len(M.Gid) == 0 {
		var b []byte
		b, err = tasklist_nofresh(_STATUS_mixed, 1)
		if err != nil {
			return
		}
		var resp _task_resp
		err = json.Unmarshal(b, &resp)
		if err != nil {
			return
		}
		M.Gid = resp.Info.User.Cookie
		M.Account = &resp.Info.User
		M.AccountInfo = &resp.UserInfo
	}
	gid = M.Gid
	log.Println("gdriveid:", gid)
	return
}

func tasklist_nofresh(tid, page int) ([]byte, error) {
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
	uri := fmt.Sprintf(SHOWTASK_UNFRESH, tid, page, _page_size, page)
	log.Println("==>", uri)
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", user_agent)
	req.Header.Add("Accept-Encoding", "gzip, deflate")
	req.AddCookie(&http.Cookie{Name: "pagenum", Value: _page_size})
retry:
	defaultConn.Lock()
	resp, err := defaultConn.Do(req)
	defaultConn.Unlock()
	if err == io.EOF {
		goto retry
	}
	if err != nil {
		return nil, err
	}
	log.Println(resp.Status)
	r, err := readBody(resp)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}
	exp := regexp.MustCompile(`rebuild\((\{.*\})\)`)
	s := exp.FindSubmatch(r)
	if s == nil {
		return nil, invalidResponseErr
	}
	return s[1], nil
}

func readExpired() ([]byte, error) {
	uri := fmt.Sprintf(EXPIRE_HOME, M.Uid)
	log.Println("==>", uri)
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", user_agent)
	req.Header.Add("Accept-Encoding", "gzip, deflate")
	req.AddCookie(&http.Cookie{Name: "lx_nf_all", Value: url.QueryEscape(_expired_ck)})
	req.AddCookie(&http.Cookie{Name: "pagenum", Value: _page_size})
retry:
	defaultConn.Lock()
	resp, err := defaultConn.Do(req)
	defaultConn.Unlock()
	if err == io.EOF {
		goto retry
	}
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	log.Println(resp.Status)
	return readBody(resp)
}

func GetExpiredTasks() ([]*Task, error) {
	r, err := readExpired()
	ts, _ := parseHistory(r, "4")
	M.invalidateGroup(_FLAG_expired)
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
	M.invalidateGroup(_FLAG_deleted)
	M.invalidateGroup(_FLAG_purged)
	M.pushTasks(tss)
	return tss, err
}

func readHistory(page int) ([]byte, error) {
	var uri string
	if page > 0 {
		uri = fmt.Sprintf(HISTORY_PAGE, M.Uid, page)
	} else {
		uri = fmt.Sprintf(HISTORY_HOME, M.Uid)
	}

	log.Println("==>", uri)
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", user_agent)
	req.Header.Add("Accept-Encoding", "gzip, deflate")
	req.AddCookie(&http.Cookie{Name: "lx_nf_all", Value: url.QueryEscape(_deleted_ck)})
	req.AddCookie(&http.Cookie{Name: "pagenum", Value: _page_size})
retry:
	defaultConn.Lock()
	resp, err := defaultConn.Do(req)
	defaultConn.Unlock()
	if err == io.EOF {
		goto retry
	}
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	log.Println(resp.Status)
	return readBody(resp)
}

func parseHistory(in []byte, ty string) ([]*Task, bool) {
	es := `<input id="d_status(\d+)"[^<>]+value="(.*)" />\s+<input id="dflag\d+"[^<>]+value="(.*)" />\s+<input id="dcid\d+"[^<>]+value="(.*)" />\s+<input id="f_url\d+"[^<>]+value="(.*)" />\s+<input id="taskname\d+"[^<>]+value="(.*)" />\s+<input id="d_tasktype\d+"[^<>]+value="(.*)" />`
	exp := regexp.MustCompile(es)
	s := exp.FindAllSubmatch(in, -1)
	ret := make([]*Task, len(s))
	for i, _ := range s {
		b, _ := strconv.Atoi(string(s[i][7]))
		ret[i] = &Task{Id: string(s[i][1]), DownloadStatus: string(s[i][2]), Cid: string(s[i][4]), URL: string(s[i][5]), TaskName: unescapeName(string(s[i][6])), TaskType: byte(b), Flag: ty}
	}
	exp = regexp.MustCompile(`<li class="next"><a href="([^"]+)">[^<>]*</a></li>`)
	return ret, exp.FindSubmatch(in) != nil
}

func DelayTask(taskid string) error {
	uri := fmt.Sprintf(TASKDELAY_URL, taskid+"_1", "task", current_timestamp())
	r, err := get(uri)
	if err != nil {
		return err
	}
	exp := regexp.MustCompile(`^task_delay_resp\((.*}),\[.*\]\)`)
	s := exp.FindSubmatch(r)
	if s == nil {
		return invalidResponseErr
	}
	var resp struct {
		K struct {
			Llt string `json:"left_live_time"`
		} `json:"0"`
		Result byte `json:"result"`
	}
	json.Unmarshal(s[1], &resp)
	log.Printf("%s: %s\n", taskid, resp.K.Llt)
	return nil
}

func redownload(tasks []*Task) error {
	form := make([]string, 0, len(tasks)+2)
	for i, _ := range tasks {
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
	uri := fmt.Sprintf(REDOWNLOAD_URL, current_timestamp())
	r, err := post(uri, strings.Join(form, "&"))
	if err != nil {
		return err
	}
	log.Printf("%s\n", r)
	return nil
}

func FillBtList(taskid, infohash string) (*bt_list, error) {
	var pgsize = _bt_page_size
retry:
	m, err := fillBtList(taskid, infohash, 1, pgsize)
	if err == io.ErrUnexpectedEOF && pgsize == _bt_page_size {
		pgsize = "100"
		goto retry
	}
	if err != nil {
		return nil, err
	}
	var list = bt_list{}
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
			log.Println("err in fillBtList()")
		}
	}
	return &list, nil
}

func fillBtList(taskid, infohash string, page int, pgsize string) (*_bt_list, error) {
	uri := fmt.Sprintf(FILLBTLIST_URL, taskid, infohash, page, M.Uid, "task", current_timestamp())
	log.Println("==>", uri)
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", user_agent)
	req.Header.Add("Accept-Encoding", "gzip, deflate")
	req.AddCookie(&http.Cookie{Name: "pagenum", Value: pgsize})
retry:
	defaultConn.Lock()
	resp, err := defaultConn.Do(req)
	defaultConn.Unlock()
	if err == io.EOF {
		goto retry
	}
	if err != nil {
		return nil, err
	}
	log.Println(resp.Status)
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
		return nil, invalidResponseErr
	}
	var bt_list _bt_list
	json.Unmarshal(s[1], &bt_list)
	exp = regexp.MustCompile(`\\`)
	for i, _ := range bt_list.Record {
		bt_list.Record[i].FileName = exp.ReplaceAllLiteralString(bt_list.Record[i].FileName, `/`)
		bt_list.Record[i].FileName = unescapeName(bt_list.Record[i].FileName)
	}
	return &bt_list, nil
}

// supported uri schemes:
// 'ed2k', 'http', 'https', 'ftp', 'bt', 'magnet', 'thunder', 'Flashget', 'qqdl'
func AddTask(req string) error {
	ttype := _TASK_TYPE
	if strings.HasPrefix(req, "magnet:") || strings.Contains(req, "get_torrent?userid=") {
		ttype = _TASK_TYPE_MAGNET
	} else if strings.HasPrefix(req, "ed2k://") {
		ttype = _TASK_TYPE_ED2K
	} else if strings.HasPrefix(req, "bt://") || strings.HasSuffix(req, ".torrent") {
		ttype = _TASK_TYPE_BT
	} else if ok, _ := regexp.MatchString(`^[a-zA-Z0-9]{40,40}$`, req); ok {
		ttype = _TASK_TYPE_BT
		req = "bt://" + req
	}
	switch ttype {
	case _TASK_TYPE, _TASK_TYPE_ED2K:
		return addSimpleTask(req)
	case _TASK_TYPE_BT:
		return addBtTask(req)
	case _TASK_TYPE_MAGNET:
		return addMagnetTask(req)
	case _TASK_TYPE_INVALID:
		fallthrough
	default:
		return unexpectedErr
	}
	panic(unexpectedErr.Error())
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
	tm := current_timestamp()
	uri := fmt.Sprintf(BATCHTASKCOMMIT_URL, tm, tm)
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
	dest := fmt.Sprintf(TASKCHECK_URL, url.QueryEscape(uri), from, current_random(), current_timestamp())
	r, err := get(dest)
	if err == nil {
		task_pre, err := getTaskPre(r)
		if err != nil {
			return err
		}
		var t_type string
		if strings.HasPrefix(uri, "http://") || strings.HasPrefix(uri, "ftp://") || strings.HasPrefix(uri, "https://") {
			t_type = strconv.Itoa(_TASK_TYPE)
		} else if strings.HasPrefix(uri, "ed2k://") {
			t_type = strconv.Itoa(_TASK_TYPE_ED2K)
		} else {
			return errors.New("Invalid protocol scheme.")
		}
		v := url.Values{}
		v.Add("callback", "ret_task")
		v.Add("uid", M.Uid)
		v.Add("cid", task_pre.Cid)
		v.Add("gcid", task_pre.GCid)
		v.Add("size", task_pre.SizeCost)
		v.Add("goldbean", task_pre.Goldbean)
		v.Add("silverbean", task_pre.Silverbean)
		v.Add("t", task_pre.FileName)
		v.Add("url", uri)
		v.Add("type", t_type)
		if len(oid) > 0 {
			v.Add("o_taskid", oid[0])
			v.Add("o_page", "history")
		} else {
			v.Add("o_page", "task")
			v.Add("o_taskid", "0")
		}
		dest = TASKCOMMIT_URL + v.Encode()
		r, err = get(dest)
		if err != nil {
			return err
		}
		if ok, _ := regexp.Match(`ret_task\(.*\)`, r); ok {
			return nil
		} else {
			return invalidResponseErr
		}
	}
	return err
}

func addBtTask(uri string) error {
	if strings.HasPrefix(uri, "bt://") {
		return addMagnetTask(fmt.Sprintf(GETTORRENT_URL, M.Uid, uri[5:]))
	}
	return addTorrentTask(uri)
}

func addMagnetTask(link string, oid ...string) error {
	uri := fmt.Sprintf(URLQUERY_URL, url.QueryEscape(link), current_random())
	r, err := get(uri)
	if err != nil {
		return err
	}
	exp := regexp.MustCompile(`queryUrl\((1,.*)\)`)
	s := exp.FindSubmatch(r)
	if s == nil {
		if ok, _ := regexp.Match(`queryUrl\(-1,'[0-9A-Za-z]{40,40}'.*`, r); ok {
			return btTaskAlreadyErr
		}
		return invalidResponseErr
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
		dest := fmt.Sprintf(BTTASKCOMMIT_URL, current_timestamp())
		r, err = post(dest, v.Encode())
		exp = regexp.MustCompile(`jsonp.*\(\{"id":"(\d+)","avail_space":"\d+".*\}\)`)
		s = exp.FindSubmatch(r)
		if s == nil {
			return invalidResponseErr
		}
	} else {
		return invalidResponseErr
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
	writer.WriteField("random", current_random())
	writer.WriteField("interfrom", "task")

	dest := TORRENTUPLOAD_URL
	log.Println("==>", dest)
	req, err := http.NewRequest("POST", dest, bytes.NewReader(body.Bytes()))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Add("User-Agent", user_agent)
	req.Header.Add("Accept-Encoding", "gzip, deflate")
retry:
	defaultConn.Lock()
	resp, err := defaultConn.Do(req)
	defaultConn.Unlock()
	if err == io.EOF {
		goto retry
	}
	if err != nil {
		return
	}
	log.Println(resp.Status)
	r, err := readBody(resp)
	resp.Body.Close()
	exp := regexp.MustCompile(`<script>document\.domain="xunlei\.com";var btResult =(\{.+\});var btRtcode = 0</script>`)
	s := exp.FindSubmatch(r)
	if s != nil {
		var result _btup_result
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
		dest = fmt.Sprintf(BTTASKCOMMIT_URL, current_timestamp())
		r, err = post(dest, v.Encode())
		exp = regexp.MustCompile(`jsonp.*\(\{"id":"(\d+)","avail_space":"\d+".*\}\)`)
		s = exp.FindSubmatch(r)
		if s == nil {
			return invalidResponseErr
		}
		// tasklist_nofresh(4, 1)
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
	return btTaskAlreadyErr
}

func ProcessTaskDaemon(ch chan byte, callback func(*Task)) {
	if len(M.Tasks) == 0 {
		GetIncompletedTasks()
	}

	go func() {
		for {
			select {
			case <-ch:
				err := process_task(M.Tasks, callback)
				if err != nil {
					log.Println("error in ProcessTask():", err)
				}
			case <-time.After(60 * time.Second):
				err := process_task(M.Tasks, callback)
				if err != nil {
					log.Println("error in ProcessTask():", err)
					time.Sleep(5 * time.Second)
					ch <- 1
				}
			}
		}
	}()
}

func ProcessTask(callback func(*Task)) error {
	return process_task(M.Tasks, callback)
}

func process_task(tasks map[string]*Task, callback func(*Task)) error {
	l := len(tasks)
	if l == 0 {
		return errors.New("No tasks in progress.")
	}
	ct := current_timestamp()
	uri := fmt.Sprintf(TASKPROCESS_URL, ct, ct)
	v := url.Values{}
	list := make([]string, 0, l)
	nm_list := make([]string, 0, l)
	bt_list := make([]string, 0, l)
	for i, _ := range tasks {
		if tasks[i].status() == _FLAG_normal && tasks[i].DownloadStatus == "1" {
			list = append(list, tasks[i].Id)
			if tasks[i].TaskType == 0 {
				bt_list = append(bt_list, tasks[i].Id)
			} else {
				nm_list = append(nm_list, tasks[i].Id)
			}
		}
	}
	v.Add("list", strings.Join(list, ","))
	v.Add("nm_list", strings.Join(nm_list, ","))
	v.Add("bt_list", strings.Join(bt_list, ","))
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
		return invalidResponseErr
	}
	var res _ptask_resp
	err = json.Unmarshal(s[1], &res)
	if err != nil {
		return err
	}
	for i, _ := range res.List {
		task := tasks[res.List[i].Id]
		task.update(&res.List[i])
		if callback != nil {
			callback(task)
		}
	}
	return nil
}

func GetTorrentByHash(hash string) ([]byte, error) {
	uri := fmt.Sprintf(GETTORRENT_URL, M.Uid, strings.ToUpper(hash))
	r, err := get(uri)
	if err != nil {
		return nil, err
	}
	exp := regexp.MustCompile(`alert\('(.*)'\)`)
	s := exp.FindSubmatch(r)
	if s != nil {
		log.Printf("%s\n", s[1])
		return nil, invalidResponseErr
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
	uri := fmt.Sprintf(TASKPAUSE_URL, tids, M.Uid, current_timestamp())
	r, err := get(uri)
	if err != nil {
		return err
	}
	if bytes.Compare(r, []byte("pause_task_resp()")) != 0 {
		return invalidResponseErr
	}
	return nil
}

func DelayAllTasks() error {
	r, err := get(DELAYONCE_URL)
	if err != nil {
		return err
	}
	log.Printf("%s\n", r)
	return nil
}

func ReAddTasks(ts map[string]*Task) {
	nbt := make([]*Task, 0, len(ts))
	bt := make([]*Task, 0, len(ts))
	for i, _ := range ts {
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
			log.Println(err)
		}
	} else if len(nbt) > 1 {
		urls, ids := extractTasks(nbt)
		if err := AddBatchTasks(urls, ids...); err != nil {
			log.Println(err)
		}
	}
	for i, _ := range bt {
		if err := addMagnetTask(fmt.Sprintf(GETTORRENT_URL, M.Uid, bt[i].Cid), bt[i].Id); err != nil {
			log.Println(err)
		}
	}
}

func RenameTask(taskid, newname string) error {
	t := M.getTaskbyId(taskid)
	if t == nil {
		return noSuchTaskErr
	}
	return rename_task(taskid, newname, t.TaskType)
}

func rename_task(taskid, newname string, tasktype byte) error {
	v := url.Values{}
	v.Add("taskid", taskid)
	if tasktype == 0 {
		v.Add("bt", "1")
	} else {
		v.Add("bt", "0")
	}
	v.Add("filename", newname)
	r, err := get(RENAME_URL + v.Encode())
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
	log.Println(resp.TaskId, "=>", resp.FileName)
	return nil
}

func DeleteTasks(ids []string) error {
	return nil
}

func DeleteTask(taskid string) error {
	t := M.getTaskbyId(taskid)
	if t == nil {
		return noSuchTaskErr
	}
	return t.Remove()
}

func PurgeTask(taskid string) error {
	t := M.getTaskbyId(taskid)
	if t == nil {
		return noSuchTaskErr
	}
	return t.Purge()
}

func ResumeTasks(pattern string) error {
	return nil
}
