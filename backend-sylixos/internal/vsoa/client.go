package vsoa

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	vsoaClient "github.com/acoinfo/vsoa/client"
	"github.com/acoinfo/vsoa/protocol"
)

// VSOA 接口路径前缀
const apiPrefix = "/api/v1/edge_setting"

// AppStatus 微应用状态
type AppStatus string

const (
	StatusRunning AppStatus = "running"
	StatusStopped AppStatus = "stopped"
	StatusError   AppStatus = "error"
)

// AppInfo 微应用信息
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

// MetricsSnapshot 系统指标快照
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

// SystemInfo 系统信息
type SystemInfo struct {
	Hostname        string `json:"hostname"`
	Arch            string `json:"arch"`
	OS              string `json:"os"`
	Platform        string `json:"platform"`
	PlatformVersion string `json:"platform_version"`
	KernelVersion   string `json:"kernel_version"`
	UptimeSystem    uint64 `json:"uptime_system"`
}

// Client VSOA 客户端
type Client struct {
	mu        sync.RWMutex
	cli       *vsoaClient.Client
	addr      string
	connected bool
	stopCh    chan struct{}
}

// New 创建 VSOA 客户端并连接到 MS
func New(addr string) (*Client, error) {
	c := &Client{
		addr:   addr,
		stopCh: make(chan struct{}),
	}

	if err := c.connect(); err != nil {
		log.Printf("VSOA 初始连接失败: %v，将后台重连", err)
		go c.reconnectLoop()
	}

	return c, nil
}

func (c *Client) connect() error {
	cli := vsoaClient.NewClient(vsoaClient.DefaultOption)
	if _, err := cli.Connect("vsoa", c.addr); err != nil {
		return err
	}
	c.mu.Lock()
	c.cli = cli
	c.connected = true
	c.mu.Unlock()
	log.Printf("VSOA 已连接: %s", c.addr)
	return nil
}

func (c *Client) reconnectLoop() {
	delays := []time.Duration{1, 2, 4, 8, 16, 30}
	idx := 0
	for {
		select {
		case <-c.stopCh:
			return
		default:
		}
		d := delays[idx] * time.Second
		time.Sleep(d)
		if err := c.connect(); err != nil {
			log.Printf("VSOA 重连失败: %v，%v 后重试", err, d)
			if idx < len(delays)-1 {
				idx++
			}
		} else {
			return
		}
	}
}

// rpcGet 发送 VSOA RPC GET 请求，返回 Param（JSON）
func (c *Client) rpcGet(url string) (json.RawMessage, error) {
	c.mu.RLock()
	cli := c.cli
	connected := c.connected
	c.mu.RUnlock()

	if !connected || cli == nil {
		return nil, fmt.Errorf("VSOA 未连接")
	}

	req := protocol.NewMessage()
	req.SetMessageType(protocol.TypeRPC)
	req.SetMessageRpcMethod(protocol.RpcMethodGet)

	reply, err := cli.Call(url, protocol.TypeRPC, protocol.RpcMethodGet, req)
	if err != nil {
		// 连接可能断开，触发重连
		c.mu.Lock()
		c.connected = false
		c.mu.Unlock()
		go c.reconnectLoop()
		return nil, fmt.Errorf("VSOA RPC 调用失败: %w", err)
	}

	if reply.StatusType() != protocol.StatusSuccess {
		return nil, fmt.Errorf("VSOA 返回错误状态: %s", reply.StatusTypeText())
	}

	return reply.Param, nil
}

// GetMetrics 从 MS 获取系统指标
// VSOA RPC GET /api/v1/edge_setting/system/metrics
func (c *Client) GetMetrics(_ context.Context) (*MetricsSnapshot, error) {
	data, err := c.rpcGet(apiPrefix + "/system/metrics")
	if err != nil {
		return nil, err
	}
	var snap MetricsSnapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return nil, fmt.Errorf("解析指标数据失败: %w", err)
	}
	return &snap, nil
}

// GetSystemInfo 从 MS 获取系统信息
// VSOA RPC GET /api/v1/edge_setting/system/info
func (c *Client) GetSystemInfo(_ context.Context) (*SystemInfo, error) {
	data, err := c.rpcGet(apiPrefix + "/system/info")
	if err != nil {
		return nil, err
	}
	var info SystemInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, fmt.Errorf("解析系统信息失败: %w", err)
	}
	return &info, nil
}

// ListApps 从 MS 获取微应用列表
// VSOA RPC GET /api/v1/edge_setting/app/list
func (c *Client) ListApps(_ context.Context) ([]*AppInfo, error) {
	data, err := c.rpcGet(apiPrefix + "/app/list")
	if err != nil {
		return nil, err
	}
	var apps []*AppInfo
	if err := json.Unmarshal(data, &apps); err != nil {
		return nil, fmt.Errorf("解析应用列表失败: %w", err)
	}
	return apps, nil
}

// Close 关闭客户端
func (c *Client) Close() {
	close(c.stopCh)
	c.mu.Lock()
	if c.cli != nil {
		c.cli.Close()
	}
	c.mu.Unlock()
}
