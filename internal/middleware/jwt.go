// internal/middleware/jwt.go
package middleware

import (
	"bill-management/pkg/jwtutil"
	"net/http"

	"github.com/gin-gonic/gin"
)

// JWTAuthMiddleware JWT鉴权中间件
// 参数：jwtService 是已初始化的JWT服务实例（包含Redis黑名单配置）
func JWTAuthMiddleware(jwtService *jwtutil.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 从Header获取Token（默认从Authorization头获取，格式：Bearer <token>）
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"msg":  "未提供token，请先登录",
			})
			c.Abort() // 终止请求链
			return
		}

		// 2. 解析Bearer Token格式
		var tokenString string
		parts := []rune(authHeader)
		if len(parts) > 7 && string(parts[:7]) == "Bearer " {
			tokenString = string(parts[7:])
		} else {
			// 兼容直接传token的情况（无Bearer前缀）
			tokenString = authHeader
		}

		// 3. 验证Token有效性（包含黑名单、签名、过期时间校验）
		claims, err := jwtService.ParseToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"msg":  "token无效：" + err.Error(),
			})
			c.Abort()
			return
		}

		// 4. 将用户信息存入Gin Context（供后续处理器使用）
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)

		// 5. 继续执行后续中间件/处理器
		c.Next()
	}
}

// GetUserIDFromContext 从Context中获取UserID（封装工具函数，方便业务层调用）
func GetUserIDFromContext(c *gin.Context) int {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0
	}
	// 类型断言，确保返回int类型
	id, ok := userID.(int)
	if !ok {
		return 0
	}
	return id
}
