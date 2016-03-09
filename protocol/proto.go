package protocol

func Login(id, passhash string) (err error)       { return defaultSession.Login(id, passhash) }
func SaveSession(cookieFile string) error         { return defaultSession.SaveSession(cookieFile) }
func ResumeSession(cookieFile string) (err error) { return defaultSession.ResumeSession(cookieFile) }
func GetAccount() *UserAccount                    { return defaultSession.Account() }
func IsOn() bool                                  { return defaultSession.IsOn() }
func GetTasks(limit ...int) ([]*Task, error)      { return defaultSession.GetTasks(limit...) }
func GetCompletedTasks() ([]*Task, error)         { return defaultSession.GetCompletedTasks() }
func GetIncompletedTasks() ([]*Task, error)       { return defaultSession.GetIncompletedTasks() }
func GetGdriveId() (gid string, err error)        { return defaultSession.GetGdriveId() }
func RawTaskList(category, page int) ([]byte, error) {
	return defaultSession.RawTaskList(category, page)
}
func RawTaskListExpired() ([]byte, error)         { return defaultSession.RawTaskListExpired() }
func RawTaskListDeleted(page int) ([]byte, error) { return defaultSession.RawTaskListDeleted(page) }
func GetExpiredTasks() ([]*Task, error)           { return defaultSession.GetExpiredTasks() }
func GetDeletedTasks() ([]*Task, error)           { return defaultSession.GetDeletedTasks() }
func DelayTask(t *Task) error                     { return defaultSession.DelayTask(t) }
func DelayTaskById(taskid string) error           { return defaultSession.DelayTaskById(taskid) }
func FillBtList(t *Task) (*btList, error) {
	return defaultSession.FillBtList(t)
}
func FillBtListById(taskid, infohash string) (*btList, error) {
	return defaultSession.FillBtListById(taskid, infohash)
}
func RawFillBtList(t *Task, page int) ([]byte, error) {
	return defaultSession.RawFillBtList(t, page)
}
func RawFillBtListById(taskid, infohash string, page int) ([]byte, error) {
	return defaultSession.RawFillBtListById(taskid, infohash, page)
}
func AddTask(req string) error { return defaultSession.AddTask(req) }
func AddBatchTasks(urls []string, oids ...string) error {
	return defaultSession.AddBatchTasks(urls, oids...)
}
func ProcessTaskDaemon(ch chan byte, callback TaskCallback) {
	defaultSession.ProcessTaskDaemon(ch, callback)
}
func ProcessTask(callback TaskCallback) error      { return defaultSession.ProcessTask(callback) }
func GetTorrentByHash(hash string) ([]byte, error) { return defaultSession.GetTorrentByHash(hash) }
func GetTorrentFileByHash(hash, file string) error {
	return defaultSession.GetTorrentFileByHash(hash, file)
}
func PauseTask(t *Task) error                  { return defaultSession.PauseTask(t) }
func PauseTasks(ids []string) error            { return defaultSession.PauseTasks(ids) }
func DelayAllTasks() error                     { return defaultSession.DelayAllTasks() }
func ReAddTask(t *Task) error                  { return defaultSession.ReAddTask(t) }
func ReAddTasks(ts map[string]*Task)           { defaultSession.ReAddTasks(ts) }
func RenameTask(t *Task, newname string) error { return defaultSession.RenameTask(t, newname) }
func RenameTaskById(taskid, newname string) error {
	return defaultSession.RenameTaskById(taskid, newname)
}
func ResumeTask(t *Task) error           { return defaultSession.ResumeTask(t) }
func ResumeTaskById(taskid string) error { return defaultSession.ResumeTaskById(taskid) }
func DeleteTask(t *Task) error           { return defaultSession.DeleteTask(t) }
func DeleteTaskById(taskid string) error { return defaultSession.DeleteTaskById(taskid) }
func PurgeTask(t *Task) error            { return defaultSession.PurgeTask(t) }
func PurgeTaskById(taskid string) error  { return defaultSession.PurgeTaskById(taskid) }
func VerifyTask(t *Task, path string) bool {
	return defaultSession.VerifyTask(t, path)
}
func FindTasks(pattern string) (map[string]*Task, error) { return defaultSession.FindTasks(pattern) }
func GetTaskById(taskid string) (t *Task, exist bool)    { return defaultSession.GetTaskById(taskid) }
func GetTasksByIds(ids []string) map[string]*Task        { return defaultSession.GetTasksByIds(ids) }
func InvalidateCache(flag byte)                          { defaultSession.InvalidateCache(flag) }
func InvalidateCacheAll()                                { defaultSession.InvalidateCacheAll() }
