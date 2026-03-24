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
	initLogger(config.Global.Log.Level)
	log.Info("Edge Setting (SylixOS) 启动")

	// ── 数据库 ────────────────────────────────────────────
	if err := database.Start(); err != nil {
		log.WithError(err).Fatal("数据库初始化失败")
	}
	defer database.Stop()

	// ── VSOA 客户端 ───────────────────────────────────────
	vsoaClient, err := vsoa.New(config.Global.VSOA.MSAddress)
	if err != nil {
		log.WithError(err).Fatal("VSOA 连接失败")
	}
	defer vsoaClient.Close()

	// ── WebSocket Hub ─────────────────────────────────────
	hub := ws.NewHub()

	// ── HTTP Handler & Router ─────────────────────────────
	h := &api.Handler{
		VSOA: vsoaClient,
		Hub:  hub,
	}
	router := api.SetupRouter(h)

	// ── 定时任务：每 5 秒通过 VSOA 从 MS 获取指标 ────────
	go runMetricsTicker(vsoaClient, hub)
	go runCleanupTicker()

	// ── HTTP Server ───────────────────────────────────────
	addr := fmt.Sprintf(":%d", config.Global.Server.Port)
	srv := &http.Server{Addr: addr, Handler: router}

	go func() {
		log.WithField("addr", addr).Info("HTTP 服务启动")
		var err error
		if config.Global.Server.HTTPS {
			err = srv.ListenAndServeTLS(config.Global.Server.CertFile, config.Global.Server.KeyFile)
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

// runMetricsTicker 每 5 秒通过 VSOA 向 MS 请求指标，存储并广播
func runMetricsTicker(vsoaClient *vsoa.Client, hub *ws.Hub) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		snap, err := vsoaClient.GetMetrics(context.Background())
		if err != nil {
			log.WithError(err).Error("VSOA 指标获取失败")
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
		retainDays := config.Global.Data.MetricsRetain
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
