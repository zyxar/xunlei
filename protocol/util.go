package protocol

import (
	"bufio"
	"bytes"
	"compress/flate"
	"compress/gzip"
	"crypto/md5"
	"fmt"
	"html"
	"io"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/zyxar/ed2k"
)

func readBody(resp *http.Response) ([]byte, error) {
	buffer := bytes.NewBuffer(make([]byte, 0, 1024))
	var rd io.ReadCloser
	var err error
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		rd, _ = gzip.NewReader(resp.Body)
		defer rd.Close()
	case "deflate":
		rd = flate.NewReader(resp.Body)
		defer rd.Close()
	default:
		rd = resp.Body
	}
	defer func() {
		e := recover()
		if e == nil {
			return
		}
		if panicErr, ok := e.(error); ok && panicErr == bytes.ErrTooLarge {
			err = panicErr
		} else {
			panic(e)
		}
	}()
	_, err = buffer.ReadFrom(rd)
	return buffer.Bytes(), err
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

func currentTimestamp() int64 {
	return time.Now().UnixNano() / 1000000
}

func currentRandom() string {
	return fmt.Sprintf("%d%v", currentTimestamp(), rand.Float64()*1000000)
}

func hashPass(pass, vcode string) string {
	h := md5.New()
	v := EncryptPass(pass)
	io.WriteString(h, v)
	io.WriteString(h, vcode)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func EncryptPass(pass string) string {
	if len(pass) == 32 {
		if ok, _ := regexp.MatchString(`[a-f0-9]{32,32}`, pass); ok {
			return pass
		}
	}
	h := md5.New()
	v := pass
	io.WriteString(h, v)
	v = fmt.Sprintf("%x", h.Sum(nil))
	h.Reset()
	io.WriteString(h, v)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func md5sum(raw interface{}) []byte {
	h := md5.New()
	switch raw.(type) {
	case []byte:
		h.Write(raw.([]byte))
	case string:
		io.WriteString(h, raw.(string))
	default:
		io.WriteString(h, fmt.Sprintf("%s", raw))
	}
	return h.Sum(nil)
}

func getTaskPre(resp []byte) (*taskPrepare, error) {
	exp := regexp.MustCompile(`queryCid\((.*)\)`)
	s := exp.FindSubmatch(resp)
	if s == nil {
		return nil, errInvalidResponse
	}
	ss := bytes.Split(s[1], []byte(","))
	j := 0
	if len(ss) >= 10 {
		j = 1
	}
	ret := taskPrepare{}
	ret.Cid = string(bytes.Trim(ss[0], "' "))
	ret.GCid = string(bytes.Trim(ss[1], "' "))
	ret.SizeCost = string(bytes.Trim(ss[2], "' "))
	ret.FileName = string(bytes.Trim(ss[j+3], "' "))
	ret.Goldbean = string(bytes.Trim(ss[j+4], "' "))
	ret.Silverbean = string(bytes.Trim(ss[j+5], "' "))
	var err error
	if ret.Goldbean != "0" || ret.Silverbean != "0" {
		err = fmt.Errorf("Task need bean: %s:%s", ret.Goldbean, ret.Silverbean)
	}
	return &ret, err
}

func parseBtQueryResponse(queryURL []byte) *btQueryResponse {
	exp := regexp.MustCompile(`'([0-9A-Za-z]{40,40})','(\d*)','(.*)','(\d)',new Array\((.*)\),new Array\((.*)\),new Array\((.*)\),new Array\((.*)\),new Array\((.*)\),new Array\((.*)\),'([\d\.]+)','(\d)'`)
	s := exp.FindSubmatch(queryURL)
	if s == nil {
		return nil
	}
	var resp btQueryResponse
	resp.InfoId = string(s[1])
	resp.Size = string(s[2])
	resp.Name = string(s[3])
	resp.IsFull = string(s[4])
	a := bytes.Split(s[5], []byte(","))
	resp.Files = make([]string, len(a))
	for i := 0; i < len(a); i++ {
		resp.Files[i] = string(bytes.Trim(a[i], "' "))
	}
	a = bytes.Split(s[6], []byte(","))
	resp.Sizesf = make([]string, len(a))
	for i := 0; i < len(a); i++ {
		resp.Sizesf[i] = string(bytes.Trim(a[i], "' "))
	}
	a = bytes.Split(s[7], []byte(","))
	resp.Sizes = make([]string, len(a))
	for i := 0; i < len(a); i++ {
		resp.Sizes[i] = string(bytes.Trim(a[i], "' "))
	}
	a = bytes.Split(s[8], []byte(","))
	resp.Picked = make([]string, len(a))
	for i := 0; i < len(a); i++ {
		resp.Picked[i] = string(bytes.Trim(a[i], "' "))
	}
	a = bytes.Split(s[9], []byte(","))
	resp.Ext = make([]string, len(a))
	for i := 0; i < len(a); i++ {
		resp.Ext[i] = string(bytes.Trim(a[i], "' "))
	}
	a = bytes.Split(s[10], []byte(","))
	resp.Index = make([]string, len(a))
	for i := 0; i < len(a); i++ {
		resp.Index[i] = string(bytes.Trim(a[i], "' "))
	}
	resp.Random = string(s[11])
	resp.Ret = string(s[12])
	return &resp
}

func extractTasks(ts []*Task) (urls []string, ids []string) {
	ids = make([]string, 0, len(ts))
	urls = make([]string, 0, len(ts))
	for i := range ts {
		ids = append(ids, ts[i].Id)
		urls = append(urls, ts[i].URL)
	}
	return
}

func getEd2kHash(filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer f.Close()
	rd := bufio.NewReader(f)
	eh := ed2k.New()
	_, err = rd.WriteTo(eh)
	return fmt.Sprintf("%x", eh.Sum(nil)), err
}

func getEd2kHashFromURL(uri string) string {
	h := strings.Split(uri, "|")
	if len(h) > 4 {
		return strings.ToLower(h[4])
	}
	return ""
}

func unescapeName(s string) string {
	var length int
	for {
		length = len(s)
		s = html.UnescapeString(s)
		if len(s) == length {
			break
		}
	}
	return s
}

func trimHTMLFontTag(raw string) string {
	exp := regexp.MustCompile(`<font color='([a-z]*)'>(.*)</font>`)
	sub := exp.FindStringSubmatch(raw)
	if sub == nil {
		return raw
	}
	return sub[2]
}
