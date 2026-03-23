# Edge Setting 接口文档

**版本**：v2.0.0  
**基础路径**：`http://<node-ip>:8080`  
**协议**：HTTP/1.1 · WebSocket  
**认证**：无（内网部署）  
**编码**：UTF-8 · JSON  

> Edge Setting 为**只读**监控系统。所有接口均为 `GET` 或 `WebSocket`，不提供任何写操作。

---

## 接口总览

| 方法 | 路径                        | 说明               |
|------|-----------------------------|--------------------|
| GET  | `/health`                   | 健康检查           |
| GET  | `/api/v1/system/info`       | 节点 & 系统信息    |
| GET  | `/api/v1/system/metrics`    | 实时指标快照       |
| WS   | `/api/v1/system/metrics/ws` | 指标实时推送       |
| GET  | `/api/v1/metrics/history`   | 历史指标查询       |
| GET  | `/api/v1/apps`              | 微应用列表         |

---

## 公共约定

### 响应格式

```json
{ "code": 0, "data": { ... } }
```

| 字段      | 类型    | 说明                          |
|-----------|---------|-------------------------------|
| `code`    | integer | `0` 成功，非 0 错误           |
| `data`    | any     | 业务数据，错误时不存在        |
| `message` | string  | 错误描述，仅错误响应时存在    |

### 时间戳

所有时间戳字段均为 **Unix 毫秒**（`int64`）。

---

## GET `/health`

健康检查，适用于探针、监控系统。

```json
{ "status": "ok", "ts": 1710000000000 }
```

---

## GET `/api/v1/system/info`

返回节点系统信息（全部来自 gopsutil 实时读取）。

### 响应

```json
{
  "code": 0,
  "data": {
    "hostname":         "edge-node-01",
    "arch":             "amd64",
    "os":               "linux",
    "platform":         "ubuntu",
    "platform_version": "22.04",
    "kernel_version":   "5.15.0-91-generic",
    "uptime_system":    1209600,
    "net_interfaces": [
      {
        "name":  "eth0",
        "addrs": ["192.168.1.42/24"],
        "flags": ["up", "broadcast", "running"]
      }
    ]
  }
}
```

| 字段               | 类型    | 来源             | 说明                     |
|--------------------|---------|------------------|--------------------------|
| `hostname`         | string  | `/proc/hostname` | 主机名                   |
| `arch`             | string  | `runtime.GOARCH` | CPU 架构                 |
| `os`               | string  | gopsutil         | 操作系统类型             |
| `platform`         | string  | gopsutil         | 发行版名称               |
| `platform_version` | string  | gopsutil         | 发行版版本               |
| `kernel_version`   | string  | gopsutil         | 内核版本                 |
| `uptime_system`    | integer | `/proc/uptime`   | 系统运行时长（秒）       |
| `net_interfaces`   | array   | gopsutil         | 网络接口列表             |

---

## GET `/api/v1/system/metrics`

返回当前时刻的系统资源指标快照（全部来自 gopsutil 实时采集）。

### 响应

```json
{
  "code": 0,
  "data": {
    "ts":         1710000000000,
    "cpu":        62.5,
    "cpu_cores":  [72.1, 54.3, 48.0, 38.7],
    "mem_used":   5368709120,
    "mem_total":  8589934592,
    "mem_pct":    62.5,
    "swap_used":  0,
    "swap_total": 2147483648,
    "disk_pct":   43.2,
    "disk_read":  204800,
    "disk_write": 102400,
    "net_in":     1048576,
    "net_out":    524288,
    "load1":      1.42,
    "load5":      1.28,
    "load15":     0.96,
    "uptime":     1209600
  }
}
```

| 字段         | 单位     | 来源                     | 说明                    |
|--------------|----------|--------------------------|-------------------------|
| `ts`         | ms       | `time.Now()`             | 采集时间戳              |
| `cpu`        | %        | `cpu.Percent()`          | 总体 CPU 占用率         |
| `cpu_cores`  | %        | `cpu.Percent(perCPU)`    | 各核心占用率            |
| `mem_used`   | bytes    | `mem.VirtualMemory()`    | 已用内存                |
| `mem_total`  | bytes    | `mem.VirtualMemory()`    | 总内存                  |
| `mem_pct`    | %        | `mem.VirtualMemory()`    | 内存使用率              |
| `swap_used`  | bytes    | `mem.SwapMemory()`       | 已用 Swap               |
| `swap_total` | bytes    | `mem.SwapMemory()`       | 总 Swap                 |
| `disk_pct`   | %        | `disk.Usage("/")`        | 根分区使用率            |
| `disk_read`  | bytes/s  | `disk.IOCounters()` 差值 | 磁盘读取速率            |
| `disk_write` | bytes/s  | `disk.IOCounters()` 差值 | 磁盘写入速率            |
| `net_in`     | bytes/s  | `net.IOCounters()` 差值  | 网络流入速率            |
| `net_out`    | bytes/s  | `net.IOCounters()` 差值  | 网络流出速率            |
| `load1`      | —        | `load.Avg()`             | 1 分钟系统负载          |
| `load5`      | —        | `load.Avg()`             | 5 分钟系统负载          |
| `load15`     | —        | `load.Avg()`             | 15 分钟系统负载         |
| `uptime`     | 秒       | `host.Info().Uptime`     | 系统运行时长（同 /info）|

