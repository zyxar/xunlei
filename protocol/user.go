package protocol

type loginResponse struct {
	Result int `json:"result"`
	Data   struct {
		UserId    string `json:"userid"`
		UserName  string `json:"usrname"`
		UserNewno string `json:"usernewno"`
		UserType  string `json:"usrtype"`
		Nickname  string `json:"nickname"`
		UserNick  string `json:"usernick"`
		VipState  string `json:"vipstate"`
	} `json:"data"`
}

type UserAccount struct {
	ExpireDate          string `json:"expire_date"`
	MaxTaskNum          string `json:"max_task_num"`
	MaxStore            string `json:"max_store"`
	VipStore            string `json:"vip_store"`
	BuyStore            string `json:"buy_store"`
	XzStore             string `json:"xz_store"`
	BuyNumTask          string `json:"buy_num_task"`
	BuyNumConn          string `json:"buy_num_connection"`
	BuyBandwith         string `json:"buy_bandwidth"`
	BuyTaskLiveTime     string `json:"buy_task_live_time"`
	XpExpireDate        string `json:"experience_expire_date"`
	AvailableSpace      string `json:"available_space"`
	TotalNum            string `json:"total_num"`
	HistoryTaskTotalNum string `json:"history_task_total_num"`
	SuspendingNum       string `json:"suspending_num"`
	DownloadingNum      string `json:"downloading_num"`
	WaitingNum          string `json:"waiting_num"`
	CompleteNum         string `json:"complete_num"`
	StorePeriod         string `json:"store_period"`
	Cookie              string `json:"cookie"`
	VipLevel            string `json:"vip_level"`
	UserType            string `json:"user_type"`
	GoldbeanNum         string `json:"goldbean_num"`
	ConvertFlag         string `json:"convert_flag"`
	SilverbeanNum       string `json:"silverbean_num"`
	SpecialNet          string `json:"special_net"`
	TotalFilterNum      string `json:"total_filter_num"`
}

type UserInfo struct {
	AllSpace       string `json:"all_space"`
	AllUsedStore   int64  `json:"all_used_store"`
	AllSpaceFormat string `json:"all_space_format"`
	AllUsedFormat  string `json:"all_used_format"`
	Isp            bool   `json:"isp"`
	Percent        string `json:"percent"`
}
