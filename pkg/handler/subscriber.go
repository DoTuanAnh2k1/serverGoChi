package handler

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/go-chi/chi"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/handler/response"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/logger"
)

// tcpDataDir trả về thư mục lưu file từ env TCP_DATA_DIR (giống tcpserver).
func tcpDataDir() string {
	if d := os.Getenv("TCP_DATA_DIR"); d != "" {
		return d
	}
	return "."
}

// ── HandlerListSubscriberFiles ────────────────────────────────────────────────

type subscriberFileInfo struct {
	Name  string `json:"name"`
	Index int    `json:"index"`
	Size  int64  `json:"size_bytes"`
}

// HandlerListSubscriberFiles handles GET /aa/subscribers/files
// Trả về danh sách các file list_subscribers_results.* đang có trên ổ cứng,
// sắp xếp theo index tăng dần.
func HandlerListSubscriberFiles(w http.ResponseWriter, r *http.Request) {
	dir := tcpDataDir()

	pattern := filepath.Join(dir, "list_subscribers_results.*")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		logger.Logger.Errorf("subscriber/files: glob %q: %v", pattern, err)
		response.InternalError(w, "failed to list files")
		return
	}

	var files []subscriberFileInfo
	for _, path := range matches {
		base := filepath.Base(path)
		// tách index từ "list_subscribers_results.<index>"
		parts := strings.SplitN(base, ".", 2)
		if len(parts) != 2 {
			continue
		}
		idx, err := strconv.Atoi(parts[1])
		if err != nil {
			continue
		}
		info, err := os.Stat(path)
		if err != nil {
			logger.Logger.WithField("file", path).Errorf("subscriber/files: stat: %v", err)
			continue
		}
		files = append(files, subscriberFileInfo{
			Name:  base,
			Index: idx,
			Size:  info.Size(),
		})
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Index < files[j].Index
	})

	logger.Logger.WithField("count", len(files)).Debug("subscriber/files: list")
	response.Write(w, http.StatusOK, files)
}

// ── HandlerViewSubscriberFile ─────────────────────────────────────────────────

type subscriberFileContent struct {
	Name  string   `json:"name"`
	Index int      `json:"index"`
	Lines []string `json:"lines"`
	Total int      `json:"total"`
}

// HandlerViewSubscriberFile handles GET /aa/subscribers/files/{index}
// Trả về nội dung của file list_subscribers_results.<index>.
func HandlerViewSubscriberFile(w http.ResponseWriter, r *http.Request) {
	idxStr := chi.URLParam(r, "index")
	idx, err := strconv.Atoi(idxStr)
	if err != nil || idx < 0 {
		logger.Logger.Warnf("subscriber/files/view: invalid index %q", idxStr)
		response.BadRequest(w, fmt.Sprintf("invalid index: %q", idxStr))
		return
	}

	name := fmt.Sprintf("list_subscribers_results.%d", idx)
	path := filepath.Join(tcpDataDir(), name)

	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Logger.WithField("file", name).Warn("subscriber/files/view: file not found")
			response.NotFound(w, fmt.Sprintf("file %q not found", name))
			return
		}
		logger.Logger.WithField("file", name).Errorf("subscriber/files/view: open: %v", err)
		response.InternalError(w, "failed to open file")
		return
	}
	defer f.Close()

	var lines []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	if err := sc.Err(); err != nil {
		logger.Logger.WithField("file", name).Errorf("subscriber/files/view: read: %v", err)
		response.InternalError(w, "failed to read file")
		return
	}

	logger.Logger.WithField("file", name).WithField("lines", len(lines)).Debug("subscriber/files/view: served")
	response.Write(w, http.StatusOK, subscriberFileContent{
		Name:  name,
		Index: idx,
		Lines: lines,
		Total: len(lines),
	})
}
