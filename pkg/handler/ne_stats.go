package handler

import (
	"net/http"
	"runtime"
	"time"

	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/handler/response"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/store"
)

var startTime = time.Now()

type neStatEntry struct {
	ID        int64  `json:"id"`
	Namespace string `json:"namespace"`
	NeType    string `json:"ne_type"`
	SiteName  string `json:"site_name"`
	MasterIP  string `json:"master_ip"`
	ConfMode  string `json:"conf_mode"`
	Commands  int    `json:"commands"`
	NAGroups  int    `json:"ne_access_groups"`
	CEGroups  int    `json:"cmd_exec_groups"`
	Users     int    `json:"users"`
}

type dashboardStatsResp struct {
	NEs      []neStatEntry      `json:"nes"`
	Metrics  appMetrics         `json:"metrics"`
	Users    userStats          `json:"users"`
	Security securitySummary    `json:"security"`
	NeTypes  map[string]int     `json:"ne_types"`
	Services map[string]int     `json:"services"`
}

type appMetrics struct {
	GoRoutines   int     `json:"goroutines"`
	HeapAllocMB  float64 `json:"heap_alloc_mb"`
	HeapSysMB    float64 `json:"heap_sys_mb"`
	SysMemMB     float64 `json:"sys_mem_mb"`
	StackMB      float64 `json:"stack_mb"`
	NumGC        uint32  `json:"num_gc"`
	NumCPU       int     `json:"num_cpu"`
	GoMaxProcs   int     `json:"gomaxprocs"`
	GoVersion    string  `json:"go_version"`
	UptimeSeconds int64  `json:"uptime_seconds"`
}

type userStats struct {
	Total        int `json:"total"`
	Enabled      int `json:"enabled"`
	Disabled     int `json:"disabled"`
	Locked       int `json:"locked"`
	SuperAdmins  int `json:"super_admins"`
	Admins       int `json:"admins"`
	RegularUsers int `json:"regular_users"`
	RecentLogins int `json:"recent_logins"`
	ExpiredPw    int `json:"expired_pw"`
}

type securitySummary struct {
	BlacklistCount int  `json:"blacklist_count"`
	WhitelistCount int  `json:"whitelist_count"`
	PwPolicySet    bool  `json:"pw_policy_set"`
	PwMinLength    int32 `json:"pw_min_length"`
	PwMaxAge       int32 `json:"pw_max_age_days"`
	PwRequireUpper bool  `json:"pw_require_upper"`
	PwRequireLower bool  `json:"pw_require_lower"`
	PwRequireDigit bool  `json:"pw_require_digit"`
	PwRequireSpec  bool  `json:"pw_require_special"`
	LockoutThresh  int32 `json:"lockout_threshold"`
	LockoutMins    int32 `json:"lockout_minutes"`
}

