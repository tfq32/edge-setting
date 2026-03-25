package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SetupRouter 配置路由（全部只读，无认证）
func SetupRouter(h *Handler) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	r.Use(recoveryMiddleware(h.Log))
	r.Use(corsMiddleware())
	r.Use(loggerMiddleware(h.Log))

	r.GET("/health", h.Health)

	v1 := r.Group("/api/v1")
	{
		v1.GET("/system/info",       h.SystemInfo)
		v1.GET("/system/metrics",    h.MetricsSnapshot)
		v1.GET("/system/metrics/ws", h.MetricsWS)
		v1.GET("/metrics/history",   h.MetricsHistory)
		v1.GET("/app/list",          h.AppList)
	}

	return r
}

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

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

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
