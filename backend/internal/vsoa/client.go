package vsoa

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
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

// AppInfo 微应用信息（来自 MS）
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

// Client VSOA 客户端
type Client struct {
	mu        sync.RWMutex
	cli       *vsoaClient.Client
	addr      string
	connected bool
	stopCh    chan struct{}
	// mock 数据，MS 未连接时使用
	mockApps []*AppInfo
}

// New 创建 VSOA 客户端并连接到 MS
func New(addr string) (*Client, error) {
	c := &Client{
		addr:   addr,
		stopCh: make(chan struct{}),
	}
	c.initMockApps()

	if err := c.connect(); err != nil {
		log.Printf("VSOA 初始连接失败: %v，将后台重连，使用 Mock 数据", err)
		go c.reconnectLoop()
	}

	return c, nil
}

func (c *Client) connect() error {
	cli := vsoaClient.NewClient(vsoaClient.DefaultOption)
	if _, err := cli.Connect("tcp", c.addr); err != nil {
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

// ListApps 查询所有微应用状态
// 优先通过 VSOA RPC GET /api/v1/edge_setting/app/list 获取
// MS 未连接时返回 Mock 数据
func (c *Client) ListApps(_ context.Context) ([]*AppInfo, error) {
	data, err := c.rpcGet(apiPrefix + "/app/list")
	if err == nil {
		var apps []*AppInfo
		if err := json.Unmarshal(data, &apps); err == nil {
			return apps, nil
		}
	}

	// fallback: 返回 Mock 数据
	return c.getMockApps(), nil
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

// ── Mock 数据（MS 未连接时使用，后续删除）──────────────────

func (c *Client) initMockApps() {
	now := time.Now().UnixMilli()
	c.mockApps = []*AppInfo{
		{ID: "edge-setting", Name: "Edge Setting", Type: "system", HasUI: true, Status: StatusRunning, PID: 1001, CPU: 0.3, Mem: 29360128, Port: []int{8080}, Version: "v2.0.0", StartTime: now - 86400000, WorkDir: "/opt/edge-setting"},
		{ID: "data-collector", Name: "数据采集服务", Type: "system", HasUI: false, Status: StatusRunning, PID: 1002, CPU: 2.1, Mem: 67108864, Port: []int{9100}, Version: "v1.3.2", StartTime: now - 86400000, WorkDir: "/opt/data-collector"},
		{ID: "modbus-gateway", Name: "Modbus 网关", Type: "user", HasUI: false, Status: StatusRunning, PID: 1003, CPU: 1.2, Mem: 47185920, Port: []int{502}, Version: "v1.1.0", StartTime: now - 3600000, WorkDir: "/opt/modbus-gw"},
		{ID: "rtsp-proxy", Name: "RTSP 视频代理", Type: "user", HasUI: false, Status: StatusStopped, PID: 0, CPU: 0, Mem: 0, Port: []int{554}, Version: "v1.0.3", StartTime: 0, WorkDir: "/opt/rtsp-proxy"},
		{ID: "ota-agent", Name: "OTA 升级代理", Type: "system", HasUI: false, Status: StatusRunning, PID: 1004, CPU: 0.1, Mem: 15728640, Port: []int{9200}, Version: "v0.8.1", StartTime: now - 86400000, WorkDir: "/opt/ota-agent"},
	}
}

func (c *Client) getMockApps() []*AppInfo {
	c.mu.Lock()
	defer c.mu.Unlock()
	result := make([]*AppInfo, len(c.mockApps))
	for i, app := range c.mockApps {
		cp := *app
		// 模拟小幅波动
		if cp.Status == StatusRunning {
			cp.CPU = round2(cp.CPU + (rand.Float64()-0.5)*0.5)
			if cp.CPU < 0 {
				cp.CPU = 0
			}
		}
		result[i] = &cp
	}
	return result
}

func round2(f float64) float64 {
	return float64(int(f*100)) / 100
}
