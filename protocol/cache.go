package api

import (
	"errors"
	"net/url"
	"regexp"
	"sync"
)

type cache struct {
	Uid         string
	Gid         string
	Account     *_user
	AccountInfo *_userinfo
	Tasks       map[string]*Task
	sync.Mutex
}

var M cache
var invalidQueryErr = errors.New("Invalid query string.")

func init() {
	M.Tasks = make(map[string]*Task)
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
	var ts map[string]*Task = M.Tasks
	n := v.Get("name")
	gg := v["group"]
	ss := v["status"]
	tt := v["type"]
	if len(tt) > 0 {
		tr := make(map[string]*Task)
		for k, _ := range tt {
			switch tt[k] {
			case "bt":
				for i, _ := range ts {
					if ts[i].IsBt() {
						tr[i] = ts[i]
					}
				}
			case "nbt":
				for i, _ := range ts {
					if !ts[i].IsBt() {
						tr[i] = ts[i]
					}
				}
			default:
				return nil, invalidQueryErr
			}
		}
		ts = tr
	}
	if len(ss) > 0 {
		tr := make(map[string]*Task)
		for k, _ := range ss {
			switch ss[k] {
			case "normal":
				for i, _ := range ts {
					if ts[i].normal() {
						tr[i] = ts[i]
					}
				}
			case "expired":
				for i, _ := range ts {
					if ts[i].expired() {
						tr[i] = ts[i]
					}
				}
			case "deleted":
				for i, _ := range ts {
					if ts[i].deleted() {
						tr[i] = ts[i]
					}
				}
			case "purged": // for rescue
				for i, _ := range ts {
					if ts[i].purged() {
						tr[i] = ts[i]
					}
				}
			default:
				return nil, invalidQueryErr
			}
		}
		ts = tr
	}
	if len(gg) > 0 {
		tr := make(map[string]*Task)
		for k, _ := range gg {
			switch gg[k] {
			case "waiting":
				for i, _ := range ts {
					if ts[i].waiting() {
						tr[i] = ts[i]
					}
				}
			case "downloading":
				for i, _ := range ts {
					if ts[i].downloading() {
						tr[i] = ts[i]
					}
				}
			case "completed":
				for i, _ := range ts {
					if ts[i].completed() {
						tr[i] = ts[i]
					}
				}
			case "failed":
				for i, _ := range ts {
					if ts[i].failed() {
						tr[i] = ts[i]
					}
				}
			case "pending":
				for i, _ := range ts {
					if ts[i].pending() {
						tr[i] = ts[i]
					}
				}
			default:
				return nil, invalidQueryErr
			}
		}
		ts = tr
	}
	if len(n) > 0 {
		exp, err := regexp.Compile(`(?i)` + n)
		if err != nil {
			return nil, invalidQueryErr
		}
		tr := make(map[string]*Task)
		for i, _ := range ts {
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
	for i, _ := range ids {
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
	for i, _ := range ts {
		c.Tasks[ts[i].Id] = ts[i]
	}
	c.Unlock()
}

func (c *cache) invalidateGroup(flag byte) {
	c.Lock()
	for i, _ := range c.Tasks {
		if c.Tasks[i].status() == flag {
			delete(c.Tasks, i)
		}
	}
	c.Unlock()
}

func (c *cache) invalidateAll() {
	c.Lock()
	c.Tasks = make(map[string]*Task)
	c.Unlock()
}

func (c *cache) replaceAll(ts []*Task) {
	m := make(map[string]*Task)
	for i, _ := range ts {
		m[ts[i].Id] = ts[i]
	}
	c.Lock()
	c.Tasks = m
	c.Unlock()
}
