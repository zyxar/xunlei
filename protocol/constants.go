package protocol

const (
	userAgent  = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_10_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/48.0.2564.116 Safari/537.36"
	pageSize   = "100"
	btPageSize = "999"
)

const (
	taskTypeOrdinary = iota
	taskTypeBt
	taskTypeEd2k
	taskTypeInvalid
	taskTypeMagnet
)

const (
	statusWaiting = iota
	statusDownloading
	statusCompleted
	statusFailed
	statusMixed
	statusPending
)

const (
	flagNormal byte = iota
	flagDeleted
	flagPurged
	flagInvalid
	flagExpired
)

const (
	deletedCk = `page_check_all=history&fltask_all_guoqi=1&class_check=0&page_check=task&fl_page_id=0&class_check_new=0&set_tab_status=11`
	expiredCk = `page_check_all=history&class_check=0&page_check=task&fl_page_id=0&class_check_new=0&set_tab_status=13`
)

const (
	domainLixianURI    = "http://dynamic.cloud.vip.xunlei.com/"
	taskBaseURI        = domainLixianURI + "user_task?userid=%s"
	taskHomeURI        = domainLixianURI + "user_task?userid=%s&st=4"
	taskPageURI        = domainLixianURI + "user_task?userid=%s&st=%s&p=%s"
	historyHomeURI     = domainLixianURI + "user_history?userid=%s"
	expireHomeURI      = domainLixianURI + "user_history?type=1&userid=%s"
	historyPageURI     = domainLixianURI + "user_history?userid=%s&p=%d"
	applyHomeURI       = domainLixianURI + "user_apply?userid=%s"
	applyPageURI       = domainLixianURI + "user_apply?userid=%s&p=%s"
	loginURI           = domainLixianURI + "login"
	interfaceURI       = domainLixianURI + "interface"
	verifyLoginURI     = interfaceURI + "/verify_login"
	taskdelayURI       = interfaceURI + "/task_delay?taskids=%s&interfrom=%s&noCacheIE=%d"
	gettorrentURI      = interfaceURI + "/get_torrent?userid=%s&infoid=%s"
	taskpauseURI       = interfaceURI + "/task_pause?tid=%s&uid=%s&noCacheIE=%d"
	redownloadURI      = interfaceURI + "/redownload?callback=jsonp%d"
	showclassURI       = interfaceURI + "/show_class?callback=jsonp%d&type_id=%d"
	menugetURI         = interfaceURI + "/menu_get"
	fillbtlistURI      = interfaceURI + "/fill_bt_list?callback=fill_bt_list&tid=%s&infoid=%s&g_net=1&p=%d&uid=%s&interfrom=%s&noCacheIE=%d"
	taskcheckURI       = interfaceURI + "/task_check?callback=queryCid&url=%s&interfrom=%s&random=%s&tcache=%d"
	taskcommitURI      = interfaceURI + "/task_commit?"
	batchtaskcheckURI  = interfaceURI + "/batch_task_check"
	batchtaskcommitURI = interfaceURI + "/batch_task_commit?callback=jsonp%d&t=%d"
	torrentuploadURI   = interfaceURI + "/torrent_upload"
	bttaskcommitURI    = interfaceURI + "/bt_task_commit?callback=jsonp%d"
	urlqueryURI        = interfaceURI + "/url_query?callback=queryUrl&u=%s&random=%s"
	delayonceURI       = interfaceURI + "/delay_once?callback=anything"
	renameURI          = interfaceURI + "/rename?"
	taskprocessURI     = interfaceURI + "/task_process?callback=jsonp%d&t=%d"
	taskdeleteURI      = interfaceURI + "/task_delete?callback=jsonp%d&type=%d&noCacheIE=%d"
	showtaskUnfreshURI = interfaceURI + "/showtask_unfresh?type_id=%d&page=%d&tasknum=%s&p=%d&interfrom=task"
	getplayurlURI      = interfaceURI + "/get_play_url?callback=jsonp%d&t=%d"
	vodBaseURI         = "http://i.vod.xunlei.com"
	reqGetMethodVodURI = vodBaseURI + "/req_get_method_vod?"
	historyPlayURI     = vodBaseURI + "/req_history_play_list/req_num/%d/req_offset/%d?type=%s&order=%s&t=%d"
	subbtURI           = vodBaseURI + "/req_subBT/info_hash/%s/req_num/%d/req_offset/%d"
	progressURI        = vodBaseURI + "/req_progress_query?&t=%d"
	lxtaskListURI      = vodBaseURI + "/req_lxtask_list"
)
