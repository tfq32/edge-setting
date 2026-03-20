# Edge Setting v2.0.0

运行在 Linux 边缘节点上的设置类微应用，提供**系统资源实时监控**、**微应用生命周期管控**和**节点基础信息展示**三大核心能力。

---

## 快速开始

### 环境要求

| 工具 | 最低版本 |
|------|----------|
| Go   | 1.21     |
| Node | 18.x     |
| pnpm / npm | 任意 |

### 1. 配置

```bash
cp backend/configs/config.yaml my-config.yaml
```

编辑 `my-config.yaml`，至少修改以下内容：

```yaml
auth:
  tokens:
    - token: "your-readonly-token"      # 只读 Token
      permission: readonly
    - token: "your-readwrite-token"     # 读写 Token（可执行启停操作）
      permission: readwrite

vsoa:
  ms_address: "vsoa://localhost:3000"   # MS 地址

data:
  dir: /var/lib/edge-setting            # 数据目录
```

### 2. 构建

```bash
# 构建前端 + 后端
make build

# 或分别构建
make build-frontend
make build-backend
```

### 3. 运行

```bash
CONFIG_PATH=my-config.yaml ./dist/edge-setting
# 默认监听 :8080
```

浏览器访问 `http://<节点IP>:8080`，输入 Token 登录。

---

## 开发模式

```bash
# 终端 1：后端（自动重载）
make dev-backend

# 终端 2：前端 Vite 开发服务器（代理到 :8080）
make dev-frontend
# 访问 http://localhost:5173
```

---

## 测试

```bash
# 所有测试
make test

# 仅后端
make test-backend

# 仅前端
make test-frontend

# 覆盖率报告
make test-coverage
```

### 后端测试覆盖模块

| 模块 | 测试文件 | 覆盖内容 |
|------|----------|----------|
| `auth` | `auth_test.go` | Token 认证、权限分级、Bearer/Query 两种方式 |
| `store` | `sqlite_test.go` | 指标读写、日志查询过滤、审计记录、过期清理 |
| `vsoa` | `client_test.go` | 应用列表/详情、启停重启、订阅事件 |
| `collector` | `metrics_test.go` | CPU/内存/磁盘/负载采集合法性 |
| `api` | `handler_test.go` | 全部 REST 接口（正常/异常/权限分支） |

### 前端测试覆盖

| 测试对象 | 内容 |
|----------|------|
| `MetricCard` | 三态告警阈值逻辑、Label/Value 渲染 |
| `AppCard` | 状态标签、内存格式化、CPU 显示 |
| `useMetricsStore` | 初始状态、错误容忍 |
| `useAppsStore` | 列表拉取、loading 状态、错误处理 |
| Router 守卫 | 未登录重定向、已登录正常跳转 |
| 工具函数 | `formatUptime`、`formatMem` |

---

## 项目结构

```
edge-setting/
├── backend/
│   ├── cmd/server/main.go        # 入口
│   ├── configs/config.yaml       # 配置示例
│   └── internal/
│       ├── api/     handler.go   # REST 接口 + 路由
│       │           routes.go
│       ├── auth/    auth.go      # Token 认证中间件
│       ├── collector/metrics.go  # gopsutil 指标采集
│       ├── config/  config.go    # Viper 配置管理
│       ├── store/   sqlite.go    # SQLite 持久化
│       ├── vsoa/    client.go    # VSOA/MS 通信
│       └── ws/      hub.go       # WebSocket 广播
│
├── frontend/
│   ├── src/
│   │   ├── api/        index.ts  # Axios + API 模块
│   │   ├── components/           # MetricCard / AppCard / TabBar / ECharts
│   │   ├── composables/          # （可扩展）
│   │   ├── router/     index.ts  # Vue Router
│   │   ├── stores/               # Pinia（metrics / apps）
│   │   ├── views/                # 5 个页面视图
│   │   └── styles/   theme.css   # CSS 变量主题
│   ├── __tests__/  app.spec.ts   # Vitest 测试
│   └── vite.config.ts
│
├── Makefile
└── README.md
```

---

## API 一览

| 方法 | 路径 | 权限 | 说明 |
|------|------|------|------|
| GET  | `/health` | 公开 | 健康检查 |
| GET  | `/api/v1/system/info` | 只读 | 系统信息 |
| GET  | `/api/v1/system/metrics` | 只读 | 实时指标快照 |
| WS   | `/api/v1/system/metrics/ws` | 只读 | 指标实时推送 |
| GET  | `/api/v1/metrics/history` | 只读 | 历史指标查询 |
| GET  | `/api/v1/apps` | 只读 | 微应用列表 |
| GET  | `/api/v1/apps/:id` | 只读 | 微应用详情 |
| POST | `/api/v1/apps/:id/start` | **读写** | 启动 |
| POST | `/api/v1/apps/:id/stop` | **读写** | 停止 |
| POST | `/api/v1/apps/:id/restart` | **读写** | 重启 |
| GET  | `/api/v1/apps/:id/logs` | 只读 | 历史日志 |
| WS   | `/api/v1/apps/:id/logs/ws` | 只读 | 实时日志流 |
| GET  | `/api/v1/apps/:id/logs/export` | 只读 | 日志导出 |
| GET  | `/api/v1/audit` | 只读 | 审计日志 |
| GET  | `/api/v1/version` | 只读 | 版本信息 |

所有接口均需 `Authorization: Bearer <token>` 请求头。

---

## 技术栈

**后端**：Go 1.21 · Gin · gorilla/websocket · gopsutil · SQLite (modernc) · Viper · Zap

**前端**：Vue 3 · Vite · TypeScript · Pinia · Vue Router · Vant 4 · ECharts · Axios

**通信**：REST API + WebSocket 推送 + VSOA 软总线（对接 MS）

---

## 部署说明

- 产物为**单一静态二进制** + `config.yaml`
- 支持 `Linux amd64 / arm64`（`make release` 交叉编译）
- 由 MS（系统管理服务）统一启停，无需独立 systemd
- 数据默认存储于 `/var/lib/edge-setting/edge-setting.db`
- Schema 变更时自动迁移，替换二进制即可升级