---

## WS `/api/v1/system/metrics/ws`

建立 WebSocket 长连接，服务端每 **3 秒**推送一次最新指标快照。

### 推送消息格式

```json
{
  "type": "metrics",
  "data": { /* 与 /system/metrics 的 data 字段完全一致 */ }
}
```

### 心跳

- 服务端每 50 秒发送 Ping 帧
- 客户端需在 60 秒内响应 Pong，否则断开
- 客户端应实现自动重连（推荐指数退避，上限 30s）

---

## GET `/api/v1/metrics/history`

查询历史指标，数据来自 SQLite，超过 7 天自动清理。  
若查询时间段内无数据，返回空数组 `[]`（前端绘制 0% 平线）。

### Query 参数

| 参数    | 默认   | 枚举值                         |
|---------|--------|-------------------------------|
| `type`  | `cpu`  | `cpu` / `mem` / `disk` / `net` |
| `range` | `1h`   | `1h` / `6h` / `24h` / `7d`    |

### 响应

```json
{
  "code": 0,
  "data": [
    { "ts": 1709996400000, "cpu": 45.2, "mem_pct": 60.1, "disk_pct": 43.0, "net_in": 102400, "net_out": 51200 }
  ]
}
```

### 数据保留策略

- 写入频率：每 3 秒一条
- 保留时长：7 天（`config.yaml` 中 `metrics_retain_days` 可调）
- 清理时机：每小时执行一次定时任务

---

## GET `/api/v1/apps`

返回所有微应用的当前状态。  
数据来源：VSOA/MS 实时查询（**目前为 Mock**，接入真实 MS 后自动替换）。

### 响应

```json
{
  "code": 0,
  "data": [
    {
      "id":         "data-collector",
      "name":       "数据采集服务",
      "type":       "system",
      "has_ui":     false,
      "status":     "running",
      "pid":        1002,
      "cpu":        2.1,
      "mem":        67108864,
      "ports":      [9100],
      "version":    "v1.3.2",
      "start_time": 1709913600000,
      "work_dir":   "/opt/data-collector"
    }
  ]
}
```

### AppInfo 字段

| 字段         | 类型    | 说明                                    |
|--------------|---------|-----------------------------------------|
| `id`         | string  | 微应用唯一 ID                           |
| `name`       | string  | 显示名称                                |
| `type`       | string  | `system`（系统）/ `user`（用户）        |
| `has_ui`     | boolean | 是否有 Web UI                           |
| `status`     | string  | `running` / `stopped` / `error`         |
| `pid`        | integer | 进程 PID，停止时为 `0`                  |
| `cpu`        | float   | CPU 占用率（%）                         |
| `mem`        | integer | RSS 内存（bytes）                       |
| `ports`      | array   | 监听端口                                |
| `version`    | string  | 版本号                                  |
| `start_time` | integer | 启动时间戳（ms），停止时为 `0`          |
| `work_dir`   | string  | 工作目录                                |

### 错误

MS 不可达时返回 `503`：

```json
{ "code": 1010, "message": "MS 不可达: ..." }
```

---

## 错误码

| code | HTTP | 说明               |
|------|------|--------------------|
| 0    | 200  | 成功               |
| 1010 | 503  | MS 不可达          |
| 9999 | 500  | 服务内部错误       |

---

## VSOA 内部接口（开发参考）

Edge Setting 通过 VSOA 软总线与 MS 通信，**仅调用以下只读接口**：

| 操作             | 方向    | 对应 HTTP 接口      |
|------------------|---------|---------------------|
| 查询微应用列表   | RPC GET | `GET /api/v1/apps`  |

断线后自动重连，指数退避：1s → 2s → 4s → 8s → 16s → 30s（上限）。

---

*Edge Setting v2.0.0 · 只读监控 · 2024*
