package list

import (
	"net/http"
	"serverGoChi/src/log"
	"serverGoChi/src/router/response"
)

func HandlerListNe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		log.Logger.Error("Method not allowed")
		response.Write(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

}

type NeResponse struct {
	Status     string   `json:"status"`
	Code       string   `json:"code"`
	Message    string   `json:"message"`
	NeDataList []NeData `json:"neDataList"`
}

type NeData struct {
	Site        string `json:"site"`
	Ne          string `json:"ne"`
	Ip          string `json:"ip"`
	Description string `json:"description"`
	Namespace   string `json:"namespace"`
	Port        int    `json:"port"`
	UrlList     []Url  `json:"urlList"`
}

type Url struct {
	IpAddress string `json:"ipAddress"`
	Port      int    `json:"port"`
}
