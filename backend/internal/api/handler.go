package api

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"strconv"
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
	Collector  *collector.Collector
	Store      *store.DB
	VSOA       *vsoa.Client
	Hub        *ws.Hub
	Log        *zap.Logger
	BuildVer   string
	BuildFE    string
	StartTime  time.Time
}

// ── 健康检查 ──────────────────────────────────────────────
func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok", "ts": time.Now().UnixMilli()})
}

// ── 系统信息 ──────────────────────────────────────────────
func (h *Handler) SystemInfo(c *gin.Context) {
	info, _ := host.Info()
	sysInfo, _ := h.VSOA.GetSystemInfo(c.Request.Context())

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

	platform := ""
	if sysInfo != nil {
		platform = sysInfo.PlatformVersion
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"os":               info.OS,
			"platform":         info.Platform,
			"platform_version": info.PlatformVersion,
			"kernel_version":   info.KernelVersion,
			"arch":             runtime.GOARCH,
			"hostname":         info.Hostname,
			"platform_ver":     platform,
			"backend_ver":      h.BuildVer,
			"frontend_ver":     h.BuildFE,
			"uptime_backend":   int64(time.Since(h.StartTime).Seconds()),
			"uptime_system":    info.Uptime,
			"net_interfaces":   ifaces,
		},
	})
}

// ── 版本信息 ──────────────────────────────────────────────
func (h *Handler) Version(c *gin.Context) {
	sysInfo, _ := h.VSOA.GetSystemInfo(c.Request.Context())
	platform := ""
	if sysInfo != nil {
		platform = sysInfo.PlatformVersion
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"backend":  h.BuildVer,
			"frontend": h.BuildFE,
			"platform": platform,
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
	rangeStr := c.DefaultQuery("range", "1h")

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

// ── 微应用列表 ────────────────────────────────────────────
func (h *Handler) AppList(c *gin.Context) {
	apps, err := h.VSOA.ListApps(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": 1010, "message": "MS 不可达: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": apps})
}

// ── 微应用详情 ────────────────────────────────────────────
func (h *Handler) AppDetail(c *gin.Context) {
	id := c.Param("id")
	app, err := h.VSOA.GetApp(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1004, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": app})
}

// ── 微应用启动 ────────────────────────────────────────────
func (h *Handler) AppStart(c *gin.Context) {
	id := c.Param("id")
	token := c.GetHeader("X-Operator")
	err := h.VSOA.Start(context.Background(), id)
	result := "success"
	if err != nil {
		result = "failed: " + err.Error()
	}
	_ = h.Store.InsertAudit(store.AuditRow{
		Token: token, AppID: id, Action: "start", Result: result,
		Timestamp: time.Now().UnixMilli(),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 9999, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "启动成功"})
}

// ── 微应用停止 ────────────────────────────────────────────
func (h *Handler) AppStop(c *gin.Context) {
	id := c.Param("id")
	token := c.GetHeader("X-Operator")
	err := h.VSOA.Stop(context.Background(), id)
	result := "success"
	if err != nil {
		result = "failed: " + err.Error()
	}
	_ = h.Store.InsertAudit(store.AuditRow{
		Token: token, AppID: id, Action: "stop", Result: result,
		Timestamp: time.Now().UnixMilli(),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 9999, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "停止成功"})
}

// ── 微应用重启 ────────────────────────────────────────────
func (h *Handler) AppRestart(c *gin.Context) {
	id := c.Param("id")
	token := c.GetHeader("X-Operator")
	err := h.VSOA.Restart(context.Background(), id)
	result := "success"
	if err != nil {
		result = "failed: " + err.Error()
	}
	_ = h.Store.InsertAudit(store.AuditRow{
		Token: token, AppID: id, Action: "restart", Result: result,
		Timestamp: time.Now().UnixMilli(),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 9999, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "重启成功"})
}

// ── 实时日志 WebSocket ────────────────────────────────────
func (h *Handler) AppLogsWS(c *gin.Context) {
	id := c.Param("id")
	if err := h.Hub.Upgrade(c.Writer, c.Request); err != nil {
		h.Log.Error("日志 WebSocket 升级失败", zap.String("app", id), zap.Error(err))
	}
}

// ── 历史日志查询 ──────────────────────────────────────────
func (h *Handler) AppLogsHistory(c *gin.Context) {
	id := c.Param("id")
	keyword := c.Query("keyword")
	level := c.Query("level")
	limitStr := c.DefaultQuery("limit", "200")
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 || limit > 1000 {
		limit = 200
	}

	to := time.Now().UnixMilli()
	fromStr := c.DefaultQuery("from", fmt.Sprintf("%d", to-3600*1000))
	from, _ := strconv.ParseInt(fromStr, 10, 64)

	toStr := c.Query("to")
	if toStr != "" {
		if v, err := strconv.ParseInt(toStr, 10, 64); err == nil {
			to = v
		}
	}

	rows, err := h.Store.QueryLogs(id, keyword, level, from, to, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 9999, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": rows})
}

// ── 日志导出 ──────────────────────────────────────────────
func (h *Handler) AppLogsExport(c *gin.Context) {
	id := c.Param("id")
	to := time.Now().UnixMilli()
	from := to - 24*3600*1000

	rows, err := h.Store.QueryLogs(id, "", "", from, to, 5000)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 9999, "message": err.Error()})
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s-%d.log"`, id, time.Now().Unix()))
	c.Header("Content-Type", "text/plain; charset=utf-8")
	for _, row := range rows {
		t := time.UnixMilli(row.Timestamp).Format("2006-01-02 15:04:05")
		fmt.Fprintf(c.Writer, "[%s] [%s] %s\n", t, row.Level, row.Content)
	}
}

// ── 审计日志查询 ──────────────────────────────────────────
func (h *Handler) AuditQuery(c *gin.Context) {
	op := c.Query("op")
	limitStr := c.DefaultQuery("limit", "100")
	limit, _ := strconv.Atoi(limitStr)
	to := time.Now().UnixMilli()
	from := to - 30*24*3600*1000

	rows, err := h.Store.QueryAudit(op, from, to, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 9999, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": rows})
}
