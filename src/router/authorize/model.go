package authorize

// Request
type (
	NeSetReq struct {
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
)
