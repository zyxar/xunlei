package protocol

func GetTasksAsync(callback func([]*Task, error)) {
	go func() {
		ts, err := GetTasks()
		if callback != nil {
			callback(ts, err)
		}
	}()
}

func GetCompletedTasksAsync(callback func([]*Task, error)) {
	go func() {
		ts, err := GetCompletedTasks()
		if callback != nil {
			callback(ts, err)
		}
	}()
}

func GetIncompletedTasksAsync(callback func([]*Task, error)) {
	go func() {
		ts, err := GetIncompletedTasks()
		if callback != nil {
			callback(ts, err)
		}
	}()
}

func GetExpiredTasksAsync(callback func([]*Task, error)) {
	go func() {
		ts, err := GetExpiredTasks()
		if callback != nil {
			callback(ts, err)
		}
	}()
}

func GetDeletedTasksAsync(callback func([]*Task, error)) {
	go func() {
		ts, err := GetDeletedTasks()
		if callback != nil {
			callback(ts, err)
		}
	}()
}

func GetGdriveIdAsync(callback func(string, error)) {
	go func() {
		gid, err := GetGdriveId()
		if callback != nil {
			callback(gid, err)
		}
	}()
}

func DelayTaskAsync(id string, callback func(error)) {
	go func() {
		err := DelayTask(id)
		if callback != nil {
			callback(err)
		}
	}()
}

func FillBtListAsync(taskid, infohash string, callback func(*btList, error)) {
	go func() {
		l, err := FillBtList(taskid, infohash)
		if callback != nil {
			callback(l, err)
		}
	}()
}

func AddTaskAsync(req string, callback func(error)) {
	go func() {
		err := AddTask(req)
		if callback != nil {
			callback(err)
		}
	}()
}

func AddBatchTasksAsync(req []string, callback func(error), oids ...string) {
	go func() {
		err := AddBatchTasks(req, oids...)
		if callback != nil {
			callback(err)
		}
	}()
}

func GetTorrentByHashAsync(hash string, callback func([]byte, error)) {
	go func() {
		b, err := GetTorrentByHash(hash)
		if callback != nil {
			callback(b, err)
		}
	}()
}

func PauseTasksAsync(ids []string, callback func(error)) {
	go func() {
		err := PauseTasks(ids)
		if callback != nil {
			callback(err)
		}
	}()
}

func DelayAllTasksAsync(callback func(error)) {
	go func() {
		err := DelayAllTasks()
		if callback != nil {
			callback(err)
		}
	}()
}

func RenameTaskAsync(id, name string, callback func(error)) {
	go func() {
		err := RenameTask(id, name)
		if callback != nil {
			callback(err)
		}
	}()
}

func DeleteTaskAsync(id string, callback func(error)) {
	go func() {
		err := DeleteTask(id)
		if callback != nil {
			callback(err)
		}
	}()
}

func PurgeTaskAsync(id string, callback func(error)) {
	go func() {
		err := PurgeTask(id)
		if callback != nil {
			callback(err)
		}
	}()
}

func ResumeTaskAsync(id string, callback func(error)) {
	go func() {
		t := M.getTaskbyId(id)
		var err error
		if t != nil {
			err = t.Resume()
		} else {
			err = errNoSuchTask
		}
		if callback != nil {
			callback(err)
		}
	}()
}
