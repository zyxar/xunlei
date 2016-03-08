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
	"github.com/zyxar/image2ascii/ascii"
	"github.com/zyxar/taipei"
)

var (
	urlXunleiCom                *url.URL
	urlVipXunleiCom             *url.URL
	urlDynamicCloudVipXunleiCom *url.URL
	defaultSession              Session
	errInvalidSession           = errors.New("invalid session")
	errInvalidResponse          = errors.New("invalid response")
	errInvalidAccount           = errors.New("invalid login account")
	errUnexpected               = errors.New("unexpected error")
	errLoginFailed              = errors.New("login failed")
	errTimeout                  = errors.New("request time out")
	errSessionExpired           = errors.New("previous session exipred")
	errBtTaskExisted            = errors.New("bt task already exists")
	errNoTasksInProgress        = errors.New("no tasks in progress")
	errInvalidTaskFlag          = errors.New("invalid flag in task")
	errTaskNoRedownCap          = errors.New("task not capable for restart")
	errTaskNotFound             = errors.New("no such taskid in list")
	errTaskNotCompleted         = errors.New("task not completed")
	errTaskAlreadyQueued        = errors.New("task already in queue")
	errTaskAlreadyPurged        = errors.New("task already purged")
	errTaskAlreadyRemoved       = errors.New("task already deleted")
	errTaskSubmissionFailed     = errors.New("task submission failed")
	errTaskNeedVerification     = errors.New("task submission need verification")
	verifyBaseURLs              = []string{
		"http://verify2.xunlei.com/image?t=MVA&cachetime=%d",
		"http://verify.xunlei.com/image?t=MVA&cachetime=%d",
		"http://verify3.xunlei.com/image?t=MVA&cachetime=%d",
	}
)

func init() {
	urlXunleiCom, _ = url.Parse("http://xunlei.com")
	urlVipXunleiCom, _ = url.Parse("http://vip.xunlei.com")
	urlDynamicCloudVipXunleiCom, _ = url.Parse("http://dynamic.cloud.vip.xunlei.com")
	defaultSession = newSession(3000 * time.Millisecond)
	log.SetHandler(text.New(os.Stderr))
	log.SetLevel(log.InfoLevel)
}

type Session interface {
	Login(id, passhash string) (err error)
	SaveSession(cookieFile string) error
	ResumeSession(cookieFile string) (err error)
	Account() (ua *UserAccount)
	IsOn() bool
	GetTasks() ([]*Task, error)
	GetCompletedTasks() ([]*Task, error)
	GetIncompletedTasks() ([]*Task, error)
	GetGdriveId() (gid string, err error)
	RawTaskList(category, page int) ([]byte, error)
	RawTaskListExpired() ([]byte, error)
	RawTaskListDeleted(page int) ([]byte, error)
	GetExpiredTasks() ([]*Task, error)
	GetDeletedTasks() ([]*Task, error)
	DelayTask(t *Task) error
	DelayTaskById(taskid string) error
	FillBtList(t *Task) (*btList, error)
	FillBtListById(taskid, infohash string) (*btList, error)
	RawFillBtList(t *Task, page int) ([]byte, error)
	RawFillBtListById(taskid, infohash string, page int) ([]byte, error)
	AddTask(req string) error
	AddBatchTasks(urls []string, oids ...string) error
	ProcessTaskDaemon(ch chan byte, callback TaskCallback)
	ProcessTask(callback TaskCallback) error
	GetTorrentByHash(hash string) ([]byte, error)
	GetTorrentFileByHash(hash, file string) error
	PauseTask(t *Task) error
	PauseTasks(ids []string) error
	DelayAllTasks() error
	ReAddTask(t *Task) error
	ReAddTasks(ts map[string]*Task)
	RenameTask(t *Task, newname string) error
	RenameTaskById(taskid, newname string) error
	ResumeTask(t *Task) error
	ResumeTaskById(taskid string) error
	DeleteTask(t *Task) error
	DeleteTaskById(taskid string) error
	PurgeTask(t *Task) error
	PurgeTaskById(taskid string) error
	VerifyTask(t *Task, path string) bool
	FindTasks(pattern string) (map[string]*Task, error)
	GetTaskById(taskid string) (t *Task, exist bool)
	GetTasksByIds(ids []string) map[string]*Task
	InvalidateCache(flag byte)
	InvalidateCacheAll()
}

