package protocol

func GetTasksAsync(callback func([]*Task, error)) {
	go func() {
		ts, err := defaultSession.GetTasks()
		if callback != nil {
			callback(ts, err)
		}
	}()
}

func GetCompletedTasksAsync(callback func([]*Task, error)) {
	go func() {
		ts, err := defaultSession.GetCompletedTasks()
		if callback != nil {
			callback(ts, err)
		}
	}()
}

func GetIncompletedTasksAsync(callback func([]*Task, error)) {
	go func() {
		ts, err := defaultSession.GetIncompletedTasks()
		if callback != nil {
			callback(ts, err)
		}
	}()
}

func GetExpiredTasksAsync(callback func([]*Task, error)) {
	go func() {
		ts, err := defaultSession.GetExpiredTasks()
		if callback != nil {
			callback(ts, err)
		}
	}()
}

func GetDeletedTasksAsync(callback func([]*Task, error)) {
	go func() {
		ts, err := defaultSession.GetDeletedTasks()
		if callback != nil {
			callback(ts, err)
		}
	}()
}

func GetGdriveIdAsync(callback func(string, error)) {
	go func() {
		gid, err := defaultSession.GetGdriveId()
		if callback != nil {
			callback(gid, err)
		}
	}()
}

func DelayTaskAsync(id string, callback func(error)) {
	go func() {
		err := defaultSession.DelayTaskById(id)
		if callback != nil {
			callback(err)
		}
	}()
}

func FillBtListAsync(taskid, infohash string, callback func(*btList, error)) {
	go func() {
		l, err := defaultSession.FillBtListById(taskid, infohash)
		if callback != nil {
			callback(l, err)
		}
	}()
}

func AddTaskAsync(req string, callback func(error)) {
	go func() {
		err := defaultSession.AddTask(req)
		if callback != nil {
			callback(err)
		}
	}()
}

func AddBatchTasksAsync(req []string, callback func(error), oids ...string) {
	go func() {
		err := defaultSession.AddBatchTasks(req, oids...)
		if callback != nil {
			callback(err)
		}
	}()
}

func GetTorrentByHashAsync(hash string, callback func([]byte, error)) {
	go func() {
		b, err := defaultSession.GetTorrentByHash(hash)
		if callback != nil {
			callback(b, err)
		}
	}()
}

func PauseTasksAsync(ids []string, callback func(error)) {
	go func() {
		err := defaultSession.PauseTasks(ids)
		if callback != nil {
			callback(err)
		}
	}()
}

func DelayAllTasksAsync(callback func(error)) {
	go func() {
		err := defaultSession.DelayAllTasks()
		if callback != nil {
			callback(err)
		}
	}()
}

func RenameTaskAsync(id, name string, callback func(error)) {
	go func() {
		err := defaultSession.RenameTaskById(id, name)
		if callback != nil {
			callback(err)
		}
	}()
}

func DeleteTaskAsync(id string, callback func(error)) {
	go func() {
		err := defaultSession.DeleteTaskById(id)
		if callback != nil {
			callback(err)
		}
	}()
}

func PurgeTaskAsync(id string, callback func(error)) {
	go func() {
		err := defaultSession.PurgeTaskById(id)
		if callback != nil {
			callback(err)
		}
	}()
}

func ResumeTaskAsync(id string, callback func(error)) {
	go func() {
		err := defaultSession.ResumeTaskById(id)
		if callback != nil {
			callback(err)
		}
	}()
}
