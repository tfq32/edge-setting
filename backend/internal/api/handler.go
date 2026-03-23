package api

import (
	"net/http"
	"runtime"
	"time"

	"github.com/edge-setting/backend/internal/collector"
	"github.com/edge-setting/backend/internal/store"
	"github.com/edge-setting/backend/internal/vsoa"
	"github.com/edge-setting/backend/internal/ws"
	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/net"
	"go.uber.org/zap"
)

// Handler 汇聚所有依赖
type Handler struct {
	Collector *collector.Collector
	Store     *store.DB
	VSOA      *vsoa.Client
	Hub       *ws.Hub
	Log       *zap.Logger
}

// ── 健康检查 ──────────────────────────────────────────────
func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok", "ts": time.Now().UnixMilli()})
}

// ── 系统信息 ──────────────────────────────────────────────
func (h *Handler) SystemInfo(c *gin.Context) {
	info, _ := host.Info()

	var ifaces []gin.H
	if nets, err := net.Interfaces(); err == nil {
		for _, n := range nets {
			addrs := make([]string, 0)
			for _, a := range n.Addrs {
				addrs = append(addrs, a.Addr)
			}
			ifaces = append(ifaces, gin.H{
				"name":  n.Name,
				"addrs": addrs,
				"flags": n.Flags,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"hostname":         info.Hostname,
			"arch":             runtime.GOARCH,
			"os":               info.OS,
			"platform":         info.Platform,
			"platform_version": info.PlatformVersion,
			"kernel_version":   info.KernelVersion,
			"uptime_system":    info.Uptime,
			"net_interfaces":   ifaces,
		},
	})
}

// ── 实时指标快照 ──────────────────────────────────────────
func (h *Handler) MetricsSnapshot(c *gin.Context) {
	snap, err := h.Collector.Collect()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 9999, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": snap})
}

// ── 历史指标查询 ──────────────────────────────────────────
func (h *Handler) MetricsHistory(c *gin.Context) {
	metricType := c.DefaultQuery("type", "cpu")
	rangeStr   := c.DefaultQuery("range", "1h")

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

	rows, err := h.Store.QueryMetrics(metricType, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 9999, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": rows})
}

// ── WebSocket 指标推送 ────────────────────────────────────
func (h *Handler) MetricsWS(c *gin.Context) {
	if err := h.Hub.Upgrade(c.Writer, c.Request); err != nil {
		h.Log.Error("WebSocket 升级失败", zap.Error(err))
	}
}

// ── 微应用列表（只读） ────────────────────────────────────
func (h *Handler) AppList(c *gin.Context) {
	apps, err := h.VSOA.ListApps(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": 1010, "message": "MS 不可达: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": apps})
}
