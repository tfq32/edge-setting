.PHONY: all build build-backend build-frontend test test-backend test-frontend clean dev

# ── 目标平台 ──────────────────────────────────────────────
PLATFORMS = linux/amd64 linux/arm64

all: build

# ── 构建 ──────────────────────────────────────────────────
build: build-frontend build-backend

build-frontend:
	@echo "> build frontend..."
	cd frontend && pnpm install && pnpm run build

build-backend:
	@echo "> build backend ($(GOOS)/$(GOARCH))..."
	cd backend && go mod tidy && go build -ldflags "-s -w" -o ./dist/edge-setting ./cmd/server

# 交叉编译全平台
release:
	@mkdir -p dist
	@for platform in $(PLATFORMS); do \
		os=$$(echo $$platform | cut -d/ -f1); \
		arch=$$(echo $$platform | cut -d/ -f2); \
		echo "> compile $$os/$$arch ..."; \
		cd backend && GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 \
			go build -ldflags "-s -w" -o ../dist/edge-setting-$$os-$$arch ./cmd/server; \
		cd ..; \
	done
	@echo "✓ done, binaries in dist/"

# ── 测试 ──────────────────────────────────────────────────
test: test-backend test-frontend

test-backend:
	@echo "> run backend tests..."
	cd backend && go test -v -race -count=1 ./internal/...

test-frontend:
	@echo "> run frontend tests..."
	cd frontend && pnpm install && pnpm test

test-coverage:
	@echo "> backend coverage..."
	cd backend && go test -coverprofile=coverage.out ./internal/...
	cd backend && go tool cover -html=coverage.out -o coverage.html
	@echo "✓ report: backend/coverage.html"

# ── 本地开发 ──────────────────────────────────────────────
dev-backend:
	cd backend && go mod tidy && CONFIG_PATH=configs/config.yaml go run ./cmd/server

dev-frontend:
	cd frontend && pnpm run dev

# ── 清理 ──────────────────────────────────────────────────
clean:
	rm -rf dist/ frontend/dist/ backend/coverage.out backend/coverage.html
	@echo "✓ cleaned"

help:
	@echo ""
	@echo "  make build          构建前端 + 后端"
	@echo "  make build-backend  仅构建后端"
	@echo "  make build-frontend 仅构建前端"
	@echo "  make release        交叉编译 amd64/arm64"
	@echo "  make test           运行所有测试"
	@echo "  make dev-backend    启动后端开发服务"
	@echo "  make dev-frontend   启动前端开发服务"
	@echo "  make clean          清理构建产物"
	@echo ""
