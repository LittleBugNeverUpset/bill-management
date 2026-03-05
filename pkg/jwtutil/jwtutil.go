package jwtutil

import (
	"bill-management/pkg/config"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type CustomClaims struct {
	UserID               int    `json:"user_id"`
	Username             string `json:"username"`
	Role                 string `json:"role"`
	jwt.RegisteredClaims `标准Claims(包含过期时间、签发时间等)`
}

func NewJWTConfig(config *config.Config) *config.JWTConfig {
	return &config.Jwt
}

func GenerateToken(userID int, username, role string, c *config.JWTConfig) (string, error) {
	claims := CustomClaims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(c.ExpireDuration) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.GetSigningMethod(c.SignMethod), claims)
	signedToken, err := token.SignedString([]byte(c.SecretKey))
	if err != nil {
		return "", err
	}
	return signedToken, nil
}

// 并析验证 JWT token
func ParseToken(tokenString string, c *config.JWTConfig) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&CustomClaims{},
		func(t *jwt.Token) (interface{}, error) {
			return []byte(c.SecretKey), nil
		},
	)
	if err != nil {
		return nil, err
	}
	//
	if claims, ok := token.Claims.(*CustomClaims); ok {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}
