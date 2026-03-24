package api

import (
	"encoding/json"
	"math"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"go-ser/internal/collector"
	"go-ser/internal/database"
	"go-ser/internal/vsoa"
	"go-ser/internal/ws"

	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
)

// Handler 汇聚所有依赖
type Handler struct {
	Collector *collector.Collector
	VSOA      *vsoa.Client
	Hub       *ws.Hub
}

// JSON 响应辅助
func jsonOK(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"code": 0, "data": data})
}

func jsonErr(w http.ResponseWriter, status int, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{"code": code, "message": message})
}

// ── 健康检查 ──────────────────────────────────────────────
func (h *Handler) Health(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "ok", "ts": time.Now().UnixMilli()})
}

// ── 系统信息 ──────────────────────────────────────────────
func (h *Handler) SystemInfo(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	hostname, _ := os.Hostname()

	kernelVersion := ""
	if data, err := os.ReadFile("/proc/version"); err == nil {
		kernelVersion = strings.TrimSpace(string(data))
		if len(kernelVersion) > 200 {
			kernelVersion = kernelVersion[:200]
		}
	}

	var uptime uint64
	if data, err := os.ReadFile("/proc/uptime"); err == nil {
		fields := strings.Fields(string(data))
		if len(fields) >= 1 {
			val, err := strconv.ParseFloat(fields[0], 64)
			if err == nil {
				uptime = uint64(math.Floor(val))
			}
		}
	}

	type ifaceInfo struct {
		Name  string   `json:"name"`
		Addrs []string `json:"addrs"`
		Flags []string `json:"flags"`
	}
	var ifaces []ifaceInfo

	jsonOK(w, map[string]interface{}{
		"hostname":         hostname,
		"arch":             runtime.GOARCH,
		"os":               runtime.GOOS,
		"platform":         "SylixOS",
		"platform_version": "",
		"kernel_version":   kernelVersion,
		"uptime_system":    uptime,
		"net_interfaces":   ifaces,
	})
}

// ── 实时指标快照 ──────────────────────────────────────────
func (h *Handler) MetricsSnapshot(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	snap, err := h.Collector.Collect()
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, 9999, err.Error())
		return
	}
	jsonOK(w, snap)
}

// ── 历史指标查询 ──────────────────────────────────────────
func (h *Handler) MetricsHistory(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	rangeStr := r.URL.Query().Get("range")
	if rangeStr == "" {
		rangeStr = "1h"
	}

	to := time.Now().UnixMilli()
	var from int64
	switch rangeStr {
	case "6h":
		from = to - 6*3600*1000
	case "24h":
		from = to - 24*3600*1000
	case "7d":
		from = to - 7*24*3600*1000
	default:
		from = to - 3600*1000
	}

	if database.EdgeDB == nil {
		jsonErr(w, http.StatusInternalServerError, 9999, "数据库未初始化")
		return
	}

	rows, err := database.EdgeDB.QueryMetrics(from, to)
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, 9999, err.Error())
		return
	}
	jsonOK(w, rows)
}

// ── WebSocket 指标推送 ────────────────────────────────────
func (h *Handler) MetricsWS(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if err := h.Hub.Upgrade(w, r); err != nil {
		log.WithError(err).Error("WebSocket 升级失败")
	}
}

// ── 微应用列表（只读） ────────────────────────────────────
func (h *Handler) AppList(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	apps, err := h.VSOA.ListApps(r.Context())
	if err != nil {
		jsonErr(w, http.StatusServiceUnavailable, 1010, "MS 不可达: "+err.Error())
		return
	}
	jsonOK(w, apps)
}
