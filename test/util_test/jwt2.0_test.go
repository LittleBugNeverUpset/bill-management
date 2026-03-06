package util_test

import (
	"bill-management/internal/middleware"
	"bill-management/pkg/config"
	"bill-management/pkg/jwtutil"
	"bill-management/pkg/logger"
	"bill-management/pkg/redisutil"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// 初始化测试用的JWTService（包含Redis黑名单）
func setupTestJWTService(t *testing.T) *jwtutil.JWTService {
	// 1. 初始化Redis客户端（测试环境建议用本地Redis，或使用mock）
	redisClient := redisutil.NewRedisClient(&config.GetConfig().Database.Redis)
	// 测试Redis连接
	_, err := redisClient.Ping(context.Background()).Result()
	if err != nil {
		t.Fatalf("连接Redis失败：%v", err)
	}

	// 2. 初始化JWT配置
	rbls := jwtutil.NewRedisBlacklistStorage(redisClient)

	// 4. 创建JWTService实例
	jwtService := jwtutil.NewJWTService(config.GetConfig(), rbls)
	return jwtService
}

// TestJWTAuthMiddleware_NoToken 测试无token请求
func TestJWTAuthMiddleware_NoToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtService := setupTestJWTService(t)

	// 初始化Gin路由
	r := gin.Default()
	r.GET("/test", middleware.JWTAuthMiddleware(jwtService), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"msg": "success"})
	})

	// 构造无token的请求
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// 验证结果
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "未提供token")
}

// TestJWTAuthMiddleware_InvalidToken 测试无效token请求
func TestJWTAuthMiddleware_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtService := setupTestJWTService(t)

	// 初始化Gin路由
	r := gin.Default()
	r.GET("/test", middleware.JWTAuthMiddleware(jwtService), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"msg": "success"})
	})

	// 构造无效token的请求
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token-123")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// 验证结果
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	logger.Infof("响应内容: %s", w.Body.String())
	assert.Contains(t, w.Body.String(), "token无效")
}

// TestJWTAuthMiddleware_ValidToken 测试有效token请求
func TestJWTAuthMiddleware_ValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtService := setupTestJWTService(t)

	// 生成有效token
	validToken, err := jwtService.GenerateToken(1001, "test_user", "admin")
	assert.NoError(t, err)

	// 初始化Gin路由
	r := gin.Default()
	r.GET("/test", middleware.JWTAuthMiddleware(jwtService), func(c *gin.Context) {
		// 从Context获取UserID并返回
		userID := middleware.GetUserIDFromContext(c)
		c.JSON(http.StatusOK, gin.H{
			"user_id": userID,
			"msg":     "success",
		})
	})

	// 构造有效token的请求
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+validToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// 验证结果
	assert.Equal(t, http.StatusOK, w.Code)
	logger.Infof("响应内容: %s", w.Body.String())
	assert.Contains(t, w.Body.String(), "1001") // 验证UserID正确

	assert.Contains(t, w.Body.String(), "success")
}

// 额外测试：黑名单token的情况（可选）
func TestJWTAuthMiddleware_BlacklistedToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtService := setupTestJWTService(t)

	// 生成token并加入黑名单
	validToken, err := jwtService.GenerateToken(1001, "test_user", "admin")
	assert.NoError(t, err)
	err = jwtService.BlacklistToken(validToken)
	assert.NoError(t, err)

	// 初始化Gin路由
	r := gin.Default()
	r.GET("/test", middleware.JWTAuthMiddleware(jwtService), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"msg": "success"})
	})

	// 构造黑名单token的请求
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+validToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// 验证结果
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	logger.Infof("响应内容: %s", w.Body.String())
	assert.Contains(t, w.Body.String(), "token 已失效")
	logger.Infof("响应内容: %s", w.Body.String())
}
