package services

import (
	"context"
	"fmt"

	"github.com/Pro100x3mal/yp-gophermart.git/internal/models"
	"golang.org/x/crypto/bcrypt"
)

type UsersRepository interface {
	CreateUser(ctx context.Context, login string, passHash []byte) (int64, error)
	GetUserByLogin(ctx context.Context, login string) (*models.User, error)
}

type TokenGenerator interface {
	Generate(id int64) (string, error)
}
type AuthService struct {
	repo   UsersRepository
	tokens TokenGenerator
}

func NewAuthService(repo UsersRepository, tg TokenGenerator) *AuthService {
	return &AuthService{
		repo:   repo,
		tokens: tg,
	}
}

func (as *AuthService) RegisterUser(ctx context.Context, creds *models.Creds) (string, error) {
	passHash, err := hashPassword(creds.Password)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	id, err := as.repo.CreateUser(ctx, creds.Login, passHash)
	if err != nil {
		return "", fmt.Errorf("failed to create user: %w", err)
	}

	token, err := as.tokens.Generate(id)
	if err != nil {
		return "", fmt.Errorf("failed to generate JWT: %w", err)
	}

	return token, nil
}

func (as *AuthService) AuthenticateUser(ctx context.Context, creds *models.Creds) (string, error) {
	user, err := as.repo.GetUserByLogin(ctx, creds.Login)
	if err != nil {
		return "", fmt.Errorf("failed to get user: %w", err)
	}

	if err = checkPasswordHash(user.PasswordHash, creds.Password); err != nil {
		return "", models.ErrUserInvalidCredentials
	}

	token, err := as.tokens.Generate(user.ID)
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
