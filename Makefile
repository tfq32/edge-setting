.PHONY: all build build-backend build-frontend test test-backend test-frontend clean dev

# ── 版本 ──────────────────────────────────────────────────
VERSION  ?= $(shell git describe --tags --always 2>/dev/null || echo "dev")
FE_VER   ?= $(VERSION)
LDFLAGS   = -X main.BuildVersion=$(VERSION) -X main.BuildFE=$(FE_VER) -s -w

# ── 目标平台 ──────────────────────────────────────────────
PLATFORMS = linux/amd64 linux/arm64

all: build

# ── 构建 ──────────────────────────────────────────────────
build: build-frontend build-backend

build-frontend:
	@echo "▶ 构建前端..."
	cd frontend && npm install && npm run build

build-backend:
	@echo "▶ 构建后端 ($(GOOS)/$(GOARCH))..."
	cd backend && go build -ldflags "$(LDFLAGS)" -o ../dist/edge-setting ./cmd/server

# 交叉编译全平台
release:
	@mkdir -p dist
	@for platform in $(PLATFORMS); do \
		os=$$(echo $$platform | cut -d/ -f1); \
		arch=$$(echo $$platform | cut -d/ -f2); \
		echo "▶ 编译 $$os/$$arch ..."; \
		cd backend && GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 \
			go build -ldflags "$(LDFLAGS)" -o ../dist/edge-setting-$$os-$$arch ./cmd/server; \
		cd ..; \
	done
	@echo "✓ 编译完成，产物在 dist/"

# ── 测试 ──────────────────────────────────────────────────
test: test-backend test-frontend

test-backend:
	@echo "▶ 运行后端单元测试..."
	cd backend && go test -v -race -count=1 ./internal/... 2>&1

test-frontend:
	@echo "▶ 运行前端单元测试..."
	cd frontend && npm install && npx vitest run --reporter=verbose 2>&1

test-coverage:
	@echo "▶ 后端覆盖率报告..."
	cd backend && go test -coverprofile=coverage.out ./internal/...
	cd backend && go tool cover -html=coverage.out -o coverage.html
	@echo "✓ 覆盖率报告: backend/coverage.html"

# ── 本地开发 ──────────────────────────────────────────────
dev-backend:
	cd backend && CONFIG_PATH=configs/config.yaml go run ./cmd/server

dev-frontend:
	cd frontend && npm run dev

# ── 清理 ──────────────────────────────────────────────────
clean:
	rm -rf dist/ frontend/dist/ backend/coverage.out backend/coverage.html
	@echo "✓ 清理完成"

# ── 帮助 ──────────────────────────────────────────────────
help:
	@echo ""
	@echo "  make build          构建前端 + 后端"
	@echo "  make build-backend  仅构建后端"
	@echo "  make build-frontend 仅构建前端"
	@echo "  make release        交叉编译 amd64/arm64"
	@echo "  make test           运行所有测试"
	@echo "  make test-backend   仅运行后端测试"
	@echo "  make test-frontend  仅运行前端测试"
	@echo "  make test-coverage  生成覆盖率报告"
	@echo "  make dev-backend    启动后端开发服务"
	@echo "  make dev-frontend   启动前端开发服务"
	@echo "  make clean          清理构建产物"
	@echo ""
