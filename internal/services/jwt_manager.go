package services

import (
	"strconv"
	"time"

	"github.com/Pro100x3mal/yp-gophermart.git/internal/configs"
	"github.com/golang-jwt/jwt/v5"
)

type JWTManager struct {
	secret string
	ttl    time.Duration
}

func NewJWTManager(cfg *configs.ServerConfig) *JWTManager {
	return &JWTManager{
		secret: cfg.Secret,
		ttl:    cfg.TokenTTL,
	}
}

func (m *JWTManager) Generate(id int64) (string, error) {
	now := time.Now().UTC()
	claims := jwt.RegisteredClaims{
		Subject:   strconv.FormatInt(id, 10),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(m.secret))
}

func (m *JWTManager) Validate(tokenString string) (int64, error) {
	var claims jwt.RegisteredClaims
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(m.secret), nil
	})
	if err != nil {
		return 0, err
	}

	if !token.Valid {
		return 0, jwt.ErrSignatureInvalid
	}

	id, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil || id <= 0 {
		return 0, jwt.ErrTokenInvalidSubject
	}

	return id, nil
}
