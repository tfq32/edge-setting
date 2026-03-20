package vsoa

import (
	"context"
	"fmt"
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

// AppInfo 微应用信息（来自 MS）
type AppInfo struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`      // system | user
	HasUI     bool      `json:"has_ui"`
	Status    AppStatus `json:"status"`
	PID       int       `json:"pid"`
	CPU       float64   `json:"cpu"`
	Mem       uint64    `json:"mem"`       // RSS 字节
	Port      []int     `json:"ports"`
	Version   string    `json:"version"`
	StartTime int64     `json:"start_time"` // Unix ms
	WorkDir   string    `json:"work_dir"`
}

// SystemInfo 节点系统信息（来自 MS）
type SystemInfo struct {
	PlatformVersion string `json:"platform_version"`
}

// StatusChangeEvent 状态变更事件
type StatusChangeEvent struct {
	AppID  string    `json:"id"`
	Status AppStatus `json:"status"`
}

// EventHandler 状态变更处理函数
type EventHandler func(event StatusChangeEvent)

// Client VSOA 客户端（此处为 Mock 实现）
type Client struct {
	mu       sync.RWMutex
	apps     map[string]*AppInfo
	handlers []EventHandler
	stopCh   chan struct{}
}

// New 创建 VSOA 客户端并连接
func New(address string) (*Client, error) {
	c := &Client{
		apps:   make(map[string]*AppInfo),
		stopCh: make(chan struct{}),
	}
	c.seedMockApps()
	go c.simulateChanges()
	return c, nil
}

// seedMockApps 初始化模拟微应用列表
func (c *Client) seedMockApps() {
	now := time.Now().UnixMilli()
	mockApps := []*AppInfo{
		{ID: "edge-setting", Name: "Edge Setting", Type: "system", HasUI: true, Status: StatusRunning, PID: 1001, CPU: 0.3, Mem: 29360128, Port: []int{8080}, Version: "v2.0.0", StartTime: now - 86400000, WorkDir: "/opt/edge-setting"},
		{ID: "data-collector", Name: "数据采集服务", Type: "system", HasUI: false, Status: StatusRunning, PID: 1002, CPU: 2.1, Mem: 67108864, Port: []int{9100}, Version: "v1.3.2", StartTime: now - 86400000, WorkDir: "/opt/data-collector"},
		{ID: "modbus-gateway", Name: "Modbus 网关", Type: "user", HasUI: false, Status: StatusRunning, PID: 1003, CPU: 1.2, Mem: 47185920, Port: []int{502}, Version: "v1.1.0", StartTime: now - 3600000, WorkDir: "/opt/modbus-gw"},
		{ID: "rtsp-proxy", Name: "RTSP 视频代理", Type: "user", HasUI: false, Status: StatusStopped, PID: 0, CPU: 0, Mem: 0, Port: []int{554}, Version: "v1.0.3", StartTime: 0, WorkDir: "/opt/rtsp-proxy"},
		{ID: "ota-agent", Name: "OTA 升级代理", Type: "system", HasUI: false, Status: StatusRunning, PID: 1004, CPU: 0.1, Mem: 15728640, Port: []int{9200}, Version: "v0.8.1", StartTime: now - 86400000, WorkDir: "/opt/ota-agent"},
	}
	for _, app := range mockApps {
		c.apps[app.ID] = app
	}
}

// simulateChanges 模拟微应用状态和资源的随机波动
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
					// CPU 小幅波动
					app.CPU = round2(app.CPU + (rand.Float64()-0.5)*0.5)
					if app.CPU < 0 {
						app.CPU = 0
					}
					// 内存轻微增长
					app.Mem += uint64(rand.Intn(102400))
				}
			}
			c.mu.Unlock()
		}
	}
}

// ListApps 查询所有微应用状态
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

// GetApp 查询单个微应用
func (c *Client) GetApp(_ context.Context, id string) (*AppInfo, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	app, ok := c.apps[id]
	if !ok {
		return nil, fmt.Errorf("微应用 %s 不存在", id)
	}
	cp := *app
	return &cp, nil
}

// Start 启动微应用
func (c *Client) Start(_ context.Context, id string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	app, ok := c.apps[id]
	if !ok {
		return fmt.Errorf("微应用 %s 不存在", id)
	}
	if app.Status == StatusRunning {
		return fmt.Errorf("微应用 %s 已在运行中", id)
	}
	app.Status = StatusRunning
	app.PID = 1000 + rand.Intn(9000)
	app.StartTime = time.Now().UnixMilli()
	app.CPU = 0.5
	app.Mem = 20971520
	c.notifyAll(StatusChangeEvent{AppID: id, Status: StatusRunning})
	return nil
}

// Stop 停止微应用
func (c *Client) Stop(_ context.Context, id string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	app, ok := c.apps[id]
	if !ok {
		return fmt.Errorf("微应用 %s 不存在", id)
	}
	app.Status = StatusStopped
	app.PID = 0
	app.CPU = 0
	app.Mem = 0
	app.StartTime = 0
	c.notifyAll(StatusChangeEvent{AppID: id, Status: StatusStopped})
	return nil
}

// Restart 重启微应用
func (c *Client) Restart(ctx context.Context, id string) error {
	if err := c.Stop(ctx, id); err != nil {
		return err
	}
	time.Sleep(200 * time.Millisecond)
	return c.Start(ctx, id)
}

// GetSystemInfo 获取平台系统信息
func (c *Client) GetSystemInfo(_ context.Context) (*SystemInfo, error) {
	return &SystemInfo{PlatformVersion: "v3.1.2"}, nil
}

// Subscribe 订阅状态变更事件
func (c *Client) Subscribe(handler EventHandler) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.handlers = append(c.handlers, handler)
}

func (c *Client) notifyAll(event StatusChangeEvent) {
	for _, h := range c.handlers {
		go h(event)
	}
}

// Close 关闭客户端
func (c *Client) Close() {
	close(c.stopCh)
}

func round2(f float64) float64 {
	return float64(int(f*100)) / 100
}
