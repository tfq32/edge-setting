# Edge Setting VSOA 接口文档

**版本**：v2.1.0
**协议**：VSOA 软总线
**方向**：Edge Setting → MS
**SDK**：`github.com/acoinfo/vsoa`
**接口前缀**：`/api/v1/edge_setting`

---

## 概述

Edge Setting 通过 VSOA 软总线与 MS 通信。响应数据统一通过 `payload.Param`（JSON）返回。

---

## 接口列表

| VSOA URL                                    | 方法     | 说明         | 调用频率 | 平台            |
|---------------------------------------------|----------|-------------|---------|-----------------|
| `/api/v1/edge_setting/system/metrics`       | RPC GET  | 系统指标快照 | 每 5 秒  | SylixOS         |
| `/api/v1/edge_setting/system/info`          | RPC GET  | 系统信息     | 按需     | SylixOS         |
| `/api/v1/edge_setting/app/list`             | RPC GET  | 微应用列表   | 按需     | Linux / SylixOS |

> **Linux** 的系统监控数据由 gopsutil 本地采集，不通过 VSOA 获取。Linux 仅使用 `app/list` 接口。
> **SylixOS** 不做本地采集，所有监控数据均通过 VSOA 从 MS 获取。

---

## 1. 获取系统指标

**平台**：SylixOS

每 5 秒定时轮询，获取后写入 SQLite 并通过 WebSocket 广播给前端。

### 请求

```
VSOA RPC GET /api/v1/edge_setting/system/metrics
```

### 响应（payload.Param）

```json
{
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
```

| 字段         | 类型    | 单位    | 说明             |
|--------------|---------|---------|------------------|
| `ts`         | int64   | ms      | 采集时间戳       |
| `cpu`        | float64 | %       | 总体 CPU 占用率  |
| `cpu_cores`  | []float | %       | 各核心占用率     |
| `mem_used`   | uint64  | bytes   | 已用内存         |
| `mem_total`  | uint64  | bytes   | 总内存           |
| `mem_pct`    | float64 | %       | 内存使用率       |
| `swap_used`  | uint64  | bytes   | 已用 Swap        |
| `swap_total` | uint64  | bytes   | 总 Swap          |
| `disk_pct`   | float64 | %       | 磁盘使用率       |
| `disk_read`  | uint64  | bytes/s | 磁盘读取速率     |
| `disk_write` | uint64  | bytes/s | 磁盘写入速率     |
| `net_in`     | uint64  | bytes/s | 网络流入速率     |
| `net_out`    | uint64  | bytes/s | 网络流出速率     |
| `load1`      | float64 | —       | 1 分钟负载       |
| `load5`      | float64 | —       | 5 分钟负载       |
| `load15`     | float64 | —       | 15 分钟负载      |
| `uptime`     | uint64  | 秒      | 系统运行时长     |

---

## 2. 获取系统信息

**平台**：SylixOS

前端请求 `/api/v1/system/info` 时按需调用。

### 请求

```
VSOA RPC GET /api/v1/edge_setting/system/info
```

### 响应（payload.Param）

```json
{
  "hostname": "sylixos-edge",
  "arch": "arm64",
  "os": "SylixOS",
  "platform": "SylixOS",
  "platform_version": "3.6.5",
  "kernel_version": "SylixOS 3.6.5",
  "uptime_system": 1209600
}
```

| 字段               | 类型   | 说明           |
|--------------------|--------|----------------|
| `hostname`         | string | 主机名         |
| `arch`             | string | CPU 架构       |
| `os`               | string | 操作系统       |
| `platform`         | string | 平台名称       |
| `platform_version` | string | 平台版本       |
| `kernel_version`   | string | 内核版本       |
| `uptime_system`    | uint64 | 运行时长（秒） |

---

## 3. 获取微应用列表

**平台**：Linux / SylixOS

前端请求 `/api/v1/app/list` 时按需调用。

### 请求

```
VSOA RPC GET /api/v1/edge_setting/app/list
```

### 响应（payload.Param）

```json
[
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
```

| 字段         | 类型    | 说明                            |
|--------------|---------|---------------------------------|
| `id`         | string  | 微应用唯一 ID                   |
| `name`       | string  | 显示名称                        |
| `type`       | string  | `system` / `user`               |
| `has_ui`     | boolean | 是否有 Web UI                   |
| `status`     | string  | `running` / `stopped` / `error` |
| `pid`        | integer | 进程 PID                        |
| `cpu`        | float   | CPU 占用率（%）                 |
| `mem`        | integer | RSS 内存（bytes）               |
| `ports`      | array   | 监听端口                        |
| `version`    | string  | 版本号                          |
| `start_time` | integer | 启动时间戳（ms）                |
| `work_dir`   | string  | 工作目录                        |

---

## 连接管理

- 连接地址：`config.yaml` → `vsoa.ms_address`（默认 `vsoa://localhost:3000`）
- 自动重连：指数退避 1s → 2s → 4s → 8s → 16s → 30s（上限）
- 连接失败不阻塞服务启动，后台持续重连

---

*Edge Setting v2.1.0 · VSOA 接口 · 2025*