type session struct {
	*http.Client
	mutex       *sync.Mutex
	cache       *cache
	account     *UserAccount
	accountInfo *UserInfo
	uid         string
	gid         string
	timeout     time.Duration
}

func NewSession(timeout time.Duration) Session {
	return newSession(timeout)
}

func newSession(timeout time.Duration) *session {
	jar, _ := cookiejar.New(nil)
	return &session{
		timeout: timeout,
		Client: &http.Client{
			Jar: jar,
			Transport: &http.Transport{
				Dial:  (&net.Dialer{Timeout: timeout}).Dial,
				Proxy: http.ProxyFromEnvironment,
			},
		},
		mutex: &sync.Mutex{},
		cache: newCache(),
	}
}

func (s *session) Login(id, passhash string) (err error) {
	var vcode string
	if len(id) == 0 {
		err = errInvalidAccount
		return
	}
	loginURL := fmt.Sprintf("http://login.xunlei.com/check?u=%s&cachetime=%d", id, currentTimestamp())
loop:
	if _, err = s.get(loginURL); err != nil {
		return
	}
	cks := s.Client.Jar.Cookies(urlXunleiCom)
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
	if _, err = s.post("http://login.xunlei.com/sec2login/", v.Encode()); err != nil {
		return
	}
	s.uid = s.getCookie("userid")
	log.Infof("uid: %s\n", s.uid)
	if len(s.uid) == 0 {
		err = errLoginFailed
		return
	}
	var r []byte
	if r, err = s.get(fmt.Sprintf("%slogin?cachetime=%d&from=0", domainLixianURI, currentTimestamp())); err != nil || len(r) < 512 {
		err = errUnexpected
	}
	return
}

func (s *session) SaveSession(cookieFile string) error {
	session := [][]*http.Cookie{
		s.Client.Jar.Cookies(urlXunleiCom),
		s.Client.Jar.Cookies(urlVipXunleiCom),
		s.Client.Jar.Cookies(urlDynamicCloudVipXunleiCom)}
	r, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(cookieFile, r, 0644)
}

func (s *session) ResumeSession(cookieFile string) (err error) {
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
			err = errInvalidSession
			return
		}
		s.Client.Jar.SetCookies(urlXunleiCom, session[0])
		s.Client.Jar.SetCookies(urlVipXunleiCom, session[1])
		s.Client.Jar.SetCookies(urlDynamicCloudVipXunleiCom, session[2])
	}
	if !s.IsOn() {
		err = errSessionExpired
	}
	return
}

func (s *session) Account() (ua *UserAccount) {
	return s.account
}

func (s *session) IsOn() bool {
	uid := s.getCookie("userid")
	if len(uid) == 0 {
		return false
	}
	r, err := s.get(fmt.Sprintf(taskHomeURI, uid))
	if err != nil {
		return false
	}
	if ok, _ := regexp.Match(`top.location='http://cloud.vip.xunlei.com/task.html\?error=`, r); ok {
		return false
	}
	if len(s.uid) == 0 {
		s.uid = uid
	}
	return true
}

func (s *session) GetTasks() ([]*Task, error) {
	accumulated := 0
	page := 1
	var ts []*Task
round:
	b, err := s.tasklistNofresh(statusMixed, page)
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
	s.cache.InvalidateGroup(flagNormal)
	s.cache.pushTasks(ts)
	return ts, err
}

func (s *session) GetCompletedTasks() ([]*Task, error) {
	accumulated := 0
	page := 1
	var ts []*Task
round:
	b, err := s.tasklistNofresh(statusCompleted, page)
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
	s.cache.pushTasks(ts)
	return ts, err
}

func (s *session) GetIncompletedTasks() ([]*Task, error) {
	accumulated := 0
	page := 1
	var ts []*Task
round:
	b, err := s.tasklistNofresh(statusDownloading, page)
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
	s.cache.pushTasks(ts)
	return ts, err
}

