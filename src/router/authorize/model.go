package authorize

// Request
type (
	NeSetReq struct {
		Username string `json:"username"`
		NeId     string `json:"neid"`
	}

	NeDeleteReq struct {
		Username string `json:"username"`
		NeId     string `json:"neid"`
	}

	UserDeleteReq struct {
		Username   string `json:"username"`
		Permission string `json:"permission"`
	}

	UserSetReq struct {
		Username   string `json:"username"`
		Permission string `json:"permission"`
	}

	AuthorizeReq struct {
		Ne         string `json:"ne"`
		Site       string `json:"site"`
		SystemType string `json:"system_type"`
	}
)

// Response
type (
	UserShowResp struct {
		Username    string `json:"username"`
		Permissions string `json:"permissions"`
	}

	NeShowResp struct {
		Name        string `json:"name"`
		SiteName    string `json:"site_name"`
		IpAddress   string `json:"ip_address"`
		Port        int32  `json:"port"`
		Description string `json:"description"`
		Id          int64  `json:"id"`
	}

	AuthorizeResp struct {
		Status           string         `json:"status"`
		ResponseCode     string         `json:"response_code"`
		ResponseDataList []ResponseData `json:"responseDataList"`
	}
)

// Sub Struct
type (
	ResponseData struct {
		Command string `json:"command"`
	}

	FiveGResponseData struct {
		fiveGObject FiveGObject `json:"params"`
	}

	FiveGObject struct {
	}
)
