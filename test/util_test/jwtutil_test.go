package util_test

import (
	"bill-management/pkg/config"
	"bill-management/pkg/jwtutil"
	"bill-management/pkg/logger"
	"bill-management/pkg/redisutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJwt(t *testing.T) {
	config.InitConfig("config") // 加载测试配置
	logger.InitLogger()         // 初始化日志（依赖配置）
	logger.Info("测试 JWT 配置")
	rbls := jwtutil.NewRedisBlacklistStorage(redisutil.NewRedisClient(&config.GetConfig().Database.Redis))

	assert.NotEmpty(t, jwtutil.NewJWTService(config.GetConfig(), rbls), "JWT 服务实例不能为空")
	logger.Infof("JWT 配置：%+v", jwtutil.NewJWTService(config.GetConfig(), rbls).Config)
}

func TestJwtToken(t *testing.T) {

	// 1. 创建 JWT 配置和实例
	rbls := jwtutil.NewRedisBlacklistStorage(redisutil.NewRedisClient(&config.GetConfig().Database.Redis))
	JWTService := jwtutil.NewJWTService(config.GetConfig(), rbls)
	token, err := JWTService.GenerateToken(1, "littleBug", "User")
	assert.NotEmpty(t, token, "生成的 JWT 令牌不能为空")
	assert.NoError(t, err, "生成 JWT 令牌失败")
	logger.Infof("生成的 JWT 令牌: %s", token)

	// 2. 验证 JWT 令牌
	claims, err := JWTService.ParseToken(token)
	assert.NoError(t, err, "解析 JWT 令牌失败")
	assert.NotNil(t, claims, "解析后的 Claims 不能为空")
	assert.Equal(t, 1, claims.UserID, "Claims 中的 UserID 不匹配")
	assert.Equal(t, "littleBug", claims.Username, "Claims 中的 Username 不匹配")
	assert.Equal(t, "User", claims.Role, "Claims 中的 Role 不匹配")
	logger.Infof("解析后的 Claims: %+v", claims)

}

// TestJwtTokenGenerateAndParse 测试Token生成与解析（对应原有TestJwtToken）
func TestJwtTokenGenerateAndParse(t *testing.T) {
	rbls := jwtutil.NewRedisBlacklistStorage(redisutil.NewRedisClient(&config.GetConfig().Database.Redis))
	JWTService := jwtutil.NewJWTService(config.GetConfig(), rbls)
	// 1. 生成JWT Token
	userID := 1
	username := "littleBug"
	role := "User"
	token, err := JWTService.GenerateToken(userID, username, role)
	assert.NotEmpty(t, token, "生成的 JWT 令牌不能为空")
	assert.NoError(t, err, "生成 JWT 令牌失败")
	logger.Infof("生成的 JWT 令牌: %s", token)

	// 2. 解析并验证JWT Token
	claims, err := JWTService.ParseToken(token)
	assert.NoError(t, err, "解析 JWT 令牌失败")
	assert.NotNil(t, claims, "解析后的 Claims 不能为空")
	assert.Equal(t, userID, claims.UserID, "Claims 中的 UserID 不匹配")
	assert.Equal(t, username, claims.Username, "Claims 中的 Username 不匹配")
	assert.Equal(t, role, claims.Role, "Claims 中的 Role 不匹配")
	logger.Infof("解析后的 Claims: %+v", claims)
}

// TestJwtTokenBlacklist 新增：测试Token黑名单功能
func TestJwtTokenBlacklist(t *testing.T) {
	rbls := jwtutil.NewRedisBlacklistStorage(redisutil.NewRedisClient(&config.GetConfig().Database.Redis))
	JWTService := jwtutil.NewJWTService(config.GetConfig(), rbls)
	// 1. 生成测试Token
	token, err := JWTService.GenerateToken(2, "testUser", "Admin")
	assert.NoError(t, err, "生成测试Token失败")
	logger.Infof("测试黑名单Token: %s", token)

	// 2. 验证Token未加入黑名单时可正常解析
	claims, err := JWTService.ParseToken(token)
	assert.NoError(t, err, "未拉黑的Token解析失败")
	assert.Equal(t, 2, claims.UserID, "解析的UserID不匹配")

	// 3. 将Token加入黑名单
	err = JWTService.BlacklistToken(token)
	assert.NoError(t, err, "将Token加入黑名单失败")
	logger.Info("Token已加入黑名单")

	// 4. 验证拉黑后的Token解析失败（提示已失效）
	blackClaims, err := JWTService.ParseToken(token)
	assert.Nil(t, blackClaims, "拉黑后的Token应解析出空Claims")
	assert.EqualError(t, err, "token 已失效", "拉黑后的Token应提示失效")
	logger.Info("验证通过：拉黑的Token解析返回「token已失效」")

}
