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
}

func TestQueryMetrics_TimeRange(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	now := time.Now().UnixMilli()
	_ = db.InsertMetric(store.MetricRow{Timestamp: now - 10000, CPU: 5.0})
	_ = db.InsertMetric(store.MetricRow{Timestamp: now, CPU: 50.0})

	result, _ := db.QueryMetrics("cpu", now-1000, now+1000)
	if len(result) != 1 {
		t.Errorf("时间范围过滤失败，期望 1 条，实际 %d", len(result))
	}
}

func TestCleanup(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	now := time.Now().UnixMilli()
	old := now - 8*24*3600*1000

	_ = db.InsertMetric(store.MetricRow{Timestamp: old, CPU: 1.0})
	_ = db.InsertMetric(store.MetricRow{Timestamp: now, CPU: 2.0})

	if err := db.Cleanup(7); err != nil {
		t.Fatalf("Cleanup 失败: %v", err)
	}
	metrics, _ := db.QueryMetrics("cpu", old-1000, now+1000)
	if len(metrics) != 1 {
		t.Errorf("清理后应剩 1 条指标，实际 %d", len(metrics))
	}
}
