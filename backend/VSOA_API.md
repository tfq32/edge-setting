# VSOA 内部接口文档（Linux）

**版本**：v2.1.0
**协议**：VSOA 软总线
**方向**：Edge Setting → MS
**状态**：**Mock 实现**，接入真实 MS 后替换

---

## 概述

Linux 版 Edge Setting 的系统监控数据由 **gopsutil 本地采集**，不通过 VSOA 获取。

VSOA 仅用于与 MS 通信，获取**微应用列表**信息。

---

## 接口列表

| 操作             | 方向              | 方法    | 对应 HTTP 接口      | 当前状态 |
|------------------|-------------------|---------|---------------------|----------|
| 查询微应用列表   | Edge Setting → MS | RPC GET | `GET /api/v1/apps`  | Mock     |

---

## 查询微应用列表

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

### 字段说明

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

*Edge Setting v2.1.0 · Linux · VSOA 内部接口 · 2025*
