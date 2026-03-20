package store_test

import (
	"os"
	"testing"
	"time"

	"github.com/edge-setting/backend/internal/store"
)

func newTestDB(t *testing.T) (*store.DB, func()) {
	dir := t.TempDir()
	db, err := store.Open(dir)
	if err != nil {
		t.Fatalf("打开测试数据库失败: %v", err)
	}
	return db, func() {
		db.Close()
		os.RemoveAll(dir)
	}
}

// ── 指标存储 ──────────────────────────────────────────────
func TestInsertAndQueryMetric(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	now := time.Now().UnixMilli()
	rows := []store.MetricRow{
		{Timestamp: now - 2000, CPU: 10.0, MemPct: 20.0, DiskPct: 30.0, NetIn: 1000, NetOut: 500},
		{Timestamp: now - 1000, CPU: 20.0, MemPct: 40.0, DiskPct: 30.0, NetIn: 2000, NetOut: 600},
		{Timestamp: now,        CPU: 30.0, MemPct: 60.0, DiskPct: 30.0, NetIn: 3000, NetOut: 700},
	}
	for _, r := range rows {
		if err := db.InsertMetric(r); err != nil {
			t.Fatalf("InsertMetric 失败: %v", err)
		}
	}

	result, err := db.QueryMetrics("cpu", now-5000, now+1000)
	if err != nil {
		t.Fatalf("QueryMetrics 失败: %v", err)
	}
	if len(result) != 3 {
		t.Errorf("期望 3 条记录，实际 %d", len(result))
	}
	if result[0].CPU != 10.0 {
		t.Errorf("第一条 CPU 应为 10.0，实际 %f", result[0].CPU)
	}
}

func TestQueryMetrics_TimeRange(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	now := time.Now().UnixMilli()
	_ = db.InsertMetric(store.MetricRow{Timestamp: now - 10000, CPU: 5.0})
	_ = db.InsertMetric(store.MetricRow{Timestamp: now, CPU: 50.0})

	result, err := db.QueryMetrics("cpu", now-1000, now+1000)
	if err != nil {
		t.Fatalf("QueryMetrics 失败: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("时间范围过滤失败，期望 1 条，实际 %d", len(result))
	}
}

// ── 日志存储 ──────────────────────────────────────────────
func TestInsertAndQueryLog(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	now := time.Now().UnixMilli()
	logs := []store.LogRow{
		{AppID: "app-a", Level: "INFO",  Content: "started",      Timestamp: now - 2000},
		{AppID: "app-a", Level: "WARN",  Content: "high cpu",     Timestamp: now - 1000},
		{AppID: "app-a", Level: "ERROR", Content: "conn refused",  Timestamp: now},
		{AppID: "app-b", Level: "INFO",  Content: "app-b log",    Timestamp: now},
	}
	for _, l := range logs {
		if err := db.InsertLog(l); err != nil {
			t.Fatalf("InsertLog 失败: %v", err)
		}
	}

	// 全量查询 app-a
	rows, err := db.QueryLogs("app-a", "", "", now-5000, now+1000, 100)
	if err != nil {
		t.Fatalf("QueryLogs 失败: %v", err)
	}
	if len(rows) != 3 {
		t.Errorf("期望 3 条，实际 %d", len(rows))
	}

	// 按级别过滤
	errRows, _ := db.QueryLogs("app-a", "", "ERROR", now-5000, now+1000, 100)
	if len(errRows) != 1 {
		t.Errorf("ERROR 过滤期望 1 条，实际 %d", len(errRows))
	}

	// 关键字搜索
	kwRows, _ := db.QueryLogs("app-a", "conn", "", now-5000, now+1000, 100)
	if len(kwRows) != 1 {
		t.Errorf("关键字过滤期望 1 条，实际 %d", len(kwRows))
	}

	// app 隔离
	bRows, _ := db.QueryLogs("app-b", "", "", now-5000, now+1000, 100)
	if len(bRows) != 1 {
		t.Errorf("app-b 应有 1 条日志，实际 %d", len(bRows))
	}
}

// ── 审计日志 ──────────────────────────────────────────────
func TestInsertAndQueryAudit(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	now := time.Now().UnixMilli()
	audits := []store.AuditRow{
		{Token: "tok1", AppID: "app-a", Action: "start",   Result: "success", Timestamp: now - 2000},
		{Token: "tok1", AppID: "app-a", Action: "restart", Result: "success", Timestamp: now - 1000},
		{Token: "tok2", AppID: "app-b", Action: "stop",    Result: "failed",  Timestamp: now},
	}
	for _, a := range audits {
		if err := db.InsertAudit(a); err != nil {
			t.Fatalf("InsertAudit 失败: %v", err)
		}
	}

	rows, err := db.QueryAudit("", now-5000, now+1000, 100)
	if err != nil {
		t.Fatalf("QueryAudit 失败: %v", err)
	}
	if len(rows) != 3 {
		t.Errorf("期望 3 条，实际 %d", len(rows))
	}

	// 按操作类型过滤
	startRows, _ := db.QueryAudit("start", now-5000, now+1000, 100)
	if len(startRows) != 1 {
		t.Errorf("start 操作期望 1 条，实际 %d", len(startRows))
	}
}

// ── 数据清理 ──────────────────────────────────────────────
func TestCleanup(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	now := time.Now().UnixMilli()
	old := now - 8*24*3600*1000 // 8天前（超过默认7天保留期）

	_ = db.InsertMetric(store.MetricRow{Timestamp: old,  CPU: 1.0})
	_ = db.InsertMetric(store.MetricRow{Timestamp: now,  CPU: 2.0})
	_ = db.InsertLog(store.LogRow{AppID: "a", Level: "INFO", Content: "old", Timestamp: old})
	_ = db.InsertLog(store.LogRow{AppID: "a", Level: "INFO", Content: "new", Timestamp: now})

	if err := db.Cleanup(7, 7, 30); err != nil {
		t.Fatalf("Cleanup 失败: %v", err)
	}

	metrics, _ := db.QueryMetrics("cpu", old-1000, now+1000)
	if len(metrics) != 1 {
		t.Errorf("清理后应剩 1 条指标，实际 %d", len(metrics))
	}

	logs, _ := db.QueryLogs("a", "", "", old-1000, now+1000, 100)
	if len(logs) != 1 {
		t.Errorf("清理后应剩 1 条日志，实际 %d", len(logs))
	}
}