func (s *session) GetGdriveId() (gid string, err error) {
	if len(s.gid) == 0 {
		var b []byte
		b, err = s.tasklistNofresh(statusMixed, 1)
		if err != nil {
			return
		}
		var resp taskResponse
		err = json.Unmarshal(b, &resp)
		if err != nil {
			return
		}
		s.gid = resp.Info.User.Cookie
		s.account = &resp.Info.User
		s.accountInfo = &resp.UserInfo
		log.Debugf("gdriveid: %s", s.gid)
	}
	gid = s.gid
	return
}

// bellow three funcs are for RESTful calling;
// we do not parse content here.
func (s *session) RawTaskList(category, page int) ([]byte, error) {
	return s.tasklistNofresh(category, page)
}

func (s *session) RawTaskListExpired() ([]byte, error) {
	return s.readExpired()
}

func (s *session) RawTaskListDeleted(page int) ([]byte, error) {
	return s.readHistory(page)
}

func (s *session) GetExpiredTasks() ([]*Task, error) {
	r, err := s.readExpired()
	ts, _ := parseHistory(r, "4")
	s.cache.InvalidateGroup(flagExpired)
	s.cache.pushTasks(ts)
	return ts, err
}

func (s *session) GetDeletedTasks() ([]*Task, error) {
	j := 0
	next := true
	var err error
	var r []byte
	var ts []*Task
	tss := make([]*Task, 0, 10)
	for next {
		j++
		r, err = s.readHistory(j)
		ts, next = parseHistory(r, "1")
		tss = append(tss, ts...)
	}
	s.cache.InvalidateGroup(flagDeleted)
	s.cache.InvalidateGroup(flagPurged)
	s.cache.pushTasks(tss)
	return tss, err
}

func (s *session) DelayTask(t *Task) error {
	if t == nil {
		return errTaskNotFound
	}
	return s.DelayTaskById(t.Id)
}

func (s *session) DelayTaskById(taskid string) error {
	r, err := s.get(fmt.Sprintf(taskdelayURI, taskid+"_1", "task", currentTimestamp()))
	if err != nil {
		return err
	}
	exp := regexp.MustCompile(`^task_delay_resp\((.*}),\[.*\]\)`)
	sub := exp.FindSubmatch(r)
	if sub == nil {
		return errInvalidResponse
	}
	var resp struct {
		K struct {
			Llt string `json:"left_live_time"`
		} `json:"0"`
		Result byte `json:"result"`
	}
	json.Unmarshal(sub[1], &resp)
	log.Infof("%s: %s\n", taskid, resp.K.Llt)
	return nil
}

func (s *session) FillBtList(t *Task) (*btList, error) {
	if t == nil {
		return nil, errTaskNotFound
	}
	return s.FillBtListById(t.Id, t.Cid)
}

