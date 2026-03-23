# Edge Setting v2.1.0

运行在 Linux 边缘节点上的设置类微应用，提供**系统资源实时监控**和**微应用状态查看**两大核心能力。

---

## 快速开始

### 环境要求

| 工具 | 最低版本 |
|------|----------|
| Go   | 1.21     |
| Node | 18.x     |
| pnpm | 任意     |

### 1. 配置

```bash
cp backend/configs/config.yaml my-config.yaml
```

按需修改 `my-config.yaml`：

```yaml
server:
  port: 8080        # 监听端口

vsoa:
  ms_address: "vsoa://localhost:3000"  # MS 地址

data:
  dir: /var/lib/edge-setting           # 数据目录
  metrics_retain_days: 7               # 历史数据保留天数
```

### 2. 构建

```bash
make build          # 构建前端 + 后端
make build-backend  # 仅后端
make build-frontend # 仅前端
```

### 3. 运行

```bash
CONFIG_PATH=my-config.yaml ./dist/edge-setting
```

浏览器访问 `http://<节点IP>:8080`。

---

## 开发模式

```bash
# 终端 1：后端
make dev-backend

# 终端 2：前端 Vite 开发服务器（代理到 :8080）
make dev-frontend
# 访问 http://localhost:5173
```

---

## 测试

```bash
make test           # 全部测试
make test-backend   # 仅后端
make test-frontend  # 仅前端
make test-coverage  # 覆盖率报告
```

### 后端测试模块

| 模块        | 说明                               |
|-------------|------------------------------------|
| `collector` | CPU/内存/磁盘/网络/负载采集合法性   |
| `store`     | 指标写入/查询/过期清理              |
| `vsoa`      | 微应用列表查询字段校验              |
| `api`       | 全部 HTTP 接口（正常/404 验证）     |

---

## 项目结构

```
edge-setting/
├── backend/
│   ├── API.md                    # 接口文档
│   ├── cmd/server/main.go        # 入口（采集定时器 / 优雅退出）
│   ├── configs/config.yaml       # 配置示例
│   └── internal/
│       ├── api/      handler.go  # HTTP 处理器
│       │             routes.go   # 路由注册
│       ├── collector/metrics.go  # gopsutil 采集
│       ├── config/   config.go   # Viper 配置
│       ├── store/    sqlite.go   # SQLite 持久化
│       ├── vsoa/     client.go   # VSOA Mock 客户端
│       └── ws/       hub.go      # WebSocket 广播
│
├── frontend/
│   ├── src/
│   │   ├── api/         index.ts       # Axios + API 模块
│   │   ├── components/                 # MetricCard / AppCard / TabBar / AppLayout
│   │   ├── composables/ useLayout.ts   # 响应式断点
│   │   ├── router/      index.ts       # Vue Router（4 个页面）
│   │   ├── stores/                     # Pinia（metrics / apps）
│   │   ├── views/                      # Dashboard / Monitor / Apps / About
│   │   └── styles/      theme.css      # CSS 变量主题
│   └── __tests__/  app.spec.ts         # Vitest 测试
│
├── Makefile
└── README.md
```

---

## API 一览

详见 `backend/API.md`，接口均无需认证。

| 方法 | 路径                        | 说明               |
|------|-----------------------------|--------------------|
| GET  | `/health`                   | 健康检查           |
| GET  | `/api/v1/system/info`       | 节点 & 系统信息    |
| GET  | `/api/v1/system/metrics`    | 实时指标快照       |
| WS   | `/api/v1/system/metrics/ws` | 指标实时推送       |
| GET  | `/api/v1/metrics/history`   | 历史指标查询       |
| GET  | `/api/v1/apps`              | 微应用列表（只读） |

---

## 技术栈

**后端**：Go 1.21 · Gin · gorilla/websocket · gopsutil · SQLite (modernc) · Viper · Zap

**前端**：Vue 3 · Vite · TypeScript · Pinia · Vue Router · Vant 4 · ECharts · Axios

**通信**：REST + WebSocket + VSOA 软总线

---

## 部署说明

- 产物为**单一静态二进制**，前端静态文件通过 `embed.FS` 内嵌
- 支持 `Linux amd64 / arm64`（`make release` 交叉编译）
- 由 MS（系统管理服务）统一管理生命周期，无需独立 systemd
- 历史数据存储于 `<data.dir>/edge-setting.db`，超期自动清理
