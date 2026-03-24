# Edge Setting 接口文档（SylixOS）

**版本**：v2.1.0
**基础路径**：`http://<node-ip>:10000`
**协议**：HTTP/1.1 · WebSocket
**认证**：无（内网部署）
**编码**：UTF-8 · JSON
**框架**：httprouter
**数据采集**：通过 VSOA 从 MS 获取，每 5 秒一次

> Edge Setting 为**只读**监控系统。所有接口均为 `GET` 或 `WebSocket`，不提供任何写操作。
> SylixOS 版本不做本地系统指标采集，所有监控数据通过 VSOA 软总线从 MS 服务获取。

---

## 接口总览

| 方法 | 路径                        | 说明               | 数据来源   |
|------|-----------------------------|--------------------|-----------|
| GET  | `/health`                   | 健康检查           | —         |
| GET  | `/api/v1/system/info`       | 节点 & 系统信息    | VSOA/MS   |
| GET  | `/api/v1/system/metrics`    | 实时指标快照       | VSOA/MS   |
| WS   | `/api/v1/system/metrics/ws` | 指标实时推送       | VSOA/MS   |
| GET  | `/api/v1/metrics/history`   | 历史指标查询       | SQLite    |
| GET  | `/api/v1/apps`              | 微应用列表         | VSOA/MS   |

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

健康检查。

```json
{ "status": "ok", "ts": 1710000000000 }
```

---

## GET `/api/v1/system/info`

返回节点系统信息（通过 VSOA 从 MS 获取）。

### 响应

```json
{
  "code": 0,
  "data": {
    "hostname": "sylixos-edge",
    "arch": "arm64",
    "os": "SylixOS",
    "platform": "SylixOS",
    "platform_version": "3.6.5",
    "kernel_version": "SylixOS 3.6.5",
    "uptime_system": 1209600,
    "net_interfaces": []
  }
}
```

| 字段               | 类型    | 说明                     |
|--------------------|---------|--------------------------|
| `hostname`         | string  | 主机名                   |
| `arch`             | string  | CPU 架构                 |
| `os`               | string  | 操作系统（SylixOS）      |
| `platform`         | string  | 平台名称                 |
| `platform_version` | string  | 平台版本                 |
| `kernel_version`   | string  | 内核版本                 |
| `uptime_system`    | integer | 系统运行时长（秒）       |
| `net_interfaces`   | array   | 网络接口列表             |

---

## GET `/api/v1/system/metrics`

返回当前时刻的系统资源指标快照（通过 VSOA 从 MS 获取）。

### 响应

```json
{
  "code": 0,
  "data": {
    "ts": 1710000000000,
    "cpu": 45.2,
    "cpu_cores": [52.1, 38.3, 41.0, 49.7],
    "mem_used": 3221225472,
    "mem_total": 8589934592,
    "mem_pct": 37.5,
    "swap_used": 0,
    "swap_total": 0,
    "disk_pct": 28.6,
    "disk_read": 102400,
    "disk_write": 51200,
    "net_in": 102400,
    "net_out": 51200,
    "load1": 1.42,
    "load5": 1.28,
    "load15": 0.96,
    "uptime": 1209600
  }
}
```

| 字段         | 单位     | 说明                    |
|--------------|----------|-------------------------|
| `ts`         | ms       | 采集时间戳              |
| `cpu`        | %        | 总体 CPU 占用率         |
| `cpu_cores`  | %        | 各核心占用率            |
| `mem_used`   | bytes    | 已用内存                |
| `mem_total`  | bytes    | 总内存                  |
| `mem_pct`    | %        | 内存使用率              |
| `swap_used`  | bytes    | 已用 Swap               |
| `swap_total` | bytes    | 总 Swap                 |
| `disk_pct`   | %        | 根分区使用率            |
| `disk_read`  | bytes/s  | 磁盘读取速率            |
| `disk_write` | bytes/s  | 磁盘写入速率            |
| `net_in`     | bytes/s  | 网络流入速率            |
| `net_out`    | bytes/s  | 网络流出速率            |
| `load1`      | —        | 1 分钟系统负载          |
| `load5`      | —        | 5 分钟系统负载          |
| `load15`     | —        | 15 分钟系统负载         |
| `uptime`     | 秒       | 系统运行时长            |

---

## WS `/api/v1/system/metrics/ws`

建立 WebSocket 长连接，服务端每 **5 秒**推送一次最新指标快照。

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

### Query 参数

| 参数    | 默认   | 枚举值                          |
|---------|--------|--------------------------------|
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

- 写入频率：每 5 秒一条（从 VSOA 获取后写入）
- 保留时长：7 天（`config.yaml` 中 `metrics_retain_days` 可调）
- 清理时机：每小时执行一次

---

## GET `/api/v1/apps`

返回所有微应用的当前状态（通过 VSOA 从 MS 获取）。

### 响应

```json
{
  "code": 0,
  "data": [
    {
      "id": "data-collector",
      "name": "数据采集服务",
      "type": "system",
      "has_ui": false,
      "status": "running",
      "pid": 1002,
      "cpu": 2.1,
      "mem": 67108864,
      "ports": [9100],
      "version": "v1.3.2",
      "start_time": 1709913600000,
      "work_dir": "/opt/data-collector"
    }
  ]
}
```

| 字段         | 类型    | 说明                                    |
|--------------|---------|-----------------------------------------|
| `id`         | string  | 微应用唯一 ID                           |
| `name`       | string  | 显示名称                                |
| `type`       | string  | `system` / `user`                       |
| `has_ui`     | boolean | 是否有 Web UI                           |
| `status`     | string  | `running` / `stopped` / `error`         |
| `pid`        | integer | 进程 PID，停止时为 `0`                  |
| `cpu`        | float   | CPU 占用率（%）                         |
| `mem`        | integer | RSS 内存（bytes）                       |
| `ports`      | array   | 监听端口                                |
| `version`    | string  | 版本号                                  |
| `start_time` | integer | 启动时间戳（ms），停止时为 `0`          |
| `work_dir`   | string  | 工作目录                                |

---

## 错误码

| code | HTTP | 说明               |
|------|------|--------------------|
| 0    | 200  | 成功               |
| 1010 | 503  | MS 不可达          |
| 9999 | 500  | 服务内部错误       |

---

*Edge Setting v2.1.0 · SylixOS · 只读监控 · 2025*
