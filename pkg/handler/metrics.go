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

// HandlerMetrics returns runtime metrics (memory, goroutines, CPU).
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
