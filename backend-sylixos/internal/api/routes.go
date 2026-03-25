package api

import (
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
)

// SetupRouter 配置路由（全部只读，无认证）
func SetupRouter(h *Handler) http.Handler {
	router := httprouter.New()

	router.GET("/health", h.Health)
	router.GET("/api/v1/system/info", h.SystemInfo)
	router.GET("/api/v1/system/metrics", h.MetricsSnapshot)
	router.GET("/api/v1/system/metrics/ws", h.MetricsWS)
	router.GET("/api/v1/metrics/history", h.MetricsHistory)
	router.GET("/api/v1/app/list", h.AppList)

	// 包装 CORS 和日志中间件
	return chainMiddleware(router, corsMiddleware, loggerMiddleware, recoveryMiddleware)
}

type middleware func(http.Handler) http.Handler

func chainMiddleware(h http.Handler, middlewares ...middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.WithField("error", err).Error("panic recovered")
				jsonErr(w, http.StatusInternalServerError, 9999, "服务内部错误")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func loggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)
		log.WithFields(log.Fields{
			"method":  r.Method,
			"path":    r.URL.Path,
			"status":  rw.status,
			"latency": time.Since(start).String(),
		}).Info("http")
	})
}

// statusWriter 用于捕获响应状态码
type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}
