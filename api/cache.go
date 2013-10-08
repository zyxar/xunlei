package api

import (
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

func init() {
	M.Tasks = make(map[string]*Task)
}

func (c *cache) getTaskbyId(taskid string) *Task {
	return c.Tasks[taskid]
}

func (c *cache) getTasksByIds(ids []string) []*Task {
	ts := make([]*Task, 0, len(ids))
	for i, _ := range ids {
		ts = append(ts, c.Tasks[ids[i]])
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