func (s *session) FillBtListById(taskid, infohash string) (*btList, error) {
	var pgsize = btPageSize
retry:
	m, err := s.fillBtList(taskid, infohash, 1, pgsize)
	if err == io.ErrUnexpectedEOF && pgsize == btPageSize {
		pgsize = "100"
		goto retry
	}
	if err != nil {
		return nil, err
	}
	var list = *m
	total, _ := strconv.Atoi(list.BtNum)
	size, _ := strconv.Atoi(pgsize)
	pageNum := total/size + 1
	next := 2
	for next <= pageNum {
		m, err = s.fillBtList(taskid, infohash, next, pgsize)
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

func (s *session) RawFillBtList(t *Task, page int) ([]byte, error) {
	if t == nil {
		return nil, errTaskNotFound
	}
	return s.RawFillBtListById(t.Id, t.Cid, page)
}

func (s *session) RawFillBtListById(taskid, infohash string, page int) ([]byte, error) {
	var pgsize = btPageSize
retry:
	uri := fmt.Sprintf(fillbtlistURI, taskid, infohash, page, s.uid, "task", currentTimestamp())
	log.Debugf("==> %s", uri)
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", userAgent)
	req.Header.Add("Accept-Encoding", "gzip, deflate")
	req.AddCookie(&http.Cookie{Name: "pagenum", Value: pgsize})
	resp, err := s.routine(req)
	if err != nil {
		return nil, err
	}
	log.Debug(resp.Status)
	defer resp.Body.Close()
	r, err := readBody(resp)
	if err == io.ErrUnexpectedEOF && pgsize == btPageSize {
		pgsize = "100"
		goto retry
	}
	return r, nil
}

// supported uri schemes:
// 'ed2k', 'http', 'https', 'ftp', 'bt', 'magnet', 'thunder', 'Flashget', 'qqdl'
func (s *session) AddTask(req string) error {
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
		return s.addSimpleTask(req)
	case taskTypeBt:
		return s.addBtTask(req)
	case taskTypeMagnet:
		return s.addMagnetTask(req)
	case taskTypeInvalid:
		fallthrough
	default:
	}
	return errUnexpected
}

func (s *session) AddBatchTasks(urls []string, oids ...string) error {
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
	r, err := s.post(fmt.Sprintf(batchtaskcommitURI, tm, tm), v.Encode())
	fmt.Printf("%s\n", r)
	return err
}

func (s *session) ProcessTaskDaemon(ch chan byte, callback TaskCallback) {
	if len(s.cache.Tasks) == 0 {
		s.GetIncompletedTasks()
	}
	go func() {
		for {
			select {
			case <-ch:
				err := s.processTask(callback)
				if err != nil {
					log.Errorf("error in ProcessTask(): %v", err)
				}
			case <-time.After(60 * time.Second):
				err := s.processTask(callback)
				if err != nil {
					log.Errorf("error in ProcessTask(): %v", err)
					time.Sleep(5 * time.Second)
					ch <- 1
				}
			}
		}
	}()
}

func (s *session) ProcessTask(callback TaskCallback) error {
	return s.processTask(callback)
}

func (s *session) GetTorrentByHash(hash string) ([]byte, error) {
	r, err := s.get(fmt.Sprintf(gettorrentURI, s.uid, strings.ToUpper(hash)))
	if err != nil {
		return nil, err
	}
	exp := regexp.MustCompile(`alert\('(.*)'\)`)
	sub := exp.FindSubmatch(r)
	if sub != nil {
		log.Infof("%s\n", sub[1])
		return nil, errInvalidResponse
	}
	return r, nil
}

func (s *session) GetTorrentFileByHash(hash, file string) error {
	if stat, err := os.Stat(file); err == nil || stat != nil {
		return os.ErrExist
	}
	r, err := s.GetTorrentByHash(hash)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(file, r, 0644)
}

func (s *session) PauseTask(t *Task) error {
	tids := t.Id + ","
	uri := fmt.Sprintf(taskpauseURI, tids, s.uid, currentTimestamp())
	r, err := s.get(uri)
	if err != nil {
		return err
	}
	if bytes.Compare(r, []byte("pause_task_resp()")) != 0 {
		return errInvalidResponse
	}
	return nil
}

func (s *session) PauseTasks(ids []string) error {
	tids := strings.Join(ids, ",")
	tids += ","
	r, err := s.get(fmt.Sprintf(taskpauseURI, tids, s.uid, currentTimestamp()))
	if err != nil {
		return err
	}
	if bytes.Compare(r, []byte("pause_task_resp()")) != 0 {
		return errInvalidResponse
	}
	return nil
}

func (s *session) DelayAllTasks() error {
	r, err := s.get(delayonceURI)
	if err != nil {
		return err
	}
	log.Infof("%s\n", r)
	return nil
}

func (s *session) ReAddTask(t *Task) error {
	if t == nil {
		return errTaskNotFound
	}
	if t.normal() {
		return errTaskAlreadyQueued
	}
	if t.purged() {
		return s.addSimpleTask(t.URL)
	}
	return s.addSimpleTask(t.URL, t.Id)
}

func (s *session) ReAddTasks(ts map[string]*Task) {
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
		if err := s.ReAddTask(nbt[0]); err != nil {
			log.Error(err.Error())
		}
	} else if len(nbt) > 1 {
		urls, ids := extractTasks(nbt)
		if err := s.AddBatchTasks(urls, ids...); err != nil {
			log.Error(err.Error())
		}
	}
	for i := range bt {
		if err := s.addMagnetTask(fmt.Sprintf(gettorrentURI, s.uid, bt[i].Cid), bt[i].Id); err != nil {
			log.Error(err.Error())
		}
	}
}

