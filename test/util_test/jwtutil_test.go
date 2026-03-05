package util_test

import (
	"bill-management/pkg/config"
	"bill-management/pkg/jwtutil"
	"bill-management/pkg/logger"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJwt(t *testing.T) {
	logger.Info("测试 JWT 配置")
	assert.NotEmpty(t, jwtutil.NewJWTConfig(config.GetConfig()))
	logger.Infof("JWT 配置：%+v", jwtutil.NewJWTConfig(config.GetConfig()))
}

func TestJwtToken(t *testing.T) {

	// 1. 创建 JWT 配置和实例
	jwtConfig := jwtutil.NewJWTConfig(config.GetConfig())
	token, err := jwtutil.GenerateToken(1, "littleBug", "User", jwtConfig)
	assert.NotEmpty(t, token, "生成的 JWT 令牌不能为空")
	assert.NoError(t, err, "生成 JWT 令牌失败")
	logger.Infof("生成的 JWT 令牌: %s", token)

	// 2. 验证 JWT 令牌
	claims, err := jwtutil.ParseToken(token, jwtConfig)
	assert.NoError(t, err, "解析 JWT 令牌失败")
	assert.NotNil(t, claims, "解析后的 Claims 不能为空")
	assert.Equal(t, 1, claims.UserID, "Claims 中的 UserID 不匹配")
	assert.Equal(t, "littleBug", claims.Username, "Claims 中的 Username 不匹配")
	assert.Equal(t, "User", claims.Role, "Claims 中的 Role 不匹配")
	logger.Infof("解析后的 Claims: %+v", claims)

}
