package collector

import (
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

// Snapshot 单次采集快照
type Snapshot struct {
	Timestamp int64   `json:"ts"`
	CPU       float64 `json:"cpu"`       // 总体占用率 %
	CPUCores  []float64 `json:"cpu_cores"` // 各核心
	MemUsed   uint64  `json:"mem_used"`  // 字节
	MemTotal  uint64  `json:"mem_total"`
	MemPct    float64 `json:"mem_pct"`
	SwapUsed  uint64  `json:"swap_used"`
	SwapTotal uint64  `json:"swap_total"`
	DiskPct   float64 `json:"disk_pct"`  // 根分区使用率
	DiskRead  uint64  `json:"disk_read"` // bytes/s
	DiskWrite uint64  `json:"disk_write"`
	NetIn     uint64  `json:"net_in"`   // bytes/s
	NetOut    uint64  `json:"net_out"`
	Load1     float64 `json:"load1"`
	Load5     float64 `json:"load5"`
	Load15    float64 `json:"load15"`
	Uptime    uint64  `json:"uptime"` // 秒
}

// Collector 指标采集器
type Collector struct {
	prevDiskIO map[string]disk.IOCountersStat
	prevNetIO  map[string]net.IOCountersStat
	prevTime   time.Time
}

// New 创建采集器
func New() *Collector {
	return &Collector{
		prevDiskIO: make(map[string]disk.IOCountersStat),
		prevNetIO:  make(map[string]net.IOCountersStat),
		prevTime:   time.Now(),
	}
}

// Collect 执行一次采集，返回快照
func (c *Collector) Collect() (*Snapshot, error) {
	now := time.Now()
	elapsed := now.Sub(c.prevTime).Seconds()
	if elapsed < 0.1 {
		elapsed = 1.0
	}
	s := &Snapshot{Timestamp: now.UnixMilli()}

	// CPU
	if pcts, err := cpu.Percent(0, false); err == nil && len(pcts) > 0 {
		s.CPU = round2(pcts[0])
	}
	if cores, err := cpu.Percent(0, true); err == nil {
		for _, v := range cores {
			s.CPUCores = append(s.CPUCores, round2(v))
		}
	}

	// 内存
	if vm, err := mem.VirtualMemory(); err == nil {
		s.MemUsed = vm.Used
		s.MemTotal = vm.Total
		s.MemPct = round2(vm.UsedPercent)
	}
	if sw, err := mem.SwapMemory(); err == nil {
		s.SwapUsed = sw.Used
		s.SwapTotal = sw.Total
	}

	// 磁盘使用率（根分区）
	if du, err := disk.Usage("/"); err == nil {
		s.DiskPct = round2(du.UsedPercent)
	}

	// 磁盘 I/O 速率
	if ioMap, err := disk.IOCounters(); err == nil {
		var totalRead, totalWrite uint64
		for name, cur := range ioMap {
			if prev, ok := c.prevDiskIO[name]; ok {
				if cur.ReadBytes >= prev.ReadBytes {
					totalRead += uint64(float64(cur.ReadBytes-prev.ReadBytes) / elapsed)
				}
				if cur.WriteBytes >= prev.WriteBytes {
					totalWrite += uint64(float64(cur.WriteBytes-prev.WriteBytes) / elapsed)
				}
			}
		}
		s.DiskRead = totalRead
		s.DiskWrite = totalWrite
		c.prevDiskIO = ioMap
	}

	// 网络 I/O 速率
	if netList, err := net.IOCounters(false); err == nil && len(netList) > 0 {
		cur := netList[0]
		if prev, ok := c.prevNetIO["all"]; ok {
			if cur.BytesRecv >= prev.BytesRecv {
				s.NetIn = uint64(float64(cur.BytesRecv-prev.BytesRecv) / elapsed)
			}
			if cur.BytesSent >= prev.BytesSent {
				s.NetOut = uint64(float64(cur.BytesSent-prev.BytesSent) / elapsed)
			}
		}
		c.prevNetIO["all"] = cur
	}

	// 系统负载
	if avg, err := load.Avg(); err == nil {
		s.Load1 = round2(avg.Load1)
		s.Load5 = round2(avg.Load5)
		s.Load15 = round2(avg.Load15)
	}

	// Uptime
	if info, err := host.Info(); err == nil {
		s.Uptime = info.Uptime
	}

	c.prevTime = now
	return s, nil
}

func round2(f float64) float64 {
	return float64(int(f*100)) / 100
}