func (s *session) RenameTask(t *Task, newname string) error {
	if t == nil {
		return errTaskNotFound
	}
	return s.RenameTaskById(t.Id, newname)
}

func (s *session) RenameTaskById(taskid, newname string) error {
	t := s.cache.getTaskbyId(taskid)
	if t == nil {
		return errTaskNotFound
	}
	return s.renameTask(taskid, newname, t.TaskType)
}

func (s *session) ResumeTask(t *Task) error {
	if t == nil {
		return errTaskNotFound
	}
	return s.ResumeTaskById(t.Id)
}

func (s *session) ResumeTaskById(taskid string) error {
	t := s.cache.getTaskbyId(taskid)
	if t == nil {
		return errTaskNotFound
	}
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
	r, err := s.post(uri, strings.Join(form, "&"))
	if err != nil {
		return err
	}
	log.Debugf("resume task: %s", r)
	return nil
}

func (s *session) DeleteTask(t *Task) error {
	if t == nil {
		return errTaskNotFound
	}
	return s.DeleteTaskById(t.Id)
}

func (s *session) DeleteTaskById(taskid string) error {
	t := s.cache.getTaskbyId(taskid)
	if t == nil {
		return errTaskNotFound
	}
	return s.removeTask(t, 0)
}

func (s *session) PurgeTask(t *Task) error {
	if t == nil {
		return errTaskNotFound
	}
	return s.PurgeTaskById(t.Id)
}

func (s *session) PurgeTaskById(taskid string) error {
	t := s.cache.getTaskbyId(taskid)
	if t == nil {
		return errTaskNotFound
	}
	if t.deleted() {
		return s.removeTask(t, 1)
	}
	if err := s.removeTask(t, 0); err != nil {
		return err
	}
	return s.removeTask(t, 1)
}

