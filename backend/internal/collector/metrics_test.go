package collector_test

import (
	"testing"

	"github.com/edge-setting/backend/internal/collector"
)

func TestCollect_ReturnsSnapshot(t *testing.T) {
	c := collector.New()
	snap, err := c.Collect()
	if err != nil {
		t.Fatalf("Collect 失败: %v", err)
	}
	if snap == nil {
		t.Fatal("Snapshot 不应为 nil")
	}
	if snap.Timestamp == 0 {
		t.Error("Timestamp 应非零")
	}
}

func TestCollect_CPURange(t *testing.T) {
	c := collector.New()
	snap, _ := c.Collect()
	if snap.CPU < 0 || snap.CPU > 100 {
		t.Errorf("CPU 应在 [0,100]，实际 %f", snap.CPU)
	}
}

func TestCollect_MemoryConsistency(t *testing.T) {
	c := collector.New()
	snap, _ := c.Collect()
	if snap.MemTotal == 0 {
		t.Error("MemTotal 应大于 0")
	}
	if snap.MemUsed > snap.MemTotal {
		t.Errorf("MemUsed(%d) 不能超过 MemTotal(%d)", snap.MemUsed, snap.MemTotal)
	}
	if snap.MemPct < 0 || snap.MemPct > 100 {
		t.Errorf("MemPct 应在 [0,100]，实际 %f", snap.MemPct)
	}
}

func TestCollect_DiskRange(t *testing.T) {
	c := collector.New()
	snap, _ := c.Collect()
	if snap.DiskPct < 0 || snap.DiskPct > 100 {
		t.Errorf("DiskPct 应在 [0,100]，实际 %f", snap.DiskPct)
	}
}

func TestCollect_LoadAvg(t *testing.T) {
	c := collector.New()
	snap, _ := c.Collect()
	if snap.Load1 < 0 {
		t.Errorf("Load1 应 >= 0，实际 %f", snap.Load1)
	}
}

func TestCollect_Uptime(t *testing.T) {
	c := collector.New()
	snap, _ := c.Collect()
	if snap.Uptime == 0 {
		t.Error("Uptime 应大于 0")
	}
}

func TestCollect_MultipleCalls(t *testing.T) {
	c := collector.New()
	s1, _ := c.Collect()
	s2, _ := c.Collect()
	if s2.Timestamp < s1.Timestamp {
		t.Error("第二次采集时间戳不应早于第一次")
	}
}
