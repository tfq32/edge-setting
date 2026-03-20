package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SetupRouter 配置路由
func SetupRouter(h *Handler) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// 全局中间件
	r.Use(recoveryMiddleware(h.Log))
	r.Use(corsMiddleware())
	r.Use(loggerMiddleware(h.Log))

	// 健康检查
	r.GET("/health", h.Health)

	// API v1（无需认证）
	v1 := r.Group("/api/v1")
	{
		// 系统信息
		v1.GET("/system/info", h.SystemInfo)
		v1.GET("/system/metrics", h.MetricsSnapshot)
		v1.GET("/system/metrics/ws", h.MetricsWS)
		v1.GET("/metrics/history", h.MetricsHistory)
		v1.GET("/version", h.Version)

		// 微应用
		v1.GET("/apps", h.AppList)
		v1.GET("/apps/:id", h.AppDetail)
		v1.GET("/apps/:id/logs", h.AppLogsHistory)
		v1.GET("/apps/:id/logs/ws", h.AppLogsWS)
		v1.GET("/apps/:id/logs/export", h.AppLogsExport)
		v1.POST("/apps/:id/start", h.AppStart)
		v1.POST("/apps/:id/stop", h.AppStop)
		v1.POST("/apps/:id/restart", h.AppRestart)

		// 审计日志
		v1.GET("/audit", h.AuditQuery)
	}

	return r
}

// recoveryMiddleware panic 恢复
func recoveryMiddleware(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Error("panic recovered", zap.Any("error", err))
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"code": 9999, "message": "服务内部错误",
				})
			}
		}()
		c.Next()
	}
}

// corsMiddleware CORS 头
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization,Content-Type")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// loggerMiddleware 请求日志
func loggerMiddleware(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		log.Info("http",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", time.Since(start)),
		)
	}
}
