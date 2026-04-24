package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/handler/response"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/logger"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/service"
	"github.com/go-chi/chi"
)

// ── request / response types ─────────────────────────────────────────────────

type configBackupSaveReq struct {
	NeName    string `json:"ne_name"`
	NeIP      string `json:"ne_ip"`
	ConfigXML string `json:"config_xml"`
}

type configBackupSaveResp struct {
	Status string `json:"status"`
	ID     int64  `json:"id"`
}

type configBackupListResp struct {
	Status  string                    `json:"status"`
	Backups []*db_models.ConfigBackup `json:"backups"`
}

type configBackupGetResp struct {
	Status    string `json:"status"`
	ID        int64  `json:"id"`
	NeName    string `json:"ne_name"`
	NeIP      string `json:"ne_ip"`
	CreatedAt string `json:"created_at"`
	Size      int64  `json:"size"`
	ConfigXML string `json:"config_xml"`
}

// ── handlers ──────────────────────────────────────────────────────────────────

// HandlerConfigBackupSave lưu một bản backup config NETCONF vào disk và DB.
//
// Input : POST body JSON { "ne_name": string (bắt buộc), "ne_ip": string,
//         "config_xml": string (XML nội dung config, bắt buộc) }
// Output: 201 { "status": "ok", "id": <int64> }
//         400 nếu thiếu ne_name hoặc config_xml
//         500 nếu lỗi ghi file hoặc DB
// Flow  : decode body → validate → SaveConfigBackup (ghi XML ra disk, insert metadata vào DB) →
//         trả id của bản ghi vừa tạo
func HandlerConfigBackupSave(w http.ResponseWriter, r *http.Request) {
	var req configBackupSaveReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Write(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if strings.TrimSpace(req.NeName) == "" {
		response.Write(w, http.StatusBadRequest, "ne_name is required")
		return
	}
	if strings.TrimSpace(req.ConfigXML) == "" {
		response.Write(w, http.StatusBadRequest, "config_xml is required")
		return
	}

	b, err := service.SaveConfigBackup(req.NeName, req.NeIP, req.ConfigXML)
	if err != nil {
		logger.Logger.Errorf("config-backup/save: %v", err)
		response.InternalError(w, "failed to save config backup")
		return
	}

	response.Write(w, http.StatusCreated, configBackupSaveResp{Status: "ok", ID: b.ID})
}

// HandlerConfigBackupList liệt kê các bản backup của một NE (hoặc tất cả NE).
//
// Input : GET query param ?ne_name=<string> (tuỳ chọn — bỏ trống để lấy tất cả)
// Output: 200 { "status": "ok", "backups": [ { id, ne_name, ne_ip, created_at, size } ] }
//         500 nếu lỗi DB
// Flow  : đọc ne_name từ query → ListConfigBackups → trả danh sách metadata
//         (không bao gồm nội dung XML)
func HandlerConfigBackupList(w http.ResponseWriter, r *http.Request) {
	neName := r.URL.Query().Get("ne_name")

	list, err := service.ListConfigBackups(neName)
	if err != nil {
		logger.Logger.Errorf("config-backup/list: %v", err)
		response.InternalError(w, "failed to list config backups")
		return
	}
	if list == nil {
		list = []*db_models.ConfigBackup{}
	}

	response.Write(w, http.StatusOK, configBackupListResp{Status: "ok", Backups: list})
}

// HandlerConfigBackupGet lấy metadata và nội dung XML của một bản backup theo ID.
//
// Input : GET path param {id} — int64
// Output: 200 { "status": "ok", "id", "ne_name", "ne_ip", "created_at", "size", "config_xml" }
//         400 nếu id không hợp lệ
//         404 nếu không tìm thấy bản ghi
//         500 nếu lỗi DB hoặc không đọc được file XML
// Flow  : parse id từ URL → GetConfigBackupById (lấy metadata từ DB + đọc XML từ disk) →
//         trả full record kèm config_xml
func HandlerConfigBackupGet(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		response.Write(w, http.StatusBadRequest, "invalid backup id")
		return
	}

	b, configXML, err := service.GetConfigBackupByID(id)
	if err != nil {
		logger.Logger.Errorf("config-backup/get id=%d: %v", id, err)
		response.InternalError(w, "failed to get config backup")
		return
	}
	if b == nil {
		response.NotFound(w, "config backup not found")
		return
	}

	response.Write(w, http.StatusOK, configBackupGetResp{
		Status:    "ok",
		ID:        b.ID,
		NeName:    b.NeName,
		NeIP:      b.NeIP,
		CreatedAt: b.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		Size:      b.Size,
		ConfigXML: configXML,
	})
}
