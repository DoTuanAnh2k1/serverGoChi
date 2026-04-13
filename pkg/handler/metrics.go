package handler

import (
	"net/http"
	"runtime"

	"github.com/DoTuanAnh2k1/serverGoChi/pkg/handler/response"
)

type metricsResp struct {
	GoRoutines  int     `json:"goroutines"`
	HeapAllocMB float64 `json:"heap_alloc_mb"`
	HeapSysMB   float64 `json:"heap_sys_mb"`
	SysMemMB    float64 `json:"sys_mem_mb"`
	NumGC       uint32  `json:"num_gc"`
	NumCPU      int     `json:"num_cpu"`
	GoMaxProcs  int     `json:"gomaxprocs"`
	GoVersion   string  `json:"go_version"`
}

// HandlerMetrics trả về các chỉ số runtime của tiến trình Go.
//
// Input : GET (không có body/query params)
// Output: 200 { goroutines, heap_alloc_mb, heap_sys_mb, sys_mem_mb,
//         num_gc, num_cpu, gomaxprocs, go_version }
// Flow  : runtime.ReadMemStats → runtime.NumGoroutine/NumCPU/GOMAXPROCS →
//         gộp vào metricsResp → trả JSON
func HandlerMetrics(w http.ResponseWriter, r *http.Request) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	resp := metricsResp{
		GoRoutines:  runtime.NumGoroutine(),
		HeapAllocMB: float64(m.HeapAlloc) / 1024 / 1024,
		HeapSysMB:   float64(m.HeapSys) / 1024 / 1024,
		SysMemMB:    float64(m.Sys) / 1024 / 1024,
		NumGC:       m.NumGC,
		NumCPU:      runtime.NumCPU(),
		GoMaxProcs:  runtime.GOMAXPROCS(0),
		GoVersion:   runtime.Version(),
	}

	response.Write(w, http.StatusOK, resp)
}
