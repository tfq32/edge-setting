package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/edge-setting/backend/internal/api"
	"github.com/edge-setting/backend/internal/collector"
	"github.com/edge-setting/backend/internal/config"
	"github.com/edge-setting/backend/internal/store"
	"github.com/edge-setting/backend/internal/vsoa"
	"github.com/edge-setting/backend/internal/ws"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// 编译时注入
var (
	BuildVersion = "dev"
	BuildFE      = "dev"
)

func main() {
	// ── 配置加载 ─────────────────────────────────────────
	cfgPath := "configs/config.yaml"
	if v := os.Getenv("CONFIG_PATH"); v != "" {
		cfgPath = v
	}
	if err := config.Load(cfgPath); err != nil {
		fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// ── 日志初始化 ────────────────────────────────────────
	log := newLogger(config.Global.Log.Level)
	defer log.Sync()
	log.Info("Edge Setting 启动", zap.String("version", BuildVersion))

	// ── 数据库 ────────────────────────────────────────────
	db, err := store.Open(config.Global.Data.Dir)
	if err != nil {
		log.Fatal("数据库初始化失败", zap.Error(err))
	}
	defer db.Close()

	// ── VSOA 客户端 ───────────────────────────────────────
	vsoaClient, err := vsoa.New(config.Global.VSOA.MSAddress)
	if err != nil {
		log.Fatal("VSOA 连接失败", zap.Error(err))
	}
	defer vsoaClient.Close()

	// ── 指标采集器 ────────────────────────────────────────
	coll := collector.New()

	// ── WebSocket Hub ─────────────────────────────────────
	hub := ws.NewHub(log)

	// ── HTTP Handler & Router ─────────────────────────────
	h := &api.Handler{
		Collector: coll,
		Store:     db,
		VSOA:      vsoaClient,
		Hub:       hub,
		Log:       log,
		BuildVer:  BuildVersion,
		BuildFE:   BuildFE,
		StartTime: time.Now(),
	}
	router := api.SetupRouter(h)

	// ── 定时任务 ──────────────────────────────────────────
	go runMetricsTicker(coll, db, hub, log)
	go runCleanupTicker(db, log)

	// ── HTTP Server ───────────────────────────────────────
	addr := fmt.Sprintf(":%d", config.Global.Server.Port)
	srv := &http.Server{Addr: addr, Handler: router}

	go func() {
		log.Info("HTTP 服务启动", zap.String("addr", addr))
		var err error
		if config.Global.Server.HTTPS {
			err = srv.ListenAndServeTLS(config.Global.Server.CertFile, config.Global.Server.KeyFile)
		} else {
			err = srv.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			log.Fatal("HTTP 服务异常", zap.Error(err))
		}
	}()

	// ── 优雅退出 ──────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("正在关闭服务...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	log.Info("服务已退出")
}

// runMetricsTicker 每 3 秒采集一次指标并广播
func runMetricsTicker(coll *collector.Collector, db *store.DB, hub *ws.Hub, log *zap.Logger) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		snap, err := coll.Collect()
		if err != nil {
			log.Error("指标采集失败", zap.Error(err))
			continue
		}
		// 持久化到 SQLite
		_ = db.InsertMetric(store.MetricRow{
			Timestamp: snap.Timestamp,
			CPU:       snap.CPU,
			MemPct:    snap.MemPct,
			DiskPct:   snap.DiskPct,
			NetIn:     snap.NetIn,
			NetOut:    snap.NetOut,
		})
		// 广播给 WebSocket 客户端
		hub.Broadcast(ws.Message{Type: "metrics", Data: snap})
	}
}

// runCleanupTicker 每小时执行一次数据清理
func runCleanupTicker(db *store.DB, log *zap.Logger) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	for range ticker.C {
		if err := db.Cleanup(
			config.Global.Data.MetricsRetain,
			config.Global.Data.LogsRetain,
			config.Global.Data.AuditRetain,
		); err != nil {
			log.Error("数据清理失败", zap.Error(err))
		} else {
			log.Info("数据清理完成")
		}
	}
}

// newLogger 按配置创建 zap logger
func newLogger(level string) *zap.Logger {
	lvl := zapcore.InfoLevel
	_ = lvl.UnmarshalText([]byte(level))
	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(lvl)
	cfg.DisableStacktrace = true
	log, _ := cfg.Build()
	return log
}
