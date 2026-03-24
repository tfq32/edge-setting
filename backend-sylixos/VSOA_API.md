# VSOA 内部接口文档（SylixOS）

**版本**：v2.1.0
**协议**：VSOA 软总线
**方向**：Edge Setting → MS
**状态**：**Mock 实现**，接入真实 MS 后替换

---

## 概述

SylixOS 版 Edge Setting **不做本地系统采集**，所有数据通过 VSOA 软总线从 MS 服务获取。

每 5 秒轮询一次 MS，获取系统指标后：
1. 写入 SQLite 持久化
2. 通过 WebSocket 广播给前端

---

## 接口列表

| 操作             | 方向              | 方法    | 对应 HTTP 接口          | 频率  | 当前状态 |
|------------------|-------------------|---------|------------------------|-------|----------|
| 获取系统指标     | Edge Setting → MS | RPC GET | `GET /system/metrics`  | 5s    | Mock     |
| 获取系统信息     | Edge Setting → MS | RPC GET | `GET /system/info`     | 按需  | Mock     |
| 查询微应用列表   | Edge Setting → MS | RPC GET | `GET /apps`            | 按需  | Mock     |

---

## 1. 获取系统指标

定时任务每 5 秒调用一次，获取系统资源快照。

### 请求

```
VSOA RPC GET /ms/system/metrics
```

### 响应

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

### 字段说明

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

### 数据流转

```
MS → VSOA → Edge Setting → SQLite (持久化)
                         → WebSocket (实时推送)
                         → HTTP API (按需查询)
```

---

## 2. 获取系统信息

前端请求 `/api/v1/system/info` 时按需调用。

### 请求

```
VSOA RPC GET /ms/system/info
```

### 响应

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

## 3. 查询微应用列表

前端请求 `/api/v1/apps` 时按需调用。

### 请求

```
VSOA RPC GET /ms/apps/list
```

### 响应

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

- 连接地址：`config.yaml` 中 `vsoa.ms_address`（默认 `vsoa://localhost:3000`）
- 断线重连：指数退避 1s → 2s → 4s → 8s → 16s → 30s（上限）

---

## 架构图

```
┌─────────┐    VSOA     ┌────────┐    HTTP/WS    ┌──────────┐
│   MS    │ ──────────→ │ Edge   │ ─────────────→ │ Frontend │
│ Service │  metrics    │Setting │  /api/v1/*     │   (Vue)  │
│         │  sysinfo    │(SylixOS│  /ws           │          │
│         │  apps       │        │                │          │
└─────────┘             └───┬────┘                └──────────┘
                            │
                            ↓
                       ┌─────────┐
                       │ SQLite  │
                       │ (持久化) │
                       └─────────┘
```

---

*Edge Setting v2.1.0 · SylixOS · VSOA 内部接口 · 2025*
