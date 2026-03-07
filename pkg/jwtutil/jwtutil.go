package jwtutil

import (
	"bill-management/pkg/config"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type CustomClaims struct {
	UserID               uint64 `json:"user_id"`
	Username             string `json:"username"`
	Role                 string `json:"role"`
	jwt.RegisteredClaims `标准Claims(包含过期时间、签发时间等)`
}

// JWTService 提供生成和解析 JWT token 的功能
type JWTService struct {
	Config  *config.JWTConfig
	storage TokenBlackListStorage // 可选：用于存储黑名单 token 的接口
}

func NewJWTService(config *config.Config, storage TokenBlackListStorage) *JWTService {
	//校验配置
	if config == nil || config.Jwt.SecretKey == "" {
		panic("JWT 配置无效")
	}
	if config.Jwt.ExpireDuration <= 0 {
		panic("JWT 过期时间必须大于0")
	}
	if jwt.GetSigningMethod(config.Jwt.SignMethod) == nil {
		panic("JWT 签名方法无效")
	}
	return &JWTService{
		Config:  &config.Jwt,
		storage: storage,
	}
}

func (s *JWTService) GenerateToken(userID uint64, username, role string) (string, error) {
	claims := CustomClaims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(s.Config.ExpireDuration) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.GetSigningMethod(s.Config.SignMethod), claims)
	signedToken, err := token.SignedString([]byte(s.Config.SecretKey))
	if err != nil {
		return "", err
	}
	return signedToken, nil
}

// ParseToken 解析并验证JWT Token（包含黑名单校验）
func (s *JWTService) ParseToken(tokenString string) (*CustomClaims, error) {
	// 1. 检查黑名单（若配置了存储）
	if s.storage != nil {
		exists, err := s.storage.Exists(tokenString)
		if err != nil {
			return nil, errors.New("failed to check blacklist: " + err.Error())
		}
		if exists {
			return nil, errors.New("token 已失效")
		}
	}

	// 2. 解析Token
	token, err := jwt.ParseWithClaims(
		tokenString,
		&CustomClaims{},
		func(t *jwt.Token) (interface{}, error) {
			if t.Method.Alg() != s.Config.SignMethod {
				return nil, errors.New("invalid signing algorithm")
			}
			return []byte(s.Config.SecretKey), nil
		},
	)
	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			switch {
			case ve.Errors&jwt.ValidationErrorExpired != 0:
				return nil, errors.New("token 已过期")
			case ve.Errors&jwt.ValidationErrorSignatureInvalid != 0:
				return nil, errors.New("token 签名无效")
			default:
				return nil, errors.New("invalid token: " + err.Error())
			}
		}
		return nil, err
	}

	// 3. 验证Token有效性
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

// BlacklistToken 将Token加入黑名单
func (s *JWTService) BlacklistToken(tokenString string) error {
	if s.storage == nil {
		return errors.New("blacklist storage not configured")
	}

	// 解析Token获取过期时间
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &CustomClaims{})
	if err != nil {
		return err
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return errors.New("failed to parse token claims")
	}

	// 计算剩余有效期
	now := time.Now()
	expireTime := claims.ExpiresAt.Time
	ttl := expireTime.Sub(now)
	if ttl <= 0 {
		ttl = 60 * time.Second // 已过期的Token仍保留60秒黑名单
	}

	// 调用存储接口加入黑名单
	return s.storage.Add(tokenString, ttl)
}

// 保留原有兼容方法（可选，逐步迁移后可删除）
// func NewJWTConfig(config *config.Config) *config.JWTConfig {
// 	return &config.Jwt
// }
