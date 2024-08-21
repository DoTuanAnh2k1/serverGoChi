package list

type (
	NeResponse struct {
		Status     string   `json:"status"`
		Code       string   `json:"code"`
		Message    string   `json:"message"`
		NeDataList []NeData `json:"neDataList"`
	}

	NeData struct {
		Site        string `json:"site"`
		Ne          string `json:"ne"`
		Ip          string `json:"ip"`
		Description string `json:"description"`
		Namespace   string `json:"namespace"`
		Port        int32  `json:"port"`
		UrlList     []Url  `json:"urlList"`
	}

	Url struct {
		IpAddress string `json:"ipAddress"`
		Port      int    `json:"port"`
	}

	NeMonitorInfo struct {
		Site         string `json:"site"`
		Ne           string `json:"ne"`
		Ip           string `json:"ip"`
		Description  string `json:"description"`
		Namespace    string `json:"namespace"`
		Port         int32  `json:"port"`
		NeMonitorURL string `json:"ne_monitor_url"`
	}
)
