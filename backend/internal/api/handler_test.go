package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/edge-setting/backend/internal/api"
	"github.com/edge-setting/backend/internal/collector"
	"github.com/edge-setting/backend/internal/store"
	"github.com/edge-setting/backend/internal/vsoa"
	"github.com/edge-setting/backend/internal/ws"
	"go.uber.org/zap"
)

func setupHandler(t *testing.T) (*api.Handler, func()) {
	db, err := store.Open(t.TempDir())
	if err != nil {
		t.Fatalf("DB 初始化失败: %v", err)
	}
	vsoaClient, _ := vsoa.New("vsoa://localhost:3000")
	log := zap.NewNop()
	h := &api.Handler{
		Collector: collector.New(),
		Store:     db,
		VSOA:      vsoaClient,
		Hub:       ws.NewHub(log),
		Log:       log,
	}
	return h, func() { db.Close(); vsoaClient.Close() }
}

func req(t *testing.T, h *api.Handler, method, path string) *httptest.ResponseRecorder {
	r := api.SetupRouter(h)
	w := httptest.NewRecorder()
	rr, _ := http.NewRequest(method, path, nil)
	r.ServeHTTP(w, rr)
	return w
}

func TestHealth(t *testing.T) {
	h, cleanup := setupHandler(t)
	defer cleanup()
	if w := req(t, h, "GET", "/health"); w.Code != 200 {
		t.Errorf("期望 200，实际 %d", w.Code)
	}
}

func TestSystemInfo(t *testing.T) {
	h, cleanup := setupHandler(t)
	defer cleanup()
	w := req(t, h, "GET", "/api/v1/system/info")
	if w.Code != 200 {
		t.Errorf("期望 200，实际 %d", w.Code)
	}
	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	// 验证只包含实际字段，不含已删除的 platform_ver/backend_ver/frontend_ver/uptime_backend
	for _, forbidden := range []string{"platform_ver", "backend_ver", "frontend_ver", "uptime_backend"} {
		if _, ok := data[forbidden]; ok {
			t.Errorf("响应中不应包含已删除字段 %q", forbidden)
		}
	}
	// 验证必要字段存在
	for _, required := range []string{"hostname", "arch", "uptime_system", "kernel_version"} {
		if _, ok := data[required]; !ok {
			t.Errorf("响应中缺少必要字段 %q", required)
		}
	}
}

func TestMetricsSnapshot(t *testing.T) {
	h, cleanup := setupHandler(t)
	defer cleanup()
	w := req(t, h, "GET", "/api/v1/system/metrics")
	if w.Code != 200 {
		t.Errorf("期望 200，实际 %d", w.Code)
	}
	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["code"] != float64(0) {
		t.Errorf("code 应为 0，实际 %v", resp["code"])
	}
}

func TestMetricsHistory(t *testing.T) {
	h, cleanup := setupHandler(t)
	defer cleanup()
	if w := req(t, h, "GET", "/api/v1/metrics/history?type=cpu&range=1h"); w.Code != 200 {
		t.Errorf("期望 200，实际 %d", w.Code)
	}
}

func TestAppList(t *testing.T) {
	h, cleanup := setupHandler(t)
	defer cleanup()
	w := req(t, h, "GET", "/api/v1/apps")
	if w.Code != 200 {
		t.Errorf("期望 200，实际 %d", w.Code)
	}
	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	data, ok := resp["data"].([]interface{})
	if !ok || len(data) == 0 {
		t.Error("应返回非空应用列表")
	}
}

// 已删除的路由应返回 404
func TestRemovedRoutes(t *testing.T) {
	h, cleanup := setupHandler(t)
	defer cleanup()
	for _, p := range []string{
		"/api/v1/version",
		"/api/v1/apps/edge-setting",
		"/api/v1/apps/edge-setting/start",
		"/api/v1/audit",
	} {
		if w := req(t, h, "GET", p); w.Code == 200 {
			t.Errorf("路由 %s 应已删除，不应返回 200，实际 %d", p, w.Code)
		}
	}
}
