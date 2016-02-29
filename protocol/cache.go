package protocol

import (
	"errors"
	"net/url"
	"regexp"
	"sync"
)

var errInvalidQuery = errors.New("Invalid query string.")

type cache struct {
	Uid         string
	Gid         string
	Account     *userAccount
	AccountInfo *userInfo
	Tasks       map[string]*Task
	sync.Mutex
}

func newCache() *cache {
	return &cache{
		Tasks: make(map[string]*Task),
	}
}

func (c *cache) getTaskbyId(taskid string) *Task {
	return c.Tasks[taskid]
}

// Find tasks in local cache by query: `name=xxx`, `group=g`, `status=s`, `type=t`
// name: `xxx` is considered as a regular expression.
// group: waiting, downloading, completed, failed, pending
// status: normal, expired, deleted, purged
// type: bt, nbt
// e.g. pattern == "name=abc&group=completed&status=normal&type=bt"
func FindTasks(pattern string) (map[string]*Task, error) {
	v, err := url.ParseQuery(pattern)
	if err != nil {
		return nil, err
	}
	var ts = M.Tasks
	n := v.Get("name")
	gg := v["group"]
	ss := v["status"]
	tt := v["type"]
	if len(tt) > 0 {
		tr := make(map[string]*Task)
		for k := range tt {
			switch tt[k] {
			case "bt":
				for i := range ts {
					if ts[i].IsBt() {
						tr[i] = ts[i]
					}
				}
			case "nbt":
				for i := range ts {
					if !ts[i].IsBt() {
						tr[i] = ts[i]
					}
				}
			default:
				return nil, errInvalidQuery
			}
		}
		ts = tr
	}
	if len(ss) > 0 {
		tr := make(map[string]*Task)
		for k := range ss {
			switch ss[k] {
			case "normal":
				for i := range ts {
					if ts[i].normal() {
						tr[i] = ts[i]
					}
				}
			case "expired":
				for i := range ts {
					if ts[i].expired() {
						tr[i] = ts[i]
					}
				}
			case "deleted":
				for i := range ts {
					if ts[i].deleted() {
						tr[i] = ts[i]
					}
				}
			case "purged": // for rescue
				for i := range ts {
					if ts[i].purged() {
						tr[i] = ts[i]
					}
				}
			default:
				return nil, errInvalidQuery
			}
		}
		ts = tr
	}
	if len(gg) > 0 {
		tr := make(map[string]*Task)
		for k := range gg {
			switch gg[k] {
			case "waiting":
				for i := range ts {
					if ts[i].waiting() {
						tr[i] = ts[i]
					}
				}
			case "downloading":
				for i := range ts {
					if ts[i].downloading() {
						tr[i] = ts[i]
					}
				}
			case "completed":
				for i := range ts {
					if ts[i].completed() {
						tr[i] = ts[i]
					}
				}
			case "failed":
				for i := range ts {
					if ts[i].failed() {
						tr[i] = ts[i]
					}
				}
			case "pending":
				for i := range ts {
					if ts[i].pending() {
						tr[i] = ts[i]
					}
				}
			default:
				return nil, errInvalidQuery
			}
		}
		ts = tr
	}
	if len(n) > 0 {
		exp, err := regexp.Compile(`(?i)` + n)
		if err != nil {
			return nil, errInvalidQuery
		}
		tr := make(map[string]*Task)
		for i := range ts {
			if exp.MatchString(ts[i].TaskName) {
				tr[i] = ts[i]
			}
		}
		ts = tr
	}
	return ts, nil
}

func (c *cache) GetTasksByIds(ids []string) map[string]*Task {
	ts := make(map[string]*Task)
	for i := range ids {
		if _, ok := c.Tasks[ids[i]]; ok {
			ts[ids[i]] = c.Tasks[ids[i]]
		}
	}
	return ts
}

func (c *cache) dropTask(taskid string) {
	c.Lock()
	if _, Ok := c.Tasks[taskid]; Ok {
		delete(c.Tasks, taskid)
	}
	c.Unlock()
}

func (c *cache) pushTask(t *Task) {
	c.Lock()
	c.Tasks[t.Id] = t
	c.Unlock()
}

func (c *cache) pushTasks(ts []*Task) {
	c.Lock()
	for i := range ts {
		c.Tasks[ts[i].Id] = ts[i]
	}
	c.Unlock()
}

func (c *cache) InvalidateGroup(flag byte) {
	c.Lock()
	for i := range c.Tasks {
		if c.Tasks[i].status() == flag {
			delete(c.Tasks, i)
		}
	}
	c.Unlock()
}

func (c *cache) InvalidateAll() {
	c.Lock()
	c.Tasks = make(map[string]*Task)
	c.Unlock()
}

func (c *cache) replaceAll(ts []*Task) {
	m := make(map[string]*Task)
	for i := range ts {
		m[ts[i].Id] = ts[i]
	}
	c.Lock()
	c.Tasks = m
	c.Unlock()
}