func HandlerDashboardStats(w http.ResponseWriter, r *http.Request) {
	s := store.GetSingleton()

	nes, err := s.ListNEs()
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}

	cmds, err := s.ListCommands(0, "")
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}

	nags, err := s.ListNeAccessGroups()
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}

	cegs, err := s.ListCmdExecGroups()
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}

	users, err := s.ListUsers()
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}

	// User stats.
	now := time.Now().UTC()
	oneDayAgo := now.Add(-24 * time.Hour)
	us := userStats{Total: len(users)}
	for _, u := range users {
		if u.IsEnabled {
			us.Enabled++
		} else {
			us.Disabled++
		}
		if u.LockedAt != nil {
			us.Locked++
		}
		switch u.Role {
		case db_models.RoleSuperAdmin:
			us.SuperAdmins++
		case db_models.RoleAdmin:
			us.Admins++
		default:
			us.RegularUsers++
		}
		if u.LastLoginAt != nil && u.LastLoginAt.After(oneDayAgo) {
			us.RecentLogins++
		}
		if u.PasswordExpiresAt != nil && now.After(*u.PasswordExpiresAt) {
			us.ExpiredPw++
		}
	}

	// Commands per NE.
	cmdCountByNE := make(map[int64]int)
	services := make(map[string]int)
	for _, c := range cmds {
		cmdCountByNE[c.NeID]++
		if c.Service != "" {
			services[c.Service]++
		}
	}

	// NE access groups per NE + users per NE.
	nagCountByNE := make(map[int64]int)
	usersPerNE := make(map[int64]map[int64]bool)
	for _, ne := range nes {
		usersPerNE[ne.ID] = make(map[int64]bool)
	}
	for _, g := range nags {
		neIDs, _ := s.ListNEsInNeAccessGroup(g.ID)
		userIDs, _ := s.ListUsersInNeAccessGroup(g.ID)
		for _, neID := range neIDs {
			nagCountByNE[neID]++
			if m, ok := usersPerNE[neID]; ok {
				for _, uid := range userIDs {
					m[uid] = true
				}
			}
		}
	}

	// Cmd exec groups per NE.
	cegCountByNE := make(map[int64]int)
	for _, g := range cegs {
		cmdIDs, _ := s.ListCommandsInCmdExecGroup(g.ID)
		seen := make(map[int64]bool)
		for _, cmdID := range cmdIDs {
			for _, c := range cmds {
				if c.ID == cmdID && !seen[c.NeID] {
					seen[c.NeID] = true
					cegCountByNE[c.NeID]++
				}
			}
		}
	}

	// NE type distribution.
	neTypes := make(map[string]int)
	entries := make([]neStatEntry, 0, len(nes))
	for _, ne := range nes {
		t := ne.NeType
		if t == "" {
			t = "unknown"
		}
		neTypes[t]++
		entries = append(entries, neStatEntry{
			ID:        ne.ID,
			Namespace: ne.Namespace,
			NeType:    ne.NeType,
			SiteName:  ne.SiteName,
			MasterIP:  ne.MasterIP,
			ConfMode:  ne.ConfMode,
			Commands:  cmdCountByNE[ne.ID],
			NAGroups:  nagCountByNE[ne.ID],
			CEGroups:  cegCountByNE[ne.ID],
			Users:     len(usersPerNE[ne.ID]),
		})
	}

	// Security / password policy.
	sec := securitySummary{}
	blEntries, _ := s.ListAccessListEntries("blacklist")
	wlEntries, _ := s.ListAccessListEntries("whitelist")
	sec.BlacklistCount = len(blEntries)
	sec.WhitelistCount = len(wlEntries)

	if pp, err := s.GetPasswordPolicy(); err == nil && pp != nil {
		sec.PwPolicySet = true
		sec.PwMinLength = pp.MinLength
		sec.PwMaxAge = pp.MaxAgeDays
		sec.PwRequireUpper = pp.RequireUppercase
		sec.PwRequireLower = pp.RequireLowercase
		sec.PwRequireDigit = pp.RequireDigit
		sec.PwRequireSpec = pp.RequireSpecial
		sec.LockoutThresh = pp.MaxLoginFailure
		sec.LockoutMins = pp.LockoutMinutes
	}

	// Runtime metrics.
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	resp := dashboardStatsResp{
		NEs: entries,
		Metrics: appMetrics{
			GoRoutines:    runtime.NumGoroutine(),
			HeapAllocMB:   float64(m.HeapAlloc) / 1024 / 1024,
			HeapSysMB:     float64(m.HeapSys) / 1024 / 1024,
			SysMemMB:      float64(m.Sys) / 1024 / 1024,
			StackMB:       float64(m.StackInuse) / 1024 / 1024,
			NumGC:         m.NumGC,
			NumCPU:        runtime.NumCPU(),
			GoMaxProcs:    runtime.GOMAXPROCS(0),
			GoVersion:     runtime.Version(),
			UptimeSeconds: int64(time.Since(startTime).Seconds()),
		},
		Users:    us,
		Security: sec,
		NeTypes:  neTypes,
		Services: services,
	}

	response.Write(w, http.StatusOK, resp)
}
