package collector

import (
	"bufio"
	"math"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// Snapshot 单次采集快照
type Snapshot struct {
	Timestamp int64     `json:"ts"`
	CPU       float64   `json:"cpu"`
	CPUCores  []float64 `json:"cpu_cores"`
	MemUsed   uint64    `json:"mem_used"`
	MemTotal  uint64    `json:"mem_total"`
	MemPct    float64   `json:"mem_pct"`
	SwapUsed  uint64    `json:"swap_used"`
	SwapTotal uint64    `json:"swap_total"`
	DiskPct   float64   `json:"disk_pct"`
	DiskRead  uint64    `json:"disk_read"`
	DiskWrite uint64    `json:"disk_write"`
	NetIn     uint64    `json:"net_in"`
	NetOut    uint64    `json:"net_out"`
	Load1     float64   `json:"load1"`
	Load5     float64   `json:"load5"`
	Load15    float64   `json:"load15"`
	Uptime    uint64    `json:"uptime"`
}

// cpuTime CPU 时间数据
type cpuTime struct {
	User   uint64
	Nice   uint64
	System uint64
	Idle   uint64
	Total  uint64
	Used   uint64
}

// netIO 网络 IO 数据
type netIO struct {
	BytesRecv uint64
	BytesSent uint64
}

// diskIO 磁盘 IO 数据
type diskIO struct {
	ReadBytes  uint64
	WriteBytes uint64
}

// Collector 指标采集器
// 通过读取 /proc 虚拟文件系统采集指标，兼容 SylixOS 和 Linux
type Collector struct {
	prevCPU     []cpuTime
	prevCPUAll  cpuTime
	prevNetIO   netIO
	prevDiskIO  diskIO
	prevTime    time.Time
	initialized bool
}

// New 创建采集器
func New() *Collector {
	return &Collector{
		prevTime: time.Now(),
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

	c.collectCPU(s)
	c.collectMemory(s)
	c.collectDiskUsage(s)
	c.collectDiskIO(s, elapsed)
	c.collectNetIO(s, elapsed)
	c.collectLoadAvg(s)
	c.collectUptime(s)

	c.prevTime = now
	c.initialized = true
	return s, nil
}

// ── CPU 采集（/proc/stat）──────────────────────────────────

func (c *Collector) collectCPU(s *Snapshot) {
	f, err := os.Open("/proc/stat")
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var allCPU cpuTime
	var perCore []cpuTime

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "cpu ") {
			allCPU = parseCPULine(line)
		} else if strings.HasPrefix(line, "cpu") {
			perCore = append(perCore, parseCPULine(line))
		}
	}

	if c.initialized && c.prevCPUAll.Total > 0 {
		totalDelta := allCPU.Total - c.prevCPUAll.Total
		usedDelta := allCPU.Used - c.prevCPUAll.Used
		if totalDelta > 0 {
			s.CPU = round2(float64(usedDelta) / float64(totalDelta) * 100)
		}
	}
	c.prevCPUAll = allCPU

	if c.initialized && len(c.prevCPU) == len(perCore) {
		for i, cur := range perCore {
			prev := c.prevCPU[i]
			totalDelta := cur.Total - prev.Total
			usedDelta := cur.Used - prev.Used
			if totalDelta > 0 {
				s.CPUCores = append(s.CPUCores, round2(float64(usedDelta)/float64(totalDelta)*100))
			} else {
				s.CPUCores = append(s.CPUCores, 0)
			}
		}
	} else {
		for range perCore {
			s.CPUCores = append(s.CPUCores, 0)
		}
	}
	c.prevCPU = perCore
}

func parseCPULine(line string) cpuTime {
	fields := strings.Fields(line)
	var ct cpuTime
	if len(fields) >= 5 {
		ct.User, _ = strconv.ParseUint(fields[1], 10, 64)
		ct.Nice, _ = strconv.ParseUint(fields[2], 10, 64)
		ct.System, _ = strconv.ParseUint(fields[3], 10, 64)
		ct.Idle, _ = strconv.ParseUint(fields[4], 10, 64)
	}
	var total uint64
	for _, f := range fields[1:] {
		v, _ := strconv.ParseUint(f, 10, 64)
		total += v
	}
	ct.Total = total
	ct.Used = total - ct.Idle
	if len(fields) >= 6 {
		iowait, _ := strconv.ParseUint(fields[5], 10, 64)
		ct.Used -= iowait
	}
	return ct
}

// ── 内存采集（/proc/meminfo）──────────────────────────────

