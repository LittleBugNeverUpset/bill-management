package util_test

import (
	"bill-management/pkg/crypto"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCrypto(t *testing.T) {
	// config.InitConfig("config") // 确保配置初始化成功，若 InitConfig 返回 error 可添加错误处理逻辑
	// logger.InitLogger()         // 初始化日志系统，确保日志功能可用
	// 1. 测试加密和解密
	originalText := "Hello, World!"

	// 加密
	encryptedText, err := crypto.BcryptEncrypt(originalText)
	assert.NoError(t, err, "加密失败")
	assert.NotEmpty(t, encryptedText, "加密结果不能为空")

	// 解密
	isEqual := crypto.BcryptVerify(originalText, encryptedText)
	assert.True(t, isEqual, "解密失败")

	// 2. 测试使用错误的密钥解密
	wrongKey := "wrongkey"
	isEqual = crypto.BcryptVerify(wrongKey, encryptedText)
	assert.False(t, isEqual, "使用错误的密钥解密应该失败")
}
