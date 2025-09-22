package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/Pro100x3mal/yp-gophermart.git/internal/models"
	"go.uber.org/zap"
)

type AuthService interface {
	RegisterUser(ctx context.Context, creds *models.Creds) (string, error)
	AuthenticateUser(ctx context.Context, creds *models.Creds) (string, error)
}
type AuthHandler struct {
	authSvc AuthService
	logger  *zap.Logger
}

func NewAuthHandler(authSvc AuthService, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{
		authSvc: authSvc,
		logger:  logger.With(zap.String("handler", "auth")),
	}
}

func (ah *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if !strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var c models.Creds
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if c.Login == "" || c.Password == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	token, err := ah.authSvc.RegisterUser(r.Context(), &c)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}
		if errors.Is(err, context.DeadlineExceeded) {
			http.Error(w, http.StatusText(http.StatusRequestTimeout), http.StatusRequestTimeout)
			return
		}
		if errors.Is(err, models.ErrUserAlreadyExists) {
			http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
			return
		}
		ah.logger.Error("failed to register user", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	setAuthCookie(w, token)
	w.WriteHeader(http.StatusOK)
}

func (ah *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if !strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var c models.Creds
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if c.Login == "" || c.Password == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	token, err := ah.authSvc.AuthenticateUser(r.Context(), &c)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}
		if errors.Is(err, context.DeadlineExceeded) {
			http.Error(w, http.StatusText(http.StatusRequestTimeout), http.StatusRequestTimeout)
			return
		}
		if errors.Is(err, models.ErrUserNotFound) || errors.Is(err, models.ErrInvalidCredentials) {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		ah.logger.Error("failed to authenticate user", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	setAuthCookie(w, token)
	w.WriteHeader(http.StatusOK)
}

func setAuthCookie(w http.ResponseWriter, accessToken string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(time.Hour * 24),
	})
}
