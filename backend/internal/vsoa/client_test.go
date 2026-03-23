package vsoa_test

import (
	"context"
	"testing"

	"github.com/edge-setting/backend/internal/vsoa"
)

func newClient(t *testing.T) *vsoa.Client {
	c, err := vsoa.New("vsoa://localhost:3000")
	if err != nil {
		t.Fatalf("创建 VSOA 客户端失败: %v", err)
	}
	return c
}

func TestListApps(t *testing.T) {
	c := newClient(t)
	defer c.Close()

	apps, err := c.ListApps(context.Background())
	if err != nil {
		t.Fatalf("ListApps 失败: %v", err)
	}
	if len(apps) == 0 {
		t.Error("期望至少 1 个微应用")
	}
}

func TestListApps_Fields(t *testing.T) {
	c := newClient(t)
	defer c.Close()

	apps, _ := c.ListApps(context.Background())
	for _, app := range apps {
		if app.ID == "" {
			t.Error("app.ID 不应为空")
		}
		if app.Version == "" {
			t.Error("app.Version 不应为空")
		}
	}
}
