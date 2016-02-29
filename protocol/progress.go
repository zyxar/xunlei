package protocol

type progressResponse struct {
	List []struct {
		Progress string `json:"progress"`
		UrlHash  string `json:"url_hash"`
	} `json:"progress_info_list"`
	Uid string `json:"userid"`
	Ret byte   `json:"ret"`
}
