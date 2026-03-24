package vsoa

import (
	"context"
	"math/rand"
	"sync"
	"time"
)

// AppStatus 微应用状态
type AppStatus string

const (
	StatusRunning AppStatus = "running"
	StatusStopped AppStatus = "stopped"
	StatusError   AppStatus = "error"
)

// AppInfo 微应用信息（来自 MS，只读）
type AppInfo struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	HasUI     bool      `json:"has_ui"`
	Status    AppStatus `json:"status"`
	PID       int       `json:"pid"`
	CPU       float64   `json:"cpu"`
	Mem       uint64    `json:"mem"`
	Port      []int     `json:"ports"`
	Version   string    `json:"version"`
	StartTime int64     `json:"start_time"`
	WorkDir   string    `json:"work_dir"`
}

// MetricsSnapshot 系统指标快照（从 MS 获取）
type MetricsSnapshot struct {
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

// SystemInfo 系统信息（从 MS 获取）
type SystemInfo struct {
	Hostname        string `json:"hostname"`
	Arch            string `json:"arch"`
	OS              string `json:"os"`
	Platform        string `json:"platform"`
	PlatformVersion string `json:"platform_version"`
	KernelVersion   string `json:"kernel_version"`
	UptimeSystem    uint64 `json:"uptime_system"`
}

// Client VSOA 客户端（目前为 Mock 实现，接入真实 MS 后替换）
type Client struct {
	mu     sync.RWMutex
	apps   map[string]*AppInfo
	stopCh chan struct{}
}

// New 创建 VSOA 客户端并连接
func New(_ string) (*Client, error) {
	c := &Client{
		apps:   make(map[string]*AppInfo),
		stopCh: make(chan struct{}),
	}
	c.seedMockApps()
	go c.simulateChanges()
	return c, nil
}

func (c *Client) seedMockApps() {
	now := time.Now().UnixMilli()
	for _, app := range []*AppInfo{
		{ID: "edge-setting", Name: "Edge Setting", Type: "system", HasUI: true, Status: StatusRunning, PID: 1001, CPU: 0.3, Mem: 29360128, Port: []int{8080}, Version: "v2.0.0", StartTime: now - 86400000, WorkDir: "/opt/edge-setting"},
		{ID: "data-collector", Name: "数据采集服务", Type: "system", HasUI: false, Status: StatusRunning, PID: 1002, CPU: 2.1, Mem: 67108864, Port: []int{9100}, Version: "v1.3.2", StartTime: now - 86400000, WorkDir: "/opt/data-collector"},
		{ID: "modbus-gateway", Name: "Modbus 网关", Type: "user", HasUI: false, Status: StatusRunning, PID: 1003, CPU: 1.2, Mem: 47185920, Port: []int{502}, Version: "v1.1.0", StartTime: now - 3600000, WorkDir: "/opt/modbus-gw"},
		{ID: "rtsp-proxy", Name: "RTSP 视频代理", Type: "user", HasUI: false, Status: StatusStopped, PID: 0, CPU: 0, Mem: 0, Port: []int{554}, Version: "v1.0.3", StartTime: 0, WorkDir: "/opt/rtsp-proxy"},
		{ID: "ota-agent", Name: "OTA 升级代理", Type: "system", HasUI: false, Status: StatusRunning, PID: 1004, CPU: 0.1, Mem: 15728640, Port: []int{9200}, Version: "v0.8.1", StartTime: now - 86400000, WorkDir: "/opt/ota-agent"},
	} {
		c.apps[app.ID] = app
	}
}

// simulateChanges CPU / 内存小幅随机漂移
func (c *Client) simulateChanges() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-c.stopCh:
			return
		case <-ticker.C:
			c.mu.Lock()
			for _, app := range c.apps {
				if app.Status == StatusRunning {
					app.CPU = round2(app.CPU + (rand.Float64()-0.5)*0.5)
					if app.CPU < 0 {
						app.CPU = 0
					}
					app.Mem += uint64(rand.Intn(102400))
				}
			}
			c.mu.Unlock()
		}
	}
}

// ListApps 查询所有微应用状态（只读）
func (c *Client) ListApps(_ context.Context) ([]*AppInfo, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make([]*AppInfo, 0, len(c.apps))
	for _, app := range c.apps {
		cp := *app
		result = append(result, &cp)
	}
	return result, nil
}

// GetMetrics 从 MS 获取系统指标快照
// TODO: 接入真实 VSOA RPC 调用，当前为 Mock
func (c *Client) GetMetrics(_ context.Context) (*MetricsSnapshot, error) {
	now := time.Now()
	return &MetricsSnapshot{
		Timestamp: now.UnixMilli(),
		CPU:       round2(20 + rand.Float64()*60),
		CPUCores:  []float64{round2(rand.Float64() * 100), round2(rand.Float64() * 100), round2(rand.Float64() * 100), round2(rand.Float64() * 100)},
		MemUsed:   uint64(2+rand.Intn(4)) * 1024 * 1024 * 1024,
		MemTotal:  8 * 1024 * 1024 * 1024,
		MemPct:    round2(30 + rand.Float64()*40),
		SwapUsed:  0,
		SwapTotal: 0,
		DiskPct:   round2(20 + rand.Float64()*30),
		DiskRead:  uint64(rand.Intn(1024 * 1024)),
		DiskWrite: uint64(rand.Intn(512 * 1024)),
		NetIn:     uint64(rand.Intn(1024 * 1024)),
		NetOut:    uint64(rand.Intn(512 * 1024)),
		Load1:     round2(rand.Float64() * 4),
		Load5:     round2(rand.Float64() * 3),
		Load15:    round2(rand.Float64() * 2),
		Uptime:    uint64(now.Unix()),
	}, nil
}

// GetSystemInfo 从 MS 获取系统信息
// TODO: 接入真实 VSOA RPC 调用，当前为 Mock
func (c *Client) GetSystemInfo(_ context.Context) (*SystemInfo, error) {
	return &SystemInfo{
		Hostname:        "sylixos-edge",
		Arch:            "arm64",
		OS:              "SylixOS",
		Platform:        "SylixOS",
		PlatformVersion: "3.6.5",
		KernelVersion:   "SylixOS 3.6.5",
		UptimeSystem:    uint64(time.Now().Unix()),
	}, nil
}

// Close 关闭客户端
func (c *Client) Close() {
	close(c.stopCh)
}

func round2(f float64) float64 {
	return float64(int(f*100)) / 100
}
