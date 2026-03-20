package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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
		BuildVer:  "test",
		BuildFE:   "test",
		StartTime: time.Now(),
	}
	return h, func() { db.Close(); vsoaClient.Close() }
}

func doRequest(t *testing.T, handler *api.Handler, method, path string) *httptest.ResponseRecorder {
	r := api.SetupRouter(handler)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(method, path, nil)
	r.ServeHTTP(w, req)
	return w
}

func TestHealth(t *testing.T) {
	h, cleanup := setupHandler(t)
	defer cleanup()
	w := doRequest(t, h, "GET", "/health")
	if w.Code != 200 {
		t.Errorf("期望 200，实际 %d", w.Code)
	}
}

func TestSystemInfo(t *testing.T) {
	h, cleanup := setupHandler(t)
	defer cleanup()
	w := doRequest(t, h, "GET", "/api/v1/system/info")
	if w.Code != 200 {
		t.Errorf("期望 200，实际 %d", w.Code)
	}
}

func TestMetricsSnapshot(t *testing.T) {
	h, cleanup := setupHandler(t)
	defer cleanup()
	w := doRequest(t, h, "GET", "/api/v1/system/metrics")
	if w.Code != 200 {
		t.Errorf("期望 200，实际 %d", w.Code)
	}
	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["code"] != float64(0) {
		t.Errorf("响应 code 应为 0，实际 %v", resp["code"])
	}
}

func TestAppList(t *testing.T) {
	h, cleanup := setupHandler(t)
	defer cleanup()
	w := doRequest(t, h, "GET", "/api/v1/apps")
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

func TestAppDetail_NotFound(t *testing.T) {
	h, cleanup := setupHandler(t)
	defer cleanup()
	w := doRequest(t, h, "GET", "/api/v1/apps/not-exist")
	if w.Code != 404 {
		t.Errorf("不存在应用应 404，实际 %d", w.Code)
	}
}

func TestAppStart(t *testing.T) {
	h, cleanup := setupHandler(t)
	defer cleanup()
	w := doRequest(t, h, "POST", "/api/v1/apps/rtsp-proxy/start")
	if w.Code != 200 {
		t.Errorf("启动应 200，实际 %d", w.Code)
	}
}

func TestAppStop(t *testing.T) {
	h, cleanup := setupHandler(t)
	defer cleanup()
	w := doRequest(t, h, "POST", "/api/v1/apps/data-collector/stop")
	if w.Code != 200 {
		t.Errorf("停止应 200，实际 %d", w.Code)
	}
}

func TestAppRestart(t *testing.T) {
	h, cleanup := setupHandler(t)
	defer cleanup()
	w := doRequest(t, h, "POST", "/api/v1/apps/data-collector/restart")
	if w.Code != 200 {
		t.Errorf("重启应 200，实际 %d", w.Code)
	}
}

func TestMetricsHistory(t *testing.T) {
	h, cleanup := setupHandler(t)
	defer cleanup()
	w := doRequest(t, h, "GET", "/api/v1/metrics/history?type=cpu&range=1h")
	if w.Code != 200 {
		t.Errorf("历史指标查询应 200，实际 %d", w.Code)
	}
}

func TestVersion(t *testing.T) {
	h, cleanup := setupHandler(t)
	defer cleanup()
	w := doRequest(t, h, "GET", "/api/v1/version")
	if w.Code != 200 {
		t.Errorf("版本接口应 200，实际 %d", w.Code)
	}
	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	if data["backend"] != "test" {
		t.Errorf("后端版本期望 'test'，实际 %v", data["backend"])
	}
}

func TestAuditQuery(t *testing.T) {
	h, cleanup := setupHandler(t)
	defer cleanup()
	_ = doRequest(t, h, "POST", "/api/v1/apps/rtsp-proxy/start")
	w := doRequest(t, h, "GET", "/api/v1/audit")
	if w.Code != 200 {
		t.Errorf("审计查询应 200，实际 %d", w.Code)
	}
}
