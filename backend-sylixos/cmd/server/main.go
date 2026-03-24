package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-ser/internal/api"
	"go-ser/internal/collector"
	"go-ser/internal/config"
	"go-ser/internal/database"
	"go-ser/internal/vsoa"
	"go-ser/internal/ws"

	log "github.com/sirupsen/logrus"
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
	initLogger(config.AppConfig.Log.Level)
	log.Info("Edge Setting (SylixOS) 启动")

	// ── 数据库 ────────────────────────────────────────────
	if err := database.Start(); err != nil {
		log.WithError(err).Fatal("数据库初始化失败")
	}
	defer database.Stop()

	// ── VSOA 客户端 ───────────────────────────────────────
	vsoaClient, err := vsoa.New(config.AppConfig.VSOA.MSAddress)
	if err != nil {
		log.WithError(err).Fatal("VSOA 连接失败")
	}
	defer vsoaClient.Close()

	// ── 指标采集器 ────────────────────────────────────────
	coll := collector.New()

	// ── WebSocket Hub ─────────────────────────────────────
	hub := ws.NewHub()

	// ── HTTP Handler & Router ─────────────────────────────
	h := &api.Handler{
		Collector: coll,
		VSOA:      vsoaClient,
		Hub:       hub,
	}
	router := api.SetupRouter(h)

	// ── 定时任务 ──────────────────────────────────────────
	go runMetricsTicker(coll, hub)
	go runCleanupTicker()

	// ── HTTP Server ───────────────────────────────────────
	addr := fmt.Sprintf(":%d", config.AppConfig.Server.Port)
	srv := &http.Server{Addr: addr, Handler: router}

	go func() {
		log.WithField("addr", addr).Info("HTTP 服务启动")
		var err error
		if config.AppConfig.Server.HTTPS {
			err = srv.ListenAndServeTLS(config.AppConfig.Server.CertFile, config.AppConfig.Server.KeyFile)
		} else {
			err = srv.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			log.WithError(err).Fatal("HTTP 服务异常")
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
func runMetricsTicker(coll *collector.Collector, hub *ws.Hub) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		snap, err := coll.Collect()
		if err != nil {
			log.WithError(err).Error("指标采集失败")
			continue
		}
		// 持久化到 SQLite
		if database.EdgeDB != nil {
			database.EdgeDB.Exec(
				`INSERT INTO metrics(ts,cpu,mem_pct,disk_pct,net_in,net_out) VALUES(?,?,?,?,?,?)`,
				snap.Timestamp, snap.CPU, snap.MemPct, snap.DiskPct, snap.NetIn, snap.NetOut,
			)
		}
		// 广播给 WebSocket 客户端
		hub.Broadcast(ws.Message{Type: "metrics", Data: snap})
	}
}

// runCleanupTicker 每小时执行一次数据清理
func runCleanupTicker() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	for range ticker.C {
		if database.EdgeDB == nil {
			continue
		}
		retainDays := config.AppConfig.Data.MetricsRetain
		if retainDays <= 0 {
			retainDays = 7
		}
		cutoff := time.Now().UnixMilli() - int64(retainDays)*86400*1000
		if _, err := database.EdgeDB.Exec(`DELETE FROM metrics WHERE ts<?`, cutoff); err != nil {
			log.WithError(err).Error("数据清理失败")
		} else {
			log.Info("数据清理完成")
		}
	}
}

// initLogger 按配置初始化 logrus
func initLogger(level string) {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
	lvl, err := log.ParseLevel(level)
	if err != nil {
		lvl = log.InfoLevel
	}
	log.SetLevel(lvl)
}
