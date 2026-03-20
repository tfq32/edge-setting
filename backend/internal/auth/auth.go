package auth

import (
	"net/http"
	"strings"

	"github.com/edge-setting/backend/internal/config"
	"github.com/gin-gonic/gin"
)

const (
	PermReadOnly  = "readonly"
	PermReadWrite = "readwrite"
	CtxPermKey    = "permission"
	CtxTokenKey   = "token"
)

// resolveToken 从配置中查找 token 对应权限
func resolveToken(token string) (string, bool) {
	for _, entry := range config.Global.Auth.Tokens {
		if entry.Token == token {
			return entry.Permission, true
		}
	}
	return "", false
}

// extractBearer 从 Authorization: Bearer <token> 提取 token
func extractBearer(c *gin.Context) string {
	header := c.GetHeader("Authorization")
	if strings.HasPrefix(header, "Bearer ") {
		return strings.TrimPrefix(header, "Bearer ")
	}
	// 兼容 query 参数 ?token=xxx（调试用）
	return c.Query("token")
}

// RequireAuth 需要有效 token（任意权限）
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractBearer(c)
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code": 1002, "message": "缺少认证 Token",
			})
			return
		}
		perm, ok := resolveToken(token)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code": 1002, "message": "Token 无效",
			})
			return
		}
		c.Set(CtxPermKey, perm)
		c.Set(CtxTokenKey, token)
		c.Next()
	}
}

// RequireReadWrite 需要 readwrite 权限
func RequireReadWrite() gin.HandlerFunc {
	return func(c *gin.Context) {
		perm, _ := c.Get(CtxPermKey)
		if perm != PermReadWrite {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code": 1003, "message": "权限不足，需要 ReadWrite 权限",
			})
			return
		}
		c.Next()
	}
}

// GetToken 从上下文获取当前 token
func GetToken(c *gin.Context) string {
	v, _ := c.Get(CtxTokenKey)
	s, _ := v.(string)
	return s
}
