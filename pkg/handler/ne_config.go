package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/handler/middleware"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/handler/response"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/logger"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/service"
)

// HandlerNeConfigCreate tạo mới một bản ghi cấu hình kết nối cho NE (mode ne-config).
//
// Input : POST body JSON { "ne_id": int64, "ip_address": string,
//         "port": int32, "username": string, "password": string,
//         "protocol": string (SSH/TELNET/NETCONF/RESTCONF), "description": string }
// Output: 201 nếu tạo thành công
//         400 nếu thiếu ne_id hoặc ip_address
//         500 nếu lỗi DB
// Flow  : decode body → validate ne_id > 0 và ip_address không rỗng →
//         mặc định protocol="SSH" → lấy actor từ context →
//         CreateNeConfig → ghi operation history
func HandlerNeConfigCreate(w http.ResponseWriter, r *http.Request) {
	var req db_models.CliNeConfig
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.NeID == 0 || req.IPAddress == "" {
		response.Write(w, http.StatusBadRequest, "ne_id and ip_address are required")
		return
	}
	if req.Protocol == "" {
		req.Protocol = "SSH"
	}

	user := mustUser(r)
	op := opHistory("ne config create", fmt.Sprintf("ne_id=%d ip=%s", req.NeID, req.IPAddress), user.Username)

	if err := service.CreateNeConfig(&req); err != nil {
		logger.Logger.Error("ne_config create: ", err)
		saveHistory(op, "failure")
		response.InternalError(w, "failed to create NE config")
		return
	}
	saveHistory(op, "success")
	response.Created(w)
}

// HandlerNeConfigList lấy danh sách cấu hình kết nối của một NE cụ thể.
//
// Input : GET query param ?ne_id=<int64>
// Output: 200 [ ...CliNeConfig ] (mảng rỗng nếu chưa có cấu hình)
//         400 nếu ne_id không hợp lệ hoặc bằng 0
//         500 nếu lỗi DB
// Flow  : parse ne_id từ query → GetNeConfigByNeId → trả danh sách
func HandlerNeConfigList(w http.ResponseWriter, r *http.Request) {
	neIdStr := r.URL.Query().Get("ne_id")
	neId, err := strconv.ParseInt(neIdStr, 10, 64)
	if err != nil || neId == 0 {
		response.Write(w, http.StatusBadRequest, "ne_id query param is required")
		return
	}

	list, err := service.GetNeConfigByNeId(neId)
	if err != nil {
		response.InternalError(w, "failed to get NE config list")
		return
	}
	if list == nil {
		list = []*db_models.CliNeConfig{}
	}
	response.Write(w, http.StatusOK, list)
}

// HandlerNeConfigUpdate cập nhật thông tin một bản ghi cấu hình kết nối NE.
//
// Input : POST body JSON { "id": int64 (bắt buộc), và bất kỳ trường CliNeConfig nào cần sửa }
// Output: 200 "NE config updated" nếu thành công
//         400 nếu thiếu id hoặc body không hợp lệ
//         500 nếu lỗi DB
// Flow  : decode body → validate id > 0 → lấy actor từ context →
//         UpdateNeConfig → ghi operation history
func HandlerNeConfigUpdate(w http.ResponseWriter, r *http.Request) {
	var req db_models.CliNeConfig
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ID == 0 {
		response.Write(w, http.StatusBadRequest, "id is required")
		return
	}

	user := mustUser(r)
	op := opHistory("ne config update", fmt.Sprintf("id=%d", req.ID), user.Username)

	if err := service.UpdateNeConfig(&req); err != nil {
		logger.Logger.Error("ne_config update: ", err)
		saveHistory(op, "failure")
		response.InternalError(w, "failed to update NE config")
		return
	}
	saveHistory(op, "success")
	response.Success(w, "NE config updated")
}

// HandlerNeConfigDelete xoá một bản ghi cấu hình kết nối NE theo ID.
//
// Input : POST body JSON { "id": int64 }
// Output: 200 "NE config deleted" nếu thành công
//         400 nếu thiếu id hoặc body không hợp lệ
//         500 nếu lỗi DB
// Flow  : decode body → validate id > 0 → lấy actor từ context →
//         DeleteNeConfigById → ghi operation history
func HandlerNeConfigDelete(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID int64 `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ID == 0 {
		response.Write(w, http.StatusBadRequest, "id is required")
		return
	}

	user := mustUser(r)
	op := opHistory("ne config delete", fmt.Sprintf("id=%d", req.ID), user.Username)

	if err := service.DeleteNeConfigById(req.ID); err != nil {
		logger.Logger.Error("ne_config delete: ", err)
		saveHistory(op, "failure")
		response.InternalError(w, "failed to delete NE config")
		return
	}
	saveHistory(op, "success")
	response.Success(w, "NE config deleted")
}

// HandlerListNeConfig trả về toàn bộ cấu hình kết nối (ne-config) của các NE thuộc user hiện tại.
//
// Input : GET (không có body/query params; user lấy từ JWT context)
// Output: 200 [ { ne_name, ne_ip, site_name, config_list: [...CliNeConfig] } ]
//         (mảng rỗng nếu chưa có NE hoặc chưa có config)
//         500 nếu lỗi DB khi lấy user/mapping
// Flow  : lấy actor từ context → GetUserByUserName → GetAllCliNeOfUserByUserId →
//         với mỗi NE: GetNeByNeId → GetNeConfigByNeId → gộp thành neConfigEntry
func HandlerListNeConfig(w http.ResponseWriter, r *http.Request) {
	userMiddleware, ok := r.Context().Value(middleware.UserContextKey).(*middleware.User)
	if !ok {
		response.InternalError(w, "Internal Server Error")
		return
	}

	account, err := service.GetUserByUserName(userMiddleware.Username)
	if err != nil {
		response.InternalError(w, "Cannot get user")
		return
	}

	mappings, err := service.GetAllCliNeOfUserByUserId(account.AccountID)
	if err != nil {
		response.InternalError(w, "Cannot get NE mappings")
		return
	}

	type neConfigEntry struct {
		NeName    string                 `json:"ne_name"`
		NeIP      string                 `json:"ne_ip"`
		SiteName  string                 `json:"site_name"`
		ConfigList []*db_models.CliNeConfig `json:"config_list"`
	}

	var result []neConfigEntry
	for _, m := range mappings {
		ne, err := service.GetNeByNeId(m.TblNeID)
		if err != nil || ne == nil {
			continue
		}
		cfgList, err := service.GetNeConfigByNeId(ne.ID)
		if err != nil {
			cfgList = nil
		}
		if cfgList == nil {
			cfgList = []*db_models.CliNeConfig{}
		}
		result = append(result, neConfigEntry{
			NeName:     ne.Name,
			NeIP:       ne.IPAddress,
			SiteName:   ne.SiteName,
			ConfigList: cfgList,
		})
	}

	if result == nil {
		result = []neConfigEntry{}
	}
	response.Write(w, http.StatusOK, result)
}

// helpers shared across ne handlers

func mustUser(r *http.Request) *middleware.User {
	u, _ := r.Context().Value(middleware.UserContextKey).(*middleware.User)
	return u
}

func opHistory(cmd, detail, username string) db_models.CliOperationHistory {
	return db_models.CliOperationHistory{
		CmdName:     fmt.Sprintf("%s %s", cmd, detail),
		CreatedDate: time.Now(),
		Scope:       "cli-config",
		Account:     username,
	}
}
