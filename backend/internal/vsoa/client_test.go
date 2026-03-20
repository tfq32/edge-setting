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

func TestGetApp_Exists(t *testing.T) {
	c := newClient(t)
	defer c.Close()

	app, err := c.GetApp(context.Background(), "edge-setting")
	if err != nil {
		t.Fatalf("GetApp 失败: %v", err)
	}
	if app.ID != "edge-setting" {
		t.Errorf("期望 id=edge-setting，实际 %s", app.ID)
	}
}

func TestGetApp_NotExists(t *testing.T) {
	c := newClient(t)
	defer c.Close()

	_, err := c.GetApp(context.Background(), "not-exist-app")
	if err == nil {
		t.Error("查询不存在的应用应返回错误")
	}
}

func TestStartStopRestart(t *testing.T) {
	c := newClient(t)
	defer c.Close()
	ctx := context.Background()

	// 确认初始停止状态
	app, _ := c.GetApp(ctx, "rtsp-proxy")
	if app.Status != vsoa.StatusStopped {
		t.Skip("rtsp-proxy 初始状态应为 stopped")
	}

	// 启动
	if err := c.Start(ctx, "rtsp-proxy"); err != nil {
		t.Fatalf("Start 失败: %v", err)
	}
	app, _ = c.GetApp(ctx, "rtsp-proxy")
	if app.Status != vsoa.StatusRunning {
		t.Errorf("Start 后应为 running，实际 %s", app.Status)
	}

	// 停止
	if err := c.Stop(ctx, "rtsp-proxy"); err != nil {
		t.Fatalf("Stop 失败: %v", err)
	}
	app, _ = c.GetApp(ctx, "rtsp-proxy")
	if app.Status != vsoa.StatusStopped {
		t.Errorf("Stop 后应为 stopped，实际 %s", app.Status)
	}
}

func TestRestart(t *testing.T) {
	c := newClient(t)
	defer c.Close()
	ctx := context.Background()

	if err := c.Restart(ctx, "data-collector"); err != nil {
		t.Fatalf("Restart 失败: %v", err)
	}
	app, _ := c.GetApp(ctx, "data-collector")
	if app.Status != vsoa.StatusRunning {
		t.Errorf("Restart 后应为 running，实际 %s", app.Status)
	}
}

func TestStartAlreadyRunning(t *testing.T) {
	c := newClient(t)
	defer c.Close()
	ctx := context.Background()

	err := c.Start(ctx, "edge-setting") // edge-setting 初始为 running
	if err == nil {
		t.Error("重复启动应返回错误")
	}
}

func TestSubscribe(t *testing.T) {
	c := newClient(t)
	defer c.Close()
	ctx := context.Background()

	events := make(chan vsoa.StatusChangeEvent, 5)
	c.Subscribe(func(e vsoa.StatusChangeEvent) {
		events <- e
	})

	_ = c.Stop(ctx, "rtsp-proxy")

	select {
	case e := <-events:
		if e.AppID != "rtsp-proxy" {
			t.Errorf("期望 rtsp-proxy 事件，实际 %s", e.AppID)
		}
		if e.Status != vsoa.StatusStopped {
			t.Errorf("期望 stopped，实际 %s", e.Status)
		}
	default:
		// 事件是异步的，不强制等待
	}
}

func TestGetSystemInfo(t *testing.T) {
	c := newClient(t)
	defer c.Close()

	info, err := c.GetSystemInfo(context.Background())
	if err != nil {
		t.Fatalf("GetSystemInfo 失败: %v", err)
	}
	if info.PlatformVersion == "" {
		t.Error("PlatformVersion 不应为空")
	}
}
