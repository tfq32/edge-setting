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

// Client VSOA 客户端（只读，Mock 实现）
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
	ticker := time.NewTicker(3 * time.Second)
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

// Close 关闭客户端
func (c *Client) Close() {
	close(c.stopCh)
}

func round2(f float64) float64 {
	return float64(int(f*100)) / 100
}
