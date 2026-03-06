package jwtutil

import "time"

type TokenBlackListStorage interface {
	Add(token string, ttl time.Duration) error
	Exists(token string) (bool, error)
}
