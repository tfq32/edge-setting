.PHONY: build-frontend \
        build-linux-amd64 build-linux-arm64 \
        test test-backend test-frontend test-coverage \
        dev-backend dev-frontend clean

# ── 产物目录 ──────────────────────────────────────────────
BIN = output/server

# ── 本地开发 ──────────────────────────────────────────────
dev-backend:
	cd backend && go mod tidy && CONFIG_PATH=configs/config.yaml go run ./cmd/server

dev-frontend:
	cd frontend && pnpm run dev

# ── 清理 ──────────────────────────────────────────────────
clean:
	rm -rf $(BIN)/ frontend/dist/ backend/coverage.out backend/coverage.html
	@echo "✓ cleaned"


build-frontend:
	@echo "> build frontend..."
	cd frontend && pnpm install && pnpm run build

build-linux-amd64:
	@echo "> build Linux amd64..."
	@mkdir -p $(BIN)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
		go build -C backend -ldflags "-s -w" \
		-o ../$(BIN)/edge-setting-linux-amd64 ./cmd/server
	cp -r backend/configs $(BIN)/

build-linux-arm64:
	@echo "> build Linux arm64..."
	@mkdir -p $(BIN)
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 \
		go build -C backend -ldflags "-s -w" \
		-o ../$(BIN)/edge-setting-linux-arm64 ./cmd/server
	cp -r backend/configs $(BIN)/

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