func (s *session) VerifyTask(t *Task, path string) bool {
	if t.IsBt() {
		fmt.Println("Verifying [BT]", path)
		var b []byte
		var err error
		if b, err = s.GetTorrentByHash(t.Cid); err != nil {
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

func (s *session) FindTasks(pattern string) (map[string]*Task, error) {
	return s.cache.FindTasks(pattern)
}
func (s *session) GetTaskById(taskid string) (t *Task, exist bool) { return s.cache.GetTaskById(taskid) }
func (s *session) GetTasksByIds(ids []string) map[string]*Task     { return s.cache.GetTasksByIds(ids) }
func (s *session) InvalidateCache(flag byte)                       { s.cache.InvalidateGroup(flag) }
func (s *session) InvalidateCacheAll()                             { s.cache.InvalidateAll() }

// ---- private ----

func (s *session) lock() {
	s.mutex.Lock()
}

func (s *session) unlock() {
	s.mutex.Unlock()
}

func (s *session) routine(req *http.Request) (*http.Response, error) {
	timeout := false
retry:
	s.lock()
	timer := time.AfterFunc(s.timeout, func() {
		s.Client.Transport.(*http.Transport).CancelRequest(req)
		timeout = true
	})
	resp, err := s.Do(req)
	if timer != nil {
		timer.Stop()
	}
	s.unlock()
	if err == io.EOF && !timeout {
		goto retry
	}
	if timeout {
		err = errTimeout
	}
	return resp, err
}

func (s *session) get(dest string) ([]byte, error) {
	log.Debugf("==> %s", dest)
	req, err := http.NewRequest("GET", dest, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", userAgent)
	req.Header.Add("Accept-Encoding", "gzip, deflate")
	resp, err := s.routine(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	log.Debug(resp.Status)
	if resp.StatusCode/100 > 3 {
		return nil, errors.New(resp.Status)
	}
	return readBody(resp)
}

func (s *session) post(dest string, data string) ([]byte, error) {
	log.Debugf("==> %s", dest)
	req, err := http.NewRequest("POST", dest, strings.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("User-Agent", userAgent)
	req.Header.Add("Accept-Encoding", "gzip, deflate")
	resp, err := s.routine(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	log.Debug(resp.Status)
	if resp.StatusCode/100 > 3 {
		return nil, errors.New(resp.Status)
	}
	return readBody(resp)
}

func (s *session) getCookie(name string) string {
	cks := s.Client.Jar.Cookies(urlXunleiCom)
	for i := range cks {
		if cks[i].Name == name {
			return cks[i].Value
		}
	}
	return ""
}

func (s *session) renameTask(taskid, newname string, tasktype byte) error {
	v := url.Values{}
	v.Add("taskid", taskid)
	if tasktype == 0 {
		v.Add("bt", "1")
	} else {
		v.Add("bt", "0")
	}
	v.Add("filename", newname)
	r, err := s.get(renameURI + v.Encode())
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

func (s *session) readExpired() ([]byte, error) {
	uri := fmt.Sprintf(expireHomeURI, s.uid)
	log.Debugf("==> %s", uri)
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", userAgent)
	req.Header.Add("Accept-Encoding", "gzip, deflate")
	req.AddCookie(&http.Cookie{Name: "lx_nf_all", Value: url.QueryEscape(expiredCk)})
	req.AddCookie(&http.Cookie{Name: "pagenum", Value: pageSize})
	resp, err := s.routine(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	log.Debug(resp.Status)
	return readBody(resp)
}

func (s *session) readHistory(page int) ([]byte, error) {
	var uri string
	if page > 0 {
		uri = fmt.Sprintf(historyPageURI, s.uid, page)
	} else {
		uri = fmt.Sprintf(historyHomeURI, s.uid)
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
	resp, err := s.routine(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	log.Debug(resp.Status)
	return readBody(resp)
}

func (s *session) redownload(tasks []*Task) error {
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
		return errTaskAlreadyQueued
	}
	form = append(form, "type=1")
	form = append(form, "interfrom=task")
	r, err := s.post(fmt.Sprintf(redownloadURI, currentTimestamp()), strings.Join(form, "&"))
	if err != nil {
		return err
	}
	log.Infof("%s\n", r)
	return nil
}

func (s *session) fillBtList(taskid, infohash string, page int, pgsize string) (*btList, error) {
	uri := fmt.Sprintf(fillbtlistURI, taskid, infohash, page, s.uid, "task", currentTimestamp())
	log.Debugf("==> %s", uri)
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", userAgent)
	req.Header.Add("Accept-Encoding", "gzip, deflate")
	req.AddCookie(&http.Cookie{Name: "pagenum", Value: pgsize})
	resp, err := s.routine(req)
	if err != nil {
		return nil, err
	}
	log.Debug(resp.Status)
	r, err := readBody(resp)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}
	exp := regexp.MustCompile(`fill_bt_list\({"Result":(.*)}\)`)
	sub := exp.FindSubmatch(r)
	if sub == nil {
		exp = regexp.MustCompile(`alert\('(.*)'\);.*`)
		sub = exp.FindSubmatch(r)
		if sub != nil {
			return nil, errors.New(string(sub[1]))
		}
		log.Debugf("fill_bt_list result: %s", r)
		return nil, errInvalidResponse
	}
	var btlist btList
	json.Unmarshal(sub[1], &btlist)
	exp = regexp.MustCompile(`\\`)
	for i := range btlist.Record {
		btlist.Record[i].FileName = exp.ReplaceAllLiteralString(btlist.Record[i].FileName, `/`)
		btlist.Record[i].FileName = unescapeName(btlist.Record[i].FileName)
	}
	return &btlist, nil
}

func (s *session) addSimpleTask(uri string, oid ...string) error {
	var from string
	if len(oid) > 0 {
		from = "history"
	} else {
		from = "task"
	}
	exp := regexp.MustCompile(`%2C|,`)
	uri = exp.ReplaceAllLiteralString(uri, `.`)
	dest := fmt.Sprintf(taskcheckURI, url.QueryEscape(uri), from, currentRandom(), currentTimestamp())
	r, err := s.get(dest)
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
			// return errors.New("invalid protocol scheme")
		}
		v := url.Values{}
		v.Add("callback", "ret_task")
		v.Add("uid", s.uid)
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
		if r, err = s.get(dest); err != nil {
			return err
		}
		if ok, _ := regexp.Match(`ret_task\(.*\)`, r); ok {
			return nil
		}
		return errInvalidResponse
	}
	return err
}

func (s *session) addBtTask(uri string) error {
	if strings.HasPrefix(uri, "bt://") {
		return s.addMagnetTask(fmt.Sprintf(gettorrentURI, s.uid, uri[5:]))
	}
	return s.addTorrentTask(uri)
}

func (s *session) addMagnetTask(link string, oid ...string) error {
	r, err := s.get(fmt.Sprintf(urlqueryURI, url.QueryEscape(link), currentRandom()))
	if err != nil {
		return err
	}
	exp := regexp.MustCompile(`queryUrl\((1,.*)\)`)
	sub := exp.FindSubmatch(r)
	if sub == nil {
		if ok, _ := regexp.Match(`queryUrl\(-1,'[0-9A-Za-z]{40,40}'.*`, r); ok {
			return errBtTaskExisted
		}
		log.Debugf("bt query response: %s", r)
		return errInvalidResponse
	}
	if resp := parseBtQueryResponse(sub[1]); resp != nil {
		log.Debugf("parsed bt query response: %v", resp)
		v := url.Values{}
		v.Add("uid", s.uid)
		v.Add("btname", resp.Name)
		v.Add("cid", resp.InfoId)
		v.Add("tsize", resp.Size)
		findex := strings.Join(resp.Index, "_")
		size := strings.Join(resp.Sizes, "_")
		v.Add("findex", findex)
		v.Add("size", size)
		if len(oid) > 0 {
			v.Add("from", "history")
			v.Add("o_taskid", oid[0])
			v.Add("o_page", "history")
		} else {
			v.Add("from", "task")
		}
		log.Debugf("submit bt: %s", v.Encode())
		dest := fmt.Sprintf(bttaskcommitURI, currentTimestamp())
		r, err = s.post(dest, v.Encode())
		exp = regexp.MustCompile(`jsonp.*\((\{"id":.*,"avail_space":.*\})\)`)
		log.Debugf("bt submission response: %s", r)
		sub = exp.FindSubmatch(r)
		if sub == nil {
			return errInvalidResponse
		}
		var submresp btSumbissionResponse
		if err = json.Unmarshal(sub[1], &submresp); err != nil {
			return err
		}
		switch submresp.Progress {
		case 1:
			return nil
		case 2:
			if submresp.ErrorMessage != nil {
				return errors.New(submresp.Message)
			}
			return errTaskSubmissionFailed
		case -11, -12:
			return errTaskNeedVerification
		default:
			return errUnexpected
		}
	}
	return errInvalidResponse
}

func (s *session) addTorrentTask(filename string) (err error) {
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
	resp, err := s.routine(req)
	if err != nil {
		return
	}
	log.Debug(resp.Status)
	r, err := readBody(resp)
	resp.Body.Close()
	exp := regexp.MustCompile(`<script>document\.domain="xunlei\.com";var btResult =(\{.+\});(var btRtcode = 0)*</script>`)
	sub := exp.FindSubmatch(r)
	if sub != nil {
		var result btUploadResponse
		json.Unmarshal(sub[1], &result)
		v := url.Values{}
		v.Add("uid", s.uid)
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
		r, err = s.post(dest, v.Encode())
		exp = regexp.MustCompile(`jsonp.*\(\{"id":"(\d+)","avail_space":"\d+".*\}\)`)
		sub = exp.FindSubmatch(r)
		if sub == nil {
			fmt.Printf("%s\n", r)
			return errInvalidResponse
		}
		// s.tasklistNofresh(4, 1)
		// FillBtList(string(sub[1]))
		return nil
	}
	exp = regexp.MustCompile(`parent\.edit_bt_list\((\{.*\}),'`)
	sub = exp.FindSubmatch(r)
	if sub == nil {
		return errUnexpected
	}
	// var result _btup_result
	// json.Unmarshal(sub[1], &result)
	return errBtTaskExisted
}

func (s *session) processTask(callback TaskCallback) error {
	tasks := s.cache.Tasks
	l := len(tasks)
	if l == 0 {
		return errNoTasksInProgress
	}
	ct := currentTimestamp()
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
	v.Add("uid", s.uid)
	v.Add("interfrom", "task")
	var r []byte
	var err error
	if r, err = s.post(fmt.Sprintf(taskprocessURI, ct, ct), v.Encode()); err != nil {
		return err
	}
	exp := regexp.MustCompile(`jsonp\d+\(\{"Process":(.*)\}\)`)
	sub := exp.FindSubmatch(r)
	if sub == nil {
		return errInvalidResponse
	}
	var res ptaskResponse
	err = json.Unmarshal(sub[1], &res)
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

func (s *session) removeTask(t *Task, flag byte) error {
	var delType = t.status()
	if delType == flagInvalid {
		return errInvalidTaskFlag
	} else if delType == flagPurged {
		return errTaskAlreadyPurged
	} else if flag == 0 && delType == flagDeleted {
		return errTaskAlreadyRemoved
	}
	ct := currentTimestamp()
	uri := fmt.Sprintf(taskdeleteURI, ct, delType, ct)
	data := url.Values{}
	data.Add("taskids", t.Id+",")
	data.Add("databases", "0,")
	data.Add("interfrom", "task")
	r, err := s.post(uri, data.Encode())
	if err != nil {
		return err
	}
	if ok, _ := regexp.Match(`\{"result":1,"type":`, r); ok {
		log.Debugf("remove task: %s", r)
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

func (s *session) tasklistNofresh(tid, page int) ([]byte, error) {
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
	resp, err := s.routine(req)
	if err != nil {
		return nil, err
	}
	log.Debug(resp.Status)
	r, err := readBody(resp)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}
	exp := regexp.MustCompile(`rebuild\((\{.*\})\)`)
	sub := exp.FindSubmatch(r)
	if sub == nil {
		return nil, errInvalidResponse
	}
	return sub[1], nil
}

func (s *session) getVerifyImage() (w io.WriterTo, err error) {
	uri := fmt.Sprintf(getVerifyURL(), currentTimestamp())
	log.Debugf("==> %s", uri)
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return
	}
	req.Header.Add("User-Agent", userAgent)
	resp, err := s.routine(req)
	if err != nil {
		return
	}
	log.Debug(resp.Status)
	if resp.StatusCode/100 > 3 {
		err = errors.New(resp.Status)
		return
	}
	w, err = ascii.Decode(resp.Body, ascii.Options{
		Invert: true,
		Color:  true,
	})
	resp.Body.Close()
	if err != nil {
		return
	}
	return
}

// func verifyLogin() bool {
// 	r, err := defaultSession.get(verifyLoginURI)
// 	if err != nil {
// 		log.Error(err.Error())
// 		return false
// 	}
// 	exp := regexp.MustCompile(`.*\((\{.*\})\)`)
// 	s := exp.FindSubmatch(r)
// 	if s == nil {
// 		log.Debugf("Response: %s", r)
// 		return false
// 	}
// 	var resp loginResponse
// 	json.Unmarshal(s[1], &resp)
// 	if resp.Result == 0 {
// 		log.Debugf("Response: %s", s[1])
// 		return false
// 	}
// 	fmt.Printf(`
// -------------+-------------
//   user id    |  %s
//   user name  |  %s
//   user newno |  %s
//   user type  |  %s
//   nickname   |  %s
//   user nick	 |  %s
//   vip state	 |  %s
// -------------+-------------
// `, resp.Data.UserId, resp.Data.UserName,
// 		resp.Data.UserNewno, resp.Data.UserType,
// 		resp.Data.Nickname, resp.Data.UserNick, resp.Data.VipState)
// 	return true
// }
