# VSOA 接口文档（Linux → MS）

**版本**：v2.1.0
**协议**：VSOA 软总线
**方向**：Edge Setting → MS
**接口前缀**：`/api/v1/edge_setting`

---

## 概述

Linux 版 Edge Setting 的系统监控数据由 gopsutil 本地采集，不通过 VSOA 获取。
VSOA 仅用于从 MS 获取**微应用列表**。

---

## 接口列表

| VSOA URL                              | 方法     | 说明         | 调用频率 |
|---------------------------------------|----------|-------------|---------|
| `/api/v1/edge_setting/app/list`       | RPC GET  | 微应用列表   | 按需     |

---

## 获取微应用列表

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
- 断线重连：指数退避 1s → 2s → 4s → 8s → 16s → 30s（上限）

---

*Edge Setting v2.1.0 · Linux · VSOA 接口 · 2025*
