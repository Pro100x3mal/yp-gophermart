package services

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/Pro100x3mal/yp-gophermart.git/internal/configs"
	"github.com/Pro100x3mal/yp-gophermart.git/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type UsersRepository interface {
	Create(ctx context.Context, login string, passHash []byte) (int64, error)
	GetByLogin(ctx context.Context, login string) (*models.User, error)
}
type AuthService struct {
	users  UsersRepository
	secret string
	ttl    time.Duration
}

func NewAuthService(users UsersRepository, cfg *configs.ServerConfig) *AuthService {
	return &AuthService{
		users:  users,
		secret: cfg.Secret,
		ttl:    time.Hour * 24,
	}
}

func (s *AuthService) Register(ctx context.Context, creds *models.Creds) (string, error) {
	passHash, err := hashPassword(creds.Password)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	id, err := s.users.Create(ctx, creds.Login, passHash)
	if err != nil {
		return "", fmt.Errorf("failed to create user: %w", err)
	}

	token, err := generateJWT(id, s.secret, s.ttl)
	if err != nil {
		return "", fmt.Errorf("failed to generate JWT: %w", err)
	}

	return token, nil
}

func (s *AuthService) Authenticate(ctx context.Context, creds *models.Creds) (string, error) {
	user, err := s.users.GetByLogin(ctx, creds.Login)
	if err != nil {
		return "", fmt.Errorf("failed to get user: %w", err)
	}

	if err = checkPasswordHash(user.PasswordHash, creds.Password); err != nil {
		return "", models.ErrInvalidCredentials
	}

	token, err := generateJWT(user.ID, s.secret, s.ttl)
	if err != nil {
		return "", fmt.Errorf("failed to generate JWT: %w", err)
	}

	return token, nil
}

func hashPassword(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}

func checkPasswordHash(passHash []byte, password string) error {
	return bcrypt.CompareHashAndPassword(passHash, []byte(password))
}

func generateJWT(id int64, secret string, ttl time.Duration) (string, error) {
	now := time.Now().UTC()
	claims := jwt.RegisteredClaims{
		Subject:   strconv.FormatInt(id, 10),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func validateJWT(tokenString, secret string) (int64, error) {
	var claims jwt.RegisteredClaims
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
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
