package crypto

import (
	"errors"

	"bill-management/pkg/logger" // 复用日志模块

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// BcryptCost bcrypt 加密成本（数值越大越安全，耗时也越长，推荐 10-14）
const BcryptCost = 12

// BcryptEncrypt 密码加密（不可逆）
func BcryptEncrypt(password string) (string, error) {

	if password == "" {
		return "", errors.New("密码不能为空")
	}

	// 生成加密后的密码字节
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	if err != nil {
		logger.Error("密码加密失败", zap.Error(err), zap.String("password", "***")) // 日志隐藏原始密码
		return "", err
	}

	return string(hashedBytes), nil
}

// BcryptVerify 验证密码是否正确
func BcryptVerify(password, hashedPassword string) bool {
	if password == "" || hashedPassword == "" {
		logger.Warn("密码或加密密码为空，验证失败")
		return false
	}

	// 对比原始密码和加密密码
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			logger.Warn("密码验证失败：密码不匹配")
		} else {
			logger.Error("密码验证出错", zap.Error(err))
		}
		return false
	}

	return true
}