func (c *Collector) collectMemory(s *Snapshot) {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		// fallback: 用 runtime
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		s.MemUsed = m.Alloc
		s.MemTotal = m.Sys
		if s.MemTotal > 0 {
			s.MemPct = round2(float64(s.MemUsed) / float64(s.MemTotal) * 100)
		}
		return
	}
	defer f.Close()

	info := make(map[string]uint64)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		valStr := strings.TrimSpace(parts[1])
		valStr = strings.TrimSuffix(valStr, " kB")
		valStr = strings.TrimSpace(valStr)
		v, _ := strconv.ParseUint(valStr, 10, 64)
		info[key] = v * 1024
	}

	s.MemTotal = info["MemTotal"]
	memFree := info["MemFree"]
	buffers := info["Buffers"]
	cached := info["Cached"]
	s.MemUsed = s.MemTotal - memFree - buffers - cached
	if s.MemTotal > 0 {
		s.MemPct = round2(float64(s.MemUsed) / float64(s.MemTotal) * 100)
	}

	s.SwapTotal = info["SwapTotal"]
	s.SwapUsed = s.SwapTotal - info["SwapFree"]
}

// ── 磁盘使用率 ───────────────────────────────────────────
// 实现在平台文件 disk_unix.go / disk_windows.go 中

func (c *Collector) collectDiskUsage(s *Snapshot) {
	s.DiskPct = getDiskUsagePct("/")
}

// ── 磁盘 IO（/proc/diskstats）─────────────────────────────

func (c *Collector) collectDiskIO(s *Snapshot, elapsed float64) {
	f, err := os.Open("/proc/diskstats")
	if err != nil {
		return
	}
	defer f.Close()

	var curIO diskIO
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 14 {
			continue
		}
		name := fields[2]
		if strings.HasPrefix(name, "loop") || strings.HasPrefix(name, "ram") {
			continue
		}
		sectorsRead, _ := strconv.ParseUint(fields[5], 10, 64)
		sectorsWritten, _ := strconv.ParseUint(fields[9], 10, 64)
		curIO.ReadBytes += sectorsRead * 512
		curIO.WriteBytes += sectorsWritten * 512
	}

	if c.initialized && elapsed > 0 {
		if curIO.ReadBytes >= c.prevDiskIO.ReadBytes {
			s.DiskRead = uint64(float64(curIO.ReadBytes-c.prevDiskIO.ReadBytes) / elapsed)
		}
		if curIO.WriteBytes >= c.prevDiskIO.WriteBytes {
			s.DiskWrite = uint64(float64(curIO.WriteBytes-c.prevDiskIO.WriteBytes) / elapsed)
		}
	}
	c.prevDiskIO = curIO
}

// ── 网络 IO（/proc/net/dev）───────────────────────────────

func (c *Collector) collectNetIO(s *Snapshot, elapsed float64) {
	f, err := os.Open("/proc/net/dev")
	if err != nil {
		return
	}
	defer f.Close()

	var curIO netIO
	scanner := bufio.NewScanner(f)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		if lineNo <= 2 {
			continue
		}
		line := scanner.Text()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		iface := strings.TrimSpace(parts[0])
		if iface == "lo" {
			continue
		}
		fields := strings.Fields(parts[1])
		if len(fields) < 10 {
			continue
		}
		recv, _ := strconv.ParseUint(fields[0], 10, 64)
		sent, _ := strconv.ParseUint(fields[8], 10, 64)
		curIO.BytesRecv += recv
		curIO.BytesSent += sent
	}

	if c.initialized && elapsed > 0 {
		if curIO.BytesRecv >= c.prevNetIO.BytesRecv {
			s.NetIn = uint64(float64(curIO.BytesRecv-c.prevNetIO.BytesRecv) / elapsed)
		}
		if curIO.BytesSent >= c.prevNetIO.BytesSent {
			s.NetOut = uint64(float64(curIO.BytesSent-c.prevNetIO.BytesSent) / elapsed)
		}
	}
	c.prevNetIO = curIO
}

// ── 系统负载（/proc/loadavg）──────────────────────────────

func (c *Collector) collectLoadAvg(s *Snapshot) {
	data, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return
	}
	fields := strings.Fields(string(data))
	if len(fields) >= 3 {
		s.Load1, _ = strconv.ParseFloat(fields[0], 64)
		s.Load5, _ = strconv.ParseFloat(fields[1], 64)
		s.Load15, _ = strconv.ParseFloat(fields[2], 64)
		s.Load1 = round2(s.Load1)
		s.Load5 = round2(s.Load5)
		s.Load15 = round2(s.Load15)
	}
}

// ── Uptime（/proc/uptime）─────────────────────────────────

func (c *Collector) collectUptime(s *Snapshot) {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return
	}
	fields := strings.Fields(string(data))
	if len(fields) >= 1 {
		val, _ := strconv.ParseFloat(fields[0], 64)
		s.Uptime = uint64(math.Floor(val))
	}
}

// GetSysInfo 获取系统信息（通过 unix.Uname）
func (c *Collector) GetSysInfo() *SysInfo {
	info, err := getSysInfo()
	if err != nil {
		return &SysInfo{Sysname: "unknown"}
	}
	return info
}

func round2(f float64) float64 {
	return float64(int(f*100)) / 100
}
